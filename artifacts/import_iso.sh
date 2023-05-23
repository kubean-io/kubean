#!/bin/bash

set -e

MINIO_USER=${MINIO_USER:-""}
MINIO_PASS=${MINIO_PASS:-""}

function iso::add_mc_host_conf() {
  local minio_addr=$1
  local minio_user=$2
  local minio_pass=$3
  if [ -z "${minio_user}" ]; then
    echo "need MINIO_USER and MINIO_PASS"
    exit 1
  fi
  if ! mc config host add kubeaniominioserver "${minio_addr}" "${minio_user}" "${minio_pass}"; then
    echo "mc add ${minio_addr} server failed"
    exit 1
  fi
}

function iso::del_mc_host_conf() {
  echo "remove mc config"
  flock ${ISO_PARALLEL_LOCK} mc config host remove kubeaniominioserver || true
}

function iso::check_mc_cmd() {
  if which mc; then
    echo "mc check successfully"
  else
    echo "please install mc first"
    exit 1
  fi
}

function iso::ensure_kubean_bucket() {
  if ! mc ls kubeaniominioserver/kubean >/dev/null 2>&1; then
    echo "create bucket 'kubean'"
    mc mb -p kubeaniominioserver/kubean
    mc anonymous set download kubeaniominioserver/kubean
  fi
}

function iso::mk_server_path() {
  local iso_mnt_path=$1 
  for path in $(find $iso_mnt_path); do
    if [ -L "$path" ]; then
      if echo "$path" | grep 'ubuntu' &>/dev/null; then
        echo "/ubuntu-iso"
        return
      fi
    fi
    if [ -f "$path" ]; then
      if echo "$path" | grep 'ky10.x86_64.rpm' >/dev/null 2>&1; then
        echo "/kylin-iso/10/os/x86_64"
        return
      fi
      if echo "$path" | grep 'ky10.aarch64.rpm' >/dev/null 2>&1; then
        echo "/kylin-iso/10/os/aarch64"
        return
      fi
      if [ "$(basename $path)" = ".treeinfo" ]; then
        local arch=$(sed -n '/^\[general\]/,$p' $path | sed -n 's/arch = //p' | head -1)
        local os=$(sed -n '/^\[general\]/,$p' $path | sed -n 's/name = //p' | head -1)
        local version=$(sed -n '/^\[general\]/,$p' $path | sed -n 's/version = //p' | head -1 | cut -d. -f1)

        if [[ "$os" =~ "openEuler" ]]; then
          local euler_version=$(sed -n '/^\[general\]/,$p' $path | sed -n 's/version = //p' | head -1  | sed -e 's/LTS//g' | sed -e 's/SP[0-9]//g' | sed -e 's/sp[0-9]//g'  |  tr -d '-') ## 22.03LTS => 22.03 , 22.03LTS SP1 => 22.03
          echo "/openeuler-iso/$euler_version/os/$arch"
          return
        fi

        if [[ "$os" =~ "CentOS" ]]; then
          echo "/centos-iso/$version/os/$arch"
          return
        fi
        if [[ "$os" =~ "Red Hat" ]]; then
          # set var releasever to '7Server' when OS is RHEL 7.x 
          if [[ "$version" == "7" ]]; then
            echo "/redhat-iso/7Server/os/$arch"
            return
          fi
          echo "/redhat-iso/$version/os/$arch"
          return
        fi
        if [[ "$os" =~ "UnionTechOS" ]]; then
          echo "/uos-iso/$version/os/$arch"
          return
        fi
        if [[ "$os" =~ "Oracle Linux" ]]; then
          echo "/oracle-iso/$version/os/$arch"
          return
        fi
      fi
    fi
  done
  echo ""
  return
}

function iso::mount_file() {
  local iso_file_path=$1
  local iso_mnt_path=$2
  if [ -z "${iso_file_path}" ]; then
    echo "empty ISO IMAGE PATH"
    exit 1
  fi

  mkdir -p ${iso_mnt_path}

  if ls ${iso_mnt_path} | grep -i -E "EFI|images|isolinux|LiveOS|Packages|repodata|boot|dists|live|pool" >/dev/null 2>&1 ; then
    ## try to umount.
    echo "try to umount ${iso_mnt_path} first"
    iso::unmount_file "${iso_mnt_path}" || true
  fi

  echo "mount ISO file"
  if ! mount -o loop,ro "${iso_file_path}" ${iso_mnt_path}; then
    echo "mount ${iso_file_path} failed"
    exit 1
  fi
}

function iso::unmount_file() {
  local iso_mnt_path=$1
  echo "unmount ISO file"
  umount ${iso_mnt_path} || true
}

function iso::import_data() {
  local iso_file_path=$1
  local iso_mnt_path=$2
  local is_cp_path=$3
  local target_path=$4
  
  if [ "${is_cp_path}" == "false" ]; then
    echo "start push ISO data into minio"
  else
    echo "start copy ISO data into ${target_path}"
  fi
  local minio_server_path=$(iso::mk_server_path "${iso_mnt_path}")

  if [ -z "${minio_server_path}" ]; then
    echo "can not find os version and arch info from ${iso_file_path}"
    exit 1
  fi
  local minio_files_path="kubeaniominioserver/kubean${minio_server_path}"
  local path_list=()

  if [ -d "${iso_mnt_path}/Packages" ]; then
    path_list+=("${iso_mnt_path}/Packages")
  fi

  if [ -d "${iso_mnt_path}/repodata" ]; then
    path_list+=("${iso_mnt_path}/repodata")
  fi

  if [ -d "${iso_mnt_path}/BaseOS" ]; then
    path_list+=("${iso_mnt_path}/BaseOS")
  fi
  if [ -d "${iso_mnt_path}/AppStream" ]; then
    path_list+=("${iso_mnt_path}/AppStream")
  fi

  if [ -d "${iso_mnt_path}/dists" ]; then
    path_list+=("${iso_mnt_path}/dists")
  fi
  if [ -d "${iso_mnt_path}/pool" ]; then
    path_list+=("${iso_mnt_path}/pool")
  fi
  
  if [ "${#path_list[@]}" -gt 0 ]; then
    for path_name in "${path_list[@]}"; do
      if [ "${is_cp_path}" == "true" ]; then
        mkdir -p "${target_path}/${minio_server_path}"
        cp -vr "${path_name}" "${target_path}/${minio_server_path}"
      else
        ## "/mnt/kubean-temp-iso/Pkgs" => "kubeaniominioserver/kubean/centos-dvd/7/os/x86_64/"
        mc cp --no-color --recursive "${path_name}" "${minio_files_path}"
      fi
    done
  else
    echo "cannot find valid repo data from ${iso_file_path}"
    exit 1
  fi
}

readonly ISO_PARALLEL_LOCK="/var/lock/kubean-import.lock"

function import_iso::main() {
  local minio_api_addr=${1:-'http://127.0.0.1:9000'}
  local iso_file_path=${2}

  local is_cp_path=false
  local target_path="${minio_api_addr}"
  local iso_mnt_path="/mnt/kubean-temp-iso"

  if [[ "${target_path}" != "https://"* ]] && [[ "${target_path}" != "http://"* ]] ; then
    is_cp_path=true
    mkdir -p "${target_path}"
  fi

  trap "iso::unmount_file ${iso_mnt_path}" EXIT

  start=$(date +%s)

  if [ "${is_cp_path}" == "false" ]; then
    iso::check_mc_cmd
    iso::add_mc_host_conf "${minio_api_addr}" "${MINIO_USER}" "${MINIO_PASS}"
    iso::ensure_kubean_bucket
  fi
  iso::mount_file "${iso_file_path}" "${iso_mnt_path}"
  export -f iso::import_data iso::mk_server_path
  flock -s ${ISO_PARALLEL_LOCK} bash -c "iso::import_data '${iso_file_path}' '${iso_mnt_path}' '${is_cp_path}' '${target_path}'"
  if [ "${is_cp_path}" == "false" ]; then
    iso::del_mc_host_conf
  fi

  end=$(date +%s)
  take=$((end - start))
  echo "Importing ISO spends ${take} seconds"

}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    import_iso::main "$@"
fi
