#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -x
#####################################
source "${REPO_ROOT}"/hack/util.sh
## This file include functions used in airgap resource build

RELEASE_FILE_LIST_PARTNAME=( "files-amd64" "images-amd64" "files-arm64" "images-arm64" "os-pkgs-centos7" "os-pkgs-kylinv10" "os-pkgs-redhat8" "os-pkgs-redhat7" )
KUBEAN_ARTIFACTS_USED_FILE_LIST_PARANAME=("files-amd64" "images-amd64" "os-pkgs-centos7")
BASE_URL="https://files.m.daocloud.io/github.com/kubean-io/kubean/releases/download"
#####################################
  # NEW_TAG: the version of kubean e.g. v0.5.0
  # ${STEP_TYPE} : DOWNLOAD (default)
  #              : BUILD
function resource::init_vars_kubean_resource(){
  export NEW_TAG=$1
  export STEP_TYPE=${2:-"DOWNLOAD"}
  export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
  export DOWNLOAD_ROOT_FOLDER="/root/release-files-download"
  # export TAG_FILE="${DOWNLOAD_ROOT_FOLDER}/tag.txt"
  # export RESOURCE_SVC_TAG_FILE="/root/resource_svc_tag/${NEW_TAG}.txt"
  export RESOURCE_CLUSTER_PREFIX="kubean-resource"
  export RESOURCE_KUBECONFIG_FILE="/root/.kube/${RESOURCE_CLUSTER_PREFIX}.config"

  export kindRun="docker exec -i  --privileged ${RESOURCE_CLUSTER_PREFIX}-control-plane  bash -c"
  MINIO_USER="admin"
  MINIO_PASS="adminpass123"
  MINIO_URL="http://127.0.0.1:32000"
}

#####################################
#function resource::check_file_tag(){
#  local new_tag=$1
#  local tag_file=$2
#  local old_tag="v0.0"
#  echo "Check file tag..." >&2
#  if [[ ! -f ${tag_file} ]];then
#    echo "NO"
#    return 0
#  else
#    old_tag=$(cat ${tag_file}|grep "tag"|awk -F "=" '{print $2}')
#    if [[ "${old_tag}" != "${new_tag}" ]];then
#      echo "NO"
#      return 0
 #   fi
 # fi
#  echo "YES"
#  return 0
#}

#####################################
# $check_type : offline_test （default)
#             : artifacts_test
# return 0: files  complete :)

function resource::check_release_file_integrity(){
  local new_tag=$1
  local download_root_path=$2
  local check_type=${3:-"offline_test"}
  local file_name=""
  if [[ "${check_type}" == "artifacts_test" ]];then
    check_file_list=("${KUBEAN_ARTIFACTS_USED_FILE_LIST_PARANAME[@]}")
  else
    check_file_list=("${RELEASE_FILE_LIST_PARTNAME[@]}")
  fi
  if [[ ! -d ${download_root_path}/${new_tag} ]];then
    echo "NO"
    return 0
  else
    pushd ${download_root_path}/${new_tag}
    pwd
    for item in "${check_file_list[@]}";do
      file_name=${item}-${new_tag}.tar.gz
       if [[ ! -f ${file_name} ]];then
         echo "${file_name} not exsit"
         echo "NO"
         return 0
       fi
    done
    popd
  fi
  echo "YES"
}

#####################################
# delete redundant released files to save disk
function resource::delete_redundant_folders(){
  echo "delete redundant folder"
  local new_tag=$1
  local download_root_path=$2
  if [[ -d ${download_root_path} ]];then
    pushd ${download_root_path}
    folder_need_clean=true
    folder_count=$(ls -l |grep -E 'v[0-9]\.[0-9]\.[0-9]$'|wc -l||folder_need_clean=false)
    if [[ ${folder_need_clean} == true ]];then
      # shellcheck disable=SC2004
      if (( ${folder_count} > 4));then
        to_delete_num=$(( folder_count -2 ))
        echo "to_delete_num is ${to_delete_num}"  >&2
        f_list=$(ls -l|sort |grep -v "${new_tag}"|grep 'v[0-9]\.[0-9]\.[0-9]'|tail -n ${to_delete_num}|awk '{print $NF}')
        for item in "${f_list[@]}";do
          rm -fr ${item}
        done
      fi
    fi
    popd
  fi
}

#####################################

function resource::init_download_folder(){
    local new_tag=$1
    local download_root_path=$2
    # shellcheck disable=SC2115
    rm -fr ${download_root_path}/${new_tag}
    mkdir -p ${download_root_path}/${new_tag}
}

#####################################
function resource::download_resource_files(){
  local new_tag=$1
  local download_root_path=$2
  local download_type=$3
  local part_file_name=${4:-""}
  local file_url=""

  if [[ ! -d ${download_root_path}/${new_tag} ]];then
    mkdir -p ${download_root_path}/${new_tag}
  fi

  if [[ "${download_type}" == "artifacts_test" ]];then
      download_file_list=("${KUBEAN_ARTIFACTS_USED_FILE_LIST_PARANAME[@]}")
    else
      download_file_list=("${RELEASE_FILE_LIST_PARTNAME[@]}")
  fi
  echo ${download_file_list[*]}

  # shellcheck disable=SC2115

  if [[ ${part_file_name} != "" ]];then
    file_name=${part_file_name}-${new_tag}.tar.gz
    file_url=${BASE_URL}/${new_tag}/${file_name}
    wget -q -c -T 1m -P "${download_root_path}/${new_tag}" ${file_url}
  else
    for item in "${download_file_list[@]}";do
      file_name=${item}-${new_tag}.tar.gz
      file_url=${BASE_URL}/${new_tag}/${file_name}
        echo "${file_url}"
        wget -q -c -T 1m -P "${download_root_path}/${new_tag}"  ${file_url}
    done
  fi
}

#####################################
#function resource::write_tag_file(){
#   local new_tag=$1
#   local tag_file=$2
#   echo "tag=${new_tag}" > ${tag_file}
#}

#####################################
### Clean up the docker containers before test
function resource::clean_kind_cluster() {
   local cluster_prefix=$1
   echo "======= container prefix: ${cluster_prefix}"
    kubean_containers_num=$( docker ps -a |grep ${cluster_prefix}||true)
    if [ "${kubean_containers_num}" ];then
      echo "Remove exist containers name contains kubean..."
      docker ps -a |grep "${cluster_prefix}"|awk '{print $NF}'|xargs docker rm -f||true
    else
      echo "No container name contains kubean to delete."
    fi
}

#####################################
function resource::check_resource_svc_version(){
  local tag_file=$1
  local new_tag=$2
  tag_path=${tag_file%/*}
  if [[ ! -f "${tag_file}" ]];then
    echo "NO"
    return 0
  else
    old_tag=$(cat ${tag_file}|grep "tag"|awk -F "=" '{print $2}')
    if [[ "${old_tag}" != "${new_tag}" ]];then
      echo "NO"
      return 0
    fi
  fi
  echo "YES"
}

#####################################
#Check the kind cluster status, and install when necessary
function util::check_kind_cluster_by_name(){
  local kind_name=$1
  local kind_config=$2
  local kind_ready=true
  kind get clusters|grep "${kind_name}"||kind_ready=false
  if [[ ${kind_ready} == false ]];then
    echo "NO"
  else
    echo "YES"
  fi
}

#####################################
function resource::create_kind_cluster_by_config_file(){
  local kind_name=$1
  local kind_kube_config=$2
  local kind_cluster_config_path="${REPO_ROOT}/artifacts/kindClusterConfig/kubean-host-offline.yml"
  resource::clean_kind_cluster  ${kind_name}
  KIND_NODE_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.26.0"
  docker pull "${KIND_NODE_VERSION}"
  kind_version="v0.17.0"
  util::install_kind ${kind_version}

  util::create_cluster "${kind_name}" "${kind_kube_config}" "${KIND_NODE_VERSION}" ${kind_cluster_config_path}
  echo "Waiting for the host clusters to be ready..."
  util::check_clusters_ready "${kind_kube_config}" "${kind_name}"
}

#####################################
# check if the minio is ok
function resource::check_minio(){
  local minio_port=32001
  loop_time=0
  for ((;loop_time <=1; loop_time++));do
    results=$(curl http://127.0.0.1:32001 || echo -n "false")
    if  [[ ${results} == "false" ]]; then
       sleep 1
    else
       echo "YES"
       return 0
    fi
  done
  echo "NO"
}

#####################################
# check if the registry is ok
function resource::check_registry(){
  local registry_port_amd64=31500
  local registry_port_arm64=31501
  local loop_count=2
  check_result_amd=""
  check_result_arm=""
  # check registry amd64
  for ((loop_time=0;loop_time <= $loop_count; loop_time++));do
    check_result_amd=$(curl http://127.0.0.1:${registry_port_amd64}/v2/_catalog || check_result_amd=false)
    if  [[ ${check_result_amd} =~ "repositories" ]]; then
      check_result_amd=true
      break
    else
       sleep 2
    fi
  done
  # check registry arm64
  for ((loop_time=0;loop_time <= $loop_count; loop_time++));do
    check_result_arm=$(curl http://127.0.0.1:${registry_port_arm64}/v2/_catalog || check_result_arm=false)
    if  [[ ${check_result_arm} =~ "repositories" ]]; then
      check_result_arm=true
      break
    else
       sleep 2
    fi
  done
  if [[ ${check_result_amd} == true ]] && [[ ${check_result_arm} == true ]]; then
    echo "YES"
  else
    echo "NO"
  fi
}

#####################################
# create namespace in k8s cluster
function resource::create_ns(){
  local ns_name=$1
  local kind_config=$2
  kubectl create ns "${ns_name}" --kubeconfig="${kind_config}"
}

#####################################
#create pv and pvc
function resource::create_pvc() {
  local name=$1
  local storage=$2
  local kube_config=$3
  local pv_name="pv-${name}"
  local pvc_name="pvc-${name}"
  local namespace="${name}-system"
  local kind_host_path="/home/kind/${name}"

  # create path
  ${kindRun} "mkdir -p ${kind_host_path}"
  ${kindRun} "chmod -R 777 ${kind_host_path}"

  # create pv & pvc
  cat > ./pvc.yaml << EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${pv_name}
spec:
  storageClassName: standard
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: ${storage}
  hostPath:
    path: ${kind_host_path}

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ${pvc_name}
  namespace: ${namespace}
spec:
  volumeName: ${pv_name}
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: ${storage}

EOF
kubectl apply -f ./pvc.yaml --kubeconfig=${kube_config}
}

#####################################
function resource::install_minio(){
  kubeconfig_file=$1
  local minio_version="5.0.9"
  local minio_ns="minio-system"
  local minio_helm_src="daocloud-community/minio"
  local minio_img_par="--set image.repository=quay.m.daocloud.io/minio/minio --set mcImage.repository=quay.m.daocloud.io/minio/mc --version=${minio_version}"
  local helm_cmd="helm upgrade --install  --create-namespace --cleanup-on-fail --namespace ${minio_ns}"
  resource::create_ns  ${minio_ns} ${kubeconfig_file}
  resource::create_pvc "minio" "50Gi" ${kubeconfig_file}
  helm repo add daocloud-community https://release.daocloud.io/chartrepo/community --force-update
  # will be replaced by operator later
  helm upgrade --install --create-namespace --cleanup-on-fail --namespace ${minio_ns}\
          --set users[0].accessKey=${MINIO_USER} \
          --set users[0].secretKey=${MINIO_PASS} \
          --set users[0].policy=consoleAdmin \
          --set securityContext.runAsUser=0,securityContext.runAsGroup=0 \
          --set mode=standalone \
          --set service.type=NodePort \
          --set consoleService.type=NodePort \
          --set resources.requests.memory=200Mi \
          --set persistence.existingClaim=pvc-minio \
          --kubeconfig=${kubeconfig_file} \
          minio ${minio_helm_src} ${minio_img_par} --wait > /dev/null
}

#####################################
function resource::install_registry(){
  local arch=$1
  if [[ ${arch} == "AMD64" ]];then
    registry_port=31500
    registry_name="registry_amd64"
  elif [[ ${arch} == "ARM64" ]];then
    registry_port=31501
    registry_name="registry_arm64"
  fi
  local kubeconfig_file=$2
  local registry_name=$3
  local service_type="NodePort"
  echo "Start install registry..."
  local registry_version=2.1.0
  helm repo add community https://release.daocloud.io/chartrepo/community --force-update
  helm upgrade --install "${registry_name}" community/docker-registry --version ${registry_version} \
                             --set service.type=${service_type} \
                             --set service.nodePort=${registry_port} \
                             --wait \
                             --kubeconfig "${kubeconfig_file}"
}

#####################################
function resource::check_offline_resource_svc_tag(){
  tag_file=$1
  tag_version=$2
  if [[ -f ${tag_file} ]];then
    old_tag=$(cat ${tag_file}|grep "Airgap_resource_svc_tag"|awk -F "=" '{print $2'})
    if [[ "${old_tag}" == ${tag_version} ]];then
      echo "YES"
    fi
  else
      echo "NO"
  fi
}

#####################################
### Import binary files to kind minio
function resource::import_files_minio_by_arch(){
  local download_root_path=${1}
  local new_tag=${2}
  local files_folder=${download_root_path}/${new_tag}
  local tag_list=()
  if [[  $# -gt 2 ]];then
    tag_list=("$3")
  else
    tag_list=("arm64" "amd64")
  fi

  for arch in "${tag_list[@]}";do
    echo "Import binary files to minio:${arch}..."
    local files_name=files-${arch}-${new_tag}.tar.gz
    echo "file name is:${files_name}"
    local untgz_folder=files-${arch}-${new_tag}
    echo "untgz_folder: ${untgz_folder}"
    pushd "${files_folder}"
    rm -fr ${untgz_folder}
    tar -zxvf ${files_name}
    popd
    mv ${files_folder}/files ${files_folder}/${untgz_folder}
    pushd "${files_folder}/${untgz_folder}"
    rm -fr ${untgz_folder}
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS}  ./import_files.sh ${MINIO_URL} > /dev/null
    popd
  done
}

#####################################
function resource::check_files_after_import(){
  local download_root_path=${1}
  local new_tag=${2}
  file_name=files-amd64.list
  wget -q -c -T 1m -P ${download_root_path}/${new_tag} ${BASE_URL}/${new_tag}/${file_name}
  file_first=$(cat ${download_root_path}/${new_tag}/${file_name} |grep "http"|head -1)
  file_last=$(cat ${download_root_path}/${new_tag}/${file_name} |grep "http"|tail -1)
  file_list=("${file_first}" "${file_last}")
  resource::check_files_url_exist "${file_list[@]}"
}

#####################################
function resource::import_os_package_minio(){

  local download_root_path=${1}
  local new_tag=${2}
  local files_folder="${download_root_path}/${new_tag}"
  local os_list=()
  if [[ $# -gt 2 ]];then
    os_list=("os-pkgs-centos7" )
  else
    os_list=( "os-pkgs-centos7"  "os-pkgs-kylinv10" "os-pkgs-redhat8" "os-pkgs-redhat7" )
  fi
  for os_name in "${os_list[@]}";do
    echo "Import os pkgs to minio: ${os_name}..."
    pushd "${files_folder}"
    rm -fr ${os_name}
    tar -zxvf ${os_name}-${new_tag}.tar.gz
    popd
    mv "${files_folder}"/os-pkgs "${files_folder}"/${os_name}
    pushd "${files_folder}"/${os_name}
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS}  ./import_ospkgs.sh  ${MINIO_URL}  os-pkgs-amd64.tar.gz > /dev/null
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS}  ./import_ospkgs.sh  ${MINIO_URL}  os-pkgs-arm64.tar.gz > /dev/null
    rm -fr ${os_name}
    popd
  done
}

#####################################
# when the imported os package is centos7
function resource::check_os_package_minio(){
  file_list=( "kubean/centos/7/os/aarch64/repodata/repomd.xml" \
              "kubean/centos/7/os/x86_64/repodata/repomd.xml" \
              "kubean/centos/7/os/packages.list" \
              "kubean/centos/7/os/packages.yml" \
            )
  resource::check_files_url_exist "${file_list[@]}"
}

#####################################
# file_list: an array that contains all the urls to check by wget

function resource::check_files_url_exist(){
  local file_list=("$@")
  file_exist_flag="true"
      for item in "${file_list[@]}"; do
        file_exist_flag=$(wget --spider ${MINIO_URL}/${item}||file_exist_flag="false")
        if [[ ${file_exist_flag} == "false" ]];then
          echo "ERROR: ${item} not exist in minio."
          return 1
        fi
      done
  }

#####################################
# make sure the iso image file is exist
function check_iso_img() {
    local iso_image_file=$1
    if [ ! -f ${iso_image_file} ]; then
      echo "Iso image: \${iso_image_file} should exist."
      exit 1
    fi
}

#####################################
#function set_iso_unmounted(){
#  echo "Umount iso if is already mounted"
#  iso_image_file=$1
#  mount_exist_flag=$(mount|grep "${iso_image_file}"||true)
#  echo "mount_exist_flag is: ${mount_exist_flag}"
#    if [  "${mount_exist_flag}" ]; then
 #     echo "Is already mounted before import, umount now..."
 #     umount ${iso_image_file}
 #   fi
#}

#####################################
function resource::import_iso_minio(){
  local iso_file_dir="/root/iso-images"
  local shell_path="${REPO_ROOT}/artifacts"
  local iso_list=()
  if [[ $# -gt 0 ]] && [[ $1 == "artifacts_test" ]];then
    iso_list=("Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso" "CentOS-7-x86_64-DVD-2207-02.iso")
  else
    iso_list=( "rhel-server-7.9-x86_64-dvd.iso"  "rhel-8.4-x86_64-dvd.iso" "CentOS-7-x86_64-DVD-2207-02.iso" "Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso")
  fi
    for iso in "${iso_list[@]}";do
    iso_image_file=${iso_file_dir}/${iso}
    check_iso_img "${iso_image_file}"
    # set_iso_unmounted "${iso_image_file}"
    pushd "${shell_path}"
    chmod +x import_iso.sh
    echo "Start import ${iso_image_file} to Minio, wait patiently...."
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS} ./import_iso.sh ${MINIO_URL} ${iso_image_file} > /dev/null
    popd
  done
}

#####################################
function resource::import_iso_local_path_check(){
  local iso_file_dir="/root/iso-images"
  local shell_path="${REPO_ROOT}/artifacts"
  local iso_list=()
  local  local_path=$(pwd)/"iso_mount_local_path"
  rm -fr ${local_path}
  if [[ $# -gt 0 ]] && [[ $1 == "artifacts_test" ]];then
    iso_list=("Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso" "CentOS-7-x86_64-DVD-2207-02.iso")
  else
    iso_list=( "rhel-server-7.9-x86_64-dvd.iso"  "rhel-8.4-x86_64-dvd.iso" "CentOS-7-x86_64-DVD-2207-02.iso" "Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso")
  fi
    for iso in "${iso_list[@]}";do
    iso_image_file=${iso_file_dir}/${iso}
    check_iso_img "${iso_image_file}"
    folder_name=$(echo "${iso}"|awk -F ".iso" '{print $1}')

    pushd "${shell_path}"
    chmod +x import_iso.sh

    echo "Start import ${iso_image_file} to local path, wait patiently...."
    ./import_iso.sh ${local_path}  "${iso_image_file}"   > /dev/null
    popd
  done
  resource::check_iso_file_local_path ${local_path}
}

#####################################
function resource::check_iso_file_minio(){
  file_list=( "kubean/centos-iso/7/os/x86_64/Packages/389-ds-base-1.3.10.2-16.el7_9.x86_64.rpm" \
  "kubean/centos-iso/7/os/x86_64/Packages/zziplib-0.13.62-12.el7.x86_64.rpm" \
  "kubean/kylin-iso/10/os/aarch64/Packages/abattis-cantarell-fonts-0.201-1.ky10.noarch.rpm" \
  "kubean/kylin-iso/10/os/aarch64/Packages/zziplib-help-0.13.69-6.ky10.noarch.rpm" \
  "kubean/kylin-iso/10/os/aarch64/repodata/13df713badb6a33bf7517dcee436d2a565773d5035f980b8e84520bc4f7d1c50-filelists.xml.gz" \
  "kubean/kylin-iso/10/os/aarch64/repodata/TRANS.TBL" )
  resource::check_files_url_exist "${file_list[@]}"
}

#####################################
function resource::check_iso_file_local_path(){
  local father_path=$1
  file_list=( "centos-iso/7/os/x86_64/Packages/389-ds-base-1.3.10.2-16.el7_9.x86_64.rpm" \
  "centos-iso/7/os/x86_64/Packages/zziplib-0.13.62-12.el7.x86_64.rpm" \
  "kylin-iso/10/os/aarch64/Packages/abattis-cantarell-fonts-0.201-1.ky10.noarch.rpm" \
  "kylin-iso/10/os/aarch64/Packages/zziplib-help-0.13.69-6.ky10.noarch.rpm" \
  "kylin-iso/10/os/aarch64/repodata/13df713badb6a33bf7517dcee436d2a565773d5035f980b8e84520bc4f7d1c50-filelists.xml.gz" \
  "kylin-iso/10/os/aarch64/repodata/TRANS.TBL" )
  file_exist_flag="true"
  for item in "${file_list[@]}"; do
    if [[ ! -f ${father_path}/${item} ]];then
      echo "Error: ${item} not exist in local path ${father_path}"
      return 1
    fi
  done
}

#####################################
###
function util::write_offline_resource_service_tag(){
  local tag_file=$1
  local new_tag=$2
  tag_path=${tag_file%/*}
  if [[ ! -d "${tag_path}" ]];then
    mkdir -p "${tag_path}"
  fi
  echo "Offline_resource_svc_tag=${new_tag}" > ${tag_file}
}
