#!/bin/bash

# set -x
set -eo pipefail

OPTION=$1
DISTRO=${DISTRO:-''}
VERSION=${VERSION:-''}
ARCH=${ARCH:-''}

PKGS_TAR_PATH=${PKGS_TAR_PATH:-"os-pkgs-${DISTRO}-${VERSION}.tar.gz"}

SSH_MODE=${SSH_MODE:-'PWD'} # PWD or KEY
SSH_USER=${SSH_USER:-''}
SSH_CRED=${SSH_CRED:=''}
HOST_IPS=${HOST_IPS:-''}

REMOTE_REPO_PATH='/home/other_repo'

#===========================#
###### Packages config ######
#===========================#

PKGS_YML_CONF='
yum:
  build_tools:
  - createrepo
  - yum-utils
  required_pkgs:
  - libselinux-python
  - python3-libselinux
  - device-mapper-libs
  - nss-tools
  - conntrack
  - container-selinux
  - libseccomp
  - chrony
  - iputils
  - nfs-utils
  required_mods:
  - container-tools

apt:
  build_tools:
  - dpkg-dev
  required_pkgs:
  - python-apt
  - python3-apt
  - aufs-tools
  - apt-transport-https
  - software-properties-common
  - conntrack
  - apparmor
  - libseccomp2
  - ntp
  - iputils-ping
  - nfs-common

commons:
- openssl
- curl
- rsync
- socat
- unzip
- e2fsprogs
- xfsprogs
- ebtables
- bash-completion
- tar
- ipvsadm
- ipset
- sshpass
'


#============================#
###### Common Functions ######
#============================#

WHITE="\033[0m"
RED="\033[31m"
GREEN="\033[32m"
CYAN="\033[36m"

function log_warn() {
  echo -e "${GREEN}[WARN]\t${WHITE} $@${WHITE}"
}

function log_info() {
  echo -e "${CYAN}[INFO]\t${WHITE} $@${WHITE}"
}

function log_erro() {
  echo -e "${RED}[ERRO]\t${WHITE} $@${WHITE}" 1>&2
  exit -1
}

function version_le() {
  # <=
  [ "$1" == "$(echo -e "$1\n$2" | sort -V | head -n1)" ]
}

function version_lt() {
  # <
  [ "$1" == "$2" ] && return 1 || version_le $1 $2
}

function require_arch() {
  case ${ARCH} in
  "amd64" | "x86_64" | "")
    echo "amd64"
    ;;
  "aarch64" | "arm64")
    echo "arm64"
    ;;
  esac
}

function yq_install() {
  local yq_curr_ver=""
  local yq_version="v4.31.1"
  local yq_binary_name="yq_linux_$(require_arch)"
  local yq_binary_url="https://files.m.daocloud.io/github.com/mikefarah/yq/releases/download/${yq_version}/${yq_binary_name}"

  if [ ! -x "$(command -v curl)" ]; then
    log_erro "Please install the curl command line tool."
  fi

  if [ -x "$(command -v yq)" ]; then
    yq_curr_ver=$(yq --version | awk '{print $4}')
  fi

  if [ ! -x "$(command -v yq)" ] || $(version_lt "${yq_curr_ver:1}" "${yq_version:1}"); then
    log_info "install or upgrade yq to ${yq_version}"
    curl --retry 10 --retry-max-time 60 -LO ${yq_binary_url}
    mv ${yq_binary_name} /usr/local/bin/yq -f
    chmod +x /usr/local/bin/yq
    ln /usr/local/bin/yq /usr/bin/yq -f
    yq --version
  else
    log_warn "skip install yq ..."
  fi
}

function prompt_refresh_pkg_list() {
  while true; do
    read -rp "Do you want to refresh package list? (y/n): " choice
    case "${choice}" in
      [Yy]* )
          echo "true"
          return
          ;;
      [Nn]* )
          echo "false"
          return
          ;;
      *)
          echo "Please answer yes or no."
          ;;
    esac
  done
}

#============================#
###### Build OS Package ######
#============================#

function get_local_os_release() {
  local keyword=$1
  local ret=$(cat /etc/os-release | grep "^${keyword}=" | awk -F '=' '{print $2}' | sed 's/\"//g' | tr A-Z a-z)
  echo ${ret}
}

function yum_build() {
  local build_path="/${DISTRO}/${VERSION}/os"
  local build_tools=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.build_tools[]' | tr '\n' ' ')
  local packages=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.required_pkgs[],.commons[]' | tr '\n' ' ')

  mkdir -p ${build_path}
  pushd ${build_path}

  local is_refresh
  is_refresh=$(prompt_refresh_pkg_list)
  if [[ "${is_refresh}" == "true" ]]; then
    yum makecache
  fi

  yum install -y ${build_tools}
  set +e
  for item in ${packages}; do
    repotrack -p ${ARCH} ${item}
    if [ $? -ne 0 ]; then
      log_warn "failed to download package '${item}'"
    fi
  done
  set -e
  createrepo -d ${ARCH}

  popd
}

function dnf_build() {
  local build_path="/${DISTRO}/${VERSION}/os"
  local build_tools=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.build_tools[]' | tr '\n' ' ')
  local packages=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.required_pkgs[],.commons[]' | tr '\n' ' ')
  local modules=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.required_mods[]' | tr '\n' ' ')

  mkdir -p ${build_path}
  pushd ${build_path}

  local is_refresh
  is_refresh=$(prompt_refresh_pkg_list)
  if [[ "${is_refresh}" == "true" ]]; then
    dnf makecache
  fi

  dnf install -y ${build_tools}
  # why use `--alldeps` ?
  # Because it is not certain that the host running the build has the package installed,
  # if the package is installed, the downloaded offline package may be missing the underlying dependencies.
  set +e
  for item in ${packages}; do
    dnf download --resolve --alldeps --destdir=${ARCH} ${item}
    if [ $? -ne 0 ]; then
      log_warn "failed to download package '${item}'"
    fi
  done
  set -e
  createrepo -d ${ARCH}

  # create a repo in ${ARCH}/modules/ with previously downloaded packages and modular metadata
  dnf module install ${modules} --downloadonly --destdir=${ARCH}/modules/ -y
  dnf install 'dnf-command(modulesync)' -y
  dnf modulesync --destdir=${ARCH}/modules/

  popd
}

function apt_build() {
  local build_path="/${DISTRO}/$(require_arch)/${VERSION}"
  local build_tools=$(echo "${PKGS_YML_CONF}" | yq eval '.apt.build_tools[]' | tr '\n' ' ')
  local packages=$(echo "${PKGS_YML_CONF}" | yq eval '.apt.required_pkgs[],.commons[]' | tr '\n' ' ')

  mkdir -p ${build_path}
  pushd ${build_path}

  local is_refresh
  is_refresh=$(prompt_refresh_pkg_list)
  if [[ "${is_refresh}" == "true" ]]; then
    apt-get update
  fi

  apt-get install -y --no-install-recommends ${build_tools}

  set +e
  depend_pkgs_cmd="apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends"
  for item in ${packages}; do
    uris=$(apt-get download $(${depend_pkgs_cmd} ${item} | grep "^\w"))
    if [ $? -ne 0 ]; then
      log_warn "failed to download package '${item}'"
    fi
  done
  set -e
  dpkg-scanpackages "." /dev/null | gzip -9c >Packages.gz

  popd
}

function zypper_build() {
  # TODO zypper install
  log_warn "zypper is not currently supported"
  exit 1
}

function apk_build() {
  # TODO apk add --no-cache
  log_warn "apk is not currently supported"
  exit 1
}

function Build() {
  # Check if tar is installed
  if [ ! -x "$(command -v tar)" ]; then
    log_erro "Please install the tar command line tool."
  fi

  if [ -z "${DISTRO}" ]; then
    DISTRO=$(get_local_os_release 'ID')
  fi

  if [ -z "${VERSION}" ]; then
    VERSION=$(get_local_os_release 'VERSION_ID')
  fi

  if [ -z "${ARCH}" ]; then
    ARCH=$(uname -m)
  fi

  yq_install

  if [ -x "$(command -v dnf)" ]; then
    dnf_build
  elif [ -x "$(command -v yum)" ]; then
    yum_build
  elif [ -x "$(command -v apt-get)" ]; then
    apt_build
  elif [ -x "$(command -v zypper)" ]; then
    zypper_build
  elif [ -x "$(command -v apk)" ]; then
    apk_build
  else
    log_erro "FAILED TO BUILD PACKAGE: Package manager not found."
  fi

  mkdir os-pkgs/ resources/
  mv /${DISTRO} resources/${DISTRO}
  tar -zcvf os-pkgs/os-pkgs-$(require_arch).tar.gz resources --remove-files
  sha256sum os-pkgs/os-pkgs-$(require_arch).tar.gz >os-pkgs/os-pkgs.sha256sum.txt

  curl -Lo ./os-pkgs/import_ospkgs.sh https://raw.githubusercontent.com/kubean-io/kubean/main/artifacts/import_ospkgs.sh
  tar -zcvf os-pkgs-${DISTRO}-${VERSION}.tar.gz os-pkgs/ --remove-files
}

#==============================#
###### Install OS Package ######
#==============================#

function remote_exec() {
  local host=$1
  local cmd=$2

  if [[ "${SSH_MODE}" == "KEY" && -z "${SSH_CRED}" ]]; then
    ssh "${SSH_USER}@${host}" -o StrictHostKeyChecking=no "${cmd}"
  elif [[ "${SSH_MODE}" == "KEY" && -f "${SSH_CRED}" ]]; then
    ssh -i "${SSH_CRED}" "${SSH_USER}@${host}" -o StrictHostKeyChecking=no "${cmd}"
  elif [[ "${SSH_MODE}" == "PWD" ]]; then
    sshpass -p "${SSH_CRED}" ssh "${SSH_USER}@${host}" -o StrictHostKeyChecking=no "${cmd}"
  fi
}

function remote_copy() {
  local host=$1
  local local_path=$2
  local remote_path=$3

  if [[ "${SSH_MODE}" == "KEY" && -z "${SSH_CRED}" ]]; then
    scp "${local_path}" "${SSH_USER}@${host}:${remote_path}"
  elif [[ "${SSH_MODE}" == "KEY" && -f "${SSH_CRED}" ]]; then
    scp -i "${SSH_CRED}" "${local_path}" "${SSH_USER}@${host}:${remote_path}"
  elif [[ "${SSH_MODE}" == "PWD" ]]; then
    sshpass -p "${SSH_CRED}" scp "${local_path}" "${SSH_USER}@${host}:${remote_path}"
  fi
}

function get_remote_os_release() {
  local ip=$1
  local keyword=$2
  local ret=$(remote_exec "${ip}" "cat /etc/os-release |grep '^${keyword}='" | awk -F '=' '{print $2}' | sed 's/\"//g' | tr A-Z a-z)
  echo ${ret}
}

function check_deb_pkg_installed() {
  local ip=$1
  local pkg_name=$2
  remote_exec "${ip}" "dpkg-query --show --showformat='\${db:Status-Status}\n' ${pkg_name} 2>/dev/null"
}

function dnf_install() {
  local ip=$1
  local yum_repos_path='/etc/yum.repos.d'
  local yum_repo_config='other-extra.repo'
  local packages=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.required_pkgs[],.commons[]' | tr '\n' ' ')

  remote_exec "${ip}" "mv /etc/yum.repos.d/ /etc/yum.repos.d.bak/ && mkdir -p /etc/yum.repos.d/"
  # Distribute yum repo configuration
  cat >${yum_repo_config} <<EOF
[other-extra]
name=Other Extra Repo
baseurl=file://${REMOTE_REPO_PATH}/os-pkgs/resources/${DISTRO}/${VERSION}/os/\$basearch/
enabled=1
gpgcheck=0
sslverify=0
[other-module]
name=Other Extra Module
baseurl=file://${REMOTE_REPO_PATH}/os-pkgs/resources/${DISTRO}/${VERSION}/os/\$basearch/modules/
enabled=1
gpgcheck=0
sslverify=0
EOF
  remote_copy "${ip}" "${yum_repo_config}" "${yum_repos_path}"
  remote_exec "${ip}" "dnf clean all && dnf repolist"
  rm ${yum_repo_config} -rf
  # Installing yum packages
  set +e
  # install container-selinux need enable container-tools module
  remote_exec "${ip}" "dnf module enable container-tools -y"
  for item in ${packages}; do
    remote_exec "${ip}" "dnf install -y ${item} --disablerepo=* --enablerepo=other-extra"
    if [ $? -ne 0 ]; then
      log_warn "failed to install package '${item}'"
    fi
  done
  set -e
}

function yum_install() {
  local ip=$1
  local yum_repos_path='/etc/yum.repos.d'
  local yum_repo_config='other-extra.repo'
  local packages=$(echo "${PKGS_YML_CONF}" | yq eval '.yum.required_pkgs[],.commons[]' | tr '\n' ' ')

  remote_exec "${ip}" "mv /etc/yum.repos.d/ /etc/yum.repos.d.bak/ && mkdir -p /etc/yum.repos.d/"
  # Distribute yum repo configuration
  cat >${yum_repo_config} <<EOF
[other-extra]
name=Other Extra Repo
baseurl=file://${REMOTE_REPO_PATH}/os-pkgs/resources/${DISTRO}/${VERSION}/os/\$basearch/
enabled=1
gpgcheck=0
sslverify=0
EOF
  remote_copy "${ip}" "${yum_repo_config}" "${yum_repos_path}"
  remote_exec "${ip}" "yum clean all && yum repolist"
  rm ${yum_repo_config} -rf
  # Installing yum packages
  set +e
  for item in ${packages}; do
    remote_exec "${ip}" "yum install -y ${item} --disablerepo=* --enablerepo=other-extra"
    if [ $? -ne 0 ]; then
      log_warn "failed to install package '${item}'"
    fi
  done
  set -e
}

function apt_install() {
  local ip=$1
  local apt_repo_path='/etc/apt/sources.list'
  local packages=$(echo "${PKGS_YML_CONF}" | yq eval '.apt.required_pkgs[],.commons[]' | tr '\n' ' ')
  local extra_repo="deb [trusted=yes] file://${REMOTE_REPO_PATH}/os-pkgs/resources/${DISTRO}/$(require_arch)/${VERSION} ./"
  local install_failed_list=()

  # Add apt source for remote node
  remote_exec "${ip}" "mv ${apt_repo_path} ${apt_repo_path}.disabled"
  remote_exec "${ip}" "echo \"${extra_repo}\" > ${apt_repo_path}"
  remote_exec "${ip}" "apt-get update"
  # Installing deb packages
  set +e
  for item in ${packages}; do
    ret=$(check_deb_pkg_installed "${ip}" "${item}")
    if [ "${ret}" == "installed" ]; then
      log_warn "the package '${item}' has been installed"
      continue
    fi
    remote_exec "${ip}" "apt-get install -y ${item}"
    if [ $? -ne 0 ]; then
      log_warn "failed to install package '${item}'"
      install_failed_list+=(${item})
    else
      log_info "succeed to install package '${item}'"
    fi
  done
  set -e

  if [ ${#install_failed_list[@]} -ne 0 ]; then
    log_erro "the packages that failed to install are: ${install_failed_list[@]}"
  fi

  # Remove apt source for remote node
  # remote_exec "${ip}" "mv ${apt_repo_path}.disabled  ${apt_repo_path}"

}

function zypper_install() {
  # TODO zypper install
  log_warn "zypper is not currently supported"
  exit 1
}

function apk_install() {
  # TODO apk add --no-cache
  log_warn "apk is not currently supported"
  exit 1
}

function Install() {
  # Check if sshpass is installed
  if [ ! -x "$(command -v sshpass)" ]; then
    log_erro "Please install the sshpass command line tool."
  fi
  # Check if PKGS_TAR_PATH exists
  if [ -z "${PKGS_TAR_PATH}" ] || [ ! -f ${PKGS_TAR_PATH} ]; then
    log_erro "Package tar path: \${PKGS_TAR_PATH} should exist."
  fi
  # Check if HOST_IPS is empty
  if [ -z "${HOST_IPS}" ]; then
    log_erro "Host IPs: \${HOST_IPS} should not be empty."
  fi
  # Check if SSH_USER/SSH_CRED is empty
  if [[ -z "${SSH_USER}" ]]; then
    log_erro "SSH USER/PASS: \${SSH_USER} should not be empty."
  fi

  if [[ "${SSH_MODE}" == "PWD" && -z "${SSH_CRED}" ]]; then
    log_erro "SSH USER/PASS: \${SSH_CRED} should not be empty."
  fi

  yq_install

  for ip in ${HOST_IPS[@]}; do
    if [ -z "$(remote_exec "${ip}" "command -v tar")" ]; then
      log_erro "Node(${ip}) does not have the tar command line installed"
    fi

    if [ -z "${DISTRO}" ]; then
      DISTRO=$(get_remote_os_release ${ip} 'ID')
    fi
    if [ -z "${VERSION}" ]; then
      VERSION=$(get_remote_os_release ${ip} 'VERSION_ID')
    fi
    if [ -z "${ARCH}" ]; then
      ARCH=$(remote_exec "${ip}" "uname -m")
    fi
    # 1. Distribute OS packages to each node
    remote_exec "${ip}" "rm ${REMOTE_REPO_PATH} -rf && mkdir -p ${REMOTE_REPO_PATH}"
    remote_copy "${ip}" "${PKGS_TAR_PATH}" "${REMOTE_REPO_PATH}"

    # 2. Unzip the OS package
    # gunzip os-pkgs.tar.gz
    # cat os-pkgs.tar | cpio -i -d -H tar
    remote_exec "${ip}" "cd ${REMOTE_REPO_PATH} && tar -zxvf $(basename ${PKGS_TAR_PATH})"
    remote_exec "${ip}" "cd ${REMOTE_REPO_PATH}/os-pkgs/ && tar -zxvf os-pkgs-$(require_arch).tar.gz"

    # 3. Install the OS package
    if [ ! -z "$(remote_exec "${ip}" "command -v dnf")" ]; then
      dnf_install ${ip}
    elif [ ! -z "$(remote_exec "${ip}" "command -v yum")" ]; then
      yum_install ${ip}
    elif [ ! -z "$(remote_exec "${ip}" "command -v apt-get")" ]; then
      apt_install ${ip}
    elif [ ! -z "$(remote_exec "${ip}" "command -v zypper")" ]; then
      zypper_install ${ip}
    elif [ ! -z "$(remote_exec "${ip}" "command -v apk")" ]; then
      apk_install ${ip}
    else
      log_erro "FAILED TO INSTALL PACKAGE: Package manager not found."
    fi
    log_info "All packages for Node (${ip}) have been installed."
  done
}

#===========================#
###### Entry functions ######
#===========================#

function Usage() {
  local cmd=$(basename $0)
  cat <<EOF
Usage
  $cmd build
  $cmd install

Commands
  build         for building offline OS packages
  install       for installing offline OS packages
  -h, --help    show help information

Examples
  # Build OS Package
  ./$cmd build

  # Install OS Package
  export PKGS_TAR_PATH=/home/os-pkgs.tar.gz
  export SSH_USER=root
  export SSH_CRED=dangerous
  export HOST_IPS='192.168.10.11 192.168.10.12'
  ./$cmd install
EOF
}

case $OPTION in
build)
  Build
  ;;
install)
  Install
  ;;
-h | --help)
  Usage
  ;;
*)
  Usage
  log_erro "This command does not exist: $1"
  ;;
esac
