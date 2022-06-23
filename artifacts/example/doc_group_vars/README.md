# group_vars必要参数梳理

## 相关必须软件的加速配置

```
    # Download Config
    download_run_once: true
    download_container: false
    download_force_cache: true
    download_localhost: true

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

* 其中的docker是基于linux repo安装(apt/yum)，不同于containerd通过压缩包解压后即可安装
* `calicoctl_download_url`和`calicoctl_alternate_download_url`是因为calicoctl在某一版本后的git仓库地址变化了
* multus功能需要镜像在`github_image_repo`镜像仓库里

## CRI参数设置

### docker

```
container_manager: docker
```

* k8s版本低于1.24可以如此设置

### docker + cri_dockerd

```
    container_manager: docker
    cri_dockerd_enabled: true
    cri_dockerd_version: v0.2.0
```

* k8s1.24移除docker-shim组件，则k8s不能直接与docker engine集成
* cri_dockerd类似于docker-shim的代替品

### containerd

```
container_manager: containerd
etcd_deployment_type: host
```

* 当cri为containerd时，etcd只能设置为host类型

## CNI插件参数设置

### cilium

```
kube_network_plugin: cilium
cilium_version: v1.11.3
```

### calico

```
kube_network_plugin: calico
calico_version: "v3.20.4"
enable_dual_stack_networks: false #单栈
kube_pods_subnet: 10.244.0.0/16
calico_vxlan_mode: Always
calico_ipip_mode: Never
calico_iptables_backend: "NFT"
calico_pool_name: "default-pool-ipv4"

### other ipv6 setting
## 略
```

* 以上配置开启ipv4的配置
* kube_pods_subnet是配置ipv4的subnet分配范围
    * kube_pods_subnet是一个通用设置，并不是专属于calico的设置
* vxlan和ipip模式只能二选一
* calico_iptables_backend默认是Auto，但是在centos8上必须设置为NFT，否则在不同节点上的pod不能联通
* calico_pool_name是`ipv4 pool`的名称

### calico + multus

```
kube_network_plugin: calico
calico_version: "v3.20.4"
...
kube_network_plugin_multus: true
multus_version: "v3.8"
```

* multus是pod多网卡方案的一种实现

## 审计日志参数设置

```
kubernetes_audit: true
# path to audit log file
# audit_log_path: /var/log/audit/kube-apiserver-audit.log
# num days
audit_log_maxage: 30
# the num of audit logs to retain
audit_log_maxbackups: 1
# the max size in MB to retain
audit_log_maxsize: 100
```

## 集群LCM

### 新建集群

```clusterOps.Spec
  actionType: playbook
  action: cluster.yml
```

### 铲除集群

```clusterOps.Spec
  actionType: playbook
  action: reset.yml
```

### 升级集群

```clusterOps.Spec
  actionType: playbook
  action: upgrade-cluster.yml
```

## Worker Node LCM

### add worker node

```clusterOps
  actionType: playbook
  action: scale.yml
  extraArgs: --limit=node2
```

* 如果是多个节点则使用逗号分隔，比如`--limit=node2,node3`

### remove worker node

```clusterOps
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2,node3
```

### ETCD类型

* 在`container_manager: containerd`时,etcd_deployment_type只能是host类型
* etcd_deployment_type可以有host docker 和 kubeadm 三个选项
