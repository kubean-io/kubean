#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

bash "$REPO_ROOT/hack/update-codegen.sh" kubeancluster
bash "$REPO_ROOT/hack/update-codegen.sh" kubeanclusterops
bash "$REPO_ROOT/hack/update-crdgen.sh"

go mod tidy
go mod vendor