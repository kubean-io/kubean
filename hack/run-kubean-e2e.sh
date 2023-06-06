#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -x


function case::artifacts_scripts_import(){
  tag=${1}
  echo "####################### Artifacts scripts test #############################s"
  source "${REPO_ROOT}"/hack/resouce_util.sh
  resource::init_vars_kubean_resource "${tag}"
  ## check local downloaded resource files

  result=$(resource::check_release_file_integrity  ${NEW_TAG} ${DOWNLOAD_ROOT_FOLDER}  "artifacts_test")
  if [[ "${result}"  =~ "YES" ]];then
    echo "Check file integrity ok"
  else
    resource::delete_redundant_folders ${NEW_TAG} ${DOWNLOAD_ROOT_FOLDER}
    echo "Start download file..."
    resource::init_download_folder ${NEW_TAG} ${DOWNLOAD_ROOT_FOLDER}
    resource::download_resource_files ${NEW_TAG} ${DOWNLOAD_ROOT_FOLDER} "artifacts_test"
  fi
  resource::clean_kind_cluster "${RESOURCE_CLUSTER_PREFIX}"
  resource::create_kind_cluster_by_config_file  "${RESOURCE_CLUSTER_PREFIX}" "${RESOURCE_KUBECONFIG_FILE}"
  resource::install_minio "${RESOURCE_KUBECONFIG_FILE}"
  resource::install_registry "AMD64" "${RESOURCE_KUBECONFIG_FILE}" "registry-amd64"
  resource::import_files_minio_by_arch "${DOWNLOAD_ROOT_FOLDER}"  "${NEW_TAG}" "amd64"
  resource::check_files_after_import "${DOWNLOAD_ROOT_FOLDER}"  "${NEW_TAG}"
  resource::import_os_package_minio "${DOWNLOAD_ROOT_FOLDER}"  "${NEW_TAG}" "artifacts_test"
  resource::check_os_package_minio "${DOWNLOAD_ROOT_FOLDER}"  "${NEW_TAG}"
  resource::import_iso_minio "artifacts_test"
  resource::check_iso_file_minio
  resource::import_iso_local_path_check "artifacts_test"
  echo "Artifacts scripts test success :) "
  #= resource::clean_kind_cluster "${RESOURCE_CLUSTER_PREFIX}"
}

# shellcheck disable=SC1130
case::artifacts_scripts_import "${TARGET_VERSION}"