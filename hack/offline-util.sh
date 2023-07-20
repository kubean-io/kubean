#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail


#####################################
### Restore vm snapshot to only os installed state
function util::restore_vsphere_vm_snapshot() {
  VSPHERE_HOST=${1}
  VSPHERE_PASSWD=${2}
  VSPHERE_USER=${3}
  SNAPSHOT_NAME=${4}
  vm_name=${5}
  echo "Start restore vm snapshot..."
  # shell脚本不支持传数组，就改用文件方式获取虚拟机名称列表 hack/vm_name.list
  # vsphere python package
  if [ ! -d "pyvmomi-community-samples" ]; then
    pip3 install -v pyvmomi==7.0.3
    for ((i=0;i <5; i++)){
       cloneOk="true"
       git clone https://github.com/vmware/pyvmomi-community-samples.git  || cloneOk="false"
       if [ ${cloneOk} == "false" ];then
          echo "pyvmomi clone failed, try later..."
          sleep 10
       else
         echo "pyvmomi clone ok."
         break
       fi
    }


  else
    echo "vmware python repo exist"
  fi

  if [ $vm_name != "" ]; then
      echo "restore $vm_name..."
      python3 pyvmomi-community-samples/samples/snapshot_operations.py -s ${VSPHERE_HOST} -u ${VSPHERE_USER} -p ${VSPHERE_PASSWD} -nossl -v "${vm_name}" -op revert --snapshot-name ${SNAPSHOT_NAME}
  else
      echo "vm_name empty, exit."
      exit 1
  fi
  echo "Restore vm snapshot end!"
}

### Set vm name && ip init
### /official mode is  differentiated by RUNNER_NAME
### only one runner in debug mode, named: debug
### official may be several runners, but only one active CI procedure, so ignore runner name
### parameter:OS_NAME value ["CENTOS", "KYLIN", "*"]
### REDHAT7, REDHAT8 and other os vm will be setup by vagrant Serially, so reuse the same ip
function util::vm_name_ip_init_offline_by_os(){
  echo "RUNNER NAME: " $RUNNER_NAME
  declare -u  OS_NAME=$1
  echo "OS_NAME: " ${OS_NAME}
  if [ "${RUNNER_NAME}" == "debug" ]; then
    case ${OS_NAME} in
        "KYLINV10")
           vm_ip_addr1="10.5.127.167"
           vm_ip_addr2="10.5.127.168"
           vm_name1="gwt-kubean-offline-e2e-kylin-node3"
           vm_name2="gwt-kubean-offline-e2e-kylin-node4"
           ;;
        "CENTOS7")
            vm_ip_addr1="10.16.10.163"
            vm_ip_addr2="10.16.10.164"
            vm_name1="gwt-kubean-offline-e2e-node3"
            vm_name2="gwt-kubean-offline-e2e-node4"
            ;;
        "REDHAT8")
            vm_ip_addr1="10.5.127.205"
            vm_ip_addr2="10.5.127.206"
            vm_name1="gwt-kubean-offline-e2e-redhat8-node205"
            vm_name2="gwt-kubean-offline-e2e-redhat8-node206"
            ;;
        "REDHAT7")
            vm_ip_addr1="10.5.127.207"
            vm_ip_addr2="10.5.127.208"
            vm_name1="gwt-kubean-offline-e2e-redhat7-node207"
            vm_name2="gwt-kubean-offline-e2e-redhat7-node208"
            ;;
    esac
  else
    case ${OS_NAME} in
        "KYLINV10")
          vm_ip_addr1="10.5.127.165"
          vm_ip_addr2="10.5.127.166"
          vm_name1="gwt-kubean-offline-e2e-kylin-node1"
          vm_name2="gwt-kubean-offline-e2e-kylin-node2"
           ;;
        "CENTOS7")
          vm_ip_addr1="10.16.10.161"
          vm_ip_addr2="10.16.10.162"
          vm_name1="gwt-kubean-offline-e2e-node1"
          vm_name2="gwt-kubean-offline-e2e-node2"
          ;;
        "REDHAT8")
            vm_ip_addr1="10.5.127.201"
            vm_ip_addr2="10.5.127.202"
            vm_name1="gwt-kubean-offline-e2e-redhat8-node201"
            vm_name2="gwt-kubean-offline-e2e-redhat8-node202"
            ;;
        "REDHAT7")
            vm_ip_addr1="10.5.127.203"
            vm_ip_addr2="10.5.127.204"
            vm_name1="gwt-kubean-offline-e2e-redhat7-node203"
            vm_name2="gwt-kubean-offline-e2e-redhat7-node204"
    esac
  fi
  echo "vm name1:  $vm_name1"
  echo "vm name2:  $vm_name2"
  echo "vm_ip_addr1: $vm_ip_addr1"
  echo "vm_ip_addr2: $vm_ip_addr2"
}

function util::init_kylin_vm_template_map(){
  if [ "${RUNNER_NAME}" == "debug" ]; then
    template_name1="gwt-kubean-offline-e2e-kylin-template3"
    template_name2="gwt-kubean-offline-e2e-kylin-template4"
  else
    template_name1="gwt-kubean-offline-e2e-kylin-template1"
    template_name2="gwt-kubean-offline-e2e-kylin-template2"
  fi
}

### Use skopeo copy images, which used in golang case, to docker registry
function util::scope_copy_test_images(){
   dest_registry_addr=${1}
   image_name=${2:-""}
   skopeo_cmd="skopeo copy --insecure-policy --src-tls-verify=false --dest-tls-verify=false  "
   if [ "${image_name}" != "" ]; then
        echo "skopeo copy image to registry: ${image_name}"
        ${skopeo_cmd} docker://"${image_name}"  docker://"${dest_registry_addr}"/test/"${image_name}"
    else
      for image_name in $(cat hack/test_images.list);
      do
         echo "skopeo copy image to registry: ${image_name}"
         ${skopeo_cmd} docker://"${image_name}"  docker://"${dest_registry_addr}"/test/"${image_name}" > /dev/null
      done
    fi
    echo "Skopeo copy images end!"
}



### OS COMPATIBILITY TEST: Kylin os use clone template
function util::init_kylin_vm(){
  local template_name=${1}
  local vm_name=${2}
  local arm_server_ip=${3}
  local arm_server_password=${4}
  local sshpass_cmd_prefix="sshpass -p ${arm_server_password} ssh -o StrictHostKeyChecking=no root@${arm_server_ip} "
  util::delete_kylin_vm ${vm_name} ${arm_server_ip} ${arm_server_password}

  # Delete vm file before create vm
  vm_image_file="/dce/images/${vm_name}.img"
  img_file_exist_cmd=${sshpass_cmd_prefix}" ls -l /dce/images|grep ${vm_name}||echo true"
  img_file_exist=$(${img_file_exist_cmd})
  if [ ! "${img_file_exist}" == "true" ]; then
    echo "delete ${vm_image_file}"
    rm_already_vm_file_cmd=${sshpass_cmd_prefix}" rm ${vm_image_file}||echo true"
    eval "${rm_already_vm_file_cmd}"
  fi
  sleep 5
  # Clone vm
  echo "Clone vm ${vm_name}..."
  clone_cmd=${sshpass_cmd_prefix}" virt-clone --original ${template_name} --name ${vm_name} --file ${vm_image_file}  > /dev/nul"
  echo ${clone_cmd}
  eval "${clone_cmd}"

  #Start vm
  echo "Start vm ${vm_name}..."
  start_cmd=${sshpass_cmd_prefix}" virsh start ${vm_name}"
  echo ${start_cmd}
  eval "${start_cmd}"
}

### After excute cmd: virsh undefine, need wait spend a few seconds for the backend process end
function wait_kylin_vm_undefine(){
  local vm_name=${1}
  local arm_server_ip=${2}
  local arm_server_password=${3}
  local sshpass_cmd_prefix="sshpass -p ${arm_server_password} ssh -o StrictHostKeyChecking=no root@${arm_server_ip} "
  for((i=1;i<=20;i++));do
    vm_exist_status=$(${sshpass_cmd_prefix}" virsh list --all|grep ${vm_name}||echo true")
    if [ "${vm_exist_status}" == "true" ]; then
        break
    else
      sleep 5
    fi
  done
}

### Wait a node reachable
function util::wait_ip_reachable(){
    vm_ip=$1
    loop_time=$2
    echo "Wait vm_ip=$1 reachable ... "
    for ((i=1;i<=$((loop_time));i++)); do
      pingOK=0
      ping -w 2 -c 1 ${vm_ip}|grep "0%" || pingOK=false
      echo "==> ping "${vm_ip} $pingOK
      if [[ ${pingOK} == false ]];then
        sleep 10
      else
        break
      fi
      if [ $i -eq $((loop_time)) ];then
        echo "node not reachable exit!"
        exit 1
      fi
    done
}

## Delete a kylin vm if exists
function util::delete_kylin_vm(){
  local vm_name=${1}
  local arm_server_ip=${2}
  local arm_server_password=${3}
  local sshpass_cmd_prefix="sshpass -p ${arm_server_password} ssh -o StrictHostKeyChecking=no root@${arm_server_ip} "
  echo "Delete vm beore clone..."
  vm_exist_status=$(${sshpass_cmd_prefix}" virsh list --all|grep ${vm_name}||echo true")
  if [ "${vm_exist_status}" == "true" ]; then
    echo "${vm_name} not exist, no delete"
  else
    echo "${vm_name} exist"
    destroy_vm_cmd=${sshpass_cmd_prefix}" virsh destroy ${vm_name}|| echo true"
    echo "${destroy_vm_cmd}"
    eval "${destroy_vm_cmd}"
    sleep 5
    undefine_vm_cmd=${sshpass_cmd_prefix}" virsh undefine ${vm_name} --nvram"
    echo "${undefine_vm_cmd}"
    eval "${undefine_vm_cmd}"
  fi
  sleep 2
  wait_kylin_vm_undefine ${vm_name} ${arm_server_ip} ${arm_server_password}
}

function util::set_k8s_version_by_tag(){
  tag=$1
  if [[ ${tag1} == *rc* ]]; then
    echo "RC Version"
  fi
}


function util::init_yum_repo_config_when_offline(){
  local dest_yaml_path=$1
  echo "init yum repo when offline..."
  if [[ "${OFFLINE_FLAG}" == "true" ]] &&  [[ ${OS_NAME} =~ REDHAT.* ]]; then
        if [[ ${OS_NAME} == REDHAT8 ]]; then
          echo "REDHAT8 OS."
          sed -i 's#basearch#basearch/AppStream,{offline_minio_url}/kubean/redhat-iso/\\\$releasever/os/\\\$basearch/BaseOS#2' ${dest_yaml_path}/kubeanClusterOps.yml
          sed -i "s#m,{#m','{#g" ${dest_yaml_path}/kubeanClusterOps.yml
        fi
        #sed -i "s#{offline_minio_url}#${MINIO_URL}#g" ${dest_yaml_path}/kubeanClusterOps.yml
        sed -i  "s#centos#redhat#g" ${dest_yaml_path}/kubeanClusterOps.yml
        # vars-conf-cm.yml set
        sed -i "s#{{ files_repo }}/centos#{{ files_repo }}/redhat#" ${dest_yaml_path}/vars-conf-cm.yml
        sed -i "$ a\    rhel_enable_repos: false"  ${dest_yaml_path}/vars-conf-cm.yml
  fi
  if [[ "${OFFLINE_FLAG}" == "true" ]]; then
    sed -i "s#{offline_minio_url}#${MINIO_URL}#g" ${dest_yaml_path}/kubeanClusterOps.yml
    sed -i "s#registry_host:#registry_host: ${registry_addr_amd64}#"    ${dest_yaml_path}/vars-conf-cm.yml
    sed -i "s#minio_address:#minio_address: ${MINIO_URL}#"    ${dest_yaml_path}/vars-conf-cm.yml
    sed -i "s#registry_host_key#${registry_addr_amd64}#g"    ${dest_yaml_path}/vars-conf-cm.yml
  fi
}