#!/bin/bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail

MINIO_API_ADDR=${1:-'http://127.0.0.1:9000'}

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

function del_mc_host_conf() {
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

function import_files() {
  if [ ! -d "offline-files" ]; then
    tar -xvf offline-files.tar.gz
    echo "unzip successfully"
  fi

  if [ -d "offline-files/files.m.daocloud.io" ]; then
    mv offline-files/files.m.daocloud.io/*  offline-files/
  fi

  for dirName in offline-files/*; do
     mc cp --no-color --recursive "$dirName" "kubeaniominioserver/kubean/"
  done

}

start=$(date +%s)

check_mc_cmd
add_mc_host_conf
ensure_kubean_bucket
export -f import_files
flock -s $PARALLEL_LOCK bash -c "import_files"
del_mc_host_conf

end=$(date +%s)
take=$((end - start))
echo "Importing files spends ${take} seconds"
