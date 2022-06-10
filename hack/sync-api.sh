#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

# This script will sync kubean api declaration to single repo.
#
# Usage: hack/sync-api.sh

# get envorinment varible.

# source repository name (eg. kubernetes) has to be set for the sync-tags
SOURCE_REPO="https://gitlab.daocloud.cn/ndx/engineering/kubean.git"
# src branch of k8s.io/kubernetes
SRC_BRANCH="main"
# the target repo
DST_REPO="https://gitlab.daocloud.cn/ndx/kubean-api.git"
# dst branch of k8s.io/${repo}
DST_BRANCH="main"
# maps to staging/k8s.io/src/${REPO}
SUBDIR="api"
# kubean tag(will be synced to kubean-api)
KUBEAN_TAG="${1}"
# git commits
GIT_COMMITS="sync api code from kubean"
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

# This function installs a Go tools by 'go get' command.
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

git clone -b $DST_BRANCH "https://gitlab-ci-token:${GITLAB_TOKEN}@gitlab.daocloud.cn/ndx/kubean-api.git" $TMP_DIR

pushd ~
cd $TMP_DIR

git config --global user.email "shihang.yun@daocloud.io"
git config --global user.name shihang.yun


# Check if the api has been updated
ret=0
diff -Naupr -x ".git" $WORK_SPACE/$SUBDIR $TMP_DIR || ret=$?
if [[ $ret -eq 0 ]]
then
  echo "api is already up to date, ignore changes."
  create_git_tag_if_needed $KUBEAN_TAG $GIT_COMMITS
  exit 0
fi

echo "The api definition has been updated, and the corresponding repo of the api will be updated."

# sync code(1. delete api repo code ;2. copy kubean api declaration ).

git rm -r $TMP_DIR

cp -r ${WORK_SPACE}/${SUBDIR}/* $TMP_DIR

# commit(keep it consistent with the main library commit).
git add -A
git commit -m "$GIT_COMMITS"

# git tag(if needed).
# git push

git push

create_git_tag_if_needed $KUBEAN_TAG $GIT_COMMITS

popd
