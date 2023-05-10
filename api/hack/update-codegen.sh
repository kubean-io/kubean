#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -ex

SubDirName=$1 # kubeancluster

# For all commands, the working directory is the parent directory(repo root).

## pwd ## /.../kubean/api

echo "Generating with deepcopy-gen"
GO111MODULE=on go install k8s.io/code-generator/cmd/deepcopy-gen
export GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin

## github.com/kubean-io/kubean-api

deepcopy-gen \
  --go-header-file hack/boilerplate/boilerplate.go.txt \
  --input-dirs=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --output-package=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --output-file-base=zz_generated.deepcopy

echo "Generating with register-gen"
GO111MODULE=on go install k8s.io/code-generator/cmd/register-gen
register-gen \
  --go-header-file hack/boilerplate/boilerplate.go.txt \
  --input-dirs=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --output-package=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --output-file-base=zz_generated.register

echo "Generating with client-gen"
GO111MODULE=on go install k8s.io/code-generator/cmd/client-gen
client-gen \
  --go-header-file hack/boilerplate/boilerplate.go.txt \
  --input-base="" \
  --input=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --output-package=github.com/kubean-io/kubean-api/generated/$SubDirName/clientset \
  --clientset-name=versioned

echo "Generating with lister-gen"
GO111MODULE=on go install k8s.io/code-generator/cmd/lister-gen
lister-gen \
  --go-header-file hack/boilerplate/boilerplate.go.txt \
  --input-dirs=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --output-package=github.com/kubean-io/kubean-api/generated/$SubDirName/listers

echo "Generating with informer-gen"
GO111MODULE=on go install k8s.io/code-generator/cmd/informer-gen
informer-gen \
  --go-header-file hack/boilerplate/boilerplate.go.txt \
  --input-dirs=github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1 \
  --versioned-clientset-package=github.com/kubean-io/kubean-api/generated/$SubDirName/clientset/versioned \
  --listers-package=github.com/kubean-io/kubean-api/generated/$SubDirName/listers \
  --output-package=github.com/kubean-io/kubean-api/generated/$SubDirName/informers

if ls "$GOPATH"/src | grep kubean.io; then
  cp -r "$GOPATH"/src/github.com/kubean-io/kubean-api/apis/$SubDirName/v1alpha1/*.go apis/$SubDirName/v1alpha1/
  cp -r "$GOPATH"/src/github.com/kubean-io/kubean-api/generated/$SubDirName generated/$SubDirName
fi
