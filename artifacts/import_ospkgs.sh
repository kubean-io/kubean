#!/bin/bash

set -eo pipefail

MINIO_API_ADDR=${1:-'http://127.0.0.1:9000'}

TAR_GZ_FILE_PATH=${2} ## os-pkgs/kubean-v0.0.1-centos7-amd64.tar.gz

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

function import_os_packages() {
  if [ -z "$TAR_GZ_FILE_PATH" ]; then
    echo "TAR_GZ_FILE_PATH is empty, for example: os-pkgs/kubean-v0.0.1-centos7-amd64.tar.gz"
    exit 1
  fi
  echo "$TAR_GZ_FILE_PATH"

  tar -xvf "$TAR_GZ_FILE_PATH" ## got resources folder

  for bucketName in resources/*; do
    bucketName=${bucketName//resources\//} ## remove dir prefix
    if ! mc ls kubeaniominioserver/"$bucketName" >/dev/null 2>&1; then
      echo "create bucket $bucketName"
      mc mb kubeaniominioserver/"$bucketName"
      mc policy set download kubeaniominioserver/"$bucketName"
    fi
  done

  for path in $(find resources); do
    if [ -f "$path" ]; then
      ## mc cp resources/centos/7/x86_64/x.rpm kubeaniominioserver/centos/7/x86_64/x.rpm
      minioFileName=${path//resources/kubeaniominioserver}
      mc cp --no-color "$path" "$minioFileName"
    fi
  done
}

check_mc_cmd
add_mc_host_conf
import_os_packages
remove_mc_host_conf
