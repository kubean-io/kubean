# 集群安装

### 安装时，KuBeanClusterOps 的关键参数：
```yaml
apiVersion: kubeanclusterops.kubean.io/v1alpha1
kind: KuBeanClusterOps
metadata:
  name: cluster1-ops-xxx
spec:
  ...
  actionType: playbook  # 执行 action 的类型
  action: cluster.yml   # 指定 cluster playbook
  ...
```

### kubespray 安装相关参数参考：
> 注： kubean 参数设置位于 var-conf-cm.yml

以下参数参考开叔的 [customize.yml](https://gist.github.com/yankay/a863cf2e300bff6f9040ab1c6c58fbae#file-k8s-customize-yml)
```yaml
# Normal
kube_version: "v1.21.6"
cluster_name: kubean.cluster
container_manager: docker
kube_network_plugin: calico

override_system_hostname: false
kube_proxy_mode: iptables
enable_nodelocaldns: false
etcd_deployment_type: kubeadm
metrics_server_enabled: true
local_path_provisioner_enabled: true

# Download Config
download_run_once: true
download_container: false
download_force_cache: true
download_localhost: true
# skip_downloads: true
# http_proxy: "socks5://10.6.108.200:10808"

# gcr and kubernetes image repo define
gcr_image_repo: "gcr.m.daocloud.io"
kube_image_repo: "k8s-gcr.m.daocloud.io"

# docker image repo define
docker_image_repo: "docker.m.daocloud.io"

# quay image repo define
quay_image_repo: "quay.m.daocloud.io"

# github image repo define (ex multus only use that)
github_image_repo: "ghcr.m.daocloud.io"

kubelet_download_url: "https://dl.k8s.io/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"
kubectl_download_url: "https://mirror.azure.cn/kubernetes/kubectl/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
kubeadm_download_url: "https://dl.k8s.io/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
etcd_download_url: "https://ghproxy.com/https://github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"
cni_download_url: "https://ghproxy.com/https://github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"
calicoctl_download_url: "https://ghproxy.com/https://github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
calicoctl_alternate_download_url: "https://ghproxy.com/https://github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
calico_crds_download_url: "https://ghproxy.com/https://github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"
crictl_download_url: "https://ghproxy.com/https://github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
helm_download_url: "https://mirror.azure.cn/kubernetes/helm/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"
runc_download_url: "https://ghproxy.com/https://github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
crun_download_url: "https://ghproxy.com/https://github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"
youki_download_url: "https://ghproxy.com/https://github.com/containers/youki/releases/download/v{{ youki_version }}/youki_v{{ youki_version | regex_replace('\\.', '_') }}_linux.tar.gz"
kata_containers_download_url: "https://ghproxy.com/https://github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"
nerdctl_download_url: "https://ghproxy.com/https://github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
krew_download_url: "https://ghproxy.com/https://github.com/kubernetes-sigs/krew/releases/download/{{ krew_version }}/krew-{{ host_os }}_{{ image_arch }}.tar.gz"
containerd_download_url: "https://ghproxy.com/https://github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
cri_dockerd_download_url: "https://ghproxy.com/https://github.com/Mirantis/cri-dockerd/releases/download/{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}-linux-{{ image_arch }}.tar.gz"
```

> 安装相关的更多可配参数见：[inventory/sample/group_vars](https://github.com/kubernetes-sigs/kubespray/tree/master/inventory/sample/group_vars) 目录；
