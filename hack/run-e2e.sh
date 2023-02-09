#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -e

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

rm -f ~/.ssh/known_hosts
arch=amd64
os_name="CENTOS7"
util::vm_name_ip_init_online_by_os ${os_name}
util::power_on_vm_first ${os_name}
echo "before deploy display hostname: "
sshpass -p "${AMD_ROOT_PASSWORD}" ssh -o StrictHostKeyChecking=no root@${vm_ip_addr1} hostname

# prepare kubean install job yml using containerd
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
cp -f "${REPO_ROOT}"/test/common/hosts-conf-cm.yml "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/
cp -f "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanClusterOps.yml  "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/
sed -i "s/ip:/ip: ${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml

sed -i "s#image:#image: ${SPRAY_JOB}#" "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/" "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
# Run cluster function e2e

ginkgo -v -race --fail-fast ./test/kubean_deploy_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}"
ginkgo -v -race -timeout=3h --fail-fast --skip "\[bug\]" ./test/kubean_functions_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
             --clusterOperationName="${CLUSTER_OPERATION_NAME1}" --vmipaddr="${vm_ip_addr1}" \
             --isOffline="false" --arch=${arch} --vmPassword="${AMD_ROOT_PASSWORD}"

# prepare kubean reset job yml
cp "${REPO_ROOT}"/test/common/hosts-conf-cm.yml "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/
cp "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/
cp "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/
sed -i "s/ip:/ip: ${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/hosts-conf-cm.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-reset-cluster/kubeanClusterOps.yml

# prepare kubean install job yml using docker
CLUSTER_OPERATION_NAME2="cluster1-install-dcr"`date "+%H-%M-%S"`
cp -r "${REPO_ROOT}"/test/kubean_functions_e2e/e2e-install-cluster "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-install-cluster-docker
sed -i "s/${CLUSTER_OPERATION_NAME1}/${CLUSTER_OPERATION_NAME2}/" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-install-cluster-docker/kubeanClusterOps.yml
#sed -i "s/containerd/docker/" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml
sed -i "$ a\    override_system_hostname: false" "${REPO_ROOT}"/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml

ginkgo -v -race --fail-fast --skip "\[bug\]" ./test/kubean_reset_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}"  \
              --clusterOperationName="${CLUSTER_OPERATION_NAME2}" --vmipaddr="${vm_ip_addr1}" \
              --isOffline="false" --arch=${arch} --vmPassword="${AMD_ROOT_PASSWORD}"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"