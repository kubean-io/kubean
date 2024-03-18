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

    files_repo: "https://files.m.daocloud.io"
 
    kubeadm_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
    kubectl_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
    kubelet_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"
    cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"
    crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
    etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"
    calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calico_crds_download_url: "{{ files_repo }}/github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"
    flannel_cni_download_url: "{{ files_repo }}/kubernetes/flannel/{{ flannel_cni_version }}/flannel-{{ image_arch }}"
    helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"
    crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"
    kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"
    cri_dockerd_download_url: "{{ files_repo }}/github.com/Mirantis/cri-dockerd/releases/download/v{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}.{{ image_arch }}.tgz"
    runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
    containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
    nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
    gvisor_runsc_download_url: "{{ files_repo }}/storage.googleapis.com/gvisor/releases/release/{{ gvisor_version }}/{{ ansible_architecture }}/runsc"
    gvisor_containerd_shim_runsc_download_url: "{{ files_repo }}/storage.googleapis.com/gvisor/releases/release/{{ gvisor_version }}/{{ ansible_architecture }}/containerd-shim-runsc-v1"
 
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
