#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

### this script is shared by offline os compatibility and online compatibility
### only cover redhat8 and redhat7
set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

os_compability_e2e(){
    OS_NAME=${1}
    ARCH="amd64"
    if [ "${OFFLINE_FLAG}" == "false" ];then
      util::vm_name_ip_init_online_by_os ${OS_NAME}
    else
      util::vm_name_ip_init_offline_by_os ${OS_NAME}
    fi
    echo "vm_name1: ${vm_name1}"
    echo "vm_name2: ${vm_name2}"
    SNAPSHOT_NAME="os-installed"
    dest_yaml_path="${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
    sleep 20
    util::wait_ip_reachable "${vm_ip_addr1}" 10
    util::wait_ip_reachable "${vm_ip_addr2}" 10
    ping -c 5 ${vm_ip_addr1}
    ping -c 5 ${vm_ip_addr2}
    rm -f ~/.ssh/known_hosts
    echo "==> scp sonobuoy bin to master: "
    sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
    sshpass -p ${AMD_ROOT_PASSWORD} ssh root@$vm_ip_addr1 "chmod +x /usr/bin/sonobuoy"

    # prepare kubean install job yml using containerd
    CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
    echo "CLUSTER_OPERATION_NAME1: $CLUSTER_OPERATION_NAME1"
    rm -fr "${dest_yaml_path}"
    mkdir "${dest_yaml_path}"
    cp -f "${SOURCE_CONFIG_PATH}"/hosts-conf-cm-2nodes.yml "${dest_yaml_path}"/hosts-conf-cm.yml
    cp -f "${SOURCE_CONFIG_PATH}"/vars-conf-cm.yml  "${dest_yaml_path}"
    cp -f "${SOURCE_CONFIG_PATH}"/kubeanCluster.yml "${dest_yaml_path}"
    cp -f "${SOURCE_CONFIG_PATH}"/kubeanClusterOps.yml  "${dest_yaml_path}"
    sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" ${dest_yaml_path}/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" ${dest_yaml_path}/hosts-conf-cm.yml
    sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${dest_yaml_path}/hosts-conf-cm.yml
    sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  ${dest_yaml_path}/kubeanClusterOps.yml
    sed -i "s#image:.*#image: ${SPRAY_JOB}#" ${dest_yaml_path}/kubeanClusterOps.yml
    if [[ "${OFFLINE_FLAG}" == "true" ]] &&  [[ ${OS_NAME} =~ REDHAT.* ]]; then
      if [[ ${OS_NAME} == REDHAT8 ]]; then
        # kubeanClusterOps.yml sed
        echo "REDHAT8 OS."
        sed -i 's#basearch#basearch/AppStream,{offline_minio_url}/kubean/redhat-iso/\\\$releasever/os/\\\$basearch/BaseOS#2' ${dest_yaml_path}/kubeanClusterOps.yml
        sed -i "s#m,{#m','{#g" ${dest_yaml_path}/kubeanClusterOps.yml
      fi
      sed -i "s#{offline_minio_url}#${MINIO_URL}#g" ${dest_yaml_path}/kubeanClusterOps.yml
      sed -i  "s#centos#redhat#g" ${dest_yaml_path}/kubeanClusterOps.yml
      # vars-conf-cm.yml set
      sed -i "s#registry_host:#registry_host: ${registry_addr_amd64}#"    ${dest_yaml_path}/vars-conf-cm.yml
      sed -i "s#minio_address:#minio_address: ${MINIO_URL}#"    ${dest_yaml_path}/vars-conf-cm.yml
      sed -i "s#registry_host_key#${registry_addr_amd64}#g"    ${dest_yaml_path}/vars-conf-cm.yml
      sed -i "s#{{ files_repo }}/centos#{{ files_repo }}/redhat#" ${dest_yaml_path}/vars-conf-cm.yml
      sed -i "$ a\    rhel_enable_repos: false"  ${dest_yaml_path}/vars-conf-cm.yml
    fi
    # Run cluster function e2e
    ginkgo -v -timeout=10h -race --fail-fast ./test/kubean_os_compatibility_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
                     --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
                     --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"
    SNAPSHOT_NAME="power-down"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
}

###### OS compitable e2e logic ########
os_array=("REDHAT7" "REDHAT8" )
echo "OS list: ${os_array[*]}"
echo ${#os_array[@]}
for OS in "${os_array[@]}"; do
    echo "***************OS is: ${OS} ***************"
    os_compability_e2e ${OS}
done