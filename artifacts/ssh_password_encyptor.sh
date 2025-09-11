#!/bin/bash
set -e

PUBLICKEY_PATH=${PUBLICKEY_PATH:-}
KUBECONFIG=${KUBECONFIG:-"$HOME/.kube/config"}

for cmd in base64 openssl; do
  if ! which "$cmd" &>/dev/null; then
    echo "Need $cmd"
    exit 1
  fi
done

if [ -z "$PUBLICKEY_PATH" ]; then
  PUBLICKEY_PATH=$(mktemp)
  trap "rm -rf $PUBLICKEY_PATH" exit
  kubectl -n kubean-system get configmap kubean-pubkey -ojsonpath='{.data.pk}' | base64 -d > "$PUBLICKEY_PATH"
fi

if [ ! -s "$PUBLICKEY_PATH" ]; then
  echo "Cannot get public key, Check PUBLICKEY_PATH env"
  exit 1
fi

read -s -r -p "Your password: " password
echo -en "\nEncrypted password: "
echo "VAULT;$(echo -n "$password" | openssl pkeyutl -encrypt -pubin -inkey "$PUBLICKEY_PATH" | base64 -w0)"
