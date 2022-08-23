#!/bin/bash

set -eo pipefail

ISO_IMG_FILE=''
LINUX_DISTRIBUTION=''
REPO_BASE_URL=''
ISO_MOUNT_PATH=''

MARK_NAME=Kubean
ISO_REPO_CONF=${MARK_NAME}-ISO.repo
URL_REPO_CONF=${MARK_NAME}-URL.repo

YUM_REPOS_PATH=/etc/yum.repos.d
YUM_REPOS_BAK_PATH=/etc/yum.repos.d.bak

function check_iso_img() {
  if [ -z "${ISO_IMG_FILE}" ] || [ ! -f ${ISO_IMG_FILE} ]; then
    echo "iso image: \${ISO_IMG_FILE} should exist."
    exit 1
  fi
}

function check_repo_url() {
  if [ -z "${REPO_BASE_URL}" ]; then
    echo "repo url: \${REPO_BASE_URL} is empty."
    exit 1
  fi
}

function mount_iso_file() {
  ISO_MOUNT_PATH=/mnt/${LINUX_DISTRIBUTION}-iso
  mkdir -p ${ISO_MOUNT_PATH}
  mount -o loop,ro ${ISO_IMG_FILE} ${ISO_MOUNT_PATH}
}

function backup_yum_repo() {
  if [ $(ls -A /etc/yum.repos.d/ | grep ${MARK_NAME} | wc -l) -eq 0 ]; then
    mkdir -p ${YUM_REPOS_BAK_PATH}
    mv ${YUM_REPOS_PATH}/* ${YUM_REPOS_BAK_PATH}
  fi
}

function generate_yum_repo() {
  MODE=$1
  echo "MODE: $MODE"
  backup_yum_repo

  if [ ${MODE} == "iso" ]; then
    cat >${YUM_REPOS_PATH}/${ISO_REPO_CONF} <<EOF
[kubean-iso]
name=Kubean ISO Repo
baseurl=file://${ISO_MOUNT_PATH}
enabled=1
gpgcheck=0
sslverify=0
EOF
    echo "generate: ${YUM_REPOS_PATH}/${ISO_REPO_CONF}"
  fi

  if [ ${MODE} == "url" ]; then
    cat >${YUM_REPOS_PATH}/${URL_REPO_CONF} <<EOF
[kubean-extra]
name=Kubean Extra Repo
baseurl=${REPO_BASE_URL}
enabled=1
gpgcheck=0
sslverify=0
EOF
    echo "generate: ${YUM_REPOS_PATH}/${URL_REPO_CONF}"
  fi
}

function gen_repo_conf_with_iso() {
  shift
  LINUX_DISTRIBUTION=$1
  ISO_IMG_FILE=$2

  echo "LINUX_DISTRIBUTION: $LINUX_DISTRIBUTION, ISO_IMG_FILE: $ISO_IMG_FILE"

  case $LINUX_DISTRIBUTION in
  centos)
    check_iso_img
    mount_iso_file
    generate_yum_repo iso
    ;;

  debian | ubuntu)
    echo "this linux distribution is temporarily not supported."
    ;;

  *)
    echo "unknown linux distribution, currently only supports centos."
    ;;
  esac
}

function gen_repo_conf_with_url() {
  shift
  LINUX_DISTRIBUTION=$1
  REPO_BASE_URL=$2

  echo "LINUX_DISTRIBUTION: $LINUX_DISTRIBUTION, REPO_BASE_URL: $REPO_BASE_URL"

  case $LINUX_DISTRIBUTION in
  centos)
    check_repo_url
    generate_yum_repo url
    ;;

  debian | ubuntu)
    echo "this linux distribution is temporarily not supported."
    ;;

  *)
    echo "unknown linux distribution, currently only supports centos."
    ;;
  esac
}

function show_usage() {
  local cmd=$(basename $0)
  cat <<EOF
Usage:
  $cmd <command>
Examples:
# Mount the ISO image and generate the repo configuration file
./gen_repo_conf.sh --iso-mode \${linux_distribution} \${iso_image_file}
./gen_repo_conf.sh -im centos CentOS-7-x86_64-Everything-2207-02.iso

# Generate repo configuration file according to url
./gen_repo_conf.sh --url-mode \${linux_distribution} \${repo_base_url}
./gen_repo_conf.sh -um centos http://10.8.172.10:8010/centos/7/

Available Commands:
  -im, --iso-mode      use the iso image as the repo source
  -um, --url-mode      use url as repo source
EOF
}

function pdie() {
  show_usage
  echo "$0: ERROR: ${1-}" 1>&2
  exit "${2-1}"
}

while [[ $# -gt 0 ]]; do
  case $1 in
  -im | --iso-mode)
    gen_repo_conf_with_iso $@
    exit 0
    ;;
  -um | --url-mode)
    gen_repo_conf_with_url $@
    exit 0
    ;;
  -h | --help)
    show_usage
    exit 0
    ;;
  *)
    pdie "This command does not exist: $1"
    ;;
  esac
done

# handle non-option arguments
if [[ $# -ne 1 ]]; then
  pdie "Required command not specified."
fi
