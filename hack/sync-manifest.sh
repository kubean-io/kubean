#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

# This script will sync kubean manifest patch to a single repo.
#
# Usage: hack/sync_manifest.sh

# get envorinment varible.
# github user and email
GITHUB_SERVER="github.com"
GIT_USER_NAME="kubean-robot"
GIT_USER_EMAIL="${GIT_USER_NAME}@kubean.io"
DST_REPO="kubean-manifest"
DST_BRANCH="main"
SUBDIR="manifest"
# git commits
ORIGIN_COMMIT="https://${GITHUB_SERVER}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA}"
COMMIT_MESSAGE="push manifest-$TAG"
# tmp dir for code sync
TMP_DIR="/tmp/kubean-manifest"-$RANDOM
WORK_SPACE="$PWD"
mkdir -p $TMP_DIR

###### Clean Up #######
# clean_up(){
#     rm -rf $TMP_DIR
# }
# trap clean_up

# git clone.

# clone target repo by ssh key
mkdir --parents "$HOME/.ssh"
DEPLOY_KEY_FILE="$HOME/.ssh/deploy_key"
echo "${SSH_DEPLOY_KEY}" > "$DEPLOY_KEY_FILE"
chmod 600 "$DEPLOY_KEY_FILE"

SSH_KNOWN_HOSTS_FILE="$HOME/.ssh/known_hosts"
ssh-keyscan -H "${GITHUB_SERVER}" > "$SSH_KNOWN_HOSTS_FILE"

export GIT_SSH_COMMAND="ssh -i ${DEPLOY_KEY_FILE} -o UserKnownHostsFile=$SSH_KNOWN_HOSTS_FILE"

GIT_CMD_REPOSITORY="git@${GITHUB_SERVER}:${REPO_OWNER}/${DST_REPO}.git"
git clone --single-branch --depth 1 --branch "$DST_BRANCH" "$GIT_CMD_REPOSITORY" "$TMP_DIR"

pushd ~
cd $TMP_DIR

git config --global user.name ${GIT_USER_NAME}
git config --global user.email ${GIT_USER_EMAIL}

cp -f ${WORK_SPACE}/${SUBDIR}/* $TMP_DIR/manifests/

git add -A
if git diff HEAD --quiet; then 
  echo 'nothing is changed, working tree clean' && exit
fi
git commit -m "$COMMIT_MESSAGE"
git push

popd
