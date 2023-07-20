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
  for dir_name in offline-images/*; do
    if skopeo inspect dir:"$dir_name" >/dev/null 2>&1; then
      echo "check $dir_name successfully"
    else
      echo "check $dir_name and $dir_name maybe bad image file"
      exit 1
    fi
  done
}

function import_images() {
  if [ ! -d "offline-images" ]; then
    tar -xvf offline-images.tar.gz
    echo "unzip successfully"
  fi

  check_local_image_files

  for dir_name in offline-images/*; do
    copy_cmd="skopeo copy --insecure-policy -a "
    if [ "$DEST_TLS_VERIFY" = false ]; then
      copy_cmd=$copy_cmd" --dest-tls-verify=false "
    fi
    if [ -n "$DEST_USER" ]; then
      copy_cmd=$copy_cmd" --dest-creds=$DEST_USER:$DEST_PASS "
    fi
    ## dir:offline-images/abc.com%coreos%etcd:v3.5.4 docker://1.2.3.4:5000/abc.com/coreos/etcd:v3.5.4
    image_name=${dir_name#*/}      # remove dir prefix
    image_name=${image_name//%/\/} # replace % with /
    image_name="$REGISTRY_ADDR"/"$image_name"

    echo "import $dir_name to $image_name"

    $copy_cmd --retry-times=3 dir:"$dir_name" docker://"$image_name"
  done

  echo "import completed!"
}

start=$(date +%s)

check_skopeo_cmd
import_images

end=$(date +%s)
take=$((end - start))
echo "Importing images spends ${take} seconds"
