#!/bin/bash

# set -x
set -eo pipefail

# Note: 
# This script is used to check whether the contents of the current os package are the same as the contents of the last os package. 
# If the contents are the same, the last system package is directly downloaded without rebuilding. 
# The script finally returns the result to determine whether the os package needs to be rebuilt.

OS_NAME=${OS_NAME:-""}
ORG_NAME=${ORG_NAME:-""}

# Get Latest Git Tag
late_tag=$(git tag --sort=committerdate -l | grep -o 'v.*' | tail -1)
# Get Previous Git Tag (the one before the latest tag)
prev_tag=$(git tag --sort=committerdate -l | grep -o 'v.*' | tail -2 | head -1)

late_packages_yml=$(git show "${late_tag}":build/os-packages/packages.yml)
prev_packages_yml=$(git show "${prev_tag}":build/os-packages/packages.yml)

git diff --quiet "${prev_tag}" "${late_tag}" artifacts/import_ospkgs.sh || { echo "true"; exit; }

if [[ "${OS_NAME}" == "kylin-v10sp2" ]]; then
  git diff --quiet "${prev_tag}" "${late_tag}" build/os-packages/repos/kylin.v10sp2.repo || { echo "true"; exit; }
fi

if [[ "${OS_NAME}" == "kylin-v10sp3" ]]; then
  git diff --quiet "${prev_tag}" "${late_tag}" build/os-packages/repos/kylin.v10sp3.repo || { echo "true"; exit; }
fi

if [[ "${OS_NAME}" == "kylin-v112503" ]]; then
  git diff --quiet "${prev_tag}" "${late_tag}" build/os-packages/repos/kylin.v112503.repo || { echo "true"; exit; }
fi

git diff --quiet "${prev_tag}" "${late_tag}" "build/os-packages/Dockerfile.${OS_NAME}" || { echo "true"; exit; }

# centos / kylinv / redhat / rocky / oracle / tencent / openEuler
if [[ "${OS_NAME}" =~ ^(centos|kylin|redhat|rocky|oracle|tencent|openeuler|bigcloud).* ]]; then
  os_key="${OS_NAME/-v10sp*/}"
  late_digest=$(echo "${late_packages_yml}" | yq eval ".common[],.yum[],.${os_key}[]" | sort | sha1sum | awk '{print $1}')
  prev_digest=$(echo "${prev_packages_yml}" | yq eval ".common[],.yum[],.${os_key}[]" | sort | sha1sum | awk '{print $1}')
  if [ "${late_digest}" == "${prev_digest}" ]; then
    ret=0
    wget -c https://github.com/${ORG_NAME}/kubean/releases/download/${prev_tag}/os-pkgs-${OS_NAME}-${prev_tag}.tar.gz -O os-pkgs-${OS_NAME}-${late_tag}.tar.gz || ret=$?
    if [ ${ret} -eq 0 ]; then
      echo "false" && exit
    fi
  fi
fi

# ubuntu2004 / ubuntu2204 / ubuntu 2404
if [[ "${OS_NAME}" == "ubuntu"* ]]; then
  late_digest=$(echo "${late_packages_yml}" | yq eval ".common[],.apt[],.${OS_NAME}[]" | sort | sha1sum | awk '{print $1}')
  prev_digest=$(echo "${prev_packages_yml}" | yq eval ".common[],.apt[],.${OS_NAME}[]" | sort | sha1sum | awk '{print $1}')
  if [ "${late_digest}" == "${prev_digest}" ]; then
    ret=0
    wget -c https://github.com/${ORG_NAME}/kubean/releases/download/${prev_tag}/os-pkgs-${OS_NAME}-${prev_tag}.tar.gz -O os-pkgs-${OS_NAME}-${late_tag}.tar.gz || ret=$?
    if [ ${ret} -eq 0 ]; then
      echo "false" && exit
    fi
  fi
fi

echo "true"
