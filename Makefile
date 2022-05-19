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


.PHONY: update
update:
	hack/update-all.sh


.PHONY: test
test:


.PHONY: images
images:
 ## build all images


.PHONY: upload-image
upload-image:
 ## push images

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
