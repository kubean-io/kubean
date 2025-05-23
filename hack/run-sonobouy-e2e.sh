#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

check_yq_intalled(){
    yq_installed=0
    yq -V |grep 'version' && yq_installed=true || yq_installed=false
    if [ "${yq_installed}" == "false" ]; then
        wget https://github.com/mikefarah/yq/releases/download/v4.30.8/yq_linux_amd64 && \
            sudo mv yq_linux_amd64 /usr/local/bin/yq && sudo chmod +x /usr/local/bin/yq
    fi
}

generate_rsa_key(){
    echo 'y'| ssh-keygen -f id_rsa -t rsa -N ''
}

rm -f ~/.ssh/known_hosts
export ARCH=amd64
export OS_NAME="ROCKY8"
export OFFLINE_FLAG=false
export NIGHTLY_E2E_STEP=${1}
if [ "${NIGHTLY_E2E_STEP}" == "ALL" ]; then
    echo "RUN ALL SONOBOUY E2E......."
    ./hack/run-vip.sh
    ./hack/kubean_resource.sh ${TARGET_VERSION} "artifacts"
    ./hack/run-network-e2e.sh
    ./hack/run-nightly-cluster-e2e.sh
fi
if [ "${NIGHTLY_E2E_STEP}" == "STEP1" ]; then
    echo "RUN SONOBOUY E2E STEP1......."
    ./hack/run-vip.sh
fi
if [ "${NIGHTLY_E2E_STEP}" == "STEP2" ]; then
    echo "RUN SONOBOUY E2E STEP2......."
    ./hack/kubean_resource.sh ${TARGET_VERSION} "artifacts"
fi
if [ "${NIGHTLY_E2E_STEP}" == "STEP3" ]; then
    echo "RUN SONOBOUY E2E STEP3......."
    ./hack/run-nightly-cluster-e2e.sh
fi

if [[ "${NIGHTLY_E2E_STEP}" =~ "network" ]]; then
    echo "RUN SONOBOUY network E2E ......."
    ./hack/run-network-e2e.sh ${NIGHTLY_E2E_STEP}
fi




