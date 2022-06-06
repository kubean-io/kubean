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
CD_TO_ENVIRONMENT=${8:-}   # cd

LOCAL_REPO_ALIAS=kubean_release
LOCAL_RELEASE_NAME=kubean
values=""
# add repo locally
helm repo add ${LOCAL_REPO_ALIAS} ${HELM_REPO}

if [ "${DEPLOY_ENV}" == "PROD" ];then
    values="-f config/ci/demo-alpha.yaml"
elif [ "${DEPLOY_ENV}" == "DEV" ];then
    values="-f config/ci/demo-alpha.yaml"
elif [ "${DEPLOY_ENV}" == "E2E" ];then
    values="-f config/ci/e2e.yaml"
else
    values=""
fi

# replace the default values.yaml, the image repo or image revision
value_override=""
if [ "${IMG_REPO}" != "" ]; then
    value_override=" $value_override --set image.repository=${IMG_REPO}/kubean-operator "
fi
if [ "${IMG_VER}" != "" ]; then
    value_override=" $value_override --set image.tag=${IMG_VER} "
fi

#v0.1.1 --> 0.1.1 Match the helm chart version specification, remove the preceding prefix `v` character
KUBEAN_CHART_VERSION="$(echo "${HELM_VER}" |sed  's/^v//g' )"

#ensure kube.conf without group-readable
chmod 600 ${KUBE_CONF}

#################################argocd：加到helm upgrade上面###############################

if [[ $CD_TO_ENVIRONMENT == '' ]];
then
    echo "CD_TO_ENVIRONMENT is empty"
    exit 2
fi

git config --global user.name "${GITLAB_USER_NAME}"
git config --global user.email "${GITLAB_USER_EMAIL}"
# 为gitops的git仓库创建临时目录
TEMP_REPO_DIR="temp"
mkdir -p $TEMP_REPO_DIR
git clone https://gitlab-ci-token:${GITLAB_TOKEN}@gitlab.daocloud.cn/ndx/cicd-infrastructure.git $TEMP_REPO_DIR
# 为新项目创建项目目录
mkdir -p ${TEMP_REPO_DIR}/argocd/${CD_TO_ENVIRONMENT}/${CI_PROJECT_NAME}
#################################argocd：加到helm upgrade上面###############################

# install or upgrade
helm upgrade --install  --create-namespace --cleanup-on-fail \
             ${LOCAL_RELEASE_NAME}     ${LOCAL_REPO_ALIAS}/kubean   \
             ${values} ${value_override} \
             -n "${TARGET_NS}"  --version ${KUBEAN_CHART_VERSION} \
             --kubeconfig ${KUBE_CONF} --dry-run --debug > /helm.yaml

#################################argocd：加到helm upgrade下面###############################

set +e
YAML_FILE="/helm.yaml"
# 拿到yaml第一行 （# 这里由于环境问题，设置了set -e可能误报，但是能正常过滤出yaml文件）
BEGIN_LINE=`grep -n "\-\-\-" ${YAML_FILE} |head -1|awk -F ":" '{print $1}'`
let BEGIN_LINE+=1
# 拿到yaml最后一行
END_LINE=`grep -n "NOTES:" ${YAML_FILE} |head -1|awk -F ":" '{print $1}'`
let END_LINE-=1
set -e

# 把yaml文件过滤出来放到git仓库的项目目录下
sed -n "${BEGIN_LINE},${END_LINE}p" $YAML_FILE > $TEMP_REPO_DIR/argocd/${CD_TO_ENVIRONMENT}/${CI_PROJECT_NAME}/${CI_PROJECT_NAME}.yaml
# 推送代码提交代码到git仓库
cd $TEMP_REPO_DIR
git add -A
git commit -m "Update ${CI_PROJECT_NAME}"
git push https://gitlab-ci-token:${GITLAB_TOKEN}@gitlab.daocloud.cn/ndx/cicd-infrastructure.git HEAD:master
#################################argocd：加到helm upgrade下面###############################
