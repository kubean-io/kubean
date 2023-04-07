#!/bin/bash

set -eo pipefail

OPTION=${1:-'create_localartifactset'} ## create_localartifactset  create_manifest

KUBESPRAY_TAG=${KUBESPRAY_TAG:-"v2.19.0"} ## env from github action
KUBEAN_TAG=${KUBEAN_TAG:-"v0.1.0"}        ## env from github action
KUBE_VERSION=${KUBE_VERSION:-""}

CURRENT_DIR=$(cd $(dirname $0); pwd) ## artifacts dir
CURRENT_DATE=$(date +%Y%m%d)

ARTIFACTS_TEMPLATE_DIR=artifacts/template
KUBEAN_OFFLINE_VERSION_TEMPLATE=${ARTIFACTS_TEMPLATE_DIR}/localartifactset.template.yml
KUBEAN_INFO_MANIFEST_TEMPLATE=${ARTIFACTS_TEMPLATE_DIR}/manifest.template.yml

CHARTS_TEMPLATE_DIR=charts/kubean/templates
OFFLINE_PACKAGE_DIR=${KUBEAN_TAG}
KUBEAN_OFFLINE_VERSION_CR=${OFFLINE_PACKAGE_DIR}/localartifactset.cr.yaml
KUBEAN_INFO_MANIFEST_CR=${CHARTS_TEMPLATE_DIR}/manifest.cr.yaml

KUBESPRAY_DIR=kubespray
KUBESPRAY_OFFLINE_DIR=${KUBESPRAY_DIR}/contrib/offline
VERSION_VARS_YML=${KUBESPRAY_OFFLINE_DIR}/version.yml

function check_dependencies() {
  if ! which yq; then
    echo "need yq (https://github.com/mikefarah/yq)."
    exit 1
  fi
  if [ ! -d ${KUBESPRAY_DIR} ]; then
    echo "${KUBESPRAY_DIR} git repo should exist."
    exit 1
  fi
}

function extract_etcd_version() {
  kube_version=${1} ## v1.23.1
  IFS='.'
  read -ra arr <<<"${kube_version}"
  major="${arr[0]}.${arr[1]}"
  version=$(yq ".etcd_supported_versions.\"${major}\"" kubespray/roles/download/defaults/main.yml)
  echo "$version"
}

function extract_version() {
  version_name="${1}"  ## cni_version
  dir=${2:-"download"} ## kubespray-defaults  or download
  version=$(yq ".${version_name}" kubespray/roles/"${dir}"/defaults/main.*ml)
  echo "$version"
}

function extract_version_range() {
  range_path="${1}"    ## .cni_binary_checksums.amd64
  dir=${2:-"download"} ## kubespray-defaults  or download
  version=$(yq "${range_path} | keys" kubespray/roles/"${dir}"/defaults/main.*ml --output-format json)
  version=$(echo $version | tr -d '\n \r') ## ["v1","v2"]
  echo "${version}"
}

function extract_docker_version_range() {
  os=${1:-"redhat-7"}
  version=$(yq ".docker_versioned_pkg | keys" kubespray/roles/container-engine/docker/vars/${os}.yml --output-format json)
  version=$(echo $version | tr -d '\n \r') ## ["v1","v2"]
  echo "${version}"
}

function update_offline_version_cr() {
  index=$1 ## start with zero
  name=$2  ## cni containerd ...
  version_val=$3
  if [ $(yq ".spec.items[$index].name" $KUBEAN_OFFLINE_VERSION_CR) != "${name}" ]; then
    echo "error param $index $name"
    exit 1
  fi
  version_val=${version_val} yq -i ".spec.items[$index].versionRange[0]=strenv(version_val)" $KUBEAN_OFFLINE_VERSION_CR
}

function update_docker_offline_version() {
  os=$1
  version_range=$2
  OS=$os yq -i ".spec.docker |= map(select(.os == strenv(OS)).versionRange |= ${version_range} | ..style=\"double\" )" \
    $KUBEAN_OFFLINE_VERSION_CR
}

function create_offline_version_cr() {
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
  cp $KUBEAN_OFFLINE_VERSION_TEMPLATE $KUBEAN_OFFLINE_VERSION_CR
  CR_NAME=offlineversion-${CURRENT_DATE} yq -i '.metadata.name=strenv(CR_NAME)' $KUBEAN_OFFLINE_VERSION_CR
  KUBESPRAY_TAG=${KUBESPRAY_TAG} yq -i '.spec.kubespray=strenv(KUBESPRAY_TAG)' $KUBEAN_OFFLINE_VERSION_CR

  update_offline_version_cr "0" "cni" "$cni_version"
  update_offline_version_cr "1" "containerd" "$containerd_version"
  update_offline_version_cr "2" "kube" "$kube_version"
  update_offline_version_cr "3" "calico" "$calico_version"
  update_offline_version_cr "4" "cilium" "$cilium_version"
  update_offline_version_cr "5" "flannel" "$flannel_version"
  update_offline_version_cr "6" "kube-ovn" "$kube_ovn_version"
  update_offline_version_cr "7" "etcd" "$etcd_version"
  update_docker_offline_version "redhat-7" "${docker_version_range_redhat7}"
}

function update_info_manifest_cr() {
  index=$1 ## start with zero
  name=$2  ## cni containerd ...
  default_version_val=$3
  version_range=$4
  if [ $(yq ".spec.components[$index].name" $KUBEAN_INFO_MANIFEST_CR) != "${name}" ]; then
    echo "error param $index $name"
    exit 1
  fi

  yq -i ".spec.components[$index].defaultVersion=\"${default_version_val}\"" $KUBEAN_INFO_MANIFEST_CR
  yq -i ".spec.components[$index].versionRange |=  ${version_range} | ..style=\"double\" " $KUBEAN_INFO_MANIFEST_CR ## update string array
}

function update_docker_component_version() {
  os=$1
  default_version=$2
  version_range=$3

  OS=$os yq -i ".spec.docker |= map(select(.os == strenv(OS)).defaultVersion=\"${default_version}\")" \
    $KUBEAN_INFO_MANIFEST_CR

  OS=$os yq -i ".spec.docker |= map(select(.os == strenv(OS)).versionRange |= ${version_range})" \
    $KUBEAN_INFO_MANIFEST_CR
}

function update_info_manifest_cr_name() {
  kubean_version=${KUBEAN_TAG//./-}
  yq -i ".metadata.name=\"kubeaninfomanifest-${kubean_version}\"" $KUBEAN_INFO_MANIFEST_CR
}

function create_info_manifest_cr() {
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

  cp $KUBEAN_INFO_MANIFEST_TEMPLATE $KUBEAN_INFO_MANIFEST_CR
  KUBESPRAY_TAG=${KUBESPRAY_TAG} yq -i '.spec.kubesprayVersion=strenv(KUBESPRAY_TAG)' $KUBEAN_INFO_MANIFEST_CR
  KUBEAN_TAG=${KUBEAN_TAG} yq -i '.spec.kubeanVersion=strenv(KUBEAN_TAG)' $KUBEAN_INFO_MANIFEST_CR

  update_info_manifest_cr_name
  update_info_manifest_cr 0 cni "${cni_version_default}" "${cni_version_range}"
  update_info_manifest_cr 1 containerd "${containerd_version_default}" "${containerd_version_range}"
  update_info_manifest_cr 2 kube "${kube_version_default}" "${kube_version_range}"
  update_info_manifest_cr 3 calico "${calico_version_default}" "${calico_version_range}"
  update_info_manifest_cr 4 cilium "${cilium_version_default}" "${cilium_version_range}"
  update_info_manifest_cr 5 flannel "${flannel_version_default}" "${flannel_version_range}"
  update_info_manifest_cr 6 kube-ovn "${kube_ovn_version_default}" "${kube_ovn_version_range}"
  update_info_manifest_cr 7 etcd "${etcd_version_default}" "${etcd_version_range}"
  update_docker_component_version "redhat-7" "${docker_version_default}" "${docker_version_range_redhat7}"
  update_docker_component_version "debian" "${docker_version_default}" "${docker_version_range_debian}"
  update_docker_component_version "ubuntu" "${docker_version_default}" "${docker_version_range_ubuntu}"
}

case $OPTION in
create_localartifactset)
  check_dependencies
  create_offline_version_cr
  ;;

create_manifest)
  check_dependencies
  create_info_manifest_cr
  ;;

*)
  echo -n "unknown operator"
  ;;
esac
