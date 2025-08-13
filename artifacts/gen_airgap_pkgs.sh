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

function patch_temp_list() {
    local temp_dir="contrib/offline/temp"

    local containerd_download_url="$(grep 'containerd/containerd' "${temp_dir}/files.list")"

    # add containerd static binary
    echo "$(echo "$containerd_download_url" | sed -r 's|containerd-([[:digit:]]\.)|containerd-static-\1|')" >> "${temp_dir}/files.list"
    # add containerd v1.7.x related files
    echo "https://github.com/containerd/nerdctl/releases/download/v1.7.7/nerdctl-1.7.7-linux-${ARCH}.tar.gz" >> "${temp_dir}/files.list"
    echo "https://github.com/containerd/containerd/releases/download/v1.7.23/containerd-1.7.23-linux-${ARCH}.tar.gz" >> "${temp_dir}/files.list"

    # clean up unused images
    local remove_images="aws-alb|aws-ebs|cert-manager|netchecker|weave|sig-storage|external_storage|cinder-csi|kubernetesui"
    mv "${temp_dir}/images.list" "${temp_dir}/images.list.old"
    grep -E -v "${remove_images}" ${temp_dir}/images.list.old > "${temp_dir}/images.list"
    rm -f ${temp_dir}/images.list.old
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

  patch_temp_list

  cp contrib/offline/temp/*.list "${OFFLINE_PACKAGE_DIR}"
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

  local mirror_host=files.m.daocloud.io
  if [[ "${ZONE}" == "CN" ]]; then
    sed -i -r "s#https?://#https://$mirror_host/#g" temp/files.list
  fi
  sed -i "s#storage.googleapis.com/kubernetes-release#dl.k8s.io#g" temp/files.list
  export NO_HTTP_SERVER=true
  if ! bash ./manage-offline-files.sh; then
    echo "Error: manage-offline-files.sh execution failed"
    exit 1
  fi
  if [[ "${ZONE}" == "CN" ]]; then
    mv offline-files/$mirror_host/* offline-files/
    rm -rf offline-files/$mirror_host offline-files.tar.gz
    tar -czf offline-files.tar.gz offline-files
  fi
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

  local images_list_path="${CURRENT_DIR}/kubespray/contrib/offline/temp/images.list"
  local images_list_content

  sed -i '/^$/d' "${images_list_path}"

  if [[ ! -s "${images_list_path}" ]]; then
    echo "images.list is empty"
    return
  fi
  images_list_content=$(cat "${images_list_path}")

  rm -rf offline-images && mkdir offline-images

  while read -r image_name; do
    echo "download image $image_name to local"
    local target_image_name=$image_name
    target_image_name=${target_image_name/docker.m.daocloud.io/docker.io}
    target_image_name=${target_image_name/gcr.m.daocloud.io/gcr.io}
    target_image_name=${target_image_name/ghcr.m.daocloud.io/ghcr.io}
    target_image_name=${target_image_name/k8s-gcr.m.daocloud.io/k8s.gcr.io}
    target_image_name=${target_image_name/k8s.m.daocloud.io/registry.k8s.io}
    target_image_name=${target_image_name/quay.m.daocloud.io/quay.io}
    ret=0
    skopeo copy --insecure-policy --quiet --retry-times=3 --override-os linux --override-arch ${ARCH} "docker://$image_name" "oci:offline-images:$target_image_name" || ret=$?
    if [ ${ret} -ne 0 ]; then
      echo "skopeo copy image failed, image name: ${image_name}."
      exit 1
    fi
    echo "$target_image_name" >> offline-images/images.list
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
