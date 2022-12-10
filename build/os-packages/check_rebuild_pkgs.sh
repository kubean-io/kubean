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
late_tag=`git tag --sort=committerdate -l | grep -o 'v.*' | tail -1`
# Get Previous Git Tag (the one before the latest tag)
prev_tag=`git describe --abbrev=0 --tags $(git rev-list --tags --skip=1 --max-count=1)`

wget -c https://raw.githubusercontent.com/${ORG_NAME}/kubean/${late_tag}/build/os-packages/packages.yml -O late_packages.yml
wget -c https://raw.githubusercontent.com/${ORG_NAME}/kubean/${prev_tag}/build/os-packages/packages.yml -O prev_packages.yml

# centos7 / kylinv10 / redhat7 / redhat8
if [ "${OS_NAME}" == "centos7" ] || [ "${OS_NAME}" == "kylinv10" ] || [ "${OS_NAME}" == "redhat7" ] || [ "${OS_NAME}" == "redhat8" ]; then
  late_digest=`yq eval ".yum[],.common[],.docker.${OS_NAME}[]" late_packages.yml | sort | sha1sum | awk '{print $1}'`
  prev_digest=`yq eval ".yum[],.common[],.docker.${OS_NAME}[]" prev_packages.yml | sort | sha1sum | awk '{print $1}'`
  if [ ${late_digest} == ${prev_digest} ]; then
    ret=$(wget -c https://github.com/${ORG_NAME}/kubean/releases/download/${prev_tag}/os-pkgs-${OS_NAME}-${prev_tag}.tar.gz -O os-pkgs-${OS_NAME}-${late_tag}.tar.gz)
    if [ $? -eq 0 ]; then
      echo "false" && exit 0
    fi
  fi
fi

# ubuntu1804 / ubuntu2004
if [ "${OS_NAME}" == "ubuntu1804" ] || [ "${OS_NAME}" == "ubuntu2004" ]; then
  late_digest=`yq eval ".common[],.docker.${OS_NAME}[]" late_packages.yml | sort | sha1sum | awk '{print $1}'`
  prev_digest=`yq eval ".common[],.docker.${OS_NAME}[]" prev_packages.yml | sort | sha1sum | awk '{print $1}'`
  if [ ${late_digest} == ${prev_digest} ]; then
    ret=$(wget -c https://github.com/${ORG_NAME}/kubean/releases/download/${prev_tag}/os-pkgs-${OS_NAME}-${prev_tag}.tar.gz -O os-pkgs-${OS_NAME}-${late_tag}.tar.gz)
    if [ $? -eq 0 ]; then
      echo "false" && exit 0
    fi
  fi
fi

echo "true"
