#!/usr/bin/env bash

# Copyright 2026 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
TEST_TMP=$(mktemp -d)
trap 'rm -rf "$TEST_TMP"' EXIT
export CLI_LOG="$TEST_TMP/cli.log"
mkdir -p "$TEST_TMP/bin"
export PATH="$TEST_TMP/bin:$PATH"

# Only external commands are replaced; the functions under test stay unchanged.
cat > "$TEST_TMP/bin/record" <<'STUB'
#!/usr/bin/env bash
name=${0##*/}
printf '%s' "$name" >> "$CLI_LOG"
printf '<%s>' "$@" >> "$CLI_LOG"
printf '\n' >> "$CLI_LOG"
if [[ "$name" == mc && "$1" == ls ]]; then
  exit "${MC_BUCKET_EXISTS:-1}"
fi
STUB
cat > "$TEST_TMP/bin/flock" <<'STUB'
#!/usr/bin/env bash
shift
exec "$@"
STUB
chmod +x "$TEST_TMP/bin/record" "$TEST_TMP/bin/flock"
for cli in mc helm kubectl; do
  ln -s record "$TEST_TMP/bin/$cli"
done

load_function() {
  local file=$1 name=$2
  eval "$(awk -v name="$name" '
    index($0, "function " name "(") == 1 { found=1; active=1 }
    active { print }
    active && /^}/ { active=0 }
    END { if (!found) exit 1 }
  ' "$REPO_ROOT/$file")"
}

expect_line() {
  if ! grep -Fxq -- "$1" "$CLI_LOG"; then
    printf 'FAIL: missing command: %s\n' "$1" >&2
    cat "$CLI_LOG" >&2
    exit 1
  fi
}

# Retired config-host syntax or a lost anonymous-download grant must fail.
for kind in files ospkgs iso; do
  (
    : > "$CLI_LOG"
    MINIO_API_ADDR=http://127.0.0.1:9000
    MINIO_USER='test user'
    MINIO_PASS='test password'
    PARALLEL_LOCK="$TEST_TMP/import.lock"
    prefix=''
    remove=del_mc_host_conf
    [[ "$kind" != ospkgs ]] || remove=remove_mc_host_conf
    [[ "$kind" != iso ]] || prefix=iso::
    file="artifacts/import_${kind}.sh"
    load_function "$file" "${prefix}add_mc_host_conf"
    load_function "$file" "${prefix}ensure_kubean_bucket"
    load_function "$file" "${prefix}${remove}"
    "${prefix}add_mc_host_conf" "$MINIO_API_ADDR" "$MINIO_USER" "$MINIO_PASS"
    "${prefix}ensure_kubean_bucket"
    "${prefix}${remove}" "$PARALLEL_LOCK"
    expect_line 'mc<alias><set><kubeaniominioserver><http://127.0.0.1:9000><test user><test password>'
    expect_line 'mc<mb><--ignore-existing><kubeaniominioserver/kubean>'
    expect_line 'mc<anonymous><set><download><kubeaniominioserver/kubean>'
    expect_line 'mc<alias><remove><kubeaniominioserver>'
    : > "$CLI_LOG"
    export MC_BUCKET_EXISTS=0
    "${prefix}ensure_kubean_bucket"
    [[ $(wc -l < "$CLI_LOG") -eq 1 ]]
    expect_line 'mc<ls><kubeaniominioserver/kubean>'
    printf 'PASS: import_%s commands\n' "$kind"
  )
done

# Capture the real Helm invocation without sourcing scripts with top-level work.
# 提取的真实函数会间接调用这些 stub。
# shellcheck disable=SC2329
for target in resource demo; do
  (
    : > "$CLI_LOG"
    MINIO_USER='test-user'
    MINIO_PASS='test-password'
    resource::create_ns() { :; }
    resource::create_pvc() { :; }
    create_pvc() { :; }
    if [[ "$target" == resource ]]; then
      load_function hack/resouce_util.sh resource::install_minio
      resource::install_minio "$TEST_TMP/kubeconfig" test-cluster
    else
      load_function examples/debug/airgap_demo.sh install_minio_and_registry
      install_minio_and_registry
    fi
    helm_line=$(grep '^helm<upgrade>.*<minio><daocloud-community/minio>' "$CLI_LOG")
    for arg in \
      '<--version=5.4.0>' \
      '<mode=standalone>' \
      '<deploymentUpdate.type=Recreate>' \
      '<image.repository=m.daocloud.io/docker.io/pgsty/minio>' \
      '<image.tag=RELEASE.2026-06-18T00-00-00Z>' \
      '<mcImage.repository=m.daocloud.io/docker.io/pgsty/minio>' \
      '<mcImage.tag=RELEASE.2026-06-18T00-00-00Z>'; do
      if [[ "$helm_line" != *"$arg"* ]]; then
        printf 'FAIL: %s Helm command missing %s\n%s\n' "$target" "$arg" "$helm_line" >&2
        exit 1
      fi
    done
    printf 'PASS: %s Helm arguments\n' "$target"
  )
done
