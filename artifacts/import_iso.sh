#!/bin/bash

set -eo pipefail

MINIO_API_ADDR=${1:-'http://127.0.0.1:9000'}

ISO_IMG_FILE=${2} ##  CentOS-7-XXX.ISO CentOS-8XXX.ISO

ISO_MOUNT_PATH="/mnt/kubean-temp-iso"

Minio_Server_PATH=""

function add_mc_host_conf() {
  if [ -z "$MINIO_USER" ]; then
    echo "need MINIO_USER and MINIO_PASS"
    exit 1
  fi
  if ! mc config host add kubeaniominioserver "$MINIO_API_ADDR" "$MINIO_USER" "$MINIO_PASS"; then
    echo "mc add $MINIO_API_ADDR server failed"
    exit 1
  fi
}

function remove_mc_host_conf() {
  echo "remove mc config"
  mc config host remove kubeaniominioserver
}

function check_mc_cmd() {
  if which mc; then
    echo "mc check successfully"
  else
    echo "please install mc first"
    exit 1
  fi
}

function ensure_kubean_bucket() {
  if ! mc ls kubeaniominioserver/kubean >/dev/null 2>&1; then
    echo "create bucket 'kubean'"
    mc mb kubeaniominioserver/kubean
    mc policy set download kubeaniominioserver/kubean
  fi
}

function unmount_iso_file() {
  echo "unmount ISO file"
  umount ${ISO_MOUNT_PATH}
}

function iso_os_version_arch() {
  for path in $(find $ISO_MOUNT_PATH); do
    if [ -f "$path" ]; then
      if echo "$path" | grep 'el7.x86_64.rpm' >/dev/null 2>&1; then
        echo "/centos-iso/7/os/x86_64"
        return
      fi
      if echo "$path" | grep 'el8.x86_64.rpm' >/dev/null 2>&1; then
        echo "/centos-iso/8/os/x86_64"
        return
      fi
    fi
  done
  echo "not support $ISO_IMG_FILE , can not find os and arch info"
  exit 1
}

function mount_iso_file() {
  if [ -z "$ISO_IMG_FILE" ]; then
    echo "empty ISO_IMG_FILE"
    exit 1
  fi

  mkdir -p ${ISO_MOUNT_PATH}
  echo "mount ISO file"
  if ! mount -o loop,ro "${ISO_IMG_FILE}" ${ISO_MOUNT_PATH}; then
    echo "mount ${ISO_IMG_FILE} failed"
    exit 1
  fi
}

function import_iso_data() {
  Minio_Server_PATH=$(iso_os_version_arch)

  if [ -z "$Minio_Server_PATH" ]; then
    echo "can not find os version and arch info from $ISO_IMG_FILE"
    exit 1
  fi
  dirArray=()

  if [ -d "$ISO_MOUNT_PATH/Packages" ]; then
    dirArray+=("$ISO_MOUNT_PATH/Packages")
  fi

  if [ -d "$ISO_MOUNT_PATH/repodata" ]; then
    dirArray+=("$ISO_MOUNT_PATH/repodata")
  fi

  minioFileName="kubeaniominioserver/kubean$Minio_Server_PATH/"

  for dirName in "${dirArray[@]}"; do
    ## "/mnt/kubean-temp-iso/Pkgs" => "kubeaniominioserver/kubean/centos-dvd/7/os/x86_64/"
    mc cp --no-color --recursive "$dirName" "$minioFileName"
  done
}

start=$(date +%s)

check_mc_cmd
add_mc_host_conf
ensure_kubean_bucket
mount_iso_file
import_iso_data
unmount_iso_file
remove_mc_host_conf

end=$(date +%s)
take=$((end - start))
echo "It spends ${take} seconds"
