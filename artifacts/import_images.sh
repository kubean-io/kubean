#!/bin/bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail

REGISTRY_ADDR=${1}

function check_skopeo_cmd() {
  if which skopeo; then
    echo "skopeo check successfully"
  else
    echo "please install skopeo first"
    exit 1
  fi
}

function check_local_image_files() {
  ## empty dir?
  if [ $(ls -A offline-images | wc -l) -eq 0 ]; then
    return
  fi
  while read -r image_name; do
    if skopeo inspect "oci:offline-images:$image_name" >/dev/null 2>&1; then
      echo "check $image_name successfully"
    else
      echo "$image_name maybe bad image file"
      exit 1
    fi
  done < offline-images/images.list
}

function import_images() {
  if [ ! -d "offline-images" ]; then
    tar -xvf offline-images.tar.gz
    echo "unzip successfully"
  fi

  check_local_image_files

  while read -r image_name; do
    copy_cmd="skopeo copy --insecure-policy -a "
    if [ "$DEST_TLS_VERIFY" = false ]; then
      copy_cmd=$copy_cmd" --dest-tls-verify=false "
    fi
    if [ -n "$DEST_USER" ]; then
      copy_cmd=$copy_cmd" --dest-creds=$DEST_USER:$DEST_PASS "
    fi
    target_image_name="$REGISTRY_ADDR"/"$image_name"

    echo "import $image_name to $target_image_name"

    $copy_cmd --retry-times=3 "oci:offline-images:$image_name" "docker://$target_image_name"
  done < offline-images/images.list

  echo "import completed!"
}

start=$(date +%s)

check_skopeo_cmd
import_images

end=$(date +%s)
take=$((end - start))
echo "Importing images spends ${take} seconds"
