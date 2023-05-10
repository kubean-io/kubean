#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

API_REPO_ROOT=$(pwd) ## /.../kubean/api

bash "$API_REPO_ROOT/hack/update-codegen.sh" cluster
bash "$API_REPO_ROOT/hack/update-codegen.sh" clusteroperation
bash "$API_REPO_ROOT/hack/update-codegen.sh" manifest
bash "$API_REPO_ROOT/hack/update-codegen.sh" localartifactset
bash "$API_REPO_ROOT/hack/update-crdgen.sh"

# go mod tidy
# go mod vendor
