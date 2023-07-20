#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

# YQ reference： https://mikefarah.gitbook.io/yq/operators/string-operators

set -o errexit
set -o nounset
set -o pipefail
shopt -s nocasematch

function yamlUtil::prepare_CRDs() {
    case $GAP_TYPE in 
        "ONLINE")
            cp $PERRENT_DIR/examples/install/*mirror/* $crd_dir/
            ;;
        "OFFLINE")
            cp $PERRENT_DIR/examples/install/*airgap/* $crd_dir/
            ;;
        "*")
            echo "no such $GAP_TYPE, exit"
            ;;
    esac
}

function yamlUtil::prepare_hosts_cm() {
    # replace hosts yaml
    yq eval -o yaml '.data."hosts.yml"' $crd_dir/HostsConfCM.yml |tee $crd_dir/hosts.yaml  >/dev/null
    yq -i ".all.hosts.node1.ip=\"$vm_ip_addr1\""  $crd_dir/hosts.yaml
    yq -i ".all.hosts.node1.access_ip=\"$vm_ip_addr1\""  $crd_dir/hosts.yaml
    yq -i ".all.hosts.node1.ansible_host=\"$vm_ip_addr1\""  $crd_dir/hosts.yaml
    yq -i ".all.hosts.node2.ip=\"$vm_ip_addr2\""  $crd_dir/hosts.yaml
    yq -i ".all.hosts.node2.access_ip=\"$vm_ip_addr2\""  $crd_dir/hosts.yaml
    yq -i ".all.hosts.node2.ansible_host=\"$vm_ip_addr2\""  $crd_dir/hosts.yaml

    yq -i "with(.all.hosts[]; .ansible_user=\"root\")" $crd_dir/hosts.yaml
    yq -i "with(.all.hosts[]; .ansible_password=\"$AMD_ROOT_PASSWORD\")" $crd_dir/hosts.yaml

    # replace HostsConfCM.yml with new hosts.yml string
    export hosts_string=$(cat $crd_dir/hosts.yaml)
    yq -i '.data."hosts.yml" = strenv(hosts_string)' $crd_dir/HostsConfCM.yml
    rm -f $crd_dir/hosts.yaml
}

function yamlUtil::prepare_clusteroperation() {
    local CLUSTER_OPERATION_NAME1=${1}
    local SPRAY_JOB=ghcr.m.daocloud.io/kubean-io/spray-job:v0.6.1-7-ge0d10a35-e2e #for test
    yq -i ".spec.image=\"$SPRAY_JOB\""  $crd_dir/ClusterOperation.yml
    yq -i ".metadata.name=\"$CLUSTER_OPERATION_NAME1\""  $crd_dir/ClusterOperation.yml
    yq -i ".metadata.labels.clusterName=\"cluster1\""  $crd_dir/ClusterOperation.yml

    yq -i ".metadata.name=\"cluster1\""  $crd_dir/Cluster.yml
    yq -i ".spec.hostsConfRef.name=\"cluster1-hosts-conf\""  $crd_dir/Cluster.yml
    yq -i ".spec.varsConfRef.name=\"cluster1-vars-conf\""  $crd_dir/Cluster.yml
    yq -i ".metadata.labels.clusterName=\"cluster1\""  $crd_dir/Cluster.yml
    yq -i ".spec.cluster=\"cluster1\""  $crd_dir/ClusterOperation.yml

    yq -i ".metadata.name=\"cluster1-vars-conf\""  $crd_dir/VarsConfCM.yml
    yq -i ".metadata.name=\"cluster1-hosts-conf\""  $crd_dir/HostsConfCM.yml
}

function yamlUtil::update_VarsConfCM() {
    export origin_str=${1}
    export replace_str=${2}
    yq -i '.data."group_vars.yml" |= sub(strenv(origin_str), strenv(replace_str))' $crd_dir/VarsConfCM.yml
}

function yamlUtil::abstract_groupVars() {
    # create group_vars.yml
    export groupVarYml=$crd_dir/group_vars.yml # export groupVarYml as global variable
    touch $groupVarYml # create one group vars yaml
    yq eval -o yaml '.data."group_vars.yml"' $crd_dir/VarsConfCM.yml |tee $groupVarYml >/dev/null

    yq -i '.kube_apiserver_port = "6443"'  $groupVarYml
    yq -i '.metrics_server_enabled = "true"'  $groupVarYml
    yq -i '.local_path_provisioner_enabled = "true"'  $groupVarYml
    yq -i '.kubeadm_init_timeout = "600s"'  $groupVarYml
    yq -i '.kube_version = "v1.25.3"'  $groupVarYml
}

function yamlUtil::update_groupVars() {
    yq -i '.data."group_vars.yml"  |= load_str(strenv(groupVarYml))' $crd_dir/VarsConfCM.yml
    #tail -n 20 $crd_dir/VarsConfCM.yml # for test
    # delete group_vars.yml
    rm -f $groupVarYml
}
