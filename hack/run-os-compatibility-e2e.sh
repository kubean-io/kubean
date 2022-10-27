#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -e

KUBECONFIG_PATH=${KUBECONFIG_PATH:-"${HOME}/.kube"}
HOST_CLUSTER_NAME=${1:-"kubean-host"}
SPRAY_JOB_VERSION=${2:-latest}
vm_ip_addr1=${3:-"10.6.127.33"}
vm_ip_addr2=${4:-"10.6.127.36"}
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}
EXIT_CODE=0

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh

echo "==> current dir: "$(pwd)
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin

# destroy the vm if exist
trap "vm_clean_up_by_name sonobouyDefault sonobouyDefault2" EXIT

os_compability_e2e(){
    vagrantfile=${1}
    echo "currnent dir: "$(pwd)
    # Install ginkgo
    GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
    export PATH=$PATH:$GOPATH/bin

    #prepare master vm
    utils::create_os_e2e_vms $vagrantfile $vm_ip_addr1 $vm_ip_addr2

    echo "==> scp sonobuoy bin to master: "
    rm -f ~/.ssh/known_hosts
    sshpass -p root scp -o StrictHostKeyChecking=no $(pwd)/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
    sshpass -p root ssh root@$vm_ip_addr1 "chmod +x /usr/bin/sonobuoy"

    # prepare kubean install job yml using containerd
    SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
    CLUSTER_OPERATION_NAME="cluster1-install-"`date +%s`
    echo "CLUSTER_OPERATION_NAME: $CLUSTER_OPERATION_NAME"
    cp $(pwd)/test/common/kubeanCluster.yml $(pwd)/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" $(pwd)/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" $(pwd)/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
    sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
    sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME}/" $(pwd)/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
    # Run cluster function e2e
    ginkgo -v -timeout=10h -race --fail-fast ./test/kubean_os_compatibility_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}" --clusterOperationName="${CLUSTER_OPERATION_NAME}" --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}"
}


###### OS compitable e2e logic ########
utils::install_sshpass sshpass
os_array=("Vagrantfile_rhel84")
for (( i=0; i<${#os_array[@]};i++)); do
    os_compability_e2e ${os_array[$i]}
    vagrant destroy -f sonobouyDefault
    vagrant destroy -f sonobouyDefault2
done



