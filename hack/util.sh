#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# This script holds common bash variables and utility functions.

ETCD_POD_LABEL="etcd"
KUBE_CONTROLLER_POD_LABEL="kube-controller-manager"

MIN_Go_VERSION=go1.16.0

# This function installs a Go tools by 'go get' command.
# Parameters:
#  - $1: package name, such as "sigs.k8s.io/controller-tools/cmd/controller-gen"
#  - $2: package version, such as "v0.4.1"
# Note:
#   Since 'go get' command will resolve and add dependencies to current module, that may update 'go.mod' and 'go.sum' file.
#   So we use a temporary directory to install the tools.
function util::install_tools() {
	local package="$1"
	local version="$2"

	temp_path=$(mktemp -d)
	pushd "${temp_path}" >/dev/null
	GO111MODULE=on go install "${package}"@"${version}"
	GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
	export PATH=$PATH:$GOPATH/bin
	popd >/dev/null
	rm -rf "${temp_path}"
}

function util::cmd_exist {
	local CMD=$(command -v ${1})
	if [[ ! -x ${CMD} ]]; then
    	return 1
	fi
	return 0
}

# util::cmd_must_exist check whether command is installed.
function util::cmd_must_exist {
    local CMD=$(command -v ${1})
    if [[ ! -x ${CMD} ]]; then
    	echo "Please install ${1} and verify they are in \$PATH."
    	exit 1
    fi
}

function util::verify_go_version {
    local go_version
    IFS=" " read -ra go_version <<< "$(GOFLAGS='' go version)"
    if [[ "${MIN_Go_VERSION}" != $(echo -e "${MIN_Go_VERSION}\n${go_version[2]}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) && "${go_version[2]}" != "devel" ]]; then
      echo "Detected go version: ${go_version[*]}."
      echo "kubean requires ${MIN_Go_VERSION} or greater."
      echo "Please install ${MIN_Go_VERSION} or later."
      exit 1
    fi
}

# util::install_environment_check will check OS and ARCH before installing
# ARCH support list: amd64,arm64
# OS support list: linux,darwin
function util::install_environment_check {
    local ARCH=${1:-}
    local OS=${2:-}
    if [[ "$ARCH" =~ ^(amd64|arm64)$ ]]; then
        if [[ "$OS" =~ ^(linux|darwin)$ ]]; then
            return 0
        fi
    fi
    echo "Sorry, kubean installation does not support $ARCH/$OS at the moment"
    exit 1
}

# util::install_kubectl will install the given version kubectl
function util::install_kubectl {
    local KUBECTL_VERSION=${1}
    local ARCH=${2}
    local OS=${3:-linux}
    if [ -z "$KUBECTL_VERSION" ]; then
    	KUBECTL_VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt)
    fi
    echo "Installing 'kubectl ${KUBECTL_VERSION}' for you"
    curl --retry 5 -sSLo ./kubectl -w "%{http_code}" https://dl.k8s.io/release/"$KUBECTL_VERSION"/bin/"$OS"/"$ARCH"/kubectl | grep '200' > /dev/null
    ret=$?
    if [ ${ret} -eq 0 ]; then
        chmod +x ./kubectl
        mkdir -p ~/.local/bin/
        mv ./kubectl ~/.local/bin/kubectl

        export PATH=$PATH:~/.local/bin
    else
        echo "Failed to install kubectl, can not download the binary file at https://dl.k8s.io/release/$KUBECTL_VERSION/bin/$OS/$ARCH/kubectl"
        exit 1
    fi
}

# util::install_kind will install the given version kind
function util::install_kind {
	local kind_version=${1}
	echo "Installing 'kind ${kind_version}' for you"
	local os_name
	os_name=$(go env GOOS)
	local arch_name
	arch_name=$(go env GOARCH)
	curl --retry 5 -sSLo ./kind -w "%{http_code}" "https://qiniu-download-public.daocloud.io/Kind/${kind_version}/kind-${os_name:-linux}-${arch_name:-amd64}" | grep '200' > /dev/null
	ret=$?
	if [ ${ret} -eq 0 ]; then
    	chmod +x ./kind
    	mkdir -p ~/.local/bin/
    	mv ./kind ~/.local/bin/kind

    	export PATH=$PATH:~/.local/bin
	else
    	echo "Failed to install kind, can not download the binary file at https://qiniu-download-public.daocloud.io/Kind/${kind_version}/kind-${os_name:-linux}-${arch_name:-amd64}"
    	exit 1
	fi
}

# util::wait_for_condition blocks until the provided condition becomes true
# Arguments:
#  - 1: message indicating what conditions is being waited for (e.g. 'ok')
#  - 2: a string representing an eval'able condition.  When eval'd it should not output
#       anything to stdout or stderr.
#  - 3: optional timeout in seconds. If not provided, waits forever.
# Returns:
#  1 if the condition is not met before the timeout
function util::wait_for_condition() {
  local msg=$1
  # condition should be a string that can be eval'd.
  local condition=$2
  echo "condition is：$condition"
  local timeout=${3:-}

  local start_msg="Waiting for ${msg}"
  local error_msg="[ERROR] Timeout waiting for ${msg}"

  local counter=0
  while ! eval ${condition}; do
    if [[ "${counter}" = "0" ]]; then
      echo -n "${start_msg}"
    fi

    if [[ -z "${timeout}" || "${counter}" -lt "${timeout}" ]]; then
      counter=$((counter + 1))
      if [[ -n "${timeout}" ]]; then
        echo -n '.'
      fi
      sleep 1
    else
      echo -e "\n${error_msg}"
      return 1
    fi
  done

  if [[ "${counter}" != "0" && -n "${timeout}" ]]; then
    echo ' done'
  fi
}

# util::wait_file_exist checks if a file exists, if not, wait until timeout
function util::wait_file_exist() {
    local file_path=${1}
    local timeout=${2}
    for ((time=0; time<${timeout}; time++)); do
        if [[ -e ${file_path} ]]; then
            return 0
        fi
        sleep 1
    done
    return 1
}

# util::wait_pod_ready waits for pod state becomes ready until timeout.
# Parmeters:
#  - $1: pod label, such as "app.kubernetes.io/name=kubean"
#  - $2: pod namespace, such as "kubean-system"
#  - $3: time out, such as "200s"
function util::wait_pod_ready() {
    local pod_label=$1
    local pod_namespace=$2
    local timeout=$3

    echo "wait the $pod_label ready..."
    set +e
    util::kubectl_with_retry wait --for=condition=Ready --timeout=${timeout} pods -l app.kubernetes.io/name=${pod_label} -n ${pod_namespace}
    ret=$?
    set -e
    if [ $ret -ne 0 ];then
      echo "kubectl describe info: $(kubectl describe pod -l app.kubernetes.io/name=${pod_label} -n ${pod_namespace})"
    fi
    return ${ret}
}

# util::kubectl_with_retry will retry if execute kubectl command failed
# tolerate kubectl command failure that may happen before the pod is created by StatefulSet/Deployment.
function util::kubectl_with_retry() {
    local ret=0
    local count=0
    for i in {1..10}; do
        kubectl "$@"
        ret=$?
        if [[ ${ret} -ne 0 ]]; then
            echo "kubectl $@ failed, retrying(${i} times)"
            sleep 1
            continue
        else
          ((count++))
          # sometimes pod status is from running to error to running
          # so we need check it more times
          if [[ ${count} -ge 3 ]];then
            return 0
          fi
          sleep 1
          continue
        fi
    done

    echo "kubectl $@ failed"
    kubectl "$@"
    return ${ret}
}

# util::create_cluster creates a kubernetes cluster
# util::create_cluster creates a kind cluster and don't wait for control plane node to be ready.
# Parmeters:
#  - $1: cluster name, such as "host"
#  - $2: KUBECONFIG file, such as "/var/run/host.config"
#  - $3: node docker image to use for booting the cluster, such as "kindest/node:v1.19.1"
#  - $4: log file path, such as "/tmp/logs/"
function util::create_cluster() {
	local cluster_name=${1}
	local kubeconfig=${2}
	local kind_image=${3}
	local cluster_config=${4:-}

	rm -f "${kubeconfig}"
  ~/.local/bin/kind delete cluster --name="${cluster_name}"
  kind create cluster --name "${cluster_name}" --kubeconfig="${kubeconfig}" --image="${kind_image}" --config="${cluster_config}"
  echo "cluster ${cluster_name} created successfully"
}

# util::delete_cluster deletes kind cluster by name
# Parmeters:
# - $1: cluster name, such as "host"
function util::delete_cluster() {
       local cluster_name=${1}
       ~/.local/bin/kind delete cluster --name="${cluster_name}"
}

# This function returns the IP address of a docker instance
# Parameters:
#  - $1: docker instance name

function util::get_docker_native_ipaddress(){
  local container_name=$1
  docker inspect --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${container_name}"
}

# This function returns the IP address and port of a specific docker instance's host IP
# Parameters:
#  - $1: docker instance name
# Note:
#   Use for getting host IP and port for cluster
#   "6443/tcp" assumes that API server port is 6443 and protocol is TCP

function util::get_docker_host_ip_port(){
  local container_name=$1
  docker inspect --format='{{range $key, $value := index .NetworkSettings.Ports "6443/tcp"}}{{if eq $key 0}}{{$value.HostIp}}:{{$value.HostPort}}{{end}}{{end}}' "${container_name}"
}

# util::check_clusters_ready checks if a cluster is ready, if not, wait until timeout
function util::check_clusters_ready() {
	local kubeconfig_path=${1}
	local context_name=${2}

	echo "Waiting for kubeconfig file ${kubeconfig_path} and clusters ${context_name} to be ready..."
	util::wait_file_exist "${kubeconfig_path}" 300
	util::wait_for_condition 'running' "docker inspect --format='{{.State.Status}}' ${context_name}-control-plane &> /dev/null" 300

	kubectl config rename-context "kind-${context_name}" "${context_name}" --kubeconfig="${kubeconfig_path}"

	local os_name
	os_name=$(go env GOOS)
	local container_ip_port
	case $os_name in
    	linux) container_ip_port=$(util::get_docker_native_ipaddress "${context_name}-control-plane")":6443"
    	;;
    	darwin) container_ip_port=$(util::get_docker_host_ip_port "${context_name}-control-plane")
    	;;
   		*)
			echo "OS ${os_name} does NOT support for getting container ip in installation script"
			exit 1
	esac
	kubectl config set-cluster "kind-${context_name}" --server="https://${container_ip_port}" --kubeconfig="${kubeconfig_path}"

	util::wait_for_condition 'ok' "kubectl --kubeconfig ${kubeconfig_path} --context ${context_name} get --raw=/healthz &> /dev/null" 300
}

###### to get k8 cluster single node ip address based on actions-runner #######
function utils:runner_ip(){
    echo "RUNNER_NAME: "$RUNNER_NAME
    if [ "${RUNNER_NAME}" == "kubean-actions-runner1" ]; then
        vm_ip_addr1="10.6.127.33"
        vm_ip_addr2="10.6.127.36"
    fi
    if [ "${RUNNER_NAME}" == "kubean-actions-runner2" ]; then
        vm_ip_addr1="10.6.127.35"
        vm_ip_addr2="10.6.127.37"
    fi
    if [ "${RUNNER_NAME}" == "kubean-actions-runner3" ]; then
        vm_ip_addr1="10.6.127.39"
        vm_ip_addr2="10.6.127.40"
    fi
    if [ "${RUNNER_NAME}" == "kubean-actions-runner4" ]; then
        vm_ip_addr1="10.6.127.42"
        vm_ip_addr2="10.6.127.43"
    fi
    if [ "${RUNNER_NAME}" == "debug" ]; then
        vm_ip_addr1="172.30.41.75"
        vm_ip_addr2="172.30.41.76"
    fi
}

###### Clean Up #######
function utils::clean_up(){
    echo "======= cluster prefix: ${CLUSTER_PREFIX}"
    local auto_cleanup="true"
    if [ "$auto_cleanup" == "true" ];then
      bash  "${REPO_ROOT}"/hack/delete-cluster.sh "${CLUSTER_PREFIX}"-host
    fi
    if [ "$EXIT_CODE" == "0" ];then
        exit $EXIT_CODE
    fi
    exit $EXIT_CODE
}

function utils::create_os_e2e_vms(){
    # create 1master+1worker cluster
    if [ -f $(pwd)/Vagrantfile ]; then
        rm -f $(pwd)/Vagrantfile
    fi
    cp $(pwd)/hack/os_vagrantfiles/"${1}" $(pwd)/Vagrantfile
    sed -i "s/sonobouyDefault_ip/${2}/" Vagrantfile
    sed -i "s/sonobouyDefault2_ip/${3}/" Vagrantfile
    vagrant up
    vagrant status
    ATTEMPTS=0
    pingOK=0
    ping -w 2 -c 1 $2|grep "0%" && pingOK=true || pingOK=false
    until [ "${pingOK}" == "true" ] || [ $ATTEMPTS -eq 10 ]; do
    ping -w 2 -c 1 $2|grep "0%" && pingOK=true || pingOK=false
    echo "==> ping "$2 $pingOK
    ATTEMPTS=$((ATTEMPTS + 1))
    sleep 10
    done
    ping -c 5 ${2}
    ping -c 5 ${3}
}

function utils::install_sshpass(){
    local CMD=$(command -v ${1})
    if [[ ! -x ${CMD} ]]; then
        echo "Installing sshpass: "
        wget --no-check-certificate http://sourceforge.net/projects/sshpass/files/sshpass/1.05/sshpass-1.05.tar.gz
        tar xvzf sshpass-1.05.tar.gz
        cd sshpass-1.05
        ./configure
        make
        echo "root" | sudo make install
        cd ..
    fi
}

function vm_clean_up_by_name(){
    echo "$# vm to destroy"
    for vm in $@;do
        echo "start destroy: ${vm}"
        vm_id=`vagrant global-status |grep ${vm} -w|grep virtualbox|awk '{print $1}'`
        if [[ -n ${vm_id} ]]; then
            echo "destroy vm: ${vm}  ${vm_id}"
            vagrant destroy -f $vm_id
        else
            echo "${vm} not exists"
        fi
     done
   echo "destroy vagrant vm end."
}
