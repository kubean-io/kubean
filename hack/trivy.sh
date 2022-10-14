#!/bin/bash
# set -o errexit
set -o nounset
set -o pipefail

# checkout command exist
trivy -v
if [[ $? != 0 ]];then
  wget https://github.com/aquasecurity/trivy/releases/download/v0.29.0/trivy_0.29.0_Linux-64bit.tar.gz
  tar xf trivy_0.29.0_Linux-64bit.tar.gz
  mv trivy /usr/local/bin/trivy
fi

for i in $*
do
  trivy image --ignore-unfixed --exit-code 0 --severity MEDIUM,HIGH,CRITICAL ${i}
done
