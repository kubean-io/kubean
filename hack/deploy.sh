#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# This script plays as a reference script to install kubean helm release
#
# Usage: bash hack/deploy.sh v0.0.1-8-gee28ca5 latest  remote_kube_cluster.conf
#        Parameter 1: helm release version
#        Parameter 2: container image tag override in helm values.yaml
#        Parameter 3: the kube.conf of the target cluster to be installed
#        Parameter 4: the namespace which kubean being installed
#        Parameter 5: the helm chart server (harbor) and project (e.x: https://release.daocloud.io/chartrepo/kubean)
#        Parameter 6: the container image repo to be override in helm values.yaml

# specific a helm package version
HELM_VER=${1:-"v0.0.1"}
IMG_VER=${2:-$HELM_VER} # by default, $IMG_VER is the same with $HELM_VER
KUBE_CONF=${3:-"/root/.kube/config"}
TARGET_NS=${4:-"kubean-system"}
HELM_REPO=${5:-"https://release.daocloud.io/chartrepo/kubean"}
IMG_REPO=${6:-} #default using what inside helm chart
DEPLOY_ENV=${7:-}   # E2E/DEV/PROD

LOCAL_REPO_ALIAS=kubean_release
LOCAL_RELEASE_NAME=kubean

# replace the default values.yaml, the image repo or image revision
value_override=""
if [ "${IMG_REPO}" != "" ]; then
    value_override=" $value_override --set image.repository=${IMG_REPO}/kubean-operator "
fi
if [ "${IMG_VER}" != "" ]; then
    value_override=" $value_override --set image.tag=${IMG_VER} "
fi

#v0.1.1 --> 0.1.1 Match the helm chart version specification, remove the preceding prefix `v` character
# KUBEAN_CHART_VERSION="$(echo "${HELM_VER}" |sed  's/^v//g' )"
KUBEAN_CHART_VERSION=${HELM_VER}

#ensure kube.conf without group-readable
chmod 600 ${KUBE_CONF}
# install or upgrade
helm upgrade --install  --create-namespace --cleanup-on-fail \
             ${LOCAL_RELEASE_NAME}     ${LOCAL_REPO_ALIAS}/kubean   \
             ${value_override} \
             -n "${TARGET_NS}"  --version ${KUBEAN_CHART_VERSION} \
             --kubeconfig ${KUBE_CONF}

# check it
helm list -n "${TARGET_NS}" --kubeconfig ${KUBE_CONF}
