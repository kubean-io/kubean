#!/bin/bash -e

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

TARGET_ARCH=${TARGET_ARCH:-""}
OCI_PATH=${OCI_PATH:-"offline-images"}
REGISTRY_ADDR=${REGISTRY_ADDR:-""}
REGISTRY_USER=${REGISTRY_USER:-""}
REGISTRY_PASS=${REGISTRY_PASS:-""}

#============================#
###### Manifest Merge ######
#============================#

function image::log_info() { echo -e "\033[32m[info] $*\033[0m"; }
function image::log_warn() { echo -e "\033[33m[warn] $*\033[0m"; }
function image::log_erro() { echo -e "\033[31m[err] $*\033[0m" && exit 1; }

function image::get_image_archs() {
  local image_name=$1
  echo "$(podman manifest inspect "${image_name}" 2>&1 | grep architecture | awk '{print $2}' | sed 's/"\|,//g' | sed ':a;N;$!ba;s/\n/ /g')"
}

function image::merge_multi_arch() {
  local img_name=$1
  
  if [[ -z "${img_name}" ]]; then
    image::log_erro "empty image name"
  fi

  local copy_args="oci:${OCI_PATH}:${img_name}"
  local manifest_list_name="${img_name}-manifest"
  local src_image_name="${REGISTRY_ADDR}/${img_name}"
  local target_image_name="${src_image_name}-${TARGET_ARCH}"
  local copy_failed_list=$(cd "$(dirname "$0")";pwd)/copy_image_failed.list

  ### Push {registry_address}/{image_name}:{tag}-{target_arch} to registry.
  image::log_info "import image from ${copy_args} to ${target_image_name}"
  skopeo copy --insecure-policy --retry-times=3 --src-tls-verify=false --dest-tls-verify=false \
    "${copy_args}" docker://"${target_image_name}" >/dev/null
  if [[ $? -ne 0 ]]; then
    image::log_warn "skopeo copy ${copy_args} failed!"
    echo "${img_name}" >>"${copy_failed_list}"
    return 0
  fi

  ### Detect if {registry_address}/{image_name}:{tag} image is not in the registry.
  if [[ $(podman manifest inspect "${src_image_name}" 2>&1) == *"unknown"* ]]; then
    image::log_info "${src_image_name} not found in registry ${REGISTRY_ADDR}"

    image::log_info "[manifest create] manifest list name: ${manifest_list_name}"
    podman manifest create --insecure --amend "${manifest_list_name}" "${target_image_name}"

    image::log_info "push manifest ${manifest_list_name}"
    podman manifest push "${manifest_list_name}" "${src_image_name}" >/dev/null
    return 0
  fi

  local is_contain_target_arch="false"
  local ori_archs=$(image::get_image_archs "${src_image_name}")
  image::log_info "original manifest arch: [${ori_archs}]"

  if [[ "${ori_archs}" == "" ]]; then
    ### The image {registry_address}/{image_name}:{tag} is in the registry, But manifest lists is not implemented.
    image::log_info "[manifest create] the manifest is not a list, add original image into list"
    podman manifest create --amend "${manifest_list_name}" "${src_image_name}"
    local origin_arch=$(podman manifest inspect "${manifest_list_name}" | awk -F: '/"architecture"/ {print $2}' | sed 's/[", ]//g')
    if [[ "${origin_arch}" != "${TARGET_ARCH}" ]]; then
      image::log_info "image retag from ${src_image_name} to ${src_image_name}-${origin_arch}"
      skopeo copy --insecure-policy --retry-times=3 --src-tls-verify=false --dest-tls-verify=false \
        docker://"${src_image_name}" docker://"${src_image_name}-${origin_arch}" >/dev/null
      podman manifest create --amend "${manifest_list_name}" "${src_image_name}-${origin_arch}"
    fi
  else
    for arch_item in ${ori_archs}; do
      if [ "${TARGET_ARCH}" == "${arch_item}" ]; then
        image::log_warn "skipping to push since it already exists in remote"
        return
      fi
    done
    ### The image {registry_address}/{image_name}:{tag} is in the registry, And manifest lists is implemented.
    image::log_info "the manifest is a list, push original arch into list"
    for arch_item in ${ori_archs}; do
      image::log_info "adding ${arch_item} into manifest list.."
      local image_name
      if [[ "${arch_item}" == "${TARGET_ARCH}" ]]; then
        is_contain_target_arch="true"
        image_name="${target_image_name}"
      else
        image_name="${src_image_name}-${arch_item}" 
      fi
      image::log_info "[manifest create] integration of pre-existing architectures: ${image_name}"
      podman manifest create --amend "${manifest_list_name}" "${image_name}"
    done
  fi

  if [[ "${is_contain_target_arch}" == "false" ]]; then
    image::log_info "[manifest create] integrating target architecture: ${target_image_name}"
    podman manifest create --amend "${manifest_list_name}" "${target_image_name}"
  fi
  podman manifest push "${manifest_list_name}" "${src_image_name}" >/dev/null
}

function image::show_multi_arch() {
  local img_name=$1

  if [ -z "${img_name}" ]; then
    image::log_erro "emtpy image name"
  fi

  local src_image_name="${REGISTRY_ADDR}/${img_name}"
  local ori_archs=$(image::get_image_archs "${src_image_name}")
  if [ -n "${ori_archs}" ]; then
    echo "${src_image_name} [${ori_archs}]"
  else
    image::log_warn "${src_image_name} empty architecture"
  fi
}

function image::pre_processing() {
  # Check if registry addr is empty
  if [[ -z "${REGISTRY_ADDR}" ]]; then
    image::log_erro "registry address cannot be empty."
  fi
  # Check if skopeo is not installed
  if [[ ! -x "$(command -v skopeo)" ]]; then
    image::log_erro "skopeo should be installed."
  fi
  # Check if podman is not installed
  if [[ ! -x "$(command -v podman)" ]]; then
    image::log_erro "podman should be installed."
  fi
  # Check if the podman version matches
  local podman_version='4.4.4'
  local current_podman_version=$(podman --version | awk '{print $3}')
  if [ ${podman_version} != ${current_podman_version} -a $(echo -e "${current_podman_version}\n${podman_versoin}" | sort -rV | head -1) == "${podman_version}" ]; then
    image::log_erro "Check podman version mismatch. [expected version: ${podman_version}]"
  fi

  # set registry inseure config for podman
  mkdir -p /etc/containers/registries.conf.d
  cat >/etc/containers/registries.conf.d/myregistry.conf <<EOF
[[registry]]
location="${REGISTRY_ADDR}"
insecure=true
EOF
  # skopeo login
  if [[ -n "${REGISTRY_USER}" ]] && [[ -n "${REGISTRY_PASS}" ]]; then
    if ! skopeo login "${REGISTRY_ADDR}" -u "${REGISTRY_USER}" -p "${REGISTRY_PASS}" --tls-verify=false; then
      image::log_erro "failed to login to ${REGISTRY_ADDR}"
    fi
  fi

  [ ! -s "${OCI_PATH}/images.list" ] && image::log_erro "${OCI_PATH}/images.list not found or empty"
  [ "$OCI_PATH" == "offline-images" -a ! -d $OCI_PATH ] && tar -xzf offline-images.tar.gz

  [ -n "$TARGET_ARCH" ] && return
  TARGET_ARCH=$(skopeo inspect oci:$OCI_PATH:$(head -1 $OCI_PATH/images.list) | awk -F: '/Architecture/ {print $2}' | sed 's/[[:space:]",]//g')
}

function image::images_list_handler() {
  local func_name=$1
  local stime=${SECONDS}
  while read -r line; do ${func_name} "${line}"; done <<<"$(cat "${OCI_PATH}/images.list")"
  image::log_info "spend $((SECONDS - stime)) seconds"
}

function image::batch_merge_multi_arch() {
  image::pre_processing
  image::images_list_handler image::merge_multi_arch
}

function image::batch_show_multi_arch() {
  image::pre_processing
  image::images_list_handler image::show_multi_arch
}

function image::batch_remove_manifests() {
  local manifest_list=$(podman images | grep manifest)
  if [[ "${manifest_list}" == "" ]]; then
    image::log_warn "current local manifest list is empty."
    return
  fi
  while read -r line; do
    local img_repo=$(echo "${line}" | awk '{print $1}')
    local img_tag=$(echo "${line}" | awk '{print $2}')
    if [[ -z "${img_repo}" ]]; then
      image::log_erro "failed to get img_repo for line: ${line}"
    fi
    if [[ -z "${img_tag}" ]]; then
      image::log_erro "failed to get img_tag for line: ${line}"
    fi
    local manifest_list_name="${img_repo}:${img_tag}"
    image::log_info "remove manifest: ${manifest_list_name}"
    podman manifest rm "${manifest_list_name}" >/dev/null
  done <<<"${manifest_list}"
}

if [ "${BASH_SOURCE[0]}" == "$0" ]; then
  image::batch_merge_multi_arch
  image::batch_remove_manifests
fi