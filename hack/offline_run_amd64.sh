#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
export registry_addr_amd64=${RUNNER_NODE_IP}:${REGISTRY_PORT_AMD64}

### All AMD64 os resources###
export ARCH="amd64"
export OS_NAME="CENTOS7"
beginTime=`date +%s`
util::import_files_minio_by_arch ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" ${ARCH}
util::push_registry_by_arch "${registry_addr_amd64}" "${DOWNLOAD_FOLDER}" ${ARCH}
util::scope_copy_test_images ${registry_addr_amd64}
endTime=`date +%s`
echo "Import minio and push registry end. Spend time $(($endTime-$beginTime)) s"

### Import iso and os packages ####
shell_path="${REPO_ROOT}/artifacts"
iso_image_file="/root/iso-images/rhel-server-7.9-x86_64-dvd.iso"
util::import_iso  ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${shell_path}" ${iso_image_file}
util::import_os_package_minio ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" "redhat7"
iso_image_file="/root/iso-images/rhel-8.4-x86_64-dvd.iso"
util::import_iso  ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${shell_path}" ${iso_image_file}
util::import_os_package_minio ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" "redhat8"
iso_image_file="/root/iso-images/CentOS-7-x86_64-DVD-2207-02.iso"
util::import_iso  ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${shell_path}" ${iso_image_file}
util::import_os_package_minio ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" "centos7"

./hack/run-network-e2e.sh
./hack/offline_run_centos.sh
./hack/run-os-compatibility-e2e.sh
