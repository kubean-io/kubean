#!/usr/bin/env bash
IMAGE_REPO=${1:-"ghcr.io/kubean-io"}
RELEASE_NAME=${2:-"kubean"}
TARGET_NS=${3:-"kubean-system"}
KUBECONFIG_PATH=${4:-"kubeconfig"}
IMAGE_TAG=${5:-"latest"}
CHART_VERSION="version: ${6:-$IMAGE_TAG}"

sed -i "/^version/c ${CHART_VERSION}" charts/kubean/Chart.yaml

# helm upgrade
helm upgrade \
    --install \
    ${RELEASE_NAME} \
    -n ${TARGET_NS} \
    --create-namespace \
    --cleanup-on-fail \
    --set kubeanOperator.image.registry=${IMAGE_REPO%/*} \
    --set kubeanOperator.image.repository=${IMAGE_REPO#*/}/kubean-operator \
    --set kubeanOperator.image.tag=${IMAGE_TAG} \
    --kubeconfig kubeconfig \
    charts/kubean/
