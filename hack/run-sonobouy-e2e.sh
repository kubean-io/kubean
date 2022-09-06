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
vm_ip_addr1=${3:-"10.6.127.33"}
vm_ip_addr2=${4:-"10.6.127.36"}
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}
EXIT_CODE=0
echo "==> current dir: "$(pwd)
# Install ginkgo
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin

# prepare vagrant vm as k8 cluster single node
vm_clean_up(){
    vagrant destroy -f sonobouyDefault
    vagrant destroy -f sonobouyDefault2
    exit $EXIT_CODE
}

trap vm_clean_up EXIT
# create 1master+1worker cluster
cp $(pwd)/hack/Vagrantfile $(pwd)/
sed -i "s/sonobouyDefault_ip/${vm_ip_addr1}/" Vagrantfile
sed -i "s/sonobouyDefault2_ip/${vm_ip_addr2}/" Vagrantfile
vagrant up
vagrant status
ATTEMPTS=0
pingOK=0
ping -w 2 -c 1 $vm_ip_addr1|grep "0%" && pingOK=true || pingOK=false
until [ "${pingOK}" == "true" ] || [ $ATTEMPTS -eq 10 ]; do
ping -w 2 -c 1 $vm_ip_addr1|grep "0%" && pingOK=true || pingOK=false
echo "==> ping "$vm_ip_addr1 $pingOK
ATTEMPTS=$((ATTEMPTS + 1))
sleep 10
done

sshpass -p root ssh -o StrictHostKeyChecking=no root@${vm_ip_addr1} cat /proc/version
ping -c 5 ${vm_ip_addr1}
ping -c 5 ${vm_ip_addr2}
echo "==> scp sonobuoy bin to master: "
sshpass -p root scp  -o StrictHostKeyChecking=no $(pwd)/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/

# prepare kubean install job yml using docker
SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
cp $(pwd)/test/common/kubeanCluster.yml $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/
cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/hosts-conf-cm.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/kubeanClusterOps.yml
sed -i "s/containerd/docker/" $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/vars-conf-cm.yml
sed -i "s/v1.23.7/v1.22.12/" $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/vars-conf-cm.yml
sed -i "s#  \"10.6.170.10:5000\": \"http://10.6.170.10:5000\"#   - 10.6.170.10:5000#" $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/vars-conf-cm.yml

# prepare cluster upgrade job yml --> upgrade from v1.22.12 to v1.23.7
mkdir $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster
cp $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/* $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster/
sed -i "s/v1.22.12/v1.23.7/" $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster/vars-conf-cm.yml
sed -i "s/e2e-cluster1-install-sonobouy/e2e-upgrade-cluster/" $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster/kubeanClusterOps.yml

# prepare cluster upgrade job yml --> upgrade from v1.23.7 to v1.24.3
mkdir $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster24
cp $(pwd)/test/kubean_sonobouy_e2e/e2e-install-cluster-sonobouy/* $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster24/
sed -i "s/v1.22.12/v1.24.3/" $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster24/vars-conf-cm.yml
sed -i "s/e2e-cluster1-install-sonobouy/e2e-upgrade-cluster24/" $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster24/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" $(pwd)/test/kubean_sonobouy_e2e/e2e-upgrade-cluster24/kubeanClusterOps.yml

# Run nightly e2e
ginkgo -v -race --fail-fast ./test/kubean_sonobouy_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}" --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}"

# prepare kubean ops yml
cp $(pwd)/test/common/kubeanCluster.yml $(pwd)/test/kubeanOps_functions_e2e/e2e-install-cluster/
cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubeanOps_functions_e2e/e2e-install-cluster/
cp $(pwd)/test/common/kubeanCluster.yml $(pwd)/test/kubeanOps_functions_e2e/backofflimit-clusterops/
sed -i "s/cluster1/cluster2/" $(pwd)/test/kubeanOps_functions_e2e/backofflimit-clusterops/kubeanCluster.yml
cp $(pwd)/test/common/vars-conf-cm.yml $(pwd)/test/kubeanOps_functions_e2e/backofflimit-clusterops/
sed -i "s/cluster1/cluster2/" $(pwd)/test/kubeanOps_functions_e2e/backofflimit-clusterops/vars-conf-cm.yml
ginkgo -v -race --fail-fast ./test/kubeanOps_functions_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}" --vmipaddr="${vm_ip_addr1}"