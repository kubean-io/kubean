set -eo pipefail

OPTION=${1:-'all'}

CurrentDir=$(pwd)
CurrentDate=$(date +%Y%m%d)

GenerateTempList() {
  if [ ! -d "kubespray" ]; then
    git clone https://github.com/kubernetes-sigs/kubespray.git
    ## cd kubespray && git checkout somebranch
  fi
  cd "$CurrentDir"/kubespray/contrib/offline
  bash generate_list.sh
}

CreateFiles() {
  cd "$CurrentDir"/kubespray/contrib/offline
  NO_HTTP_SERVER=true bash manage-offline-files.sh
  cp offline-files.tar.gz "$CurrentDir"
}

CreateImages() {
  cd "$CurrentDir"
  bash manage_images.sh create kubespray/contrib/offline/temp/images.list
}

GetKubeSprayBranch() {
  cd "$CurrentDir"/kubespray
  KubeSprayBranch="$(git symbolic-ref --short HEAD)"
}

CreateTar() {
  GetKubeSprayBranch

  cd "$CurrentDir"
  if [ ! -d "kubeanio-offline" ]; then
    mkdir kubeanio-offline
  fi
  mv offline-files.tar.gz kubeanio-offline
  mv offline-images.tar.gz kubeanio-offline
  tar -czvf kubeanio-offline-"$CurrentDate"-"$KubeSprayBranch".tar.gz kubeanio-offline
}

case $OPTION in
all)
  GenerateTempList
  CreateFiles
  CreateImages
  CreateTar
  ;;

createtemplist)
  GenerateTempList
  ;;

createfiles)
  CreateFiles
  ;;

createimages)
  CreateImages
  ;;

createtar)
  CreateTar
  ;;

*)
  echo -n "unknown operator"
  ;;
esac
