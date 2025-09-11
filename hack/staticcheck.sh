#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
GOLANGCI_LINT_VER="v2.1.0"

cd "${REPO_ROOT}"
source "hack/util.sh"

# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin $GOLANGCI_LINT_VER

if golangci-lint run --fix --verbose; then
  echo 'Congratulations!  All Go source files have passed staticcheck.'
else
  echo # print one empty line, separate from warning messages.
  echo 'Please review the above warnings.'
  echo 'If the above warnings do not make sense, feel free to file an issue.'
  exit 1
fi
