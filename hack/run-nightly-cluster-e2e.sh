#!/usr/bin/env bash

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


### cluster add worker and remove worker ###
util::vm_name_ip_init_online_by_os ${OS_NAME}
util::power_on_2vms ${OS_NAME}
func_generate_rsa_key
func_ssh_login_no_password ${vm_ip_addr1} ${AMD_ROOT_PASSWORD}
func_ssh_login_no_password ${vm_ip_addr2} ${AMD_ROOT_PASSWORD}
echo "==> scp sonobuoy bin to master: "
sshpass -p "${AMD_ROOT_PASSWORD}" scp  -o StrictHostKeyChecking=no "${REPO_ROOT}"/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/

## prepare kubean install job yml files
source_config_path="${REPO_ROOT}"/test/common
dest_config_path="${REPO_ROOT}"/test/kubean_add_remove_worker_nightlye2e/e2e-install-1node-cluster-prikey
rm -fr "${dest_config_path}"
mkdir "${dest_config_path}"
cp -f "${source_config_path}"/hosts-conf-cm.yml  "${dest_config_path}"
cp -f "${source_config_path}"/vars-conf-cm.yml  "${dest_config_path}"
cp -f "${source_config_path}"/kubeanCluster.yml "${dest_config_path}"
cp -f "${source_config_path}"/kubeanClusterOps.yml  "${dest_config_path}"

# set hosts-conf-cm.yml
sed -i '/ansible_connection/d' "${dest_config_path}"/hosts-conf-cm.yml
sed -i '/ansible_user/d' "${dest_config_path}"/hosts-conf-cm.yml
sed -i '/ansible_password/d' "${dest_config_path}"/hosts-conf-cm.yml
sed -i "s/ip:/ip: ${vm_ip_addr1}/" "${dest_config_path}"/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr1}/" "${dest_config_path}"/hosts-conf-cm.yml

## set kubeanClusterOps.yml
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${dest_config_path}"/kubeanClusterOps.yml

# set kubeanCluster.yml
cat >> "${dest_config_path}"/kubeanCluster.yml <<EOF
  sshAuthRef:
    namespace: kubean-system
    name: sample-ssh-auth
EOF

# write the secret file for private key
ID_RSA=$(cat ~/.ssh/id_rsa|base64 -w 0)
cat > "${dest_config_path}"/ssh_auth_secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: sample-ssh-auth
  namespace: kubean-system
type: kubernetes.io/ssh-auth
data:
  ssh-privatekey: |
    ${ID_RSA}
EOF

## prepare kubean add worker job yml files
dest_config_path="${REPO_ROOT}"/test/kubean_add_remove_worker_nightlye2e/add-worker-node
rm -fr "${dest_config_path}"
mkdir "${dest_config_path}"
cp ${REPO_ROOT}/test/common/hosts-conf-cm-2nodes.yml ${dest_config_path}/hosts-conf-cm.yml
cp ${REPO_ROOT}/test/common/kubeanClusterOps.yml ${dest_config_path}

# set hosts-conf-cm.yml
sed -i '/ansible_connection/d' "${dest_config_path}"/hosts-conf-cm.yml
sed -i '/ansible_user/d' "${dest_config_path}"/hosts-conf-cm.yml
sed -i '/ansible_password/d' "${dest_config_path}"/hosts-conf-cm.yml
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${dest_config_path}"/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${dest_config_path}"/hosts-conf-cm.yml

# set kubeanClusterOps.yml
CLUSTER_OPERATION_NAME2="cluster1-add-worker"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME2}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/action: cluster.yml/action: scale.yml\n  extraArgs: --limit=node2/"  ${dest_config_path}/kubeanClusterOps.yml

## prepare kubean remove worker job yml files
dest_config_path="${REPO_ROOT}"/test/kubean_add_remove_worker_nightlye2e/remove-worker-node
rm -fr "${dest_config_path}"
mkdir "${dest_config_path}"
cp ${REPO_ROOT}/test/common/kubeanClusterOps.yml ${dest_config_path}

CLUSTER_OPERATION_NAME3="cluster1-remove-worker"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME3}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/action: cluster.yml/action: remove-node.yml\n  extraArgs: -e node=node2/"  ${dest_config_path}/kubeanClusterOps.yml
ginkgo -v -timeout=3h -race --fail-fast ./test/kubean_add_remove_worker_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

### k8s upgrade test ####
util::power_on_2vms ${OS_NAME}
echo "==> scp sonobuoy bin to master: "
sshpass -p "${AMD_ROOT_PASSWORD}" scp  -o StrictHostKeyChecking=no "${REPO_ROOT}"/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/

##prepare kubean install job yml using docker：kube_version: v1.24.7
source_config_path="${REPO_ROOT}"/test/common
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-install-cluster-sonobouy/
func_prepare_config_yaml "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml


## prepare cluster upgrade job yml --> upgrade from v1.24.7 to v1.25.3
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-upgrade-cluster-y/
func_prepare_config_yaml "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME2="cluster1-upgrade-y"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME2}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/v1.24.7/v1.25.3/"  "${dest_config_path}"/vars-conf-cm.yml

## prepare cluster upgrade job yml --> upgrade from v1.25.3 to v1.25.5
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-upgrade-cluster-z/
func_prepare_config_yaml "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME3="cluster1-upgrade-z"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME3}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/v1.24.7/v1.25.5/"  "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race -timeout=6h --fail-fast ./test/kubean_sonobouy_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"