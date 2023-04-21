#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CONTROLLER_GEN_PKG="sigs.k8s.io/controller-tools/cmd/controller-gen"
CONTROLLER_GEN_VER="v0.6.2"

source hack/util.sh

echo "Generating with controller-gen"
util::install_tools ${CONTROLLER_GEN_PKG} ${CONTROLLER_GEN_VER} >/dev/null 2>&1

# Unify the crds used by helm chart and the installation scripts
controller-gen crd paths=./apis/... output:crd:dir=./charts/crds

for f in ./charts/crds/* ; do
  ## f: "./charts/crds/kubean.io_clusteroperations.yaml"
  sed '/^[[:blank:]]*$/d' "$f" > "$f.tmp" ## remove blank
  mv "$f.tmp"  "$f"
done

cp -r charts/crds ../charts/kubean
