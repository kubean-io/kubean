#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

### Clean up the docker containers before test
function util::clean_offline_kind_cluster() {
   echo "======= container prefix: ${CONTAINERS_PREFIX}"
    kubean_containers_num=$( docker ps -a |grep ${CONTAINERS_PREFIX}||true)
    if [ "${kubean_containers_num}" ];then
      echo "Remove exist containers name contains kubean..."
      docker ps -a |grep "${CONTAINERS_PREFIX}"|awk '{print $NF}'|xargs docker rm -f
    else
      echo "No container name contains kubean to delete."
    fi
}

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

###  Install MinIO in kind
function util::install_minio(){
  local MINIO_USER=$1
  local MINIO_PASS=$2
  local kubeconfig_file=$3
  #* helm repo add minio-official https://charts.min.io
  #* helm repo update minio-official
  #* helm pull minio-official/minio --version=5.0.1

  # will be replaced by operator later
  # helm upgrade --install  --create-namespace --cleanup-on-fail \
            #--set rootUser=${MINIO_USER},rootPassword=${MINIO_PASS} \
            #--set mode="standalone" \
            #--set service.type=NodePort \
            #--set consoleService.type=NodePort \
            #--set resources.requests.memory=200Mi \
            #--set persistence.size=10Gi \
            #--kubeconfig "${kubeconfig_file}" \
            #minio minio-official/minio --wait

    helm repo add community https://release.daocloud.io/chartrepo/community
    helm repo update
    helm upgrade --install  --create-namespace --cleanup-on-fail \
              --set rootUser=${MINIO_USER},rootPassword=${MINIO_PASS} \
              --set mode="standalone" \
              --set service.type=NodePort \
              --set consoleService.type=NodePort \
              --set resources.requests.memory=200Mi \
              --set persistence.size=10Gi \
              --kubeconfig "${kubeconfig_file}" \
              minio community/minio --wait
}


### Install docker_registry in kind
### Must set the images version to 2.6.2, otherwise the cilium image will fail
### https://github.com/kubean-io/kubean/issues/246
function util::install_registry(){
  local registry_port=$1
  local kubeconfig_file=$2
  local registry_name=$3
  local registry_version=2.1.0
  local service_type="NodePort"
  echo "Start install registry..."
  # helm repo add twuni https://helm.twun.io
  # helm repo update twuni
  # helm pull twuni/docker-registry --version=${registry_version}
  # helm upgrade --install "${registry_name}" twuni/docker-registry --version ${registry_version} \
                         # --set service.type=${service_type} \
                         # --set image.tag=2.6.2 \
                         # --set service.nodePort=${registry_port} \
                         # --wait \
                         # --kubeconfig "${kubeconfig_file}"
  helm repo add community https://release.daocloud.io/chartrepo/community
  helm repo update
  helm upgrade --install "${registry_name}" community/docker-registry --version ${registry_version} \
                           --set service.type=${service_type} \
                           --set service.nodePort=${registry_port} \
                           --wait \
                           --kubeconfig "${kubeconfig_file}"
}

### Download kubean offline files
### This step only down
function util::download_offline_files_by_tag(){
  local tag=$1
  local download_folder=${2:-download_offline_files_"${tag}"}
  local base_url=https://files.m.daocloud.io/github.com/kubean-io/kubean/releases/download/${tag}
  if [ -d "${download_folder}" ]; then
    echo "Local offline_fils folder not empty, delete it."
    rm -fr "${download_folder}"
  fi
  mkdir "${download_folder}"
  # arch pkgs
  f_files_amd64_tgz=${base_url}/files-amd64-${tag}.tar.gz
  f_images_amd64_tgz=${base_url}/images-amd64-${tag}.tar.gz
  f_files_arm64__tgz=${base_url}/files-arm64-${tag}.tar.gz
  f_images_arm64_tgz=${base_url}/images-arm64-${tag}.tar.gz
  # os pkgs
  f_os_centos7=${base_url}/os-pkgs-centos7-${tag}.tar.gz
  f_os_kylin10=${base_url}/os-pkgs-kylinv10-${tag}.tar.gz
  f_os_redhat8=${base_url}/os-pkgs-redhat8-${tag}.tar.gz
  f_os_redhat7=${base_url}/os-pkgs-redhat7-${tag}.tar.gz
  # shellcheck disable=SC2206
  file_down_list=(${f_files_amd64_tgz}  ${f_images_amd64_tgz} ${f_files_arm64__tgz} ${f_images_arm64_tgz} \
                  ${f_os_centos7} ${f_os_kylin10} ${f_os_redhat8} ${f_os_redhat7})
  for (( i=0; i<${#file_down_list[@]};i++)); do
    echo "${file_down_list[$i]}"
    wget -q -c  -P  "${download_folder}"  "${file_down_list[$i]}"
  done
}

### Import binary files to kind minio
function util::import_files_minio_by_arch(){
  local minio_usr=${1:-admin}
  local minio_password=${2:-adminPassword}
  local minio_url=${3:-"http://172.18.0.2:32000"}
  local download_floder=${4}
  local tag=${TAG_VERSION}
  local arch=${5}
  echo "Import binary files to minio:${arch}..."
  local files_name=files-${arch}-${tag}.tar.gz
  echo "file name is:${files_name}"
  local untgz_folder=files-${arch}-${tag}
  echo "untgz_folder: ${untgz_folder}"
  pushd "${download_floder}"
  tar -zxvf ${files_name}
  popd
  mv ${download_floder}/files ${download_floder}/${untgz_folder}
  pushd "${download_floder}/${untgz_folder}"
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password}  ./import_files.sh ${minio_url} > /dev/null
  popd
}

### Import os packages to kind minio
### 主要用于解决 docker-ce 的安装依赖
function util::import_os_package_minio(){
  local minio_usr=${1:-admin}
  local minio_password=${2:-adminPassword}
  local minio_url=${3:-"http://172.18.0.2:32000"}
  local download_folder=${4}
  local os_name=${5}
  local tag=${TAG_VERSION}
  echo "Import os pkgs to minio: ${os_name}..."
  local untgz_folder=os-pkgs-${os_name}
  pushd "${download_folder}"
  tar -zxvf os-pkgs-${os_name}-${tag}.tar.gz
  popd
  mv "${download_folder}"/os-pkgs "${download_folder}"/"${untgz_folder}"
  pushd "${download_folder}"/"${untgz_folder}"
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password}  ./import_ospkgs.sh  ${minio_url}  os-pkgs-amd64.tar.gz > /dev/null
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password}  ./import_ospkgs.sh  ${minio_url}  os-pkgs-arm64.tar.gz > /dev/null
  popd
}

### Push images file to kind registry
function util::push_registry_by_arch(){
  registry_addr=$1
  download_folder=$2
  arch=$3
  local tag=${TAG_VERSION}
  echo "Push Registry：${arch}... "
  file_name=images-${arch}-${tag}.tar.gz

  untgz_folder=images-${arch}-${tag}
  pushd "${download_folder}"
  tar -zxvf ${file_name}
  popd
  mv ${download_folder}/images ${download_folder}/${untgz_folder}
  pushd "${download_folder}/${untgz_folder}"
  DEST_TLS_VERIFY=false ./import_images.sh ${registry_addr} > /dev/null
  popd
}

function check_iso_img() {
    ISO_IMG_FILE=$1
    if [ ! -f ${ISO_IMG_FILE} ]; then
      echo "iso image: \${ISO_IMG_FILE} should exist."
      exit 1
    fi
}

### Import iso images repo files to minio
### Work cluster node will use the image repo
function util::import_iso(){
  minio_usr=${1:-admin}
  minio_password=${2:-adminPassword}
  minio_url=${3:-"http://172.18.0.2:32000"}
  shell_path=${4}
  iso_image_file=${5}
  check_iso_img "${iso_image_file}"
  # umount before mount
  set_ios_unmounted "${iso_image_file}"
  pushd "${shell_path}"
  chmod +x import_iso.sh
  echo "Start import ${iso_image_file} to Minio, wait patiently...."
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password} ./import_iso.sh ${minio_url} ${iso_image_file} > /dev/null
  popd
}

function set_ios_unmounted(){
  echo "Umount iso if is already mounted"
  iso_image_file=$1
  mount_exist_flag=$(mount|grep "${iso_image_file}"||true)
  echo "mount_exist_flag is: ${mount_exist_flag}"
    if [  "${mount_exist_flag}" ]; then
      echo "Is already mounted before import, umount now..."
      umount ${iso_image_file}
    fi
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