#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

# YQ reference： https://mikefarah.gitbook.io/yq/operators/string-operators

#set -o errexit
set -o nounset
set -o pipefail
shopt -s nocasematch

source "${CURRENT_DIR}"/util.sh
source "${CURRENT_DIR}"/offline-util.sh
source "${CURRENT_DIR}"/yml_utils.sh

function prepare_global_vars() {
    export TARGET_VERSION=${HELM_CHART_VERSION}
    export IMAGE_VERSION=${CONTAINER_TAG}
    export SPRAY_JOB_VERSION=${CONTAINER_TAG}
    export SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
    export HELM_REPO="https://kubean-io.github.io/kubean-helm-chart"
    export IMG_REGISTRY="ghcr.m.daocloud.io"
    export VSPHERE_HOST="10.64.56.11"
    export OFFLINE_FLAG=false
    export KUBECONFIG_PATH="${HOME}/.kube"
    export CLUSTER_PREFIX="kubean-online-$RANDOM"
    export CONTAINERS_PREFIX="kubean-online"
    export KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
    export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
    export POWER_ON_SNAPSHOT_NAME="os-installed"
    export POWER_DOWN_SNAPSHOT_NAME="power-down"
    export LOCAL_REPO_ALIAS="kubean_release"
    export LOCAL_RELEASE_NAME=kubean
    export E2eInstallClusterYamlFolder="e2e-install-cluster"

    util::clean_online_kind_cluster
    KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.26.0"
    ./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REGISTRY}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host
    util::check_yq_intalled
}

function prepare_case_vars() {
    export vm_ip_addr1=$(yq ".ONLINE.debug.$OS_NAME[0].ip" $CURRENT_DIR/vm_config.yaml)
    export vm_ip_addr2=$(yq ".ONLINE.debug.$OS_NAME[1].ip" $CURRENT_DIR/vm_config.yaml)
}

function prepare_sonobuoy() {
    sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
}

####################### create ipvs cluster ################
function kubean_ipvs_cluster_e2e() {
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    yamlUtil::update_VarsConfCM "iptables" "ipvs"
    CLUSTER_OPERATION_NAME1="cluster1-ipvs-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    prepare_sonobuoy
    mkdir ${REPO_ROOT}/test/kubean_ipvs_cluster_e2e/${E2eInstallClusterYamlFolder}
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_ipvs_cluster_e2e/${E2eInstallClusterYamlFolder}/
    ginkgo -v -race -timeout=3h  --fail-fast "${REPO_ROOT}"/test/kubean_ipvs_cluster_e2e -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"
}

####################### create cilium cluster ###############
function kubean_cilium_cluster_e2e() {
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    echo "create cilium cluster....."
    echo "OS_NAME: ${OS_NAME}"
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-cilium-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.kube_network_plugin = "cilium"' $groupVarYml
    yq -i '.kube_service_addresses = "10.88.0.0/16"'  $groupVarYml
    yq -i '.kube_pods_subnet = "192.88.128.0/20"'  $groupVarYml
    yq -i '.kube_network_node_prefix = 24'  $groupVarYml
    yq -i '.cilium_kube_proxy_replacement = "partial"'  $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    mkdir ${REPO_ROOT}/test/kubean_cilium_cluster_e2e/e2e-install-cilium-cluster
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_cilium_cluster_e2e/e2e-install-cilium-cluster/
    ginkgo -v -race -timeout=3h  --fail-fast ./test/kubean_cilium_cluster_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"
}

####################### calico dual stack ###############
# Description:
# calico dual stack cluster need install on a Redhat8 os
# and the testing vm need add a ipv6 in snapshot
########################################################
function kubean_calico_dualstack_e2e() {
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    #### calico dual stuck: VXLAN_ALWAYS-VXLAN_ALWAYS ####
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-vxlan-always-vxlan-always-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ipip_mode = "Never"' $groupVarYml
    yq -i '.calico_vxlan_mode = "Always"' $groupVarYml
    yq -i '.calico_ipip_mode_ipv6 = "Never"'  $groupVarYml
    yq -i '.calico_vxlan_mode_ipv6 = "Always"'  $groupVarYml
    yq -i '.enable_dual_stack_networks = "true"'  $groupVarYml
    yq -i '.kube_pods_subnet_ipv6 = "fd89:ee78:d8a6:8608::1:0000/112"'  $groupVarYml
    yq -i '.kube_service_addresses_ipv6 = "fd89:ee78:d8a6:8608::1000/116"'  $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    mkdir ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster #TBD
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_ALWAYS-VXLAN_ALWAYS"

    #### calico dual stuck: VXLAN_CrossSubnet-VXLAN_ALWAYS ####
    # precondtion: power_on_2vms
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-vxlan-cross-vxlan-always-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ipip_mode = "Never"' $groupVarYml
    yq -i '.calico_vxlan_mode = "CrossSubnet"' $groupVarYml
    yq -i '.calico_ipip_mode_ipv6 = "Never"'  $groupVarYml
    yq -i '.calico_vxlan_mode_ipv6 = "Always"'  $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    rm -f ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/*
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet-VXLAN_ALWAYS"

    #### calico dual stuck: IPIP_ALWAYS-VXLAN_CrossSubnet ####
    # precondtion: power_on_2vms
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-ipip-always-vxlan-cross-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ipip_mode = "Always"' $groupVarYml
    yq -i '.calico_vxlan_mode = "Never"' $groupVarYml
    yq -i '.calico_ipip_mode_ipv6 = "Never"'  $groupVarYml
    yq -i '.calico_network_backend = "bird"'  $groupVarYml
    yq -i '.calico_vxlan_mode_ipv6 = "CrossSubnet"'  $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    rm -f ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/*
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always-VXLAN_CrossSubnet"

    #### calico dual stuck: IPIP_CrossSubnet-VXLAN_CrossSubnet ####
    # precondtion: power_on_2vms
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-ipip-cross-vxlan-cross-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ipip_mode = "CrossSubnet"' $groupVarYml
    yq -i '.calico_vxlan_mode = "Never"' $groupVarYml
    yq -i '.calico_ipip_mode_ipv6 = "Never"'  $groupVarYml
    yq -i '.calico_network_backend = "bird"'  $groupVarYml
    yq -i '.calico_vxlan_mode_ipv6 = "CrossSubnet"'  $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    rm -f ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/*
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet-VXLAN_CrossSubnet"
}

#######################calico single stack #######################
function kubean_calico_single_stack_e2e() {
    ### CALICO: IPIP_ALWAYS ###
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    # preparation: CRDS
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-always-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ip_auto_method = "first-found"' $groupVarYml
    yq -i '.calico_ip6_auto_method = "first-found"' $groupVarYml
    yq -i '.calico_ipip_mode = "Always"' $groupVarYml
    yq -i '.calico_vxlan_mode = "Never"' $groupVarYml
    yq -i '.calico_network_backend = "bird"' $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    mkdir ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always"

    ### CALICO: IPIP_CrossSubnet ###
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-cross-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ip_auto_method = "kubernetes-internal-ip"' $groupVarYml
    yq -i '.calico_ip6_auto_method = "kubernetes-internal-ip"' $groupVarYml
    yq -i '.calico_ipip_mode = "CrossSubnet"' $groupVarYml
    yq -i '.calico_vxlan_mode = "Never"' $groupVarYml
    yq -i '.calico_network_backend = "bird"' $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    rm -f ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/*
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet"

    ### CALICO: VXLAN_ALWAYS ###
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-always-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ip_auto_method = "kubernetes-internal-ip"' $groupVarYml
    yq -i '.calico_ip6_auto_method = "kubernetes-internal-ip"' $groupVarYml
    yq -i '.calico_ipip_mode = "Never"' $groupVarYml
    yq -i '.calico_vxlan_mode = "Always"' $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    rm -f ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/*
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_Always"


    ### CALICO: VXLAN_CrossSubnet ###
    export crd_dir=$(mktemp -d)
    echo $crd_dir
    # precondtion: power_on_2vms
    util::power_on_2vms ${OS_NAME}
    yamlUtil::prepare_CRDs
    yamlUtil::prepare_hosts_cm
    CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-cross-"`date "+%H-%M-%S"`
    yamlUtil::prepare_clusteroperation $CLUSTER_OPERATION_NAME1
    ##### Start: abstract group vars from VarsConfCM.yml
    yamlUtil::abstract_groupVars
    yq -i '.calico_ip_auto_method = "kubernetes-internal-ip"' $groupVarYml
    yq -i '.calico_ip6_auto_method = "kubernetes-internal-ip"' $groupVarYml
    yq -i '.calico_ipip_mode = "Never"' $groupVarYml
    yq -i '.calico_vxlan_mode = "CrossSubnet"' $groupVarYml
    #### End: write back group vars to VarsConfCM.yml
    yamlUtil::update_groupVars
    prepare_sonobuoy
    rm -f ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/*
    cp $crd_dir/* ${REPO_ROOT}/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster/
    ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet"
}

main() {
    tcList=$@
    prepare_global_vars
    #echo 'network testcases: ' ${tcList[@]}
    for case in ${tcList[@]}; do
        echo "🥦🥦🥦run testcase: "$case
        case $case in
            "kubean_ipvs_cluster_e2e")
            export OS_NAME="CENTOS7"
            prepare_case_vars
            kubean_ipvs_cluster_e2e
            ;;
            "kubean_cilium_cluster_e2e")
            export OS_NAME="CENTOS7-HK"
            prepare_case_vars
            kubean_cilium_cluster_e2e
            ;;
            "kubean_calico_dualstack_e2e")
            export OS_NAME="CENTOS7-HK"
            prepare_case_vars
            kubean_calico_dualstack_e2e
            ;;
            "kubean_calico_single_stack_e2e")
            export OS_NAME="CENTOS7"
            prepare_case_vars
            kubean_calico_single_stack_e2e
            ;;
            *)
            echo "no such testcase: $case, exit"
            ;;
        esac
    done
    util::clean_online_kind_cluster
}

main $@