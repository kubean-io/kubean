# Use Kubean in offline scenes

## Preparation

1. Services requiring pre-deployment:
* Document resource service [`minio`](https://docs.min.io/docs/minio-quickstart-guide.html)
* Image registry services: [`docker registry`](https://hub.docker.com/_/registry)（2.7 below）
  or [`harbor`](https://goharbor.io/docs/2.0.0/install-config/)

2. Necessary tools to be installed:

* A tool for importing image files: [`skopeo`](https://github.com/containers/skopeo/blob/main/install.md), Required >=1.9.2
* A tool for importing binary files: [`minio client`](https://docs.min.io/docs/minio-client-quickstart-guide.html)

3. Deploy Kubean by Helm[`kubean`](https://github.com/kubean-io/kubean/blob/main/charts/kubean/README.md)


## Download offline resources

The [Github Releases](https://github.com/kubean-io/kubean/releases) page allows us to download the offline resources for the version we want

Basic instructions for offline resources:

``` bash
├── files.list                                  # List of file contents
├── files-${tag}.tar.gz                         # File compression package, including import scripts
├── images.list                                 # List of image contents
├── images-${tag}.tar.gz                        # File compression package, including import scripts
└── os-pkgs-${linux_distribution}-${tag}.tar.gz # Compressed package of each system, including import script
```

## Importing offline resources into the corresponding service

### 1. Import binary resources

Please first unzip the `files-${tag}.tar.gz` file, which contains:

``` bash
files/
├── import_files.sh       # This script is used to import binary files into the minio file service
└── offline-files.tar.gz  # Compressed package of binary 
```

Execute the following command to import the binary file into the minio service:

``` bash
$ MINIO_USER=${username} MINIO_PASS=${password} ./import_files.sh ${minio_address}
```

* `minio_address` is the `minio API Server` address, typically on port 9000, e.g. `http://1.2.3.4:9000`

### 2. Images  import of resources

You need to unzip the `images-${tag}.tar.gz` file, which contains:

``` bash
images/
├── import_images.sh       # This script is used to import binary files into the minio file service
└── offline-images.tar.gz  # Compressed package of image file
```

Execute the following command to import the image file into the Docker Registry or the Harbor image repository service:

``` bash
# 1. Non-secure password-free mode
$ DEST_TLS_VERIFY=false ./import_images.sh ${registry_address}

# 2. Username password mode
$ DEST_USER=${username} DEST_PASS=${password} ./import_images.sh ${registry_address}
```

* When `DEST_TLS_VERIFY=false`, the image is uploaded in non-secure HTTP mode
* `DEST_USER` and `DEST_PASS` need to be set when username and password authentication exists for the mirror repository
* `registry_address` is the address of the mirror repository, e.g. `1.2.3.4:5000`

### 3. OS packages import of resources

Note: 
- [OS Packages](https://github.com/kubean-io/kubean/blob/main/build/os-packages/README.md) resources for Centos / Redhat / Kylin / Ubuntu distributions are currently supported.
- The OS Package of UnionTech V20 series needs to be built manually, see [README](https://github.com/kubean-io/kubean/blob/main/build/os-packages/others/uos_v20/README.md) for the build method.

You need to unzip the `os-pkgs-${linux_distribution}-${tag}.tar.gz` file, which contains:

``` bash
os-pkgs
├── import_ospkgs.sh       # This script is used to import os packages to the minio file service
├── os-pkgs-amd64.tar.gz   # os packages for the amd64 architecture
├── os-pkgs-arm64.tar.gz   # os packages for the arm64 architecture
└── os-pkgs.sha256sum.txt  # The sha256sum checksum file for the os packages
```

Execute the following command to os packages into the minio file service:

``` bash
$ MINIO_USER=${username} MINIO_PASS=${password} ./import_ospkgs.sh ${minio_address} os-pkgs-${arch}.tar.gz
```

## Create offline sources

### Create ISO image source
The following [Create local ISO image source] and [Create online ISO image source] only need to execute one of them.

#### 1.1 Create a local ISO image source

OS Packages are primarily used to resolve docker-ce installation dependencies, but for offline deployments, other packages from the distribution may be used, and a local ISO image source will need to be created.

> Note: We need to download the ISO system distribution image for the host in advance, currently only the ISO image source for the Centos distribution is supported;
> Note: This operation needs to be performed on each cluster that creates kubernetes nodes;

The script `artifacts/gen_repo_conf.sh` can be used to mount the ISO image and create the Repo configuration file by executing the following command:

``` bash
# Basic format
$ ./gen_repo_conf.sh --iso-mode ${linux_distribution} ${iso_image_file}

# Execute the script to create the ISO image source
$ ./gen_repo_conf.sh --iso-mode centos CentOS-7-x86_64-Everything-2207-02.iso
# Check ISO image mounts
$ df -h | grep mnt
/dev/loop0               9.6G  9.6G     0 100% /mnt/centos-iso
# Check ISO image source configuration
$ cat /etc/yum.repos.d/Kubean-ISO.repo
[kubean-iso]
name=Kubean ISO Repo
baseurl=file:///mnt/centos-iso
enabled=1
gpgcheck=0
sslverify=0
```

#### 1.2 Create an online ISO image source

To import the image source from the ISO into the minio server, use the script `artifacts/import_iso.sh`

```bash
MINIO_USER=${username} MINIO_PASS=${password} ./import_iso.sh ${minio_address} Centos-XXXX.ISO
```

Create the following file for the host `/etc/yum.repos.d/centos-iso-online.repo` to use the online ISO image source:

```
[kubean-iso-online]
name=Kubean ISO Repo Online
baseurl=${minio_address}/kubean/centos-iso/$releasever/os/$basearch
enabled=1
gpgcheck=0
sslverify=0
```

* Need to replace `${minio_address}` with the minio API Server address

### 2. Create extras software sources

> Currently only supported on Centos distributions

When installing a K8S cluster, it also relies on extras, such as `container-selinux`, which are not always provided in the ISO image source. This is supplemented by the OS packages offline package, which requires us to create an extra repo configuration file for each node after importing minio.

The Extra Repo can also be created using the script `artifacts/gen_repo_conf.sh`, by executing the following command:

``` bash
$ ./gen_repo_conf.sh --url-mode ${linux_distribution} ${repo_base_url}

# Execute the script to create a URL source profile
$ ./gen_repo_conf.sh --url-mode centos ${minio_address}/kubean/centos/\$releasever/os/\$basearch
# View URL source profile
$ cat /etc/yum.repos.d/Kubean-URL.repo
[kubean-extra]
name=Kubean Extra Repo
baseurl=http://10.20.30.40:9000/kubean/centos/$releasever/os/$basearch
enabled=1
gpgcheck=0
sslverify=0
```

> Note: If the `repo_base_url` parameter has a `$` symbol, it needs to be escaped `\$`

> Need to replace `${minio_address}` with the actual `minio API Server` address

### 3. ClusterOperation combined with playbook to create source profiles

> Only Centos yum repo additions are currently supported

As the process of creating a source involves all the nodes in the cluster, manual scripting is relatively tedious, so a playbook solution is provided here:

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
      action: enable-repo.yml  # Before deploying the cluster, run the enable-repo playbook to create a source configuration for each node with the specified url
      extraArgs: |
        -e "{repo_list: ['http://10.20.30.40:9000/kubean/centos/\$releasever/os/\$basearch']}"
    - actionType: playbook
      action: disable-firewalld.yml
  postHook:
    - actionType: playbook
      action: cluster-info.yml
    - actionType: playbook
      action: enable-repo.yml  # After deploying the cluster, restore the yum repo configuration for each node. (Note: This step can be added as appropriate.)
      extraArgs: |
        -e undo=true
```

## Pre-deployment cluster configuration

Offline settings need to be referred to [`kubespray`](https://github.com/kubernetes-sigs/kubespray)
Located in `kubespray/inventory/sample/group_vars/all/offline.yml` configuration file:

``` yaml
---
## Global offline configuration
### Configure the address of the private container image repository service
registry_host: "{{ registry_address }}"

### Configuring the address of the binary file service
files_repo: "{{ minio_address }}/kubean"

### If you are using CentOS / RedHat / AlmaLinux / Fedora, you need to configure the yum source file service address:
yum_repo: "{{ minio_address }}"

### If using Debian, configure:
debian_repo: "{{ minio_address }}"

### If using Ubuntu, configure:
ubuntu_repo: "{{ minio_address }}"

### If containerd uses a non-secure HTTP authentication-free method, it needs to be configured:
containerd_insecure_registries:
  "{{ registry_address }}": "http://{{ registry_address }}"

### Required if docker uses non-secure HTTP authentication-free methods:
docker_insecure_registries:
  - {{ registry_address }}

## Kubernetes components
kubeadm_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
kubectl_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
kubelet_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"

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

For offline deployment, additional parameters are required for some special operating systems:

|  OS | Additional parameters  |
|  ------  | :-----  |
| RHEL Series  | `rhel_enable_repos: false` |
| Oracle Linux Series  | `use_oracle_public_repo: false` |

We use `examples/install/3.airgap` as a template,

Adapt the offline configuration as above to your specific situation, especially if you need to replace `<registry_address>` and `<minio_address>`,

Finally add the configuration update to the `examples/install/3.airgap/VarsConfCM.yml`  file,

We also need to change the cluster node IP and username password in `examples/install/3.airgap/HostsConfCM.yml`,

Finally, the ClusterOperation task is started with `kubectl apply -f examples/install/3.airgap` to install the k8s cluster.

## Generation and use of incremental offline packages

For detailed documentation see: [Air gap patch usage](airgap_patch_usage.md).
