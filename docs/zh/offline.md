# 离线场景的使用

## 提醒事项

- 针对 RHEL 8.4 系列，由于包依赖问题，执行过程中会卸载系统预装的 fuse 包

## 准备事项

1. 需要预先部署的服务:
* 文件资源服务 [`minio`](https://docs.min.io/docs/minio-quickstart-guide.html)
* 镜像仓库服务 [`docker registry`](https://hub.docker.com/_/registry)（2.7 以下）
  或者 [`harbor`](https://goharbor.io/docs/2.0.0/install-config/)

2. 需要安装的必要工具:

* 用于导入镜像文件的工具: [`skopeo`](https://github.com/containers/skopeo/blob/main/install.md)，需要 >=1.9.2
* 用于导入二进制文件的工具: [`minio client`](https://docs.min.io/docs/minio-client-quickstart-guide.html)

3. 通过Helm部署[`kubean`](https://github.com/kubean-io/kubean/blob/main/charts/kubean/README.md)


## 下载离线资源

通过 [Github Releases](https://github.com/kubean-io/kubean/releases) 页面可以下载我们想要版本的离线资源

离线资源的基本说明:

``` bash
├── files.list                                  # 文件内容的列表
├── files-${tag}.tar.gz                         # 文件压缩包, 内含导入脚本
├── images.list                                 # 镜像内容的列表
├── images-${tag}.tar.gz                        # 镜像压缩包, 内含导入脚本
└── os-pkgs-${linux_distribution}-${tag}.tar.gz # 各系统压缩包, 内含导入脚本
```

## 将离线资源导入对应服务

### 1. Binaries 资源的导入

请先解压 `files-${tag}.tar.gz` 文件, 其内部包含:

``` bash
files/
├── import_files.sh       # 该脚本用于导入二进制文件到 minio 文件服务
└── offline-files.tar.gz  # 二进制文件的压缩包
```

执行如下命令, 将二进制文件导入到 minio 服务中:

``` bash
$ MINIO_USER=${username} MINIO_PASS=${password} ./import_files.sh ${minio_address}
```

* `minio_address` 是 `minio API Server`地址，端口一般为9000，比如 `http://1.2.3.4:9000`

### 2. Images 资源的导入

需要解压 `images-${tag}.tar.gz` 文件, 其内部包含:

``` bash
images/
├── import_images.sh       # 该脚本用于导入镜像文件到 docker registry 或 harbor 镜像仓库服务
└── offline-images.tar.gz  # 镜像文件的压缩包
```

执行如下命令, 将镜像文件导入到 docker registry 或 harbor 镜像仓库服务中:

``` bash
# 1. 非安全免密模式
$ DEST_TLS_VERIFY=false ./import_images.sh ${registry_address}

# 2. 用户名口令模式
$ DEST_USER=${username} DEST_PASS=${password} ./import_images.sh ${registry_address}
```

* 当 `DEST_TLS_VERIFY=false`, 此时采用非安全 HTTP 模式上传镜像
* 当镜像仓库存在用户名密码验证时，需要设置 `DEST_USER` 和 `DEST_PASS`
* `registry_address` 是镜像仓库的地址，比如`1.2.3.4:5000`

### 3. OS packages 资源的导入

> 注: 当前仅支持 Centos 发行版的 OS Packages 资源

需要解压 `os-pkgs-${linux_distribution}-${tag}.tar.gz` 文件, 其内部包含:

``` bash
os-pkgs
├── import_ospkgs.sh              # 该脚本用于导入 os packages 到 minio 文件服务
├── os-pkgs-amd64.tar.gz   # amd64 架构的 os packages 包
├── os-pkgs-arm64.tar.gz   # arm64 架构的 os packages 包
└── os-pkgs.sha256sum.txt  # os packages 包的 sha256sum 效验文件
```

执行如下命令, 将 os packages 包到 minio 文件服务中:

``` bash
$ MINIO_USER=${username} MINIO_PASS=${password} ./import_ospkgs.sh ${minio_address} os-pkgs-${arch}.tar.gz
```

## 建立离线源

### 1. 建立本地 ISO 镜像源

OS Packages 主要用于解决 docker-ce 的安装依赖, 但在实际的离线部署过程中, 可能还需要使用到发行版系统的其他包, 此时需要建立本地
ISO 镜像源.

> 注: 我们需要提前下载主机对应的 ISO 系统发行版镜像, 当前仅支持 Centos 发行版的 ISO 镜像源创建;

这里可以使用脚本 `artifacts/gen_repo_conf.sh`, 执行如下命令即可挂载 ISO 镜像文件, 并创建 Repo 配置文件:

``` bash
# 基本格式
$ ./gen_repo_conf.sh --iso-mode ${linux_distribution} ${iso_image_file}

# 执行脚本创建 ISO 镜像源
$ ./gen_repo_conf.sh --iso-mode centos CentOS-7-x86_64-Everything-2207-02.iso
# 查看 ISO 镜像挂载情况
$ df -h | grep mnt
/dev/loop0               9.6G  9.6G     0 100% /mnt/centos-iso
# 查看 ISO 镜像源配置
$ cat /etc/yum.repos.d/Kubean-ISO.repo
[kubean-iso]
name=Kubean ISO Repo
baseurl=file:///mnt/centos-iso
enabled=1
gpgcheck=0
sslverify=0
```

#### 1.1. 建立在线 ISO 镜像源

将 ISO 中的镜像源导入到minio server中，需要使用到脚本 `artifacts/import_iso.sh` ，执行如下面命令即可将 ISO 镜像中软件源导入到
minio server 中

```bash
MINIO_USER=${username} MINIO_PASS=${password} ./import_iso.sh ${minio_address} Centos-XXXX.ISO
```

为主机新建如下文件 `/etc/yum.repos.d/centos-iso-online.repo` 即可使用在线 ISO 镜像源:

```
[kubean-iso-online]
name=Kubean ISO Repo Online
baseurl=${minio_address}/kubean/centos-iso/$releasever/os/$basearch
enabled=1
gpgcheck=0
sslverify=0
```

此外，如果导入的是 RHEL ISO，需注意此 ISO 提供两个源：

```
[kubean-iso-online-BaseOS]
name=Kubean ISO Repo Online BaseOS
baseurl=${minio_address}/kubean/redhat-iso/$releasever/os/$basearch/BaseOS
enabled=1
gpgcheck=0
sslverify=0


[kubean-iso-online-AppStream]
name=Kubean ISO Repo Online AppStream
baseurl=${minio_address}/kubean/redhat-iso/$releasever/os/$basearch/AppStream
enabled=1
gpgcheck=0
sslverify=0
```

* 需要将 `${minio_address}` 替换为 minio API Server 地址

### 2. 建立 extras 软件源

> 当前仅支持 Centos 发行版

在安装 K8S 集群时, 还会依赖一些 extras 软件, 比如 `container-selinux`, 这些软件往往在 ISO 镜像源中并不提供. 对此 OS
packages 离线包已对其进行了补充, 其在导入 minio 之后,
我们还需要向各个节点创建 extra repo 配置文件.

同样可以使用脚本 `artifacts/gen_repo_conf.sh`, 执行如下命令即可创建 Extra Repo:

``` bash
$ ./gen_repo_conf.sh --url-mode ${linux_distribution} ${repo_base_url}

# 执行脚本创建 URL 源配置文件
$ ./gen_repo_conf.sh --url-mode centos ${minio_address}/kubean/centos/\$releasever/os/\$basearch
# 查看 URL 源配置文件
$ cat /etc/yum.repos.d/Kubean-URL.repo
[kubean-extra]
name=Kubean Extra Repo
baseurl=http://10.20.30.40:9000/kubean/centos/$releasever/os/$basearch
enabled=1
gpgcheck=0
sslverify=0
```

> 注: 若 `repo_base_url` 参数中带有 `$` 符号, 需要对其进行转义 `\$`

> 需要将 `${minio_address}` 替换为实际 `minio API Server` 的地址

### 3. ClusterOperation 结合 playbook 创建源配置文件

> 当前仅支持 Centos yum repo 的添加

由于创建源的过程涉及到集群的所有节点, 手动脚本操作相对繁琐, 这里提供了一种 playbook 的解决方式:

``` yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-ops-01
spec:
  cluster: sample
  image: ghcr.io/kubean-io/spray-job:latest
  backoffLimit: 0
  actionType: playbook
  action: cluster.yml
  preHook:
    - actionType: playbook
      action: ping.yml
    - actionType: playbook
      action: enable-repo.yml  # 在部署集群前, 先执行 enable-repo 的 playbook, 为每个节点创建指定 url 的源配置
      extraArgs: |
        -e "{yum_repo_url_list: ['http://10.20.30.40:9000/kubean/centos/\$releasever/os/\$basearch']}"
    - actionType: playbook
      action: disable-firewalld.yml
  postHook:
    - actionType: playbook
      action: cluster-info.yml
    - actionType: playbook
      action: enable-repo.yml  # 在部署集群后, 还原各节点 yum repo 配置. (注：此步骤, 可视情况添加.)
      extraArgs: |
        -e undo=true
```

## 部署集群前的配置

离线设置需要参考 [`kubespray`](https://github.com/kubernetes-sigs/kubespray)
位于 `kubespray/inventory/sample/group_vars/all/offline.yml` 的配置文件:


``` yaml
---
## 全局的离线配置
### 配置私有容器镜像仓库服务的地址
registry_host: "{{ registry_address }}"

### 配置二进制文件服务的地址
files_repo: "{{ minio_address }}/kubean"

### 如果使用 CentOS / RedHat / AlmaLinux / Fedora, 需要配置 yum 源文件服务地址:
yum_repo: "{{ minio_address }}"

### 如果使用 Debian, 则配置:
debian_repo: "{{ minio_address }}"

### 如果使用 Ubuntu, 则配置:
ubuntu_repo: "{{ minio_address }}"

### 如果 containerd 采用非安全 HTTP 免认证方式, 则需要配置:
containerd_insecure_registries:
  "{{ registry_address }}": "http://{{ registry_address }}"

### 如果 docker 采用非安全 HTTP 免认证方式, 则需要配置:
docker_insecure_registries:
  - {{ registry_address }}

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

```

**额外说明：** 对于 RHEL 系列的离线安装，需添加配置 `rhel_enable_repos: false`

我们以 `artifacts/offlineDemo` 作为模板,

将如上离线配置按照具体情况进行调整, 特别需要替换`{{ registry_address }}` 和 `{{ minio_address }}`,

最终将配置添加更新到 `artifacts/offlineDemo/vars-conf-cm.yml` 文件中,

同时我们还需要修改 `artifacts/offlineDemo/hosts-conf-cm.yml` 中的集群节点 IP 及用户名密码,

最终, 通过 `kubectl apply -f artifacts/offlineDemo` 启动 ClusterOperation 任务来安装 k8s 集群.

## 增量离线包的生成和使用

详细文档见: [Air gap patch usage](airgap_patch_usage.md).
