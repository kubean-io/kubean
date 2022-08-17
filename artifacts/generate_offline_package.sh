#!/bin/bash

set -eo pipefail

OPTION=${1:-'all'}
KUBEAN_TAG=${KUBEAN_TAG:-"v0.1.0"}

CURRENT_DIR=$(pwd)
OFFLINE_PACKAGE_DIR=${CURRENT_DIR}/${KUBEAN_TAG}

function generate_offline_dir() {
  mkdir -p $OFFLINE_PACKAGE_DIR
}

function generate_temp_list() {
  if [ ! -d "kubespray" ]; then
    echo "kubespray git repo should exist."
    exit 1
  fi
  echo "$CURRENT_DIR/kubespray"
  cd $CURRENT_DIR/kubespray
  bash contrib/offline/generate_list.sh
  cp contrib/offline/temp/*.list $OFFLINE_PACKAGE_DIR
}

function create_files() {
  cd $CURRENT_DIR/kubespray/contrib/offline/
  NO_HTTP_SERVER=true bash manage-offline-files.sh
  cp offline-files.tar.gz $OFFLINE_PACKAGE_DIR
}

function create_images() {
  cd $CURRENT_DIR/artifacts
  bash manage_images.sh create $CURRENT_DIR/kubespray/contrib/offline/temp/images.list
  cp offline-images.tar.gz $OFFLINE_PACKAGE_DIR
}

case $OPTION in
all)
  generate_offline_dir
  generate_temp_list
  create_files
  create_images
  ;;

list)
  generate_temp_list
  ;;

files)
  create_files
  ;;

images)
  create_images
  ;;

*)
  echo -n "unknown operator"
  ;;
esac
