#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

### Download kubean offline files
### This step only down
export TAG=$1
export DOWNLOAD_FOLDER="/root/download_offline_files"
export DOWNLOAD_FOLDER_TAG="/root/download_offline_files/${TAG}"
export base_url=https://files.m.daocloud.io/github.com/kubean-io/kubean/releases/download/${TAG}
rm -fr "${DOWNLOAD_FOLDER}"
mkdir -p "${DOWNLOAD_FOLDER_TAG}"

# arch pkgs
f_files_amd64_tgz=${base_url}/files-amd64-${TAG}.tar.gz
f_images_amd64_tgz=${base_url}/images-amd64-${TAG}.tar.gz
f_files_arm64__tgz=${base_url}/files-arm64-${TAG}.tar.gz
f_images_arm64_tgz=${base_url}/images-arm64-${TAG}.tar.gz
# os pkgs
f_os_centos7=${base_url}/os-pkgs-centos7-${TAG}.tar.gz
f_os_kylin10=${base_url}/os-pkgs-kylinv10-${TAG}.tar.gz
f_os_redhat8=${base_url}/os-pkgs-redhat8-${TAG}.tar.gz
f_os_redhat7=${base_url}/os-pkgs-redhat7-${TAG}.tar.gz
# shellcheck disable=SC2206
file_down_list=(${f_files_amd64_tgz}  ${f_images_amd64_tgz} ${f_files_arm64__tgz} ${f_images_arm64_tgz} \
                ${f_os_centos7} ${f_os_kylin10} ${f_os_redhat8} ${f_os_redhat7})
for (( i=0; i<${#file_down_list[@]};i++)); do
  echo "${file_down_list[$i]}"
  timeout 1m wget -q -c  -P  "${DOWNLOAD_FOLDER}"  "${file_down_list[$i]}"
done

rm -fr "${DOWNLOAD_FOLDER}"
