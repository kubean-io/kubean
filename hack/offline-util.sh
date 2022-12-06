### Clean up the docker containers before test
function util::clean_offline_kind_cluster() {
   echo "======= container prefix: ${CONTAINERS_PREFIX}"
    kubean_containers_num=$( docker ps -a |grep ${CONTAINERS_PREFIX}||true)
    if [ "${kubean_containers_num}" ];then
      echo "Remove exist containers name contains kubean..."
      docker ps -a |grep "${CONTAINERS_PREFIX}"|awk '{print $NF}'|xargs docker rm -f
    else
      echo "No container name contains kubean to delete."
    fi
}

### Restore vm snapshot to only os installed state
function util::restore_vsphere_vm_snapshot {
  VSPHERE_HOST=${1}
  VSPHERE_PASSWD=${2}
  VSPHERE_USER=${3}
  SNAPSHOT_NAME=${4}
  vm_name=${5:-""}
  echo "Start restore vm snapshot..."
  # shell脚本不支持传数组，就改用文件方式获取虚拟机名称列表 hack/vm_name.list
  # vsphere python package
  if [ ! -d "pyvmomi-community-samples" ]; then
    pip3 install -v pyvmomi==7.0.3
    git clone https://github.com/vmware/pyvmomi-community-samples.git
  else
    echo "vmware python repo exist"
  fi

  if [ $vm_name != "" ]; then
      echo "restore $vm_name..."
      python3 pyvmomi-community-samples/samples/snapshot_operations.py -s ${VSPHERE_HOST} -u ${VSPHERE_USER} -p ${VSPHERE_PASSWD} -nossl -v "${vm_name}" -op revert --snapshot-name ${SNAPSHOT_NAME}
  else
    for i in $(cat hack/vm_name.list);
    do
      # revert vm snapshot
      echo ${i}
       python3 pyvmomi-community-samples/samples/snapshot_operations.py -s ${VSPHERE_HOST} -u ${VSPHERE_USER} -p ${VSPHERE_PASSWD} -nossl -v "${i}" -op revert --snapshot-name ${SNAPSHOT_NAME}
    done
  fi
  echo "Restore vm snapshot end!"
}

###  Install MinIO in kind
function util::install_minio(){
  local MINIO_USER=$1
  local MINIO_PASS=$2
  local kubeconfig_file=$3
  helm repo add minio-official https://charts.min.io
  helm repo update minio-official
  helm pull minio-official/minio --version=5.0.1

  # will be replaced by operator later
  helm upgrade --install  --create-namespace --cleanup-on-fail \
            --set rootUser=${MINIO_USER},rootPassword=${MINIO_PASS} \
            --set mode="standalone" \
            --set service.type=NodePort \
            --set consoleService.type=NodePort \
            --set resources.requests.memory=200Mi \
            --set persistence.size=10Gi \
            --kubeconfig "${kubeconfig_file}" \
            minio minio-official/minio --wait
}


### Install docker_registry in kind
### Must set the images version to 2.6.2, otherwise the cilium image will fail
### https://github.com/kubean-io/kubean/issues/246
function util::install_registry(){
  local registry_port=$1
  local kubeconfig_file=$2
  local registry_version=2.1.0
  local service_type="NodePort"
  local registry_namespace="kube-system"
  local registry_path="registry"
  echo "Start install registry..."
  # mkdir -p ${registry_path}
  helm repo add twuni https://helm.twun.io
  helm repo update twuni
  helm pull twuni/docker-registry --version=${registry_version}
  helm upgrade --install registry twuni/docker-registry --version ${registry_version} \
                         --namespace ${registry_namespace}   \
                         --set service.type=${service_type} \
                         --set image.tag=2.6.2 \
                         --set service.nodePort=${registry_port} \
                         --wait \
                         --kubeconfig "${kubeconfig_file}"
}

### Download kubean offline files
function util::download_offline_files(){
  echo "Download offline install files...."
  local arch=${1:-amd64}
  local tag=$2
  local download_folder=${3:-download_offline_files_"${tag}"}
  local base_url=https://github.com/kubean-io/kubean/releases/download/${tag}
  if [ -d "${download_folder}" ]; then
    echo "Local offline_fils folder not empty, delete it."
    rm -fr "${download_folder}"
  fi
  mkdir "${download_folder}"

  f_file_list=${base_url}/files-${arch}.list
  f_files_tgz=${base_url}/files-${arch}-${tag}.tar.gz
  f_images_list=${base_url}/images-${arch}.list
  f_images_tgz=${base_url}/images-${arch}-${tag}.tar.gz
  f_os_pkgs=${base_url}/os-pkgs-centos7-${tag}.tar.gz
  # shellcheck disable=SC2206
  file_down_list=(${f_file_list} ${f_files_tgz} ${f_images_list} ${f_images_tgz} ${f_os_pkgs})
  for (( i=0; i<${#file_down_list[@]};i++)); do
    echo "${file_down_list[$i]}"
    wget -q -c  -P  "${download_folder}"  "${file_down_list[$i]}"
  done
}

### Uncompress files downloaded
### All the file "*.tar.gz" will be uncompressed.
function util::uncompress_tgz_files(){
  download_folder=$1
  echo "Uncompress tgz files ..."
  pushd "${download_folder}"
  tgz_list=($(ls -al |grep "tar.gz"|awk '{print $NF}'))
  for(( i=0;i<${#tgz_list[@]};i++));do
      tar -zxvf  ${tgz_list[i]}
  done
  popd
}

### Uncompress files downloaded
function util::uncompress_os_tgz_files(){
  os=${1:-centos7}
  tag=$2
  download_folder=$3
  pushd "${download_folder}"
  os-pkgs-centos7-v0.4.0-rc7.tar.gz
  os_tgz="os-pkgs--${os}-${tag}.tar.gz"
  tar -zxf "${os_tgz}"
  popd
}

### Import binary files to kind minio
function util::import_files_minio(){
  echo "Import binary files to minio ..."
  minio_usr=${1:-admin}
  minio_password=${2:-adminPassword}
  minio_url=${3:-"http://172.18.0.2:32000"}
  import_files_path=${4}
  pushd "${import_files_path}"
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password}  ./import_files.sh ${minio_url} > /dev/null
  popd
}

### Import os packages to kind minio
### 主要用于解决 docker-ce 的安装依赖
function util::import_os_package_minio_by_arch(){
  echo "Import ospkgs to minio ..."
  minio_usr=${1:-admin}
  minio_password=${2:-adminPassword}
  minio_url=${3:-"http://172.18.0.2:32000"}
  os_packeges_path=${4}
  arch=${5}
  pushd ${os_packeges_path}
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password}  ./import_ospkgs.sh  ${minio_url}  os-pkgs-${arch}.tar.gz > /dev/null
  popd
}

### Push images file to kind registry
function util::push_registry(){
  echo "Push Registry... "
  registry_addr=$1
  images_files_path=${2}
  pushd "${images_files_path}"
  DEST_TLS_VERIFY=false ./import_images.sh ${registry_addr}
  popd
}


### Use artifacts/gen_repo_conf.sh to generate local repo
### Before test, should prepare the iso file
### Centos repo only support redhat os family
function util::mount_iso_image(){
  linux_distribution=${1:-centos}
  iso_image_file=${2}
  shell_path=${3}
  mount_exist_flag=$(mount|grep "${iso_image_file}"||true)
  echo "mount_exist_flag is: ${mount_exist_flag}"
  if [ ! "${mount_exist_flag}" ]; then
    echo "Mount iso images."
    check_iso_img "${iso_image_file}"
    pushd "${shell_path}"
    echo "current dirs:"
    pwd
    #chmod +x "${shell_path}"/gen_repo_conf.sh
    echo "Start import iso to minio..."
    sh gen_repo_conf.sh --iso-mode "${linux_distribution}" "${iso_image_file}" > /dev/null
    popd
  else
    echo "Iso is already mounted, nothing to do"
  fi
}

function check_iso_img() {
    ISO_IMG_FILE=$1
    if [ ! -f ${ISO_IMG_FILE} ]; then
      echo "iso image: \${ISO_IMG_FILE} should exist."
      exit 1
    fi
}

### Import iso images repo files to minio
### Work cluster node will use the image repo
function util::import_iso(){
  minio_usr=${1:-admin}
  minio_password=${2:-adminPassword}
  minio_url=${3:-"http://172.18.0.2:32000"}
  shell_path=${4}
  iso_image_file=${5}
  check_iso_img "${iso_image_file}"
  # umount before mount
  set_ios_unmounted "${iso_image_file}"
  pushd "${shell_path}"
  pwd
  chmod +x import_iso.sh
  echo "Start import iso to Minio, wait patiently...."
  MINIO_USER=${minio_usr} MINIO_PASS=${minio_password} ./import_iso.sh ${minio_url} ${iso_image_file} > /dev/null
  popd
}

function set_ios_unmounted(){
  echo "Umount iso if is already mounted"
  iso_image_file=$1
  mount_exist_flag=$(mount|grep "${iso_image_file}"||true)
  echo "mount_exist_flag is: ${mount_exist_flag}"
    if [  "${mount_exist_flag}" ]; then
      echo "Is already mounted before import, umount now..."
      umount ${iso_image_file}
    fi
}

### Set kubean node firewalld, add registry and minio service port for worker node to fetch files
function util::firewalld_add_port_forward(){
  node_port=$1
  kind_service_port=$2
  kind_ip=$3
  cmd="firewall-cmd --add-forward-port=port=${node_port}:proto=tcp:toaddr=${kind_ip}:toport=${kind_service_port}"
  echo $cmd
  ${cmd}
}

### Set work cluster node ip
function util::vm_name_ip_init(){
  echo "RUNNER NAME: " $RUNNER_NAME
  echo "OFFLINE_FLAG" $OFFLINE_FLAG
  # Offline vm name && ip init
  if [ "${OFFLINE_FLAG}" == "true" ] || [ "${OFFLINE_FLAG}" == "True" ]; then
     if [ "${RUNNER_NAME}" == "debug" ]; then
        vm_ip_addr1="10.16.10.163"
        vm_ip_addr2="10.16.10.164"
        vm_name1="gwt-kubean-offline-e2e-node3"
        vm_name2="gwt-kubean-offline-e2e-node4"
      else
        vm_ip_addr1="10.16.10.161"
        vm_ip_addr2="10.16.10.162"
        vm_name1="gwt-kubean-offline-e2e-node1"
        vm_name2="gwt-kubean-offline-e2e-node2"
      fi
  else
    ## Online vm name && ip init
    if [ "${RUNNER_NAME}" == "debug" ]; then
          vm_ip_addr1="10.6.127.41"
          vm_ip_addr2="10.6.127.42"
          vm_name1="gwt-kubean-e2e-node5"
          vm_name1="gwt-kubean-e2e-node6"
    else
      if [ "${RUNNER_NAME}" == "kubean-actions-runner1" ]; then
          vm_ip_addr1="10.6.127.31"
          vm_ip_addr2="10.6.127.32"
          vm_name1="gwt-kubean-e2e-node1"
          vm_name1="gwt-kubean-e2e-node2"
      fi
      if [ "${RUNNER_NAME}" == "kubean-actions-runner2" ]; then
          vm_ip_addr1="10.6.127.33"
          vm_ip_addr2="10.6.127.34"
          vm_name1="gwt-kubean-e2e-node3"
          vm_name1="gwt-kubean-e2e-node4"
      else
           vm_ip_addr1="10.6.127.31"
           vm_ip_addr2="10.6.127.32"
           vm_name1="gwt-kubean-e2e-node1"
           vm_name1="gwt-kubean-e2e-node2"
      fi
    fi
  fi
  echo "vm name:  $vm_name1 $vm_name2"
  echo "vm_ip_addr:  $vm_ip_addr1 $vm_ip_addr2"
}

### Use skopeo copy images, which used in golang case, to docker registry
function util::scope_copy_test_images(){
   dest_registry_addr=${1}
   image_name=${2:-""}
   skopeo_cmd="skopeo copy --insecure-policy --src-tls-verify=false --dest-tls-verify=false  "
   if [ "${image_name}" != "" ]; then
        echo "skopeo copy image to registry: ${image_name}"
        ${skopeo_cmd} docker://"${image_name}"  docker://"${dest_registry_addr}"/test/"${image_name}"
    else
      for image_name in $(cat hack/test_images.list);
      do
         echo "skopeo copy image to registry: ${image_name}"
         ${skopeo_cmd} docker://"${image_name}"  docker://"${dest_registry_addr}"/test/"${image_name}"
      done
    fi
    echo "Skopeo copy images end!"
}

### Fetch image addr from docker registry
# ${registry_addr} format: 172.30.41.62:31500
# ${image} format: nginx
# function util::fetch_image_addr_from_registry(){
  #image=$1
 # registry_addr=$2
 # project_name=${3:-"test"}
 # cmd="curl -s  http://${registry_addr}/v2/${project_name}/${image}/tags/list"
  #cmd="curl -s  http://${registry_addr}/v2/${project_name}/${image}/tags/list"
 # echo ${cmd}
  #tag_list=$(${cmd})
 # tag=$(echo ${tag_list#*tags\":[}|awk -F ',' '{print $1}'|awk -F '"' '{print $2}')
 # echo $tag
  #echo "end"
#}

