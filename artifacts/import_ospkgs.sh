#!/bin/bash

set -eo pipefail

MINIO_API_ADDR=${1:-'http://127.0.0.1:9000'}

TAR_GZ_FILE_PATH=${2} ## os-pkgs/kubean-v0.0.1-centos7-amd64.tar.gz

readonly PARALLEL_LOCK="/var/lock/kubean-import.lock"

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

function ensure_kubean_bucket() {
  if ! mc ls kubeaniominioserver/kubean >/dev/null 2>&1; then
    echo "create bucket 'kubean'"
    mc mb -p kubeaniominioserver/kubean
    mc anonymous set download kubeaniominioserver/kubean
  fi
}

function remove_mc_host_conf() {
  echo "remove mc config"
  flock $PARALLEL_LOCK mc config host remove kubeaniominioserver || true
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

  for dirName in resources/*; do
    mc cp --no-color --recursive "$dirName" "kubeaniominioserver/kubean/"
  done
}

start=$(date +%s)

check_mc_cmd
add_mc_host_conf
ensure_kubean_bucket
export -f import_os_packages
flock -s $PARALLEL_LOCK bash -c "import_os_packages"
remove_mc_host_conf

end=$(date +%s)
take=$((end - start))
echo "Importing OS pkgs spends ${take} seconds"
