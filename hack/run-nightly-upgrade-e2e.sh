#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "ARCH: ${ARCH}"
echo "OS_NAME: ${OS_NAME}"
echo "IS_OFFLINE: ${OFFLINE_FLAG}"

function func_prepare_config_yaml() {
    local source_path=$1
    local dest_path=$2
    rm -fr "${dest_path}"
    mkdir "${dest_path}"
    cp -f "${source_path}"/hosts-conf-cm-2nodes.yml  "${dest_path}"/hosts-conf-cm.yml
    cp -f "${source_path}"/vars-conf-cm.yml  "${dest_path}"
    cp -f "${source_path}"/kubeanCluster.yml "${dest_path}"
    cp -f "${source_path}"/kubeanClusterOps.yml  "${dest_path}"
    sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${dest_path}/hosts-conf-cm.yml
    sed -i "s#image:#image: ${SPRAY_JOB}#" "${dest_path}"/kubeanClusterOps.yml
}

function func_generate_rsa_key(){
    echo 'y'| ssh-keygen -f ~/.ssh/id_rsa -t rsa -N ''
}

function func_ssh_login_no_password(){
  dest_ip=$1
  passwd=$2
  sshpass -p ${passwd} scp -o StrictHostKeyChecking=no -r /root/.ssh/id_rsa.pub  root@${dest_ip}:/root
  sshpass -p ${passwd} ssh root@${dest_ip} "mkdir /root/.ssh/; chmod 700 /root/.ssh"
  sshpass -p ${passwd} ssh root@${dest_ip} "touch  /root/.ssh/authorized_keys;chmod 600 /root/.ssh/authorized_keys"
  sshpass -p ${passwd} ssh root@${dest_ip} "cat  /root/id_rsa.pub >> /root/.ssh/authorized_keys"
}
### k8s upgrade test ####
util::power_on_2vms ${OS_NAME}
echo "==> scp sonobuoy bin to master: "
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no /home/kubernetes_e2e_images_v1.32.9.tar root@$vm_ip_addr1:/home
sshpass -p "${AMD_ROOT_PASSWORD}" scp  -o StrictHostKeyChecking=no "${REPO_ROOT}"/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/

##prepare kubean install job yml using docker：kube_version: 1.24.7
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-install-cluster-sonobouy/
func_prepare_config_yaml "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`

# set vars-conf-cm.yml
sed -i "s/1.32.9/1.32.0/"  "${dest_config_path}"/vars-conf-cm.yml

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml


## prepare cluster upgrade job yml --> upgrade from v1.31.6 to v1.31.9
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-upgrade-cluster-y/
func_prepare_config_yaml "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME2="cluster1-upgrade-y"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME2}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/1.32.9/1.32.9/"  "${dest_config_path}"/vars-conf-cm.yml

## prepare cluster upgrade job yml --> upgrade from v1.31.9 to v1.32.1
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-upgrade-cluster-z/
func_prepare_config_yaml "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME3="cluster1-upgrade-z"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME3}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/1.32.9/1.33.0/"  "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race -timeout=6h --fail-fast ./test/kubean_sonobouy_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"