#!/usr/bin/env bash

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
export OS_NAME="CENTOS7"
export OFFLINE_FLAG=false
./hack/kubean_resource.sh ${TARGET_VERSION} "artifacts"
./hack/run-network-e2e.sh
./hack/run-nightly-cluster-e2e.sh


