#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

#####################################
source "${REPO_ROOT}"/hack/util.sh

RELEASE_FILE_LIST_PARTNAME=( "files-amd64" "images-amd64" "files-arm64" "images-arm64" "os-pkgs-rocky8" "os-pkgs-kylin-v10sp2" "os-pkgs-kylin-v10sp3" "os-pkgs-redhat8" "os-pkgs-ubuntu2204" )
KUBEAN_ARTIFACTS_USED_FILE_LIST_PARANAME=("files-amd64" "images-amd64" "os-pkgs-rocky8")
BASE_URL="https://files.m.daocloud.io/github.com/kubean-io/kubean/releases/download"

#####################################
function resource::init_vars_kubean_resource(){
  export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
  MINIO_USER="admin"
  MINIO_PASS="adminpass123"
  MINIO_URL="http://127.0.0.1:32000"
}

#####################################
# $check_type : offline_test （default)
#             : artifacts_test
# return 0: files  complete :)

function resource::check_release_file_integrity(){
  local new_tag=$1
  local download_root_path=$2
  local check_type=${3:-"offline_test"}
  local file_name=""
  if [[ "${check_type}" =~ "artifact" ]];then
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
    folder_count=$(ls -l |grep -E 'v[0-9]\.[0-9]\.[0-9]$'|wc -l||echo 0)
    # shellcheck disable=SC2004
    if (( ${folder_count} > 4));then
      to_delete_num=$(( folder_count -2 ))
      echo "to_delete_num is ${to_delete_num}"  >&2
      f_list=$(ls -l|sort |grep -v "${new_tag}"|grep 'v[0-9]\.[0-9]\.[0-9]'|tail -n ${to_delete_num}|awk '{print $NF}')
      for item in "${f_list[@]}";do
        rm -fr ${item}
      done
      fi
    popd
  fi
}

#####################################
function resource::download_resource_files(){
  local new_tag=$1
  local download_root_path=$2
  local test_type=${3:-"offline_test"}

  if [[ ! -d ${download_root_path}/${new_tag} ]];then
    mkdir -p ${download_root_path}/${new_tag}
  fi

  if [[ "${test_type}" =~ "artifact" ]];then
      download_file_list=("${KUBEAN_ARTIFACTS_USED_FILE_LIST_PARANAME[@]}")
  else
      download_file_list=("${RELEASE_FILE_LIST_PARTNAME[@]}")
  fi
  echo ${download_file_list[*]}
  # download sha256sum
  sha256sum_file_name=${new_tag}-sha256sum.txt
  echo "download sha256sum:  ${BASE_URL}/${new_tag}/sha256sum.txt"
  curl --retry 10 --retry-max-time 60 -Lo "${download_root_path}/${new_tag}/${sha256sum_file_name}"  ${BASE_URL}/${new_tag}/sha256sum.txt
  # shellcheck disable=SC2115
  for item in "${download_file_list[@]}";do
    file_name=${item}-${new_tag}.tar.gz
    file_url=${BASE_URL}/${new_tag}/${file_name}
    max_attempts=3

    echo "${file_url}"

    # retry more times to download files
    attempts=0
    while [ $attempts -lt $max_attempts ]; do
      attempts=$(($attempts+1))
      echo "${attempts}th download ${file_name}"
      curl --retry 10 --retry-max-time 60 -Lo "${download_root_path}/${new_tag}/${file_name}"  ${file_url}

      expected_checksum=$(grep "${file_name}" "${download_root_path}/${new_tag}/${sha256sum_file_name}" | awk '{print $1}')
      computed_checksum=$(sha256sum "${download_root_path}/${new_tag}/${file_name}" | awk '{print $1}')

      if [ "$computed_checksum" == "$expected_checksum" ]; then
        echo "checksum success!"
        break
      else
        ls -l "${download_root_path}/${new_tag}/${file_name}"
        rm -rf "${download_root_path}/${new_tag}/${file_name}"
        echo "checksum failed"
      fi
    done
  done
}

#####################################
function resource::create_kind_cluster_by_config_file(){
  local kind_name=$1
  local kind_kube_config=$2
  local kind_cluster_config_path="${REPO_ROOT}/hack/kindClusterConfig/kubean-host-offline.yml"
  kind::clean_kind_cluster ${kind_name}
  KIND_NODE_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.26.4"
  docker pull "${KIND_NODE_VERSION}"
  kind_version="v0.19.0"
  util::install_kind ${kind_version}
  util::create_cluster "${kind_name}" "${kind_kube_config}" "${KIND_NODE_VERSION}" ${kind_cluster_config_path}
  echo "Waiting for the host clusters to be ready..."
  util::check_clusters_ready "${kind_kube_config}" "${kind_name}"
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
  local resource_cluster_prefix=$4
  local pv_name="pv-${name}"
  local pvc_name="pvc-${name}"
  local namespace="${name}-system"
  local kind_host_path="/home/kind/${name}"
  local kindRun="docker exec -i  --privileged ${resource_cluster_prefix}-control-plane  bash -c"
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
  local kubeconfig_file=$1
  local kind_cluster_prefix=$2
  local minio_version="5.0.9"
  local minio_ns="minio-system"
  local minio_helm_src="daocloud-community/minio"
  local minio_img_par="--set image.repository=quay.m.daocloud.io/minio/minio --set mcImage.repository=quay.m.daocloud.io/minio/mc --version=${minio_version}"
  local helm_cmd="helm upgrade --install  --create-namespace --cleanup-on-fail --namespace ${minio_ns}"
  resource::create_ns  ${minio_ns} ${kubeconfig_file}
  resource::create_pvc "minio" "50Gi" ${kubeconfig_file} ${kind_cluster_prefix}
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
          minio ${minio_helm_src} ${minio_img_par} --wait --timeout 600s
}

#####################################
function resource::install_registry(){
  local kubeconfig_file=$1
  local arch=$2
  local service_type="NodePort"
  local registry_version=2.1.0

  if [[ ${arch} == "amd64" ]];then
    registry_port=31500
    registry_name="registry-amd64"
  elif [[ ${arch} == "arm64" ]];then
    registry_port=31501
    registry_name="registry-arm64"
  fi

  echo "Start install registry..."

  helm repo add community https://release.daocloud.io/chartrepo/community --force-update
  helm upgrade --install "${registry_name}" community/docker-registry --version ${registry_version} \
                             --set service.type=${service_type} \
                             --set service.nodePort=${registry_port} \
                             --wait \
                             --kubeconfig "${kubeconfig_file}"
}

function version_le() {
  # <=
  [ "$1" == "$(echo -e "$1\n$2" | sort -V | head -n1)" ]
}
function version_lt() {
  # <
  [ "$1" == "$2" ] && return 1 || version_le $1 $2
}

function resource::install_podman() {
  local PODMAN_VERSION="v4.7.1"
  local PODMAN_TAR_NAME="podman-linux-amd64.tar.gz"
  local PODMAN_TAR_URL="https://files.m.daocloud.io/github.com/mgoltzsche/podman-static/releases/download/${PODMAN_VERSION}/${PODMAN_TAR_NAME}"

  # If you use the kylin linux, uninstall podman by default
  if [ $(cat /etc/os-release | grep kylin | wc -l) -eq 1 ]; then
    dnf remove podman -y
  fi
  # If you use the RHEL 8.x and runc was preinstalled, uninstall runc by default
  if [[ $(cat /etc/os-release | sed -n -e 's/^ID=//p' -e 's/^VERSION_ID=//p' | sed ':a;N;$!ba;s/[\n"]//g') == rhel8* ]] && rpm -qa | grep -q runc; then
    echo "detected preinstalled runc, uninstall it..."
    yum remove runc -y
  fi
  local curr_podman_ver=""
  if  [ -x "$(command -v podman)" ]; then
    curr_podman_ver=$(podman --version |awk '{print $3}')
  fi
  if ! [ -x "$(command -v podman)" ] || $(version_lt ${curr_podman_ver} "${PODMAN_VERSION:1}"); then
    echo "install podman or upgrade podman($curr_podman_ver)"
    curl --retry 10 --retry-max-time 60 -LO ${PODMAN_TAR_URL}
    tar -zxvf ${PODMAN_TAR_NAME}
    sudo cp -r "podman-linux-${arch}/usr" "podman-linux-${arch}/etc" / -f
    rm -rf /usr/bin/podman && ln /usr/local/bin/podman /usr/bin/podman
    podman --version
  else
    echo "skip install podman ..."
  fi
}

#####################################
# $test_type: artifacts_test (replace the scripts from artifacts to tgz decompressed path)
#             offline_test   (use scripts in the tgz)

function resource::push_registry_by_arch(){
  local download_root_path=${1}
  local new_tag=${2}
  local arch=${3:-"amd64"}
  local test_type=${4:-"offline-test"}
  local download_root_path_tag=${download_root_path}/${new_tag}

  if [[ ${arch} == "amd64" ]];then
      registry_addr="127.0.0.1:31500"
    elif [[ ${arch} == "arm64" ]];then
      registry_addr="127.0.0.1:31501"
    else
      echo "Image arch type error."
      exit 1
    fi
  echo "Push Registry：${arch}... "
  rm -fr  "${download_root_path_tag}/img-${arch}" && mkdir -p "${download_root_path_tag}/img-${arch}"
  tar -zxvf "${download_root_path_tag}/images-${arch}-${new_tag}.tar.gz" -C "${download_root_path_tag}/img-${arch}"
  chmod +x ${download_root_path_tag}/img-${arch}/images/import_images.sh

  pushd ${download_root_path_tag}/img-${arch}/images
  REGISTRY_ADDR=${registry_addr} ./import_images.sh > /dev/null
  popd
}
#####################################
# default check amd64 images
function resource::check_img_registry(){
  local download_root_path=${1}
  local new_tag=${2}
  local image_list_name="images-amd64.list"
  wget -q -c -T 1m -P ${download_root_path}/${new_tag} ${BASE_URL}/${new_tag}/${image_list_name}
  # shellcheck disable=SC2036
  image_name_first=`cat ${download_root_path}/${new_tag}/${image_list_name}|head -1|awk -F ":" '{print $1}'`
  tag_first=`cat ${download_root_path}/${new_tag}/${image_list_name}|head -1|awk -F ":" '{print $2}'`
  image_name_last=`cat ${download_root_path}/${new_tag}/${image_list_name} |tail -1|awk -F ":" '{print $1}'`
  tag_last=`cat ${download_root_path}/${new_tag}/${image_list_name}|tail -1|awk -F ":" '{print $2}'`

  cmd_list=( "${image_name_first}/tags/list" \
             "${image_name_last}/tags/list" \
           )
  for cmd in "${cmd_list[@]}";do
    result=$(curl http://127.0.0.1:31500/v2/${cmd} )
    if [[ ${result} =~ "errors" ]];then
      return 1
    else
      echo "Registry check ${cmd} ok！"
    fi
  done

}

#####################################
function resource::import_files_minio_by_arch(){
  local download_root_path=${1}
  local new_tag=${2}
  local arch=${3:-"amd64"}
  local test_type=${4:-"offline-test"}
  local download_root_path_tag="${download_root_path}/${new_tag}"
  local decompress_folder="files-${arch}"
  echo "Import binary files to minio:${arch}..."
  # shellcheck disable=SC2115
  rm -fr ${download_root_path_tag}/${decompress_folder} && mkdir -p ${download_root_path_tag}/${decompress_folder}
  tar -zxvf "${download_root_path_tag}/files-${arch}-${new_tag}.tar.gz" -C ${download_root_path_tag}/${decompress_folder}
  chmod +x "${download_root_path_tag}/${decompress_folder}"/files/import_files.sh

  pushd "${download_root_path_tag}/${decompress_folder}/files"
  MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS}  ./import_files.sh ${MINIO_URL} > /dev/null
  popd

}

#####################################
# file_list: an array that contains all the urls to check by wget
function resource::check_files_url_exist(){
  local file_list=("$@")
  file_exist_flag="true"
      for item in "${file_list[@]}"; do
        wget --spider ${item}||file_exist_flag="false"
        if [[ ${file_exist_flag} == "false" ]];then
          echo "ERROR: ${item} not exist in minio."
          return 1
        fi
        echo "check ${item} ok!"
      done
}

#####################################
function resource::check_files_after_import(){
  local download_root_path=${1}
  local new_tag=${2}
  file_name=files-amd64.list
  wget -q -c -T 1m -P ${download_root_path}/${new_tag} ${BASE_URL}/${new_tag}/${file_name}
  file_first=$(cat ${download_root_path}/${new_tag}/${file_name} |grep "http"|awk -F "https://" '{print $2}'|head -1)
  file_last=$(cat ${download_root_path}/${new_tag}/${file_name} |grep "http"|awk -F "https://" '{print $2}'|tail -1)
  file_list=( "${MINIO_URL}/kubean/${file_first}"  "${MINIO_URL}/kubean/${file_last}")
  resource::check_files_url_exist "${file_list[@]}"
}

#####################################
function resource::import_os_package_minio(){
  local download_root_path=${1}
  local new_tag=${2}
  local test_type=${3:-"offline-test"}
  local download_root_path_tag="${download_root_path}/${new_tag}"

  local os_list=()
  if [[ ${test_type} =~ "artifact" ]];then
    os_list=("os-pkgs-rocky8" )
  else
    os_list=("os-pkgs-rocky8" "os-pkgs-kylin-v10sp2" "os-pkgs-redhat8" "os-pkgs-ubuntu2204" )
  fi
  for os_name in "${os_list[@]}";do
    echo "Import os pkgs to minio: ${os_name}..."
    decompress_folder="${os_name}"
    # shellcheck disable=SC2115
    rm -fr ${download_root_path_tag}/${decompress_folder} && mkdir -p ${download_root_path_tag}/${decompress_folder}
    tar -zxvf ${download_root_path_tag}/${os_name}-${new_tag}.tar.gz -C ${download_root_path_tag}/${decompress_folder}

    # when offline-test use the import script in the tgz file
    # when artifact-test cp the import script from artifact path
    if [[ ${test_type} =~ "artifact" ]];then
      rm -f "${download_root_path_tag}/${decompress_folder}"/os-pkgs/import_ospkgs.sh
      cp -f ${REPO_ROOT}/artifacts/import_ospkgs.sh ${download_root_path_tag}/${decompress_folder}/os-pkgs/
      chmod +x ${download_root_path_tag}/${decompress_folder}/os-pkgs/import_ospkgs.sh
    fi

    pushd ${download_root_path_tag}/${decompress_folder}/os-pkgs
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS}  ./import_ospkgs.sh  ${MINIO_URL}  os-pkgs-amd64.tar.gz > /dev/null
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS}  ./import_ospkgs.sh  ${MINIO_URL}  os-pkgs-arm64.tar.gz > /dev/null
    popd
  done
}

#####################################
# when the imported os package is rocky8
function resource::check_os_package_minio(){
  file_list=( "${MINIO_URL}/kubean/rocky/8/os/aarch64/repodata/repomd.xml" \
              "${MINIO_URL}/kubean/rocky/8/os/x86_64/repodata/repomd.xml" \
              "${MINIO_URL}/kubean/rocky/8/os/packages.list" \
              "${MINIO_URL}/kubean/rocky/8/os/packages.yml" \
            )
  resource::check_files_url_exist "${file_list[@]}"
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
function resource::import_iso_minio(){
  local iso_file_dir="/root/iso-images"
  local shell_path="${REPO_ROOT}/artifacts"
  local iso_list=()
  if [[ $# -gt 0 ]] && [[ $1 =~ "artifact" ]];then
    iso_list=("Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso" "Rocky-8.10-x86_64-dvd1.iso")
  else
    iso_list=( "ubuntu-22.04.4-live-server-amd64.iso" "Rocky-8.10-x86_64-dvd1.iso" "rhel-8.4-x86_64-dvd.iso" "CentOS-7-x86_64-DVD-2207-02.iso" "Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso")
  fi
  for iso in "${iso_list[@]}";do
    iso_image_file=${iso_file_dir}/${iso}
    check_iso_img "${iso_image_file}"
    pushd "${shell_path}"
    chmod +x import_iso.sh
    echo "Start import ${iso_image_file} to Minio, wait patiently...."
    MINIO_USER=${MINIO_USER} MINIO_PASS=${MINIO_PASS} ./import_iso.sh ${MINIO_URL} ${iso_image_file} > /dev/null
    popd
  done
}

#####################################
function resource::check_iso_minio(){
    file_list=( "${MINIO_URL}/kubean/centos-iso/7/os/x86_64/Packages/389-ds-base-1.3.10.2-16.el7_9.x86_64.rpm" \
                "${MINIO_URL}/kubean/centos-iso/7/os/x86_64/Packages/zziplib-0.13.62-12.el7.x86_64.rpm" \
                "${MINIO_URL}/kubean/centos/7/os/packages.list" \
                "${MINIO_URL}/kubean/centos/7/os/packages.yml" \
              )
    resource::check_files_url_exist "${file_list[@]}"
}
#####################################
function resource::import_iso_local_path_check(){
  test_type=${1:-"offline_test"}
  local iso_file_dir="/root/iso-images"
  local shell_path="${REPO_ROOT}/artifacts"
  local iso_list=()
  local  local_path=$(pwd)/"iso_mount_local_path"
  rm -fr ${local_path}
  if [[ "${test_type}" =~ "artifact" ]];then
    iso_list=("Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso" "Rocky-8.10-x86_64-dvd1.iso")
  else
    iso_list=( "rhel-server-7.9-x86_64-dvd.iso" "Rocky-8.10-x86_64-dvd1.iso" "rhel-8.4-x86_64-dvd.iso" "Kylin-Server-10-SP2-aarch64-Release-Build09-20210524.iso")
  fi
    for iso in "${iso_list[@]}";do
    iso_image_file=${iso_file_dir}/${iso}
    check_iso_img "${iso_image_file}"
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
  file_list=( "${MINIO_URL}/kubean/rocky-iso/8/os/x86_64/AppStream/Packages/3/389-ds-base-1.4.3.39-3.module+el8.10.0+1754+15117a6d.x86_64.rpm" \
     "${MINIO_URL}/kubean/rocky-iso/8/os/x86_64/AppStream/Packages/z/zziplib-utils-0.13.68-13.el8_10.x86_64.rpm" \
     "${MINIO_URL}/kubean/kylin-iso/10/sp2/os/aarch64/Packages/abattis-cantarell-fonts-0.201-1.ky10.noarch.rpm" \
     "${MINIO_URL}/kubean/kylin-iso/10/sp2/os/aarch64/Packages/zziplib-help-0.13.69-6.ky10.noarch.rpm" \
     "${MINIO_URL}/kubean/kylin-iso/10/sp2/os/aarch64/repodata/13df713badb6a33bf7517dcee436d2a565773d5035f980b8e84520bc4f7d1c50-filelists.xml.gz" \
     "${MINIO_URL}/kubean/kylin-iso/10/sp2/os/aarch64/repodata/TRANS.TBL"
    )
  resource::check_files_url_exist "${file_list[@]}"
}

#####################################
function resource::check_iso_file_local_path(){
  local father_path=$1
  file_list=( "rocky-iso/8/os/x86_64/AppStream/Packages/3/389-ds-base-1.4.3.39-3.module+el8.10.0+1754+15117a6d.x86_64.rpm" \
  "rocky-iso/8/os/x86_64/AppStream/Packages/z/zziplib-utils-0.13.68-13.el8_10.x86_64.rpm" \

  "kylin-iso/10/sp2/os/aarch64/Packages/abattis-cantarell-fonts-0.201-1.ky10.noarch.rpm" \
  "kylin-iso/10/sp2/os/aarch64/Packages/zziplib-help-0.13.69-6.ky10.noarch.rpm" \
  "kylin-iso/10/sp2/os/aarch64/repodata/13df713badb6a33bf7517dcee436d2a565773d5035f980b8e84520bc4f7d1c50-filelists.xml.gz" \
  "kylin-iso/10/sp2/os/aarch64/repodata/TRANS.TBL" )
  for item in "${file_list[@]}"; do
    if [[ ! -f ${father_path}/${item} ]];then
      echo "Error: ${item} not exist in local path ${father_path}"
      return 1
    fi
    echo "check ${item} ok"
  done
}

function resource::import_cilium_chart(){
  authUser="rootuser"
  authPassword="rootpass123"

  kubeconfig_file=${1}

  curl -LO http://10.6.170.11:8080/tools/yq
  mv yq /usr/local/bin/yq &&  chmod +x /usr/local/bin/yq

  curl -LO http://10.6.170.11:8080/tools/helm-v3.17.2
  cp -f helm-v3.17.2 /usr/local/bin/helm &&  chmod +x /usr/local/bin/helm

  helm repo add chartmuseum https://chartmuseum.github.io/charts --force-update
  helm install chartmuseum chartmuseum/chartmuseum \
                        --namespace kube-system \
                        --set image.repository=ghcr.m.daocloud.io/helm/chartmuseum \
                        --set image.tag=v0.13.1 \
                        --set secret.name=chartmuseum-creds \
                        --set secret.key=${authUser} \
                        --set secret.password=${authPassword} \
                        --set service.type=NodePort \
                        --set service.nodePort=30081 \
                        --set resources.requests.memory=200Mi \
                        --set env.open.DISABLE_API=false \
                        --set env.open.AUTH_ANONYMOUS_GET=true\
                        --kubeconfig=${kubeconfig_file}

  sleep 60
  KUBEAN_VERSION=${2}
  registry_addr="172.30.41.63:8081"


  LocalArtifactSetName="localartifactset.cr.yaml"
  LocalArtifactSetURL="https://github.com/kubean-io/kubean/releases/download/${KUBEAN_VERSION}/${LocalArtifactSetName}"
  curl -LO ${LocalArtifactSetURL}

  CILIUM_VERSION=$(yq '.spec.items[] | select(.name == "cilium") | .versionRange[0]' ${LocalArtifactSetName})

  helm repo add cilium https://helm.cilium.io/

  helm pull cilium/cilium --version=${CILIUM_VERSION}

  curl -u ${authUser}:${authPassword} --data-binary @cilium-${CILIUM_VERSION}.tgz ${registry_addr}/api/charts -ksS

  rm -f cilium-${CILIUM_VERSION}.tgz
  rm -f ${LocalArtifactSetName}

}