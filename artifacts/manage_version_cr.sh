#!/bin/bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail

OPTION=${1:-'create_localartifactset'} ## create_localartifactset  create_manifest

SPRAY_COMMIT=${SPRAY_COMMIT:-""}
SPRAY_RELEASE=${SPRAY_RELEASE:-"master"}
SPRAY_COMMIT_TIMESTAMP=${SPRAY_COMMIT_TIMESTAMP:-""}

SPRAY_TAG=${SPRAY_TAG:-"v2.19.0"} ## env from github action
KUBEAN_TAG=${KUBEAN_TAG:-""}        ## env from github action
KUBE_VERSION=${KUBE_VERSION:-""}

CURRENT_DIR=$(cd $(dirname $0); pwd) ## artifacts dir
CURRENT_TIME=$(date +%s)

ARTIFACTS_TEMPLATE_DIR=artifacts/template
KUBEAN_LOCALARTIFACTSET_TEMPLATE=${ARTIFACTS_TEMPLATE_DIR}/localartifactset.template.yml
KUBEAN_MANIFEST_TEMPLATE=${ARTIFACTS_TEMPLATE_DIR}/manifest.template.yml

CHARTS_TEMPLATE_DIR=charts/kubean/templates
OFFLINE_PACKAGE_DIR=${KUBEAN_TAG}
KUBEAN_LOCALARTIFACTSET_CR=${OFFLINE_PACKAGE_DIR}/localartifactset.cr.yaml
KUBEAN_MANIFEST_CR=${CHARTS_TEMPLATE_DIR}/manifest.cr.yaml

SPRAY_DIR=kubespray
SPRAY_OFFLINE_DIR=${SPRAY_DIR}/contrib/offline
VERSION_VARS_YML=${SPRAY_OFFLINE_DIR}/version.yml

function check_dependencies() {
  if ! which yq; then
    echo "need yq (https://github.com/mikefarah/yq)."
    exit 1
  fi
  if [ ! -d ${SPRAY_DIR} ]; then
    echo "${SPRAY_DIR} git repo should exist."
    exit 1
  fi
}

function extract_etcd_version() {
  local kube_version=${1} ## v1.26.5
  IFS='.'
  read -ra arr <<<"${kube_version}"
  major="${arr[0]}.${arr[1]}"
  version=$(yq ".etcd_supported_versions.\"${major}\"" kubespray/roles/download/defaults/main.yml)
  echo "${version}"
}

function extract_version() {
  local version_name="${1}"  ## cni_version
  local dir="${2}" ## kubespray-defaults  or download
  local version
  version=$(yq ".${version_name}" kubespray/roles/download/defaults/main.yml)
  if [[ -n "${dir}" ]]; then
    version=$(yq ".${version_name}" kubespray/roles/"${dir}"/defaults/main.*ml)
  fi
  echo "${version}"
}

function extract_version_range() {
  local range_path="${1}"    ## .cni_binary_checksums.amd64
  local version
  version=$(yq "${range_path} | keys" kubespray/roles/download/defaults/main.yml --output-format json)
  version=$(echo "${version}" | tr -d '\n \r') ## ["v1","v2"]
  echo "${version}"
}

function extract_docker_version_range() {
  os=${1:-"redhat-7"}
  version=$(yq ".docker_versioned_pkg | keys" kubespray/roles/container-engine/docker/vars/${os}.yml --output-format json)
  version=$(echo $version | tr -d '\n \r') ## ["v1","v2"]
  echo "${version}"
}

function update_custom_resource_metadata() {
  local cr_name=${1}
  local cr_yaml_path=${2}

  local old_cr_postfix=${KUBEAN_TAG//./-}
  if [[ ${cr_name} == 'localartifactset' ]]; then
    old_cr_postfix=${CURRENT_TIME}
  fi

  if [[ "${SPRAY_RELEASE}" != 'master' ]]; then
    yq -i ".metadata.name=\"${cr_name}-${SPRAY_RELEASE}-${SPRAY_COMMIT}\"" "${cr_yaml_path}"
    yq -i ".metadata.labels.\"kubean.io/sprayRelease\"=\"${SPRAY_RELEASE}\"" "${cr_yaml_path}"
    yq -i ".metadata.annotations.\"kubean.io/sprayTimestamp\"=\"${SPRAY_COMMIT_TIMESTAMP}\"" "${cr_yaml_path}"
    yq -i ".metadata.annotations.\"kubean.io/sprayRelease\"=\"${SPRAY_RELEASE}\"" "${cr_yaml_path}"
    yq -i ".metadata.annotations.\"kubean.io/sprayCommit\"=\"${SPRAY_COMMIT}\"" "${cr_yaml_path}"
  else
    yq -i ".metadata.name=\"${cr_name}-${old_cr_postfix}\"" "${cr_yaml_path}"
    yq -i ".metadata.labels.\"kubean.io/sprayRelease\"=\"master\"" "${cr_yaml_path}"
  fi
}

function update_localartifactset_cr() {
  index=$1 ## start with zero
  name=$2  ## cni containerd ...
  version_val=$3
  if [ $(yq ".spec.items[$index].name" ${KUBEAN_LOCALARTIFACTSET_CR}) != "${name}" ]; then
    echo "error param $index $name"
    exit 1
  fi
  version_val=${version_val} yq -i ".spec.items[$index].versionRange[0]=strenv(version_val)" ${KUBEAN_LOCALARTIFACTSET_CR}
}

function update_docker_offline_version() {
  os=$1
  version_range=$2
  OS=$os yq -i ".spec.docker |= map(select(.os == strenv(OS)).versionRange |= ${version_range} | ..style=\"double\" )" \
    ${KUBEAN_LOCALARTIFACTSET_CR}
}

function create_localartifactset_cr() {
  cni_version=$(extract_version "cni_version")
  containerd_version=$(extract_version "containerd_version")

  if [ -z "${KUBE_VERSION}" ]; then
    kube_version=$(extract_version "kube_version" "kubespray-defaults")
  else
    kube_version=${KUBE_VERSION}
  fi

  calico_version=$(extract_version "calico_version")
  cilium_version=$(extract_version "cilium_version")
  flannel_version=$(extract_version "flannel_version")
  kube_ovn_version=$(extract_version "kube_ovn_version")
  etcd_version=$(extract_etcd_version "$kube_version")

  docker_version_range_redhat7=["18.09","19.03","20.10"]

  mkdir -p $OFFLINE_PACKAGE_DIR
  cp ${KUBEAN_LOCALARTIFACTSET_TEMPLATE} ${KUBEAN_LOCALARTIFACTSET_CR}
  update_custom_resource_metadata "localartifactset" "${KUBEAN_LOCALARTIFACTSET_CR}"
  SPRAY_TAG=${SPRAY_TAG} yq -i '.spec.kubespray=strenv(SPRAY_TAG)' ${KUBEAN_LOCALARTIFACTSET_CR}

  update_localartifactset_cr "0" "cni" "$cni_version"
  update_localartifactset_cr "1" "containerd" "$containerd_version"
  update_localartifactset_cr "2" "kube" "$kube_version"
  update_localartifactset_cr "3" "calico" "$calico_version"
  update_localartifactset_cr "4" "cilium" "$cilium_version"
  update_localartifactset_cr "5" "flannel" "$flannel_version"
  update_localartifactset_cr "6" "kube-ovn" "$kube_ovn_version"
  update_localartifactset_cr "7" "etcd" "$etcd_version"
  update_docker_offline_version "redhat-7" "${docker_version_range_redhat7}"
}

function update_manifest_cr() {
  index=$1 ## start with zero
  name=$2  ## cni containerd ...
  default_version_val=$3
  version_range=$4
  if [ $(yq ".spec.components[$index].name" ${KUBEAN_MANIFEST_CR}) != "${name}" ]; then
    echo "error param $index $name"
    exit 1
  fi

  yq -i ".spec.components[$index].defaultVersion=\"${default_version_val}\"" ${KUBEAN_MANIFEST_CR}
  yq -i ".spec.components[$index].versionRange |=  ${version_range} | ..style=\"double\" " ${KUBEAN_MANIFEST_CR} ## update string array
}

function update_docker_component_version() {
  os=$1
  default_version=$2
  version_range=$3

  OS=$os yq -i ".spec.docker |= map(select(.os == strenv(OS)).defaultVersion=\"${default_version}\")" \
    ${KUBEAN_MANIFEST_CR}

  OS=$os yq -i ".spec.docker |= map(select(.os == strenv(OS)).versionRange |= ${version_range})" \
    ${KUBEAN_MANIFEST_CR}
}

function create_manifest_cr() {
  cni_version_default=$(extract_version "cni_version")
  cni_version_range=$(extract_version_range ".cni_binary_checksums.amd64")

  containerd_version_default=$(extract_version "containerd_version")
  containerd_version_range=$(extract_version_range ".containerd_archive_checksums.amd64")

  if [ -z "${KUBE_VERSION}" ]; then
    kube_version_default=$(extract_version "kube_version" "kubespray-defaults")
  else
    kube_version_default=${KUBE_VERSION}
  fi

  kube_version_range=$(extract_version_range ".kubelet_checksums.amd64")

  calico_version_default=$(extract_version "calico_version")
  calico_version_range=$(extract_version_range ".calico_crds_archive_checksums")

  cilium_version_default=$(extract_version "cilium_version")
  cilium_version_range="[]" ## anything

  flannel_version_default=$(extract_version "flannel_version")
  flannel_version_range="[]"

  kube_ovn_version_default=$(extract_version "kube_ovn_version")
  kube_ovn_version_range="[]"

  etcd_version_default=$(extract_etcd_version "${kube_version_default}")
  etcd_version_range=$(extract_version_range ".etcd_binary_checksums.amd64")

  docker_version_default=$(extract_version "docker_version" "container-engine/docker")
  docker_version_range_redhat7=$(extract_docker_version_range "redhat-7")
  docker_version_range_debian=$(extract_docker_version_range "debian")
  docker_version_range_ubuntu=$(extract_docker_version_range "ubuntu")

  cp ${KUBEAN_MANIFEST_TEMPLATE} ${KUBEAN_MANIFEST_CR}
  SPRAY_TAG=${SPRAY_TAG} yq -i '.spec.kubesprayVersion=strenv(SPRAY_TAG)' ${KUBEAN_MANIFEST_CR}
  KUBEAN_TAG=${KUBEAN_TAG} yq -i '.spec.kubeanVersion=strenv(KUBEAN_TAG)' ${KUBEAN_MANIFEST_CR}

  update_custom_resource_metadata "manifest" "${KUBEAN_MANIFEST_CR}"
  update_manifest_cr 0 cni "${cni_version_default}" "${cni_version_range}"
  update_manifest_cr 1 containerd "${containerd_version_default}" "${containerd_version_range}"
  update_manifest_cr 2 kube "${kube_version_default}" "${kube_version_range}"
  update_manifest_cr 3 calico "${calico_version_default}" "${calico_version_range}"
  update_manifest_cr 4 cilium "${cilium_version_default}" "${cilium_version_range}"
  update_manifest_cr 5 flannel "${flannel_version_default}" "${flannel_version_range}"
  update_manifest_cr 6 kube-ovn "${kube_ovn_version_default}" "${kube_ovn_version_range}"
  update_manifest_cr 7 etcd "${etcd_version_default}" "${etcd_version_range}"
  update_docker_component_version "redhat-7" "${docker_version_default}" "${docker_version_range_redhat7}"
  update_docker_component_version "debian" "${docker_version_default}" "${docker_version_range_debian}"
  update_docker_component_version "ubuntu" "${docker_version_default}" "${docker_version_range_ubuntu}"
}

function merge_kubespray_offline_download_files() {
  if [ -d 'kubespray/roles/download/defaults/main' ]; then
    cat kubespray/roles/download/defaults/main/* | sed '/^---$/d' > kubespray/roles/download/defaults/main.yml
  fi
}

case $OPTION in
create_localartifactset)
  check_dependencies
  merge_kubespray_offline_download_files
  create_localartifactset_cr
  ;;

create_manifest)
  check_dependencies
  merge_kubespray_offline_download_files
  create_manifest_cr
  ;;

*)
  echo -n "unknown operator"
  ;;
esac
