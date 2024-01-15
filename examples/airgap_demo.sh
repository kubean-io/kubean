#!/bin/bash

# need helm skopeo podman mc
# https://helm.sh/docs/intro/install/
# github.com/lework/skopeo-binary/releases
# https://min.io/docs/minio/linux/reference/minio-mc.html


CURRENT_PATH=$(pwd) || true
ARCH=amd64
DISTRO=centos7
KUBEAN_VERSION="v0.10.0"

NETWORK_INTERFACE=ens192
LOCALHOST_ADDR=$(ip addr show "${NETWORK_INTERFACE}" | grep -Po 'inet \K[\d.]+' || true)

MINIO_USER="kubeanuser"
MINIO_PASS="kubeanpass123"
MINIO_ADDR="http://${LOCALHOST_ADDR}:32000"

REGISTRY_ADDR="${LOCALHOST_ADDR}:30080"

TEST_NODE_ADDR="172.30.41.162"
TEST_NODE_USER="root"
TEST_NODE_PASS="dangerous"


function create_pvc() {
  local name=$1
  local storage=$2
  local pv_name="pv-${name}"
  local pvc_name="pvc-${name}"
  local namespace="${name}-system"
  local kind_host_path="/home/storage/${name}"

  # create path
  mkdir -p "${kind_host_path}"
  chmod -R 777 "${kind_host_path}"

  # create pv & pvc
  cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${pv_name}
spec:
  storageClassName: ""
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: ${storage}
  hostPath:
    path: ${kind_host_path}

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ${pvc_name}
  namespace: ${namespace}
spec:
  volumeName: ${pv_name}
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: ${storage}

EOF
}

# install minio & registry
function install_minio_and_registry() {
  local minio_version="4.0.8"
  local registry_version="2.1.0"

  helm repo add daocloud-community https://release.daocloud.io/chartrepo/community --force-update

  kubectl create ns minio-system --dry-run=client -o yaml | kubectl apply -f - || true
  kubectl create ns registry-system --dry-run=client -o yaml | kubectl apply -f - || true

  create_pvc "minio" "50Gi"
  create_pvc "registry" "100Gi"

  helm upgrade --install  --create-namespace --cleanup-on-fail --namespace minio-system \
    --set users[0].accessKey="${MINIO_USER}" \
    --set users[0].secretKey="${MINIO_PASS}" \
    --set users[0].policy='consoleAdmin' \
    --set securityContext.runAsUser=0,securityContext.runAsGroup=0 \
    --set mode=standalone \
    --set replicas=1 \
    --set service.type=NodePort \
    --set consoleService.type=NodePort \
    --set resources.requests.memory=200Mi \
    --set persistence.existingClaim=pvc-minio \
    --set image.repository=quay.m.daocloud.io/minio/minio \
    --set mcImage.repository=quay.m.daocloud.io/minio/mc --version="${minio_version}" \
    minio daocloud-community/minio --wait

  helm upgrade --install  --create-namespace --cleanup-on-fail --namespace registry-system \
    --set service.type=NodePort \
    --set service.nodePort=30080 \
    --set persistence.enabled=true \
    --set persistence.existingClaim=pvc-registry \
    --set image.repository=docker.m.daocloud.io/registry \
    registry daocloud-community/docker-registry --version="${registry_version}" --wait
}

install_minio_and_registry

# install kubean helm chart
function install_kubean_chart() {
  local kubean_version=$1
  helm repo add community https://release.daocloud.io/chartrepo/community --force-update

  helm upgrade --install  --create-namespace --cleanup-on-fail kubean -n kubean-system community/kubean --create-namespace --wait \
    --version="${kubean_version}" \
    --set kubeanOperator.image.registry=ghcr.m.daocloud.io \
    --set kubeanAdmission.image.registry=ghcr.m.daocloud.io \
    --set sprayJob.image.registry=ghcr.m.daocloud.io \
    --set kubeanOperator.image.tag="${kubean_version}" \
    --set kubeanAdmission.image.tag="${kubean_version}" \
    --set sprayJob.image.tag="${kubean_version}" \
    --set kubeanOperator.replicaCount=1 \
    --set kubeanAdmission.replicaCount=1 \
    --set kubeanOperator.operationsBackendLimit=20
}

install_kubean_chart "${KUBEAN_VERSION}"


# https://kubean-io.github.io/kubean/en/releases/artifacts/
DEFAULT_SPRAY_JOB_IMG='ghcr.m.daocloud.io/kubean-io/spray-job:v0.10.0'
RLS_2_21_SPRAY_JOB_IMG='ghcr.io/kubean-io/spray-job:2.21-d6f688f'
RLS_2_22_SPRAY_JOB_IMG='ghcr.io/kubean-io/spray-job:2.22-d65e4e6'
RLS_2_23_SPRAY_JOB_IMG='ghcr.io/kubean-io/spray-job:2.23-72da838'

RLS_2_21_AIRGAP_PATCH_IMG='ghcr.io/kubean-io/airgap-patch:2.21-d6f688f'
RLS_2_22_AIRGAP_PATCH_IMG='ghcr.io/kubean-io/airgap-patch:2.22-d65e4e6'
RLS_2_23_AIRGAP_PATCH_IMG='ghcr.io/kubean-io/airgap-patch:2.23-72da838'

PREPARE_IMGS=()
PREPARE_IMGS+=("${DEFAULT_SPRAY_JOB_IMG}")
PREPARE_IMGS+=("${RLS_2_21_SPRAY_JOB_IMG}")
PREPARE_IMGS+=("${RLS_2_22_SPRAY_JOB_IMG}")
PREPARE_IMGS+=("${RLS_2_23_SPRAY_JOB_IMG}")
PREPARE_IMGS+=("${RLS_2_21_AIRGAP_PATCH_IMG}")
PREPARE_IMGS+=("${RLS_2_22_AIRGAP_PATCH_IMG}")
PREPARE_IMGS+=("${RLS_2_23_AIRGAP_PATCH_IMG}")


# batch import images for spray-job & airgap-patch 
function batch_sync_images() {
  local registry=$1
  local images=$2
  for image_item in ${images}; do
    skopeo copy --insecure-policy -a --dest-tls-verify=false --retry-times=3 "docker://${image_item}" "docker://${registry}/${image_item}"
  done
}

batch_sync_images "${REGISTRY_ADDR}" "${PREPARE_IMGS[*]}"



# import default airgap images and binaries for kubean v0.10.0
function import_default_airgap_pkg() {
  local kubean_version=$1
  local registry_addr=$2
  local minio_addr=$3
  local minio_user=$4
  local minio_pass=$5

  wget "https://files.m.daocloud.io/github.com/kubean-io/kubean/releases/download/${kubean_version}/files-${ARCH}-${kubean_version}.tar.gz"
  wget "https://files.m.daocloud.io/github.com/kubean-io/kubean/releases/download/${kubean_version}/images-${ARCH}-${kubean_version}.tar.gz"
  wget "https://github.com/kubean-io/kubean/releases/download/${kubean_version}/os-pkgs-${DISTRO}-${kubean_version}.tar.gz"

  tar -zxvf "files-${ARCH}-${kubean_version}.tar.gz"
  pushd files || return
  MINIO_USER="${minio_user}" MINIO_PASS="${minio_pass}" ./import_files.sh "${minio_addr}"
  popd || return
  
  tar -zxvf "images-${ARCH}-${kubean_version}.tar.gz"
  pushd images || return
  REGISTRY_ADDR="${registry_addr}" ./import_images.sh
  popd || return

  tar -zxvf "os-pkgs-${DISTRO}-${kubean_version}.tar.gz"
  pushd os-pkgs || return
  MINIO_USER=${minio_user} MINIO_PASS=${minio_pass} ./import_ospkgs.sh "${minio_addr}" os-pkgs-${ARCH}.tar.gz
  popd || return
}


import_default_airgap_pkg "${KUBEAN_VERSION}" "${REGISTRY_ADDR}" "${MINIO_ADDR}" "${MINIO_USER}" "${MINIO_PASS}"


# generate v1.23.0 version of k8s airgap images and binaries
function export_airgap_patch_pkg() {
  local kube_version=$1
  local airgap_patch_image=$2

  mkdir -p "${CURRENT_PATH}/data"
  cat > "${CURRENT_PATH}/manifest.yml" <<EOF
image_arch:
  - "${ARCH}"
kube_version:
  - "${kube_version}"
EOF
  podman run -v "${CURRENT_PATH}/manifest.yml":/manifest.yml -v "${CURRENT_PATH}/data":/data -e ZONE=CN -e MODE=FULL "${airgap_patch_image}"
  ls -lh "${CURRENT_PATH}/data"
}

function import_airgap_patch_pkg() {
  local data_path=$1
  local registry_addr=$2
  local minio_addr=$3
  local minio_user=$4
  local minio_pass=$5
  pushd "${data_path}/${ARCH}/files" || return
  MINIO_USER="${minio_user}" MINIO_PASS="${minio_pass}" ./import_files.sh "${minio_addr}"
  popd || return

  pushd "${data_path}/${ARCH}/images" || return
  REGISTRY_ADDR="${registry_addr}" ./import_images.sh
  popd || return
}

mkdir -p /etc/containers/registries.conf.d
cat >/etc/containers/registries.conf.d/myregistry.conf <<EOF
[[registry]]
location="${REGISTRY_ADDR}"
insecure=true
EOF

export_airgap_patch_pkg "v1.23.0" "${REGISTRY_ADDR}/${RLS_2_21_AIRGAP_PATCH_IMG}"
import_airgap_patch_pkg "${CURRENT_PATH}/data" "${REGISTRY_ADDR}" "${MINIO_ADDR}" "${MINIO_USER}" "${MINIO_PASS}"
rm -rf "${CURRENT_PATH}/data" "${CURRENT_PATH}/manifest.yml"

export_airgap_patch_pkg "v1.24.0" "${REGISTRY_ADDR}/${RLS_2_21_AIRGAP_PATCH_IMG}"
import_airgap_patch_pkg "${CURRENT_PATH}/data" "${REGISTRY_ADDR}" "${MINIO_ADDR}" "${MINIO_USER}" "${MINIO_PASS}"
rm -rf "${CURRENT_PATH}/data" "${CURRENT_PATH}/manifest.yml"

export_airgap_patch_pkg "v1.25.6" "${REGISTRY_ADDR}/${RLS_2_21_AIRGAP_PATCH_IMG}"
import_airgap_patch_pkg "${CURRENT_PATH}/data" "${REGISTRY_ADDR}" "${MINIO_ADDR}" "${MINIO_USER}" "${MINIO_PASS}"
rm -rf "${CURRENT_PATH}/data" "${CURRENT_PATH}/manifest.yml"

export_airgap_patch_pkg "v1.25.0" "${REGISTRY_ADDR}/${RLS_2_22_AIRGAP_PATCH_IMG}"
import_airgap_patch_pkg "${CURRENT_PATH}/data" "${REGISTRY_ADDR}" "${MINIO_ADDR}" "${MINIO_USER}" "${MINIO_PASS}"
rm -rf "${CURRENT_PATH}/data" "${CURRENT_PATH}/manifest.yml"

export_airgap_patch_pkg "v1.26.0" "${REGISTRY_ADDR}/${RLS_2_23_AIRGAP_PATCH_IMG}"
import_airgap_patch_pkg "${CURRENT_PATH}/data" "${REGISTRY_ADDR}" "${MINIO_ADDR}" "${MINIO_USER}" "${MINIO_PASS}"
rm -rf "${CURRENT_PATH}/data" "${CURRENT_PATH}/manifest.yml"


function apply_cluster() {
  local registry_addr=$1
  local minio_addr=$2

  cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mini-1-hosts-conf
  namespace: kubean-system
data:
  hosts.yml: |
    all:
      hosts:
        node1:
          ip: ${TEST_NODE_ADDR}
          access_ip: ${TEST_NODE_ADDR}
          ansible_host: ${TEST_NODE_ADDR}
          ansible_connection: ssh
          ansible_user: ${TEST_NODE_USER}
          ansible_password: ${TEST_NODE_PASS}
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
  name: mini-1-vars-conf
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

    containerd_insecure_registries:
      "${registry_addr}": "http://${registry_addr}"

    # containerd_registries_mirrors:
    # - prefix: ${registry_addr}
    #   mirrors:
    #     - host: http://${registry_addr}
    #       capabilities: ["pull", "resolve"]
    #       skip_verify: true

    gcr_image_repo: "${registry_addr}/gcr.m.daocloud.io"
    kube_image_repo: "${registry_addr}/k8s.m.daocloud.io"
    docker_image_repo: "${registry_addr}/docker.m.daocloud.io"
    quay_image_repo: "${registry_addr}/quay.m.daocloud.io"
    github_image_repo: "${registry_addr}/ghcr.m.daocloud.io"
    files_repo: "${minio_addr}/kubean"

    ## Kubernetes components
    kubeadm_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
    kubectl_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
    kubelet_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"

    ## CNI Plugins
    cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"

    ## cri-tools
    crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"

    ## [Optional] etcd: only if you **DON'T** use etcd_deployment=host
    etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"

    # [Optional] Calico: If using Calico network plugin
    calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    # [Optional] Calico with kdd: If using Calico network plugin with kdd datastore
    calico_crds_download_url: "{{ files_repo }}/github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"

    # [Optional] Flannel: If using Falnnel network plugin
    flannel_cni_download_url: "{{ files_repo }}/kubernetes/flannel/{{ flannel_cni_version }}/flannel-{{ image_arch }}"

    # [Optional] helm: only if you set helm_enabled: true
    helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"

    # [Optional] crun: only if you set crun_enabled: true
    crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"

    # [Optional] kata: only if you set kata_containers_enabled: true
    kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"

    # [Optional] cri-dockerd: only if you set container_manager: docker
    cri_dockerd_download_url: "{{ files_repo }}/github.com/Mirantis/cri-dockerd/releases/download/v{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}.{{ image_arch }}.tgz"

    # [Optional] runc,containerd: only if you set container_runtime: containerd
    runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
    containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
    nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"

    # [Optional] runsc,containerd-shim-runsc: only if you set gvisor_enabled: true
    gvisor_runsc_download_url: "{{ files_repo }}/storage.googleapis.com/gvisor/releases/release/{{ gvisor_version }}/{{ ansible_architecture }}/runsc"
    gvisor_containerd_shim_runsc_download_url: "{{ files_repo }}/storage.googleapis.com/gvisor/releases/release/{{ gvisor_version }}/{{ ansible_architecture }}/containerd-shim-runsc-v1"

---
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
  name: cluster-mini-1
  labels:
    clusterName: cluster-mini-1
spec:
  hostsConfRef:
    namespace: kubean-system
    name: mini-1-hosts-conf
  varsConfRef:
    namespace: kubean-system
    name:  mini-1-vars-conf

EOF
}

apply_cluster "${REGISTRY_ADDR}" "${MINIO_ADDR}"

function apply_cluster_for_online() {
  cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mini-hosts-conf
  namespace: kubean-system
data:
  hosts.yml: |
    all:
      hosts:
        node1:
          ip: 172.30.41.162
          access_ip: 172.30.41.162
          ansible_host: 172.30.41.162
          ansible_connection: ssh
          ansible_user: root
          ansible_password: dangerous
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
  name: mini-vars-conf
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

    coredns_version: "{{ 'v1.9.3' if (kube_version is version('v1.25.0','>=')) else 'v1.8.6' }}"

    pkg_install_retries: 1
    download_retries: 1

    gcr_image_repo: "gcr.m.daocloud.io"
    kube_image_repo: "k8s.m.daocloud.io"
    docker_image_repo: "docker.m.daocloud.io"
    quay_image_repo: "quay.m.daocloud.io"
    github_image_repo: "ghcr.m.daocloud.io"
    files_repo: "https://files.m.daocloud.io"

    ## Kubernetes components
    kubeadm_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
    kubectl_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
    kubelet_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"

    ## CNI Plugins
    cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"

    ## cri-tools
    crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"

    ## [Optional] etcd: only if you **DON'T** use etcd_deployment=host
    etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"

    # [Optional] Calico: If using Calico network plugin
    calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    # [Optional] Calico with kdd: If using Calico network plugin with kdd datastore
    calico_crds_download_url: "{{ files_repo }}/github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"

    # [Optional] Flannel: If using Falnnel network plugin
    flannel_cni_download_url: "{{ files_repo }}/kubernetes/flannel/{{ flannel_cni_version }}/flannel-{{ image_arch }}"

    # [Optional] helm: only if you set helm_enabled: true
    helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"

    # [Optional] crun: only if you set crun_enabled: true
    crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"

    # [Optional] kata: only if you set kata_containers_enabled: true
    kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"

    # [Optional] cri-dockerd: only if you set container_manager: docker
    cri_dockerd_download_url: "{{ files_repo }}/github.com/Mirantis/cri-dockerd/releases/download/v{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}.{{ image_arch }}.tgz"

    # [Optional] runc,containerd: only if you set container_runtime: containerd
    runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
    containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
    nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"

    # [Optional] runsc,containerd-shim-runsc: only if you set gvisor_enabled: true
    gvisor_runsc_download_url: "{{ files_repo }}/storage.googleapis.com/gvisor/releases/release/{{ gvisor_version }}/{{ ansible_architecture }}/runsc"
    gvisor_containerd_shim_runsc_download_url: "{{ files_repo }}/storage.googleapis.com/gvisor/releases/release/{{ gvisor_version }}/{{ ansible_architecture }}/containerd-shim-runsc-v1"

---
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
  name: cluster-mini
  labels:
    clusterName: cluster-mini
spec:
  hostsConfRef:
    namespace: kubean-system
    name: mini-hosts-conf
  varsConfRef:
    namespace: kubean-system
    name:  mini-vars-conf

EOF
}

function apply_clear_cluster_operation() {
  local spray_job_image=$1

  cat << EOF | kubectl apply -f -
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-1-clear-ops
spec:
  cluster: cluster-mini-1
  image: ${spray_job_image}
  actionType: playbook
  action: reset.yml
EOF
}
apply_clear_cluster_operation "${REGISTRY_ADDR}/${RLS_2_21_SPRAY_JOB_IMG}"

function apply_debug_cluster_operation() {
  local spray_job_image=$1

  cat << EOF | kubectl apply -f -
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-1-debug-ops
spec:
  cluster: cluster-mini-1
  image: ${spray_job_image}
  actionType: shell
  action: sleep infinity
EOF
}

function apply_init_cluster_operation() {
  local spray_job_image=$1

  cat << EOF | kubectl apply -f -
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-1-init-ops
spec:
  cluster: cluster-mini-1
  image: ${spray_job_image}
  actionType: playbook
  action: disable-firewalld.yml
  preHook:
    - actionType: playbook
      action: ping.yml
EOF
}

apply_init_cluster_operation "${REGISTRY_ADDR}/${RLS_2_21_SPRAY_JOB_IMG}"

function apply_init_cluster_operation_for_online() {
  local spray_job_image=$1

  cat << EOF | kubectl apply -f -
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-preinstall-ops
spec:
  cluster: cluster-mini
  image: ${spray_job_image}
  actionType: playbook
  action: disable-firewalld.yml
  preHook:
    - actionType: playbook
      action: ping.yml
EOF
}

function apply_cluster_operation() {
  local operation=$1
  local kube_version=$2
  local spray_job_image=$3

  local main_playbook="cluster.yml"
  if [[ "${operation}" == "upgrade" ]]; then
    main_playbook="upgrade-cluster.yml"
  fi

  cat << EOF | kubectl apply -f -
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-1-$(echo "${kube_version}" | tr . -)-ops
spec:
  cluster: cluster-mini-1
  image: ${spray_job_image}
  actionType: playbook
  action: ${main_playbook}
  extraArgs: -e kube_version=${kube_version}
  postHook:
    - actionType: playbook
      action: cluster-info.yml

EOF
}

apply_cluster_operation "install" "v1.23.0" "${REGISTRY_ADDR}/${RLS_2_21_SPRAY_JOB_IMG}"

apply_cluster_operation "upgrade" "v1.24.0" "${REGISTRY_ADDR}/${RLS_2_21_SPRAY_JOB_IMG}"

apply_cluster_operation "upgrade" "v1.25.6" "${REGISTRY_ADDR}/${RLS_2_21_SPRAY_JOB_IMG}"

apply_cluster_operation "upgrade" "v1.25.0" "${REGISTRY_ADDR}/${RLS_2_22_SPRAY_JOB_IMG}"

apply_cluster_operation "upgrade" "v1.26.0" "${REGISTRY_ADDR}/${RLS_2_23_SPRAY_JOB_IMG}"

apply_cluster_operation "upgrade" "v1.27.0" "${REGISTRY_ADDR}/${DEFAULT_SPRAY_JOB_IMG}"
