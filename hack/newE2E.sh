#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -e
shopt -s nocasematch

function init_vars() {
    export CURRENT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd) # hack
    export PERRENT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && cd .. && pwd) # kubean
    echo $GITHUB_JOB $OS_TYPE $ARCH $Network_TYPE $GAP_TYPE
}

function execute_case() {
    case $GITHUB_JOB in 
        "centos_calico_airgap")
            #todo
            ;;
        "centos_calico_online")
            bash hack/e2e.sh $HELM_CHART_VERSION $CONTAINER_TAG $runner_name $VSPHERE_USER $VSPHERE_PASSWD $AMD_ROOT_PASSWORD $KYLIN_VM_PASSWORD "NIGHTLY"
            ;;
        "centos_cilium_online")
            bash hack/offline-e2e.sh $HELM_CHART_VERSION $VSPHERE_USER $VSPHERE_PASSWD $AMD_ROOT_PASSWORD $KYLIN_VM_PASSWORD $runner_name
            ;;
        "centos_cilium_airgap")
            # kubean_ipvs_cluster_e2e
            # kubean_cilium_cluster_e2e # skip this case cause' some other testcase is imbeded in this network case that cannot decouple code in test/
            # kubean_calico_dualstack_e2e # skip this case
            # kubean_calico_single_stack_e2e
            test_case_string="kubean_ipvs_cluster_e2e,kubean_calico_single_stack_e2e"
            OLD_IFS="$IFS"
            IFS=","
            test_case_arr=($test_case_string)
            bash hack/network_testcase.sh ${test_case_arr[@]}
            ;;
        "redhat_calico_online")
            #todo
            ;;
        "redhat_calico_airgap")
            #todo
            ;;
        "redhat_cilium_airgap")
            #todo
            ;;
        "redhat_cilium_online")
            #todo
            ;;
        "kylin_calico_online")
            #todo
            ;;
        "kylin_calico_airgap")
            #todo
            ;;
        *)
            echo "no such $GITHUB_JOB, exit"
            ;;
        esac
}

function main() {
    init_vars
    execute_case
    #exit 0
}

main $@