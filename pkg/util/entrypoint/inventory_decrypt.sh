VAULT_PASSWORD_FILE="/auth/vault-password"
if [ -f $VAULT_PASSWORD_FILE ]; then
  HOSTS_FILE_RAW="/conf/hosts.yml"
  HOSTS_FILE="/conf/hosts0.yml"
  cp $HOSTS_FILE_RAW $HOSTS_FILE
  yq '.all.hosts | keys | join("\n")' $HOSTS_FILE | while read host; do
    echo "Decrypting vault for $host"
    decryptable_fields=
    if [ "$(yq ".all.hosts.$host.ansible_password | tag" $HOSTS_FILE)" = "!vault" ]; then
      decryptable_fields+="ansible_password "
    fi
    if [ "$(yq ".all.hosts.$host.ansible_ssh_pass | tag" $HOSTS_FILE)" = "!vault" ]; then
      decryptable_fields+="ansible_ssh_pass "
    fi
    if [ "$(yq ".all.hosts.$host.ansible_become_password | tag" $HOSTS_FILE)" = "!vault" ]; then
      decryptable_fields+="ansible_become_password "
    fi
    for field in $decryptable_fields; do
      yq ".all.hosts.$host.$field" $HOSTS_FILE > decrypted_field
      decrypted_field=$(ansible-vault decrypt --output - --vault-password-file $VAULT_PASSWORD_FILE decrypted_field)
      yq -i ".all.hosts.$host.$field tag = \"\"" $HOSTS_FILE
      yq -i ".all.hosts.$host.$field = \"$decrypted_field\"" $HOSTS_FILE
      rm -f decrypted_field
    done
  done
  echo "Decrypting done"
  exec 200<> $HOSTS_FILE
  rm -f $HOSTS_FILE
fi
