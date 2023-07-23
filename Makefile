# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

GOARCH ?= $(shell go env GOARCH)
BUILD_ARCH ?= linux/$(GOARCH)

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

REGISTRY_SERVER_ADDRESS?=container.io
REGISTRY_REPO?=$(REGISTRY_SERVER_ADDRESS)/kubean-ci

SPRAY_TAG?="latest"


.PHONY: test
test:
	bash hack/unit-test.sh

.PHONY: staticcheck
staticcheck:
	hack/staticcheck.sh

.PHONY: update
update:
	cd ./api/ && make update && cd .. && go mod tidy && go mod vendor

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
			--build-arg SPRAY_TAG=$(SPRAY_TAG) \
			--builder spray-job-multi-platform-builder \
			--platform $(BUILD_ARCH) \
			--tag $(REGISTRY_REPO)/spray-job:$(KUBEAN_IMAGE_VERSION)  \
			--tag $(REGISTRY_REPO)/spray-job:latest  \
			-f ./build/images/spray-job/Dockerfile \
			--load \
			.

IMAGE_REGISTRY ?= "ghcr.m.daocloud.io"
RELEASE_NAME ?= "kubean"
TARGET_NS ?= "kubean-system"
KUBECONFIG_PATH ?= "kubeconfig"
GIT_VERSION ?= $(shell git describe --tags --abbrev=8)
IMAGE_TAG ?= ${GIT_VERSION}
.PHONY: local-chart-to-deploy
local-chart-to-deploy:
	bash hack/local-chart-to-deploy.sh ${IMAGE_REGISTRY} ${RELEASE_NAME} ${TARGET_NS} ${KUBECONFIG_PATH} ${IMAGE_TAG} ${GIT_VERSION}


.PHONY: security-scanning
security-scanning:
	bash hack/trivy.sh \
	${REGISTRY}/${REPO}/spray-job:${IMAGE_TAG} \
	${REGISTRY}/${REPO}/kubean-operator:${IMAGE_TAG} \
	${REGISTRY}/${REPO}/kubespray:${SPRAY_IMAGE_TAG_SHORT_SHA}
