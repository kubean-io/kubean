#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
export registry_addr_amd64=${RUNNER_NODE_IP}:${REGISTRY_PORT_AMD64}

### All AMD64 os resources
export arch=amd64
beginTime=`date +%s`
util::import_files_minio_by_arch ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" ${arch}
util::push_registry_by_arch "${registry_addr_amd64}" "${DOWNLOAD_FOLDER}" ${arch}
util::scope_copy_test_images ${registry_addr_amd64}
endTime=`date +%s`
echo "Import minio and push registry end. Spend time $(($endTime-$beginTime)) s"


### OS ####
### Import os package
./hack/offline_run_centos.sh
