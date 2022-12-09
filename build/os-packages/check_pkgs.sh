#!/bin/bash

set -eo pipefail

# Get Latest Git Tag
late_tag=`git tag --sort=committerdate -l | grep -o 'v.*' | sort -r | head -1`
# Get Previous Git Tag (the one before the latest tag)
prev_tag=`git describe --abbrev=0 --tags $(git rev-list --tags --skip=1 --max-count=1)`

wget -c https://raw.githubusercontent.com/kubean-io/kubean/${late_tag}/build/os-packages/packages.yml -O late_packages.yml
wget -c https://raw.githubusercontent.com/kubean-io/kubean/${prev_tag}/build/os-packages/packages.yml -O prev_packages.yml

# centos7
late_digest=`yq eval '.yum[],.common[],.docker.centos7[]' late_packages.yml | sort | md5`
prev_digest=`yq eval '.yum[],.common[],.docker.centos7[]' prev_packages.yml | sort | md5`
if [ ${late_digest} == ${prev_digest} ]; then
  wget -c https://github.com/kubean-io/kubean/releases/download/${prev_tag}/os-pkgs-centos7-${prev_tag}.tar.gz -O os-pkgs-centos7-${late_tag}.tar.gz  
fi

# kylinv10
late_digest=`yq eval '.yum[],.common[],.docker.kylinv10[]' late_packages.yml | sort | md5`
prev_digest=`yq eval '.yum[],.common[],.docker.kylinv10[]' prev_packages.yml | sort | md5`
if [ ${late_digest} == ${prev_digest} ]; then
  wget -c https://github.com/kubean-io/kubean/releases/download/${prev_tag}/os-pkgs-kylinv10-${prev_tag}.tar.gz -O os-pkgs-kylinv10-${late_tag}.tar.gz  
fi

# redhat7
late_digest=`yq eval '.yum[],.common[],.docker.redhat7[]' late_packages.yml | sort | md5`
prev_digest=`yq eval '.yum[],.common[],.docker.redhat7[]' prev_packages.yml | sort | md5`
if [ ${late_digest} == ${prev_digest} ]; then
  wget -c https://github.com/kubean-io/kubean/releases/download/${prev_tag}/os-pkgs-redhat7-${prev_tag}.tar.gz -O os-pkgs-redhat7-${late_tag}.tar.gz  
fi

# redhat8
late_digest=`yq eval '.yum[],.common[],.docker.redhat8[]' late_packages.yml | sort | md5`
prev_digest=`yq eval '.yum[],.common[],.docker.redhat8[]' prev_packages.yml | sort | md5`
if [ ${late_digest} == ${prev_digest} ]; then
  wget -c https://github.com/kubean-io/kubean/releases/download/${prev_tag}/os-pkgs-redhat8-${prev_tag}.tar.gz -O os-pkgs-redhat8-${late_tag}.tar.gz  
fi

# ubuntu1804
late_digest=`yq eval '.common[],.docker.ubuntu1804[]' late_packages.yml | sort | md5`
prev_digest=`yq eval '.common[],.docker.ubuntu1804[]' prev_packages.yml | sort | md5`
if [ ${late_digest} == ${prev_digest} ]; then
  wget -c https://github.com/kubean-io/kubean/releases/download/${prev_tag}/os-pkgs-ubuntu1804-${prev_tag}.tar.gz -O os-pkgs-ubuntu1804-${late_tag}.tar.gz  
fi

# ubuntu2004
late_digest=`yq eval '.common[],.docker.ubuntu2004[]' late_packages.yml | sort | md5`
prev_digest=`yq eval '.common[],.docker.ubuntu2004[]' prev_packages.yml | sort | md5`
if [ ${late_digest} == ${prev_digest} ]; then
  wget -c https://github.com/kubean-io/kubean/releases/download/${prev_tag}/os-pkgs-ubuntu2004-${prev_tag}.tar.gz -O os-pkgs-ubuntu2004-${late_tag}.tar.gz  
fi
