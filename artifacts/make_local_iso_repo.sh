#!/bin/bash

set -eo pipefail

ISO_IMG_FILE=${ISO_IMG_FILE:-"linux.iso"}
LINUX_DISTRIBUTION=${1:-'centos'}

ISO_MOUNT_PATH=/mnt/${LINUX_DISTRIBUTION}-iso

function check_dependencies() {
  if [ ! -f ${ISO_IMG_FILE} ]; then
    echo "iso image: [${ISO_IMG_FILE}] should exist."
    exit 1
  fi
}

function mount_iso_file() {
  mkdir -p ${ISO_MOUNT_PATH}
  mount -o loop,ro ${ISO_IMG_FILE} ${ISO_MOUNT_PATH}
}

function generate_yum_repo() {
  REPOS_PATH=/etc/yum.repos.d
  REPOS_BAK_PATH=/etc/yum.repos.d.bak
  mkdir -p ${REPOS_BAK_PATH}
  mv ${REPOS_PATH}/* ${REPOS_BAK_PATH}
  cat >${REPOS_PATH}/iso.repo <<EOF
[iso]
name=Local ISO Repo
baseurl=file://${ISO_MOUNT_PATH}
enabled=1
gpgcheck=0
sslverify=0
EOF
  # yum clean all
  # yum makecache
  # yum list
}

case $LINUX_DISTRIBUTION in
centos)
  check_dependencies
  mount_iso_file
  generate_yum_repo
  ;;

debian | ubuntu)
  echo "this linux distribution is temporarily not supported."
  ;;

*)
  echo "unknown linux distribution, currently only supports centos."
  ;;
esac
