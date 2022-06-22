#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

trivy image --exit-code 999 ${1}
