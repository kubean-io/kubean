#!/bin/bash

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
  cp contrib/offline/temp/*.list $OFFLINE_PACKAGE_DIR
}

function create_files() {
  cd $CURRENT_DIR/kubespray/contrib/offline/
  NO_HTTP_SERVER=true bash manage-offline-files.sh
  cp offline-files.tar.gz $OFFLINE_FILES_DIR
}

function create_images() {
  cd $CURRENT_DIR/artifacts

  if which skopeo; then
    echo "skopeo check successfully"
  else
    echo "please install skopeo first"
    exit 1
  fi

  IMG_LIST=$CURRENT_DIR/kubespray/contrib/offline/temp/images.list

  echo "begin to download images"
  images_list_content=$(cat "$IMG_LIST")

  if [ ! -d "offline-images" ]; then
    echo "create dir offline-images"
    mkdir offline-images
  fi

  while read -r image_name; do
    ## quay.io/metallb/controller:v0.12.1 => dir:somedir/metallb%controller:v0.12.1
    ## quay.io/metallb/controller:v0.12.1 => dir:somedir/quay.io%metallb%controller:v0.12.1 ## keep host with multi harbor projects

    ## new_dir_name=${image_name#*/}     ## remote host
    new_dir_name=${image_name} ## keep host
    new_dir_name=${new_dir_name//\//%} ## replace all / with %
    echo "download image $(replace_image_name "$image_name") to local $new_dir_name"
    ret=0
    skopeo copy --insecure-policy --retry-times=3 --override-os linux --override-arch ${ARCH} docker://"$(replace_image_name "$image_name")" dir:offline-images/"$new_dir_name" || ret=$?
    if [ ${ret} -ne 0 ]; then
      echo "skopeo copy image failed, image name: ${image_name}."
      exit 1
    fi
  done <<< "$images_list_content"

  tar -czvf $OFFLINE_IMAGES_DIR/offline-images.tar.gz offline-images

  echo "zipping images completed!"
}

function replace_image_name() {
  local origin_address=$1

  if [ "$ZONE" != "CN" ]; then
    echo "$origin_address" 
    return
  fi

  origin_address=${origin_address/docker.io/docker.m.daocloud.io}
  origin_address=${origin_address/gcr.io/gcr.m.daocloud.io} 
  origin_address=${origin_address/ghcr.io/ghcr.m.daocloud.io}
  origin_address=${origin_address/k8s.gcr.io/k8s-gcr.m.daocloud.io}
  origin_address=${origin_address/registry.k8s.io/k8s.m.daocloud.io}
  origin_address=${origin_address/quay.io/quay.m.daocloud.io}
  echo "$origin_address"
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
