#!/usr/bin/env bash
# 用法: publish-image-tags.sh <repo> <digest> <version> [alias...]
# 把按 digest 推送的镜像发布为 <version>-ansible-core<x.y> 及给定别名，core 版本从镜像内读取。
set -euo pipefail

REPO="${1:-}"
DIGEST="${2:-}"
VERSION="${3:-}"

if [[ -z "${REPO}" || -z "${VERSION}" ]]; then
  echo "usage: $0 <repo> <digest> <version> [alias...]" >&2
  exit 1
fi

if [[ "${DIGEST}" != sha256:* ]]; then
  echo "build did not produce a digest: '${DIGEST}'" >&2
  exit 1
fi

shift 3
IMAGE_REF="${REPO}@${DIGEST}"
CORE_VERSION=$("$(dirname "$0")/ansible-core-version.sh" "${IMAGE_REF}")

TAG_ARGS=(--tag "${REPO}:${VERSION}-ansible-core${CORE_VERSION}")
for ALIAS in "$@"; do
  [[ -z "${ALIAS}" ]] || TAG_ARGS+=(--tag "${REPO}:${ALIAS}")
done

docker buildx imagetools create "${TAG_ARGS[@]}" "${IMAGE_REF}"
