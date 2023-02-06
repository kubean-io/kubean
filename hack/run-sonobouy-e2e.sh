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
export ISOFFLINE=false

./hack/run-nightly-cluster-e2e.sh
./hack/run-network-e2e.sh


### do add worker node senario
## precondition generate rsa key
## step1 create k8 cluster with containerd and private key
## step2 add worker node with containerd and private key
## step3 remove worker node with containerd and private key
# prepare kubean install job yml files
#* generate_rsa_key
#ID_RSA=$(cat ~/.ssh/id_rsa|base64 -w 0)
#* ID_RSA=$(cat ./id_rsa|base64 -w 0)
#* sed -i "s/ID_RSA/${ID_RSA}/" $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/ssh-auth-secret.yml
#* cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/
#* sshpass -p root ssh-copy-id -f -i ./id_rsa.pub root@$vm_ip_addr1
#* sshpass -p root ssh-copy-id -f -i ./id_rsa.pub root@$vm_ip_addr2
#* sed -i "s/vm_ip_addr1/${vm_ip_addr1}/"  $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/hosts-conf-cm.yml
#* sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/kubeanClusterOps.yml
# prepare add-worker-node yaml
#* cp $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/kubeanCluster.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node
#* cp $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/ssh-auth-secret.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node
#* cp $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/kubeanCluster.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node
#* cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node
#* sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node/hosts-conf-cm.yml
#* sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node/hosts-conf-cm.yml
#* sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_add_remove_worker_nightlye2e/add-worker-node/kubeanClusterOps.yml
## do remove worker node senario
#* cp $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/kubeanCluster.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node
#* cp $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/ssh-auth-secret.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node
#* cp $(pwd)/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey/kubeanCluster.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node
#* cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node
#* sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node/hosts-conf-cm.yml
#* sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node/hosts-conf-cm.yml
#* sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_add_remove_worker_nightlye2e/remove-worker-node/kubeanClusterOps.yml
#* ginkgo -v -race --fail-fast ./test/kubean_add_remove_worker_nightlye2e/  -- --kubeconfig="${MAIN_KUBECONFIG}" --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}"


