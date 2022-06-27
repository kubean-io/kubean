GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
SOURCES := $(shell find . -type f  -name '*.go')

BUILD_ARCH ?= linux/$(GOARCH)

# Git information
GIT_VERSION ?= $(shell git describe --tags --abbrev=8 --dirty) # attention: gitlab CI: git fetch should not use shallow
GIT_COMMIT_HASH ?= $(shell git rev-parse HEAD)
GIT_TREESTATE = "clean"
GIT_DIFF = $(shell git diff --quiet >/dev/null 2>&1; if [ $$? -eq 1 ]; then echo "1"; fi)
ifeq ($(GIT_DIFF), 1)
    GIT_TREESTATE = "dirty"
endif
BUILDDATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := "-X github.com/daocloud/kubean/pkg/version.gitVersion=$(GIT_VERSION) \
            -X github.com/daocloud/kubean/pkg/version.gitCommit=$(GIT_COMMIT_HASH) \
            -X github.com/daocloud/kubean/pkg/version.gitTreeState=$(GIT_TREESTATE) \
            -X github.com/daocloud/kubean/pkg/version.buildDate=$(BUILDDATE)"

GOBIN         = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN         = $(shell go env GOPATH)/bin
endif
GOIMPORTS     = $(GOBIN)/goimports

# Images management
REGISTRY_SERVER_ADDRESS?="release-ci.daocloud.io"
REGISTRY_REPO?="$(REGISTRY_SERVER_ADDRESS)/kubean-ci"
HELM_REPO?="https://$(REGISTRY_SERVER_ADDRESS)/chartrepo/kubean-ci"
API_PKG    := ./api

# Parameter
KUBEAN_NAMESPACE="kubean-system"
RETAIN_UI_IMAGE_WHEN_DEPLOY?="false" # on dev site, ui image in the helm chart maybe old( UI image tag will not get updated until sprint ends), so we should left ui as it was, then UI repo pipeline will CD ui image alone

# CICD
DEPLOY_ENV?="PROD"

# Set your version by env or using latest tags from git
VERSION?=""
ifeq ($(VERSION), "")
    LATEST_TAG=$(shell git describe --tags --abbrev=8)
    ifeq ($(LATEST_TAG),)
        # Forked repo may not sync tags from upstream, so give it a default tag to make CI happy.
        VERSION="unknown"
    else
        VERSION=$(LATEST_TAG)
    endif
endif

# convert to git version to semver version v0.1.1-14-gb943a40 --> v0.1.1+14-gb943a40
KUBEAN_VERSION := $(shell echo $(VERSION) | sed 's/-/+/1')

# convert to git version to semver version v0.1.1+14-gb943a40 --> v0.1.1-14-gb943a40
KUBEAN_IMAGE_VERSION := $(shell echo $(KUBEAN_VERSION) | sed 's/+/-/1')

#v0.1.1 --> 0.1.1 Match the helm chart version specification, remove the preceding prefix `v` character
KUBEAN_CHART_VERSION := $(shell echo ${KUBEAN_VERSION} |sed  's/^v//g' )

SPRAY_TAG ?= "release-2.19"


## Deploy current version of helm package to target cluster of $(YOUR_KUBE_CONF) [not defined]
.PHONY: deploy
deploy:
	bash hack/deploy.sh "$(KUBEAN_CHART_VERSION)" "$(KUBEAN_IMAGE_VERSION)" "$(YOUR_KUBE_CONF)" "$(KUBEAN_NAMESPACE)" "$(HELM_REPO)" "$(REGISTRY_REPO)" "$(DEPLOY_ENV)" "$(CD_TO_ENVIRONMENT)"

.PHONY: argocd
argocd:
	bash hack/argocd.sh "$(KUBEAN_CHART_VERSION)" "$(KUBEAN_IMAGE_VERSION)" "$(YOUR_KUBE_CONF)" "$(KUBEAN_NAMESPACE)" "$(HELM_REPO)" "$(REGISTRY_REPO)" "$(DEPLOY_ENV)" "$(CD_TO_ENVIRONMENT)"

.PHONY: kubean-imgs
kubean-imgs: kubean-operator spray-job

.PHONY: kubean-operator
kubean-operator: $(SOURCES)
	echo "Building kubean-operator for arch = $(BUILD_ARCH)"
	export DOCKER_CLI_EXPERIMENTAL=enabled ;\
	! ( docker buildx ls | grep kubean-operator-multi-platform-builder ) && docker buildx create --use --platform=$(BUILD_ARCH) --name kubean-operator-multi-platform-builder ;\
	docker buildx build \
			--build-arg kubean_version=$(KUBEAN_VERSION) \
			--builder kubean-operator-multi-platform-builder \
			--platform $(BUILD_ARCH) \
			--tag $(REGISTRY_REPO)/kubean-operator:$(KUBEAN_IMAGE_VERSION)  \
			--tag $(REGISTRY_REPO)/kubean-operator:latest  \
			-f ./build/images/kubean-operator/Dockerfile \
			--load \
			.

.PHONY: spray-job
spray-job: $(SOURCES)
	echo "Building spray-job for arch = $(BUILD_ARCH)"
	export DOCKER_CLI_EXPERIMENTAL=enabled ;\
	! ( docker buildx ls | grep spray-job-multi-platform-builder ) && docker buildx create --use --platform=$(BUILD_ARCH) --name spray-job-multi-platform-builder ;\
	docker buildx build \
			--build-arg spray_tag=$(SPRAY_TAG) \
			--builder spray-job-multi-platform-builder \
			--platform $(BUILD_ARCH) \
			--tag $(REGISTRY_REPO)/spray-job:$(KUBEAN_IMAGE_VERSION)  \
			--tag $(REGISTRY_REPO)/spray-job:latest  \
			-f ./build/images/spray-job/Dockerfile \
			--load \
			.

.PHONY: upload-image
upload-image: kubean-imgs
	@echo "push images to $(REGISTRY_REPO)"
	docker login -u ${REGISTRY_USER_NAME} -p ${REGISTRY_PASSWORD} ${REGISTRY_SERVER_ADDRESS}
	@docker push $(REGISTRY_REPO)/kubean-operator:latest
	@docker push $(REGISTRY_REPO)/kubean-operator:$(KUBEAN_IMAGE_VERSION)
	@docker push $(REGISTRY_REPO)/spray-job:latest
	@docker push $(REGISTRY_REPO)/spray-job:$(KUBEAN_IMAGE_VERSION)

.PHONY: push-chart
push-chart:
	#helm package -u ./charts/ -d ./dist/
	helm repo add kubean-release $(HELM_REPO)
	helm package ./charts/ -d dist --version $(KUBEAN_CHART_VERSION)
	helm cm-push ./dist/kubean-$(KUBEAN_CHART_VERSION).tgz  kubean-release -a $(KUBEAN_CHART_VERSION) -v $(KUBEAN_CHART_VERSION) -u $(REGISTRY_USER_NAME)  -p $(REGISTRY_PASSWORD)

.PHONY: release
release: kubean-imgs upload-image push-chart

.PHONY: test
test:
	bash hack/unit-test.sh

.PHONY: e2e-test
e2e-test:
	bash hack/e2e.sh "$(KUBEAN_VERSION)" "${KUBEAN_IMAGE_VERSION}" "${HELM_REPO}" "${REGISTRY_REPO}"

.PHONY: nightly-e2e-test
nightly-e2e-test:
	bash hack/run-nightly-e2e.sh "$(KUBEAN_VERSION)" "${KUBEAN_IMAGE_VERSION}" "${HELM_REPO}" "${REGISTRY_REPO}"

.PHONY: clear-kind
clear-kind:
	bash hack/delete-kind-cluster.sh

.PHONY: verify-import-alias
 verify-import-alias:
	bash hack/verify-import-alias.sh

.PHONY: update
update:
	cd ./api/ && make update && cd .. && go mod tidy && go mod vendor

.PHONY: test-staticcheck
test-staticcheck:
	hack/verify-staticcheck.sh


.PHONY: verify-code-gen
verify-code-gen:
	#hack/verify-codegen.sh
	#hack/verify-crdgen.sh


.PHONY: verify-vendor
verify-vendor:
	#hack/verify-vendor.sh

.PHONY: gen-release-notes
gen-release-notes:
	bash hack/release-version.sh

.PHONY: sync_api
sync_api:
	bash hack/sync-api.sh $(VERSION)


.PHONY: security-scanning
security-scanning:
	bash hack/trivy.sh $(REGISTRY_REPO)/kubean-operator:$(KUBEAN_IMAGE_VERSION)
