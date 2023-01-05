#!/bin/bash

set -eo pipefail

ISO_IMG_FILE=''
LINUX_DISTRIBUTION=''
REPO_BASE_URL=''
ISO_MOUNT_PATH=''

MARK_NAME=Kubean
YUM_ISO_REPO_CONF=${MARK_NAME}-ISO.repo
YUM_URL_REPO_CONF=${MARK_NAME}-URL.repo

YUM_REPOS_PATH=/etc/yum.repos.d
YUM_REPOS_BAK_PATH=/etc/yum.repos.d.bak

APT_REPOS_DIR=/etc/apt/sources.list.d
APT_REPOS_PATH=/etc/apt/sources.list
APT_ISO_REPO_CONF=${MARK_NAME}-ISO.list
APT_URL_REPO_CONF=${MARK_NAME}-URL.list

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
  local OS=${2:-""}
  echo "MODE: $MODE"
  backup_yum_repo

  if [ ${MODE} == "iso" ]; then
    if [ "${OS}" == "rhel" ]; then
      cat >${YUM_REPOS_PATH}/${YUM_ISO_REPO_CONF} <<EOF
[kubean-iso-BaseOS]
name=Kubean ISO Repo BaseOS
baseurl=file://${ISO_MOUNT_PATH}/BaseOS
enabled=1
gpgcheck=0
sslverify=0

[kubean-iso-AppStream]
name=Kubean ISO Repo AppStream
baseurl=file://${ISO_MOUNT_PATH}/AppStream
enabled=1
gpgcheck=0
sslverify=0
EOF
    else
    cat >${YUM_REPOS_PATH}/${YUM_ISO_REPO_CONF} <<EOF
[kubean-iso]
name=Kubean ISO Repo
baseurl=file://${ISO_MOUNT_PATH}
enabled=1
gpgcheck=0
sslverify=0
EOF
    fi
    echo "generate: ${YUM_REPOS_PATH}/${YUM_ISO_REPO_CONF}"
  fi

  if [ ${MODE} == "url" ]; then
    cat >${YUM_REPOS_PATH}/${YUM_URL_REPO_CONF} <<EOF
[kubean-extra]
name=Kubean Extra Repo
baseurl=${REPO_BASE_URL}
enabled=1
gpgcheck=0
sslverify=0
EOF
    echo "generate: ${YUM_REPOS_PATH}/${YUM_URL_REPO_CONF}"
  fi
}

function backup_apt_repo() {
  if [ $(ls -A ${APT_REPOS_DIR} | grep ${MARK_NAME} | wc -l) -eq 0 ]; then
    mv ${APT_REPOS_DIR}{,.bak}
    mkdir -p ${APT_REPOS_DIR}
  fi
  [ -f ${APT_REPOS_PATH} ] && mv ${APT_REPOS_PATH}{,.bak} || true
}

function get_apt_codename_from_os_release() {
  echo "$(cat /etc/os-release | sed -n -r 's/^VERSION_CODENAME=|^UBUNTU_CODENAME=//p' | head -n1)"
}

function get_apt_codename_from_iso() {
  for codename in $(find ${ISO_MOUNT_PATH}/dists/ -maxdepth 1 -type d -exec basename {} \;); do
    [[ "${codename}" =~ xenial|bionic|focal|jammy ]] && { echo "${codename}"; return; }
  done
  echo $(get_apt_codename_from_os_release)
}

function generate_apt_repo() {
  MODE=$1
  echo "MODE: $MODE"
  backup_apt_repo

  if [ ${MODE} == "iso" ]; then
    cat >${APT_REPOS_DIR}/${APT_ISO_REPO_CONF} <<EOF
deb file://${ISO_MOUNT_PATH} $(get_apt_codename_from_iso) main restricted
EOF
    echo "generate: ${APT_REPOS_DIR}/${APT_ISO_REPO_CONF}"
  fi

  if [ ${MODE} == "url" ]; then
    cat >${APT_REPOS_DIR}/${APT_URL_REPO_CONF} <<EOF
deb ${REPO_BASE_URL} $(get_apt_codename_from_os_release) main restricted
EOF
    echo "generate: ${APT_REPOS_DIR}/${APT_URL_REPO_CONF}"
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
  redhat | rhel)
    check_iso_img
    mount_iso_file
    generate_yum_repo iso rhel
    ;;
  debian | ubuntu)
    check_iso_img
    mount_iso_file
    generate_apt_repo iso
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
  centos | redhat | rhel)
    check_repo_url
    generate_yum_repo url
    ;;

  debian | ubuntu)
    check_repo_url
    generate_apt_repo url
    ;;

  *)
    echo "unknown linux distribution, currently only supports centos."
    ;;
  esac
}

function show_usage() {
  local cmd=$(basename $0)
  cat <<EOF
Usage
  $cmd [ -im | --iso-mode ] <linux_distribution> <iso_image_file>
  $cmd [ -um | --url-mode ] <linux_distribution> <repo_base_url>

Commands
  -im, --iso-mode       use the iso image as the repo source
  -um, --url-mode       use url as repo source

Arguments
  linux_distribution       supported for centos, redhat(rhel), ubuntu only
  iso_image_file           path to iso image file
  repo_base_url            url to access remote repo

Examples
  # Mount the ISO image and generate the repo configuration file
  ./gen_repo_conf.sh -im centos CentOS-7-x86_64-Everything-2207-02.iso

  # Generate repo configuration file according to url
  ./gen_repo_conf.sh -um centos http://10.8.172.10:8010/centos/7/
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
