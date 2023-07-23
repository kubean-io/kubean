#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

IMAGE_REGISTRY=${1:-"ghcr.m.daocloud.io"}
RELEASE_NAME=${2:-"kubean"}
TARGET_NS=${3:-"kubean-system"}
KUBECONFIG_PATH=${4:-"kubeconfig"}
IMAGE_TAG=${5:-"latest"}
CHART_VERSION="version: ${6:-${IMAGE_TAG}}"

latest_status="$(helm hist -n kubean-system kubean --kubeconfig "${KUBECONFIG_PATH}" 2>&1 | awk '{print $7}' | tail -n 1)"
if [[ "${latest_status}" == "pending-upgrade" ]]; then
    helm rollback -n kubean-system kubean
fi

sed -i "/^version/c ${CHART_VERSION}" charts/kubean/Chart.yaml
# helm upgrade
helm upgrade \
    --burst-limit=250 \
    --install \
    "${RELEASE_NAME}" \
    -n "${TARGET_NS}" \
    --create-namespace \
    --cleanup-on-fail \
    --set kubeanOperator.image.registry="${IMAGE_REGISTRY}" \
    --set sprayJob.image.registry="${IMAGE_REGISTRY}" \
    --set kubeanOperator.image.tag="${IMAGE_TAG}" \
    --set sprayJob.image.tag="${IMAGE_TAG}" \
    --kubeconfig "${KUBECONFIG_PATH}" \
    --timeout 1200s \
    charts/kubean/
