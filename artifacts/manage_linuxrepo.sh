set -eo pipefail

OPTION=${1:-'create'}
MinioAPIAddress=${2:-'http://127.0.0.1:9000'}

CheckCmd(){
  if ! which createrepo ; then
    echo "need createrepo tool"
    exit 1
  fi
  if ! which wget ; then
    echo "need wget tool"
    exit 1
  fi
}

DownloadForCentos(){
  versions=("7") ## ("7" "8")
  ##archs=("aarch64" "x86_64")
  archs=("x86_64")
  baseurl="https://download.docker.com/linux/centos/VERSION/ARCH/stable/Packages"
  rpmfiles=(
    "docker-ce-cli-19.03.15-3.elVERSION.ARCH.rpm"
    "docker-ce-19.03.15-3.elVERSION.ARCH.rpm"
    "docker-ce-20.10.11-3.elVERSION.ARCH.rpm"
    "docker-ce-cli-20.10.11-3.elVERSION.ARCH.rpm"
    "docker-ce-rootless-extras-20.10.11-3.elVERSION.ARCH.rpm"
    "docker-ce-20.10.17-3.elVERSION.ARCH.rpm"
    "docker-ce-cli-20.10.17-3.elVERSION.ARCH.rpm"
    "docker-ce-rootless-extras-20.10.17-3.elVERSION.ARCH.rpm"
    "docker-scan-plugin-0.17.0-3.elVERSION.ARCH.rpm"
    "containerd.io-1.6.4-3.1.elVERSION.ARCH.rpm"
  )

  for version in "${versions[@]}"
  do
    for arch in "${archs[@]}"
    do
      mkdir -p linuxrepo/centos/"$version"/"$arch"
      for rpmfile in "${rpmfiles[@]}"
      do
        rpmfile="${rpmfile//ARCH/$arch}"
        rpmfile="${rpmfile//VERSION/$version}"
        fileURL="${baseurl//ARCH/$arch}"/"${rpmfile}"
        fileURL="${fileURL//VERSION/$version}"
        echo "downloading $fileURL"
        if ! wget "$fileURL" -q ; then
          echo "not found $fileURL"
        else
          mv "${rpmfile}" linuxrepo/centos/"$version"/"$arch"
        fi
      done
      if ! createrepo linuxrepo/centos/"$version"/"$arch" ; then
        echo "create repo but failed"
        exit 1
      fi
    done
  done
}

CreateTar() {
    tar -czvf linuxrepo.tar.gz linuxrepo
}

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

ImportLinuxRepoToMinio() {
  if [ ! -d "linuxrepo" ]; then
    tar -xvf linuxrepo.tar.gz
    echo "unzip successfully"
  fi

  for bucketName in linuxrepo/*; do
    bucketName=${bucketName//linuxrepo\//} ## remove dir prefix
    if ! mc ls kubeaniominioserver/"$bucketName" >/dev/null 2>&1 ; then
      echo "create bucket $bucketName"
      mc mb kubeaniominioserver/"$bucketName"
      mc policy set download kubeaniominioserver/"$bucketName"
    fi
  done

  for path in $(find linuxrepo); do
    if [ -f "$path" ]; then
      ## mc cp linuxrepo/centos/7/x86_64/x.rpm kubeaniominioserver/centos/7/x86_64/x.rpm
      minioFileName=${path//linuxrepo/kubeaniominioserver}
      mc cp --no-color "$path" "$minioFileName"
    fi
  done
}

case $OPTION in
  create)
    CheckCmd
    DownloadForCentos
    CreateTar
    ;;

  import)
    CheckMCCmd
    AddMCHostConfig
    ImportLinuxRepoToMinio
    RemoveMCHostConfig
    ;;

  *)
    echo -n "unknown operator"
    ;;
esac