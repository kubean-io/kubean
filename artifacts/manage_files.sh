set -eo pipefail

OPTION=${1:-'import'}
MinioAPIAddress=${2:-'http://127.0.0.1:9000'}

AddMCHostConfig() {
  if [ -z "$Minio_User" ]; then
    echo "need Minio_User and Minio_Password"
    exit 1
  fi
  if ! mc config host add kubeaniominioserver "$MinioAPIAddress" "$Minio_User" "$Minio_Password"; then
    echo "mc add $MinioAPIAddress server failed"
    exit 1
  fi
}

RemoveMCHostConfig() {
  echo "remove mc config"
  mc config host remove kubeaniominioserver
}

CheckMCCmd() {
  if which mc; then
    echo "mc check successfully"
  else
    echo "please install mc first"
    exit 1
  fi
}

ImportFilesToMinio() {
  if [ ! -d "offline-files" ]; then
    tar -xvf offline-files.tar.gz
    echo "unzip successfully"
  fi

  for bucketName in offline-files/*; do
    bucketName=${bucketName//offline-files\//} ## remove dir prefix
    if ! mc ls kubeaniominioserver/"$bucketName" >/dev/null 2>&1 ; then
      echo "create bucket $bucketName"
      mc mb kubeaniominioserver/"$bucketName"
      mc policy set download kubeaniominioserver/"$bucketName"
    fi
  done

  for path in $(find offline-files); do
    if [ -f "$path" ]; then
      ## mc cp offline-files/a/b/c/d.txt kubeaniominioserver/a/b/c/d.txt
      minioFileName=${path//offline-files/kubeaniominioserver}
      mc cp --no-color "$path" "$minioFileName"
    fi
  done

}

case $OPTION in
import)
  CheckMCCmd
  AddMCHostConfig
  ImportFilesToMinio
  RemoveMCHostConfig
  ;;

*)
  echo -n "unknown operator"
  ;;
esac
