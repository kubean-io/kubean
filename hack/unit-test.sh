#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail


#https://www.dazhuanlan.com/erlv/topics/1058820
set -x
PATH2TEST=( ./pkg/...  ./cmd/... )
tmpDir=$(mktemp -d)
mergeF="${tmpDir}/merge.out"
rm -f ${mergeF}
for (( i=0; i<${#PATH2TEST[@]}; i++)) do
    ls $tmpDir
    cov_file="${tmpDir}/$i.cover"
    go test --race --v  -covermode=atomic -coverpkg=${PATH2TEST[i]} -coverprofile=${cov_file}    ${PATH2TEST[i]}
    # TODO: if `cat $cov_file | grep -v mode: | grep -v zz_generated` outputs null
    # the script will abort and exit. we haven't found a good solution and therefore
    # disable the `set -o errexit` to wraping the statement
    cat ${cov_file} >> coverage.txt
    set +o errexit
    bash -c "cat $cov_file | grep -v mode: | grep -v generated | grep -v zz_generated  >> ${mergeF}"
    set -o errexit
done
#merge them
header=$(head -n1 "${tmpDir}/0.cover")
echo "${header}" > coverage.out
cat ${mergeF} >> coverage.out
go tool cover -func=coverage.out
rm -rf coverage.out ${tmpDir}  ${mergeF}
