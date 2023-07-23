#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

# This script will sync kubean api declaration to a single repo.
#
# Usage: hack/sync_api.sh

# get envorinment varible.
# kubean tag(will be synced to kubean-api)
KUBEAN_TAG="${1}"
# SSH_DEPLOY_KEY="secrets.SSH_DEPLOY_KEY"
# github user and email
GITHUB_SERVER="github.com"
GIT_USER_NAME="kubean-robot"
GIT_USER_EMAIL="${GIT_USER_NAME}@kubean.io"
# the repository owner
# REPO_OWNER="kubean-io"
# the target repo
DST_REPO="kubean-api"
# dst branch of kubean-io/kubean-api
DST_BRANCH="main"
# maps to staging/k8s.io/src/${REPO}
SUBDIR="api"
# git commits
ORIGIN_COMMIT="https://${GITHUB_SERVER}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA}"
COMMIT_MESSAGE="See ${ORIGIN_COMMIT} from ${GITHUB_REF}"
# tmp dir for code sync
TMP_DIR="/tmp/kubean-api"-$RANDOM
WORK_SPACE="$PWD"
mkdir -p $TMP_DIR

###### Clean Up #######
# clean_up(){
#     rm -rf $TMP_DIR
# }
# trap clean_up

# git clone.

# This function installs a Go tool by 'go get' command.
# Parameters:
#  - $1: tag name
#  - $2: git comments

function create_git_tag_if_needed() {
	local tag="$1"
	local comments="$2"

  if [ $(git tag -l "$tag") ]; then
      echo "tag already exist~"
  else
      git tag $KUBEAN_TAG -a -m "$comments"
      git push origin $KUBEAN_TAG
  fi
}

# clone target repo by ssh key
mkdir --parents "$HOME/.ssh"
DEPLOY_KEY_FILE="$HOME/.ssh/deploy_key"
echo "${SSH_DEPLOY_KEY}" > "$DEPLOY_KEY_FILE"
chmod 600 "$DEPLOY_KEY_FILE"

SSH_KNOWN_HOSTS_FILE="$HOME/.ssh/known_hosts"
ssh-keyscan -H "${GITHUB_SERVER}" > "$SSH_KNOWN_HOSTS_FILE"

export GIT_SSH_COMMAND="ssh -i "$DEPLOY_KEY_FILE" -o UserKnownHostsFile=$SSH_KNOWN_HOSTS_FILE"

GIT_CMD_REPOSITORY="git@${GITHUB_SERVER}:${REPO_OWNER}/${DST_REPO}.git"
git clone --single-branch --depth 1 --branch "$DST_BRANCH" "$GIT_CMD_REPOSITORY" "$TMP_DIR"

pushd ~
cd $TMP_DIR

git config --global user.name ${GIT_USER_NAME}
git config --global user.email ${GIT_USER_EMAIL}

# Check if the api has been updated
ret=0
diff -Naupr -x ".git" $WORK_SPACE/$SUBDIR $TMP_DIR || ret=$?
if [[ $ret -eq 0 ]]
then
  echo "api is already up to date, ignore changes."
  create_git_tag_if_needed $KUBEAN_TAG $COMMIT_MESSAGE
  exit 0
fi

echo "The api definition has been updated, and the corresponding repo of the api will be updated."

# sync code(1. delete api repo code ;2. copy kubean api declaration ).

git rm -r $TMP_DIR

cp -r ${WORK_SPACE}/${SUBDIR}/* $TMP_DIR

# commit(keep it consistent with the main library commit).
git add -A
git commit -m "$COMMIT_MESSAGE"

# git tag(if needed).
# git push

git push

create_git_tag_if_needed $KUBEAN_TAG $COMMIT_MESSAGE

popd
