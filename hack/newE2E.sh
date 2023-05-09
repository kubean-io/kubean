#!/usr/bin/env bash
set -o nounset
set -o pipefail
set -e

function init_vars() {
    echo $JOB_NAME $OS_TYPE $ARCH $Network_TYPE $GAP_TYPE
}

function execute_case() {
    case $JOB_NAME in 
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
            #todo
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
        "*")
            echo "no such $JOB_NAME, exit"
            ;;
        esac
}

function main() {
    init_vars
    execute_case
    #exit 0
}

main $@