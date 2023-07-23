#!/bin/bash
# set -o errexit

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail

# install trivy
{ which trivy 2>/dev/null; } || { echo "install trivy now..."; curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin latest; }

trivy -v
for i in $*
do
  ## MEDIUM,HIGH,CRITICAL
  trivy image --ignore-unfixed --exit-code 1 --severity CRITICAL ${i}
done
