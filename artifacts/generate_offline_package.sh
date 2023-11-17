#!/bin/bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail

OPTION=${1:-'all'}
KUBEAN_TAG=${KUBEAN_TAG:-"v0.1.0"}
KUBE_VERSION=${KUBE_VERSION:-""}

CURRENT_DIR=$(pwd)
ARCH=${ARCH:-"amd64"}
OFFLINE_PACKAGE_DIR=${CURRENT_DIR}/${KUBEAN_TAG}/${ARCH}
OFFLINE_FILES_DIR=${OFFLINE_PACKAGE_DIR}/files
OFFLINE_IMAGES_DIR=${OFFLINE_PACKAGE_DIR}/images
OFFLINE_OSPKGS_DIR=${OFFLINE_PACKAGE_DIR}/os-pkgs

ZONE=${ZONE:-"Other"} # CN or Other

function generate_offline_dir() {
  mkdir -p $OFFLINE_FILES_DIR
  mkdir -p $OFFLINE_IMAGES_DIR
  mkdir -p $OFFLINE_OSPKGS_DIR
}

function bump_pause_version() {
  major_version=$(echo "$1" | awk -F . '{print int($1)}')
  minor_version=$(echo "$1" | awk -F . '{print int($2)}')
  patch_version=$(echo "$1" | awk -F . '{print int($3)}')

  local bump_versions=()
  bump_versions+=("${major_version}.$((minor_version + 1))")
  bump_versions+=("${major_version}.$((minor_version + 1)).1")
  bump_versions+=("${major_version}.${minor_version}.$((patch_version + 1))")
  bump_versions+=("$((major_version + 1)).0")
  local bump_pause_tag=""
  for tag in "${bump_versions[@]}"; do
    local ret=0
    curl -sL https://registry.k8s.io/v2/pause/tags/list | jq .tags | grep -v sha256 | grep -q "\"${tag}\"" || ret=$?
    if [[ "${ret}" == 0 ]]; then
      bump_pause_tag=${tag}
      break
    fi
  done
  echo "${bump_pause_tag}"
}

function add_pause_image_addr() {
  local image_list=$1
  local pause_img_addr
  pause_img_addr=$(< "${image_list}" grep pause)
  if [ -n "${pause_img_addr}" ]; then
    pause_addr=$(echo "${pause_img_addr}" | cut -d: -f1)
    pause_tag=$(echo "${pause_img_addr}" | cut -d: -f2)
    new_tag=$(bump_pause_version "${pause_tag}")
    if [ -n "${new_tag}" ]; then
      echo "${pause_addr}:${new_tag}" >> "${image_list}"
    fi
  fi
}

function generate_temp_list() {
  if [ ! -d "kubespray" ]; then
    echo "kubespray git repo should exist."
    exit 1
  fi
  echo "$CURRENT_DIR/kubespray"
  cd $CURRENT_DIR/kubespray

  if [ -z "${KUBE_VERSION}" ]; then
    bash contrib/offline/generate_list.sh -e"image_arch=${ARCH}"
  else
    bash contrib/offline/generate_list.sh -e"image_arch=${ARCH}" -e"kube_version=${KUBE_VERSION}"
  fi

  # Clean up unused images
  remove_images="aws-alb|aws-ebs|cert-manager|netchecker|weave|sig-storage|external_storage|cinder-csi|kubernetesui"
  mv contrib/offline/temp/images.list contrib/offline/temp/images.list.old
  cat contrib/offline/temp/images.list.old | egrep -v ${remove_images} > contrib/offline/temp/images.list
  add_pause_image_addr contrib/offline/temp/images.list
  cp contrib/offline/temp/*.list ${OFFLINE_PACKAGE_DIR}
}

function update_binaris_cn_mirror() {
  local file_list="${CURRENT_DIR}/kubespray/contrib/offline/temp/files.list"
  # 1. backup files list
  mv "${file_list}" "${file_list}.bak"
  # 2. update cn mirror
  local binary_mirror_addr="files.m.daocloud.io"
  while read -r binary_addr; do
    echo "${binary_addr}" | sed "s/https:\/\//&${binary_mirror_addr}\//" >> "${file_list}"
  done <<< "$(cat "${file_list}.bak" || true)"
  # 3. clear files list backup file
  rm -rf "${file_list}.bak"
}

function update_images_cn_mirror() {
  local image_list="${CURRENT_DIR}/kubespray/contrib/offline/temp/images.list"
  # 1. backup images list
  mv "${image_list}" "${image_list}.bak"
  # 2. update cn mirror
  while read -r image_addr; do
    image_addr=${image_addr/docker.io/docker.m.daocloud.io}
    image_addr=${image_addr/gcr.io/gcr.m.daocloud.io}
    image_addr=${image_addr/ghcr.io/ghcr.m.daocloud.io}
    image_addr=${image_addr/k8s.gcr.io/k8s-gcr.m.daocloud.io}
    image_addr=${image_addr/registry.k8s.io/k8s.m.daocloud.io}
    image_addr=${image_addr/quay.io/quay.m.daocloud.io}
    echo "${image_addr}" >> "${image_list}"
  done <<< "$(cat "${image_list}.bak" || true)"
  # 3. clear images list backup file
  rm -rf "${image_list}.bak"
}

function create_files() {
  cd $CURRENT_DIR/kubespray/contrib/offline/
  if [[ "${ZONE}" == "CN" ]]; then
    update_binaris_cn_mirror
  fi

  NO_HTTP_SERVER=true bash manage-offline-files.sh
  cp offline-files.tar.gz $OFFLINE_FILES_DIR
}

function create_images() {
  cd $CURRENT_DIR/artifacts
  if [[ "${ZONE}" == "CN" ]]; then
    update_images_cn_mirror
  fi

  if which skopeo; then
    echo "skopeo check successfully."
  else
    echo "please install skopeo first"
    exit 1
  fi

  echo "begin to download images."
  local images_list_content
  images_list_content=$(cat "${CURRENT_DIR}/kubespray/contrib/offline/temp/images.list")

  rm -rf offline-images && mkdir offline-images

  while read -r image_name; do
    echo "download image $image_name to local"
    ret=0
    skopeo copy --insecure-policy --retry-times=3 --override-os linux --override-arch ${ARCH} "docker://$image_name" "oci:offline-images:$image_name" || ret=$?
    if [ ${ret} -ne 0 ]; then
      echo "skopeo copy image failed, image name: ${image_name}."
      exit 1
    fi
    echo "$image_name" >> offline-images/images.list
  done <<< "${images_list_content}"

  tar -czvf $OFFLINE_IMAGES_DIR/offline-images.tar.gz offline-images

  echo "zipping images completed!"
}

function copy_import_sh() {
    cp $CURRENT_DIR/artifacts/import_files.sh $OFFLINE_FILES_DIR
    cp $CURRENT_DIR/artifacts/import_images.sh $OFFLINE_IMAGES_DIR
    cp $CURRENT_DIR/artifacts/import_ospkgs.sh $OFFLINE_OSPKGS_DIR
}

case $OPTION in
all)
  generate_offline_dir
  generate_temp_list
  create_files
  create_images
  copy_import_sh
  ;;

list)
  generate_temp_list
  ;;

offline_dir)
  generate_offline_dir
  ;;

copy_import_sh)
  copy_import_sh
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
