#!/bin/bash

# This script launches a job pod in a persistent state using 'sleep infinity', 
# allowing us to utilize the Kubespray environment within this pod for debugging Ansible playbooks.
# 
# 1. apply kubean custom resources and create pod.
# $ export DEBUG_NODE_ADDR=192.168.10.10
# $ export DEBUG_NODE_USER=root
# $ export DEBUG_NODE_PASS=1q2w3e
# $ ./debug.sh
#
# 2. enter the kubespray container environment.
# $ kubectl -n kubean-system exec -it kubean-debug-install-ops-job-xxxxx bash
#
# 3. execute the ansible-playbook command of Kubespray inside the container.
# $ ansible-playbook -i /conf/hosts.yml -b --become-user root -e "@/conf/group_vars.yml" /kubespray/cluster.yml
#
# 4. delete custom resources.
# $ kubectl delete cluster debug
#

DEBUG_NODE_ADDR=${DEBUG_NODE_ADDR:-""}
DEBUG_NODE_USER=${DEBUG_NODE_USER:-""}
DEBUG_NODE_PASS=${DEBUG_NODE_PASS:-""}
SPRAY_JOB_VERSION=${SPRAY_JOB_VERSION:-"latest"}
CLUSTER_NAME=${CLUSTER_NAME:-"debug"}

function check_params() {
  if [[ -z "${DEBUG_NODE_ADDR}" || -z "${DEBUG_NODE_USER}" || -z "${DEBUG_NODE_PASS}" ]]; then
    echo "DEBUG_NODE_ADDR: '${DEBUG_NODE_ADDR}'"
    echo "DEBUG_NODE_USER: '${DEBUG_NODE_USER}'"
    echo "DEBUG_NODE_PASS: '${DEBUG_NODE_PASS}'"
    echo "[warn] node params cannot be empty."
    exit 1
  fi
}
check_params

cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${CLUSTER_NAME}-hosts-conf
  namespace: kubean-system
data:
  hosts.yml: |
    all:
      hosts:
        node1:
          ip: ${DEBUG_NODE_ADDR}
          access_ip: ${DEBUG_NODE_ADDR}
          ansible_host: ${DEBUG_NODE_ADDR}
          ansible_connection: ssh
          ansible_user: ${DEBUG_NODE_USER}
          ansible_password: ${DEBUG_NODE_PASS}
      children:
        kube_control_plane:
          hosts:
            node1:
        kube_node:
          hosts:
            node1:
        etcd:
          hosts:
            node1:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
 
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${CLUSTER_NAME}-vars-conf
  namespace: kubean-system
data:
  group_vars.yml: |
    unsafe_show_logs: true
    container_manager: containerd
    kube_network_plugin: calico
    kube_network_plugin_multus: false
    kube_proxy_mode: iptables
    etcd_deployment_type: kubeadm
    override_system_hostname: true
    ntp_enabled: true
 
    pkg_install_retries: 1
    download_retries: 1

    gcr_image_repo: "gcr.m.daocloud.io"
    kube_image_repo: "k8s.m.daocloud.io"
    docker_image_repo: "docker.m.daocloud.io"
    quay_image_repo: "quay.m.daocloud.io"
    github_image_repo: "ghcr.m.daocloud.io"

    github_url: https://files.m.daocloud.io/github.com
    dl_k8s_io_url: https://files.m.daocloud.io/dl.k8s.io
    storage_googleapis_url: https://files.m.daocloud.io/storage.googleapis.com
    get_helm_url: https://files.m.daocloud.io/get.helm.sh

---
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
  name: ${CLUSTER_NAME}
  labels:
    clusterName: ${CLUSTER_NAME}
spec:
  hostsConfRef:
    namespace: kubean-system
    name: ${CLUSTER_NAME}-hosts-conf
  varsConfRef:
    namespace: kubean-system
    name:  ${CLUSTER_NAME}-vars-conf

---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: ${CLUSTER_NAME}-install-ops
spec:
  cluster: ${CLUSTER_NAME}
  image: ghcr.m.daocloud.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}
  actionType: shell
  action: sleep infinity
EOF
