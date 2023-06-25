#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

rm -f ~/.ssh/known_hosts
export ARCH=amd64
export OS_NAME="CENTOS7"
export OFFLINE_FLAG=false
echo "ARCH: ${ARCH}"
echo "IS_OFFLINE: ${OFFLINE_FLAG}"

function func_prepare_config_yaml_3node() {
    local source_path=$1
    local dest_path=$2
    rm -fr "${dest_path}"
    mkdir "${dest_path}"
    cp -f "${source_path}"/hosts-conf-cm-3nodes.yml  "${dest_path}"/hosts-conf-cm.yml
    cp -f "${source_path}"/vars-conf-cm.yml  "${dest_path}"
    cp -f "${source_path}"/kubeanCluster.yml "${dest_path}"
    cp -f "${source_path}"/kubeanClusterOps.yml  "${dest_path}"
    sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr3/${vm_ip_addr3}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${dest_path}/hosts-conf-cm.yml
    sed -i "s#image:#image: ${SPRAY_JOB}#" "${dest_path}"/kubeanClusterOps.yml 

}

####################### create vip cluster ################
echo "create vip cluster....."
export OS_NAME="CENTOS7"
echo "OS_NAME: ${OS_NAME}"
util::power_on_2vms ${OS_NAME}

export OS_NAME="CENTOS7-3MASTER"
echo "OS_NAME: ${OS_NAME}"
util::power_on_third_vms ${OS_NAME}
echo "this:"
go_test_path="test/kubean_3master_vip_e2e"
dest_config_path="${REPO_ROOT}"/test/kubean_3master_vip_e2e/e2e-install-cluster
echo "dest_config_path:${dest_config_path}"

rm -f ~/.ssh/known_hosts
echo "==> scp sonobuoy bin to master: "
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
sshpass -p ${AMD_ROOT_PASSWORD} ssh root@$vm_ip_addr1 "chmod +x /usr/bin/sonobuoy"

func_prepare_config_yaml_3node "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"

CLUSTER_OPERATION_NAME1="cluster1-vip-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i '/# k8s-cluster/a\    kube_vip_address: vip_address1' "${dest_config_path}"/vars-conf-cm.yml
sed -i '/# k8s-cluster/a\    kube_vip_controlplane_enabled: true' "${dest_config_path}"/vars-conf-cm.yml
sed -i '/# k8s-cluster/a\    kube_vip_enabled: true' "${dest_config_path}"/vars-conf-cm.yml
sed -i '/# k8s-cluster/a\    kube_proxy_strict_arp: false' "${dest_config_path}"/vars-conf-cm.yml
sed -i '/# k8s-cluster/a\    kube_vip_arp_enabled: true' "${dest_config_path}"/vars-conf-cm.yml
sed -i "s/vip_address1/10.6.178.220/"  "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race -timeout=3h  --fail-fast ./${go_test_path}  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" --vmipaddr3="${vm_ip_addr3}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"
