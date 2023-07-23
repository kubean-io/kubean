#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail


#####################################
function offline_pre::prepare_kubean_resource_svc(){
  release_tag=${1}
  test_name=${2}
  resource_cluster_prefix=${3}
  kubeconfig_file=${4}

  echo "####################### Offline Resource Build #############################s"
  source "${REPO_ROOT}"/hack/resouce_util.sh
  resource::init_vars_kubean_resource "${release_tag}"

  result=$(resource::check_release_file_integrity  ${release_tag} ${download_root_folder})
  if [[ "${result}"  =~ "YES" ]];then
    echo "Check file integrity ok"
  else
    resource::delete_redundant_folders ${release_tag} ${download_root_folder}
    echo "Start download file..."
    resource::download_resource_files ${release_tag} ${download_root_folder} ${test_name}
  fi
  kind::clean_kind_cluster "${resource_cluster_prefix}"
  resource::create_kind_cluster_by_config_file  "${resource_cluster_prefix}" "${kubeconfig_file}"
  resource::install_minio "${kubeconfig_file}" ${resource_cluster_prefix}
  resource::install_registry "${kubeconfig_file}" "amd64"
  resource::install_registry "${kubeconfig_file}" "arm64"
  resource::import_files_minio_by_arch "${download_root_folder}"  "${release_tag}"  "amd64"  ${test_name}
  resource::import_files_minio_by_arch "${download_root_folder}"  "${release_tag}"  "arm64"  ${test_name}
  resource::push_registry_by_arch "${download_root_folder}"  "${release_tag}" "amd64"  ${test_name}
  resource::push_registry_by_arch "${download_root_folder}"  "${release_tag}" "arm64"  ${test_name}
  resource::import_os_package_minio "${download_root_folder}"  "${release_tag}" ${test_name}
  resource::import_iso_minio ${test_name}

  echo "Resource build end :) "
}

#####################################
function case::artifacts_import_scripts_test(){
  release_tag=${1}
  test_name=${2}
  resource_cluster_prefix=${3}
  kubeconfig_file=${4}

  echo "####################### Artifacts scripts test #############################s"
  source "${REPO_ROOT}"/hack/resouce_util.sh
  resource::init_vars_kubean_resource "${release_tag}"
  ## check local downloaded resource files

  result=$(resource::check_release_file_integrity  ${release_tag} ${download_root_folder}  ${test_name})
  if [[ "${result}"  =~ "YES" ]];then
    echo "Check file integrity ok"
  else
    resource::delete_redundant_folders ${release_tag} ${download_root_folder}
    echo "Start download file..."
   resource::download_resource_files ${release_tag} ${download_root_folder} ${test_name}
  fi
  kind::clean_kind_cluster "${resource_cluster_prefix}"
  resource::create_kind_cluster_by_config_file  "${resource_cluster_prefix}" "${kubeconfig_file}"
  resource::install_minio "${kubeconfig_file}" "${resource_cluster_prefix}"
  resource::install_registry "${kubeconfig_file}"  "amd64"
  resource::import_files_minio_by_arch "${download_root_folder}"  "${release_tag}" "amd64" ${test_name}
  resource::check_files_after_import "${download_root_folder}"  "${release_tag}"
  resource::import_os_package_minio "${download_root_folder}"  "${release_tag}" ${test_name}
  resource::check_os_package_minio "${download_root_folder}"  "${release_tag}"
  resource::push_registry_by_arch "${download_root_folder}"  "${release_tag}" "amd64" ${test_name}
  resource::check_img_registry "${download_root_folder}"  "${release_tag}"
  resource::import_iso_minio ${test_name}
  resource::check_iso_file_minio
  resource::import_iso_local_path_check ${test_name}
  echo "${test_name} scripts test success :) "
  kind::clean_kind_cluster "${resource_cluster_prefix}"
}

#####################################
# shellcheck disable=SC1130
function main(){
  tag=$1
  test_name=$2
  resource_cluster_prefix="kubean-resource"
  download_root_folder="/root/release-files-download"
  kubeconfig_file="/root/.kube/${resource_cluster_prefix}.config"

  if [[ ${test_name} =~ "artifacts" ]];then
    case::artifacts_import_scripts_test "${tag}" "${test_name}" "${resource_cluster_prefix}" "${kubeconfig_file}"
  else
    export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
    offline_pre::prepare_kubean_resource_svc "${tag}" "${test_name}" "${resource_cluster_prefix}" "${kubeconfig_file}"
  fi
}

main $@

