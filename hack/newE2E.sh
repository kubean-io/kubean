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
    echo $GITHUB_JOB $ARCH $GAP_TYPE
}

function execute_case() {
    case $GITHUB_JOB in
        "network_e2e_online")
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
        *)
            echo "unsupported E2E job: $GITHUB_JOB"
            exit 1
            ;;
    esac
}

function main() {
    init_vars
    execute_case
    #exit 0
}

main $@