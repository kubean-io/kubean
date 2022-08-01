set -eo pipefail

OPTION=${1:-'create'}
Target=${2}

CheckSkopeoCmd() {
  if which skopeo ; then
    echo "skopeo check successfully"
  else
    echo "please install skopeo first"
    exit 1
  fi
}

CheckLocalImageFiles() {
  ## empty dir?
  if [ $(ls -A offline-images | wc -l)  -eq 0 ] ; then
    return
  fi
  for DirName in offline-images/* ; do
    if skopeo inspect dir:"$DirName" >/dev/null 2>&1 ;then
      echo "check $DirName successfully"
    else
      echo "check $DirName and $DirName maybe bad image file"
      exit 1
    fi
  done
}

CreateTar(){
  ImageListFile=$Target ## temp/images.list

  echo "begin to download images"

  if [ ! -d "offline-images" ]; then
    echo "create dir offline-images"
    mkdir offline-images
  fi

  CheckLocalImageFiles

  while read ImageName; do
    ## quay.io/metallb/controller:v0.12.1 => dir:somedir/metallb%controller:v0.12.1
    NewDirName=${ImageName#*/} ## remote host
    NewDirName=${NewDirName//\//%} ## replace all / with %
    echo "download image $ImageName to local $NewDirName"
    skopeo copy --retry-times=3 --override-os linux --override-arch amd64 docker://"$ImageName" dir:offline-images/"$NewDirName"
  done < "$ImageListFile"

  tar -czvf offline-images.tar.gz offline-images

  echo "zipping images completed!"
}

ImagesToRegistry() {
  RegistryAddress=$Target # ip:port

  if [ ! -d "offline-images" ]; then
    tar -xvf offline-images.tar.gz
    echo "unzip successfully"
  fi

  CheckLocalImageFiles

  for DirName in offline-images/* ; do
    CopyCmd="skopeo copy -a "
    if [ "$Dest_TLS_Verify" = false ] ; then
      CopyCmd=$CopyCmd" --dest-tls-verify=false "
    fi
    if [ -n "$Dest_User" ]; then
      CopyCmd=$CopyCmd" --dest-creds=$Dest_User:$Dest_Password "
    fi
    ## dir:offline-images/coreos%etcd:v3.5.4 docker://1.2.3.4:5000/coreos/etcd:v3.5.4
    ImageName=${DirName#*/} # remove dir prefix
    ImageName=${ImageName//%/\/} # replace % with /
    ImageName="$RegistryAddress"/"$ImageName"

    echo "import $DirName to $ImageName"

    $CopyCmd --retry-times=3 dir:"$DirName" docker://"$ImageName"
  done

  echo "import completed!"
}

case $OPTION in
  create)
    CheckSkopeoCmd
    CreateTar
    ;;

  import)
    CheckSkopeoCmd
    ImagesToRegistry
    ;;

  *)
    echo -n "unknown operator"
    ;;
esac
