#!/usr/bin/env bash
# 打印镜像内 ansible-core 的主次版本号(如 2.16)，用于拼接镜像 tag。
set -euo pipefail

IMAGE_REF="${1:-}"
if [[ -z "${IMAGE_REF}" ]]; then
  echo "usage: $0 <image-ref>" >&2
  exit 1
fi

VERSION_FILE=/etc/kubean/ansible-core-version
WORK_DIR=$(mktemp -d)
CONTAINER=""
cleanup() {
  [[ -z "${CONTAINER}" ]] || docker rm -f "${CONTAINER}" >/dev/null 2>&1 || true
  rm -rf "${WORK_DIR}"
}
trap cleanup EXIT

# 版本由 requirements.txt 固定，与架构无关，读原生架构即可，无需 QEMU。
# 只从容器文件系统取构建期落盘的版本，不执行镜像内任何代码。
CONTAINER=$(docker create "${IMAGE_REF}")
if ! docker cp "${CONTAINER}:${VERSION_FILE}" "${WORK_DIR}/version" >/dev/null 2>&1; then
  echo "${IMAGE_REF} has no ${VERSION_FILE}; rebuild it with the current kubespray Dockerfile" >&2
  exit 1
fi

CORE_VERSION=$(tr -d '[:space:]' < "${WORK_DIR}/version")

if [[ ! "${CORE_VERSION}" =~ ^[0-9]+\.[0-9]+$ ]]; then
  echo "unexpected ansible-core version from ${IMAGE_REF}: '${CORE_VERSION}'" >&2
  exit 1
fi

echo "${CORE_VERSION}"
