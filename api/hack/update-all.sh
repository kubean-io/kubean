#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

API_REPO_ROOT=$(pwd) ## /.../kubean/api

bash "$API_REPO_ROOT/hack/update-codegen.sh" kubeancluster
bash "$API_REPO_ROOT/hack/update-codegen.sh" kubeanclusterops
bash "$API_REPO_ROOT/hack/update-codegen.sh" kubeancomponentsversion
bash "$API_REPO_ROOT/hack/update-codegen.sh" kubeanofflineversion
bash "$API_REPO_ROOT/hack/update-crdgen.sh"

go mod tidy
go mod vendor
