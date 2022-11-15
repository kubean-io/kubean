#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -e

# This script runs e2e test against on kubean control plane.
# You should prepare your environment in advance and following environment may be you need to set or use default one.
# - CONTROL_PLANE_KUBECONFIG: absolute path of control plane KUBECONFIG file.
#
# Usage: hack/run-e2e.sh

# Run e2e 
KUBECONFIG_PATH=${KUBECONFIG_PATH:-"${HOME}/.kube"}
HOST_CLUSTER_NAME=${1:-"kubean-host"}
SPRAY_JOB_VERSION=${2:-latest}
vm_ip_addr=${3:-"10.6.127.33"}
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}
EXIT_CODE=0
echo "currnent dir: "$(pwd)
# Install ginkgo
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin

# prepare vagrant vm as k8 cluster single node
vm_clean_up(){
    vagrant destroy -f default
    exit $EXIT_CODE
}

trap vm_clean_up EXIT
rm -f ~/.ssh/known_hosts
vagrant init Kiowa/kubean-e2e-vm-template --box-version 0
sed -i "$ i\  config.vm.network \"public_network\", ip: \"${vm_ip_addr}\", bridge: \"ens192\"" Vagrantfile
vagrant up
vagrant status
ping -c 5 ${vm_ip_addr}
sshpass -p root ssh -o StrictHostKeyChecking=no root@${vm_ip_addr} cat /proc/version
# print vm origin hostname
echo "before deploy display hostname: "
sshpass -p root ssh -o StrictHostKeyChecking=no root@${vm_ip_addr} hostname

# prepare kubean install job yml using containerd
SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
CLUSTER_OPERATION_NAME1="cluster1-install-"`date +%s`
cp $(pwd)/test/common/hosts-conf-cm.yml $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/
cp $(pwd)/test/common/kubeanCluster.yml $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/
cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/
sed -i "s/ip:/ip: ${vm_ip_addr}/" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr}/" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
# Run cluster function e2e

ginkgo -v -race --fail-fast ./test/kubean_deploy_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}"
ginkgo -v -race -timeout=3h --fail-fast --skip "\[bug\]" ./test/kubean_functions_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}" --clusterOperationName="${CLUSTER_OPERATION_NAME1}" --vmipaddr="${vm_ip_addr}"

# prepare kubean reset job yml
cp $(pwd)/test/common/hosts-conf-cm.yml $(pwd)/test/kubean_reset_e2e/e2e-reset-cluster/
cp $(pwd)/test/common/kubeanCluster.yml $(pwd)/test/kubean_reset_e2e/e2e-reset-cluster/
cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_reset_e2e/e2e-reset-cluster/
sed -i "s/ip:/ip: ${vm_ip_addr}/" $(pwd)/test/kubean_reset_e2e/e2e-reset-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr}/" $(pwd)/test/kubean_reset_e2e/e2e-reset-cluster/hosts-conf-cm.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_reset_e2e/e2e-reset-cluster/kubeanClusterOps.yml

# prepare kubean install job yml using docker
CLUSTER_OPERATION_NAME2="cluster1-install-dcr"`date +%s`
cp -r $(pwd)/test/kubean_functions_e2e/e2e-install-cluster $(pwd)/test/kubean_reset_e2e/e2e-install-cluster-docker
sed -i "s/${CLUSTER_OPERATION_NAME1}/${CLUSTER_OPERATION_NAME2}/" $(pwd)/test/kubean_reset_e2e/e2e-install-cluster-docker/kubeanClusterOps.yml
sed -i "s/containerd/docker/" $(pwd)/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml
sed -i "$ a\    override_system_hostname: false" $(pwd)/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml

ginkgo -v -race --fail-fast --skip "\[bug\]" ./test/kubean_reset_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}"  --clusterOperationName="${CLUSTER_OPERATION_NAME2}" --vmipaddr="${vm_ip_addr}"