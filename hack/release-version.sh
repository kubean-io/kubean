#!/usr/bin/env bash

set -ex

CUR_DIR=$(
    cd -- "$(dirname "$0")" >/dev/null 2>&1
    pwd -P
)

GITLAB_REPO='gitlab.daocloud.cn/ndx/engineering/kubean.git'


get_version() {
    bash ${CUR_DIR}/get-version.sh $1
}

if [ -z "${PRE_VERSION}" ]; then
    echo you must specify PRE_VERSION var >>/dev/stderr
    exit 1
fi

if [ -z "${NEXT_VERSION}" ]; then
    echo you must specify NEXT_VERSION var >>/dev/stderr
    exit 1
fi

if [ "${PRE_VERSION}" = "$(get_version kubean)" ]; then
    echo PRE_VERSION should not same as kubean in versions.json >>/dev/stderr
    exit 1
fi

if [ "${NEXT_VERSION}" = "$(get_version kubean)" ]; then
    echo NEXT_VERSION should not same as kubean in versions.json >>/dev/stderr
    exit 1
fi

git fetch

if ! git rev-list ${PRE_VERSION} >/dev/null; then
    echo "${PRE_VERSION} tag not exists" >/dev/stderr
    exit 1
fi

if [ -n "${CI_BUILD_REF_NAME}" ]; then
    git checkout ${CI_BUILD_REF_NAME}
fi

CUR_VERSION=$(get_version kubean)

SHORT_VERSION=$(echo ${CUR_VERSION} | awk -F . '{print $1"."$2}')

cd ${CUR_DIR}/../tools/gen-release-notes
mkdir -p ${CUR_DIR}/../changes/${SHORT_VERSION}
go run . --oldRelease ${PRE_VERSION} --newRelease ${CUR_VERSION} --notes ${CUR_DIR}/../ --outDir ${CUR_DIR}/../changes/${SHORT_VERSION}
cd ${CUR_DIR}

f=$(mktemp)
jq ".kubean = \"${NEXT_VERSION}\"" ${CUR_DIR}/../versions.json >$f && cat $f >${CUR_DIR}/../versions.json
rm $f

git add ${CUR_DIR}/..

if ! git config user.name; then
    git config user.name "Auto Release Bot"
    git config user.email "kubean-auto-release@daocloud.io"
fi

git commit -m "Release ${CUR_VERSION} and modify versions.json"
cat ${CUR_DIR}/../changes/${SHORT_VERSION}/${CUR_VERSION}-minorReleaseNotes.md | git tag -a ${CUR_VERSION} -F-


if [ -n "${GITLAB_TOKEN}" ]; then
    git remote set-url origin https://gitlab-ci-token:${GITLAB_TOKEN}@${GITLAB_REPO}
fi

if [ -z "${CI_BUILD_REF_NAME}" ]; then
    git push origin $(git rev-parse --abbrev-ref HEAD)
else
    git push origin ${CI_BUILD_REF_NAME}
fi

git push origin ${CUR_VERSION}

echo "CI_PROJECT_ID: ${CI_PROJECT_ID}"

curl -s -v \
    -H "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
    -H 'Content-Type: application/json' \
    "https://gitlab.daocloud.cn/api/v4/projects/${CI_PROJECT_ID}/releases" \
    -X POST \
    -d "$(echo '{}' | jq \
        --arg name "Release ${CUR_VERSION}" \
        --arg tag_name "${CUR_VERSION}" \
        --arg description "$(cat ${CUR_DIR}/../changes/${SHORT_VERSION}/${CUR_VERSION}-minorReleaseNotes.md)" \
        '.name = $name | .tag_name = $tag_name | .description = $description')"
