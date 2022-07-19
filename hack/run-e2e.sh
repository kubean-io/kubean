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
SPRAY_JOB_VERSION=${2:-v0.0.1}
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
vagrant init kubean-e2e-vm-template
sed -i "$ i\  config.vm.network \"public_network\", ip: \"${vm_ip_addr}\", bridge: \"ens192\"" Vagrantfile
vagrant up
vagrant status
ping -c 5 ${vm_ip_addr}
sshpass -p root ssh root@${vm_ip_addr} cat /proc/version

# prepare kubean job yml
SPRAY_JOB="ghcr.io/kubean-io/kubean/spray-job:${SPRAY_JOB_VERSION}"
sed -i "s/ip:/ip: ${vm_ip_addr}/" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr}/" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
# Run nightly e2e
ginkgo -v -race --fail-fast ./test/kubean_deploy_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}"
ginkgo -v -race --fail-fast ./test/kubean_functions_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}" --vmipaddr="${vm_ip_addr}"
