#!/usr/bin/env bash
set -o nounset
set -o pipefail
set -e
# This script schedules e2e tests
# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REPO](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from
#
TARGET_VERSION=${1:-latest}
IMAGE_VERSION=${2:-latest}
HELM_REPO=${3:-"https://release.daocloud.io/chartrepo/kubean"}
IMG_REPO=${4:-"release.daocloud.io/kubean"}
EXIT_CODE=0

CLUSTER_PREFIX=kubean-"${IMAGE_VERSION}"-$RANDOM

chmod +x ./hack/delete-cluster.sh
chmod +x ./hack/local-up-kindcluster.sh
chmod +x ./hack/run-e2e.sh

###### Clean Up #######
clean_up(){
    echo 'Removing kubean kind cluster...'
    kind delete cluster --name "${CLUSTER_PREFIX}"-host
    echo 'Done!'
}

###### e2e logic ########

trap clean_up EXIT
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "release.daocloud.io/kpanda/kindest-node:v1.21.1" "${CLUSTER_PREFIX}"-host

kubeconf_pos='/tmp/kind_cluster.conf'
current_dir=$(pwd) # 获取当前路径
echo 'current_dir: '${current_dir}
vm_ipaddr='10.6.127.12'
###### e2e install test execution ########
ginkgo run -v -race --fail-fast --focus="\[install\]" ./test/e2e/

reset_ops_name='cluster1-ops-reset-1e2w3q'
install_ops_name='cluster1-ops-install-1e2w3q'
###### e2e apply ops test execution ########
# TBD: 首先要检查vm上是否已经有k8 cluster；如果有的话则reset再安装
kubectl --kubeconfig=${kubeconf_pos} apply -f ${current_dir}/artifacts/example/e2e_reset
sleep 10
# 检查reset job执行是否成功
ATTEMPTS=0
check_cmd=2
check_cmd=`kubectl --kubeconfig=${kubeconf_pos}  -n kubean-system get job ${reset_ops_name}-job -o jsonpath='{.status.succeeded}'`
until [ "$check_cmd" == "1" ] || [ $ATTEMPTS -eq 60 ]; do
check_cmd=`kubectl --kubeconfig=${kubeconf_pos}  -n kubean-system get job ${reset_ops_name}-job -o jsonpath='{.status.succeeded}'`
echo 'check_cmd: '$check_cmd
ATTEMPTS=$((ATTEMPTS + 1))
sleep 30
done

# 安装k8 cluster
kubectl --kubeconfig=${kubeconf_pos} apply -f ${current_dir}/artifacts/example/e2e_install
sleep 10
kubectl --kubeconfig=${kubeconf_pos} -n kubean-system get jobs
# 检查install job执行是否成功
ATTEMPTS=0
check_cmd=2
check_cmd=`kubectl --kubeconfig=${kubeconf_pos}  -n kubean-system get job ${install_ops_name}-job -o jsonpath='{.status.succeeded}'`
until [ "$check_cmd" == "1" ] || [ $ATTEMPTS -eq 60 ]; do
check_cmd=`kubectl --kubeconfig=${kubeconf_pos}  -n kubean-system get job ${install_ops_name}-job -o jsonpath='{.status.succeeded}'`
echo 'check_cmd: '$check_cmd
ATTEMPTS=$((ATTEMPTS + 1))
sleep 30
done

# sshpass 免密获取k8集群的kubeconfig文件: sshpass -p 'dangerous' scp root@xx:/root/.kube/config .
# 暂定k8集群环境的ip是10.6.127.12；用户名密码是root/dangerous；此处需与hosts-conf-cm yml文件保持一致
sshpass -p 'dangerous' scp root@${vm_ipaddr}:/root/.kube/config ${current_dir}/test/tools/
sed -i  "s/127.0.0.1:.*/${vm_ipaddr}:6443/" ${current_dir}/test/tools/config
ginkgo run -v -race --fail-fast --focus="\[create\]" ${current_dir}/test/e2e/


ret=$?
if [ ${ret} -ne 0 ]; then
  EXIT_CODE=1
fi