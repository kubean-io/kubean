#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
export registry_addr_amd64=${RUNNER_NODE_IP}:${REGISTRY_PORT_AMD64}

### All AMD64 os resources###
export arch=amd64
beginTime=`date +%s`
util::import_files_minio_by_arch ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" ${arch}
util::push_registry_by_arch "${registry_addr_amd64}" "${DOWNLOAD_FOLDER}" ${arch}
util::scope_copy_test_images ${registry_addr_amd64}
endTime=`date +%s`
echo "Import minio and push registry end. Spend time $(($endTime-$beginTime)) s"

### 1. OS compatibility ####
shell_path="${REPO_ROOT}/artifacts"
iso_image_file="/root/iso-images/rhel-server-7.9-x86_64-dvd.iso"
util::import_iso  ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${shell_path}" ${iso_image_file}
util::import_os_package_minio ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" "redhat7"
iso_image_file="/root/iso-images/rhel-8.4-x86_64-dvd.iso"
util::import_iso  ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${shell_path}" ${iso_image_file}
util::import_os_package_minio ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" "redhat8"

./hack/run-os-compatibility-e2e.sh

### 2. CASE on centos
./hack/offline_run_centos.sh
