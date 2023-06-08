#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
export registry_addr_arm64=${RUNNER_NODE_IP}:${REGISTRY_PORT_ARM64}
util::scope_copy_test_images ${registry_addr_arm64}

############### Case Prepare
os_name="kylinv10"
arch="arm64"
util::vm_name_ip_init_offline_by_os  ${os_name}
util::init_kylin_vm_template_map
rm -fr "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
mkdir "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
cp -f  ${REPO_ROOT}/test/offline-common/hosts-conf-cm-2nodes.yml ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
cp -f  ${REPO_ROOT}/test/offline-common/kubeanCluster.yml ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster
cp -f  ${REPO_ROOT}/test/offline-common/kubeanClusterOps.yml ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster
cp -f  ${REPO_ROOT}/test/offline-common/vars-conf-cm.yml ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster

# host-config-cm.yaml set
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
ARM64_SERVER_IP="10.0.6.17"
ARM64_SERVER_PASSWORD="Admin@9000"
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/g" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/g" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${KYLIN_VM_PASSWORD}/g" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
# kubeanClusterOps.yml sed
sed -i "s#image:#image: ${SPRAY_JOB}#" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s#e2e-cluster1-install#${CLUSTER_OPERATION_NAME1}#" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s#{offline_minio_url}#${MINIO_URL}#g" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i  "s#centos#kylin#g" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
# vars-conf-cm.yml set
sed -i "s#registry_host:#registry_host: ${registry_addr_arm64}#"    ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#minio_address:#minio_address: ${MINIO_URL}#"    ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#registry_host_key#${registry_addr_arm64}#g"    ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#{{ files_repo }}/centos#{{ files_repo }}/kylin#" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/vars-conf-cm.yml
echo ${vm_name1}
echo ${vm_name2}
util::init_kylin_vm ${template_name1} ${vm_name1} ${ARM64_SERVER_IP} ${ARM64_SERVER_PASSWORD}
util::init_kylin_vm ${template_name2} ${vm_name2} ${ARM64_SERVER_IP} ${ARM64_SERVER_PASSWORD}
# Wait for vm ready
sleep 60
echo "wait ${vm_ip_addr1} ..."
util::wait_ip_reachable "${vm_ip_addr1}" 10
echo "wait ${vm_ip_addr2} ..."
util::wait_ip_reachable "${vm_ip_addr2}" 10

### RUN CASE
ginkgo -v -timeout=10h -race --fail-fast ./test/kubean_os_compatibility_e2e/  -- \
    --kubeconfig="${KUBECONFIG_FILE}" \
    --clusterOperationName="${CLUSTER_OPERATION_NAME1}" --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
    --isOffline="true"  --vmPassword="${KYLIN_VM_PASSWORD}"  --arch=${arch}

util::delete_kylin_vm ${vm_name1} ${ARM64_SERVER_IP} ${ARM64_SERVER_PASSWORD}
util::delete_kylin_vm ${vm_name2} ${ARM64_SERVER_IP} ${ARM64_SERVER_PASSWORD}
echo "Delete vm end!"