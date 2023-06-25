# 离线安装包的原理

Kubean CI 构建便于在离线环境中安装 Kubernetes 的离线文件。

注：有关如何使用离线文件，请参阅[离线场景的使用](../usage/airgap.md)。

本页说明离线文件包的构建原理。

- 发版页面：[Github Releases](https://github.com/kubean-io/kubean/releases)
- 主要包含：

  |  文件包 | 描述  |
  |  -----------------------  | :---------------------  |
  | files-*.tar.gz  | Kubespray 安装所需的二进制文件：kubeadm、runc |
  | images-*.tar.gz  | K8s 集群镜像和 CNI 镜像  |
  | os-pkgs-${linux_distribution}-${tag}.tar.gz | K8s 安装期间所需的 deb/rpm |

## 如何构建离线文件

1. 镜像和二进制文件

    类似 Kubespray [离线部署指南](https://github.com/kubernetes-sigs/kubespray/blob/master/contrib/offline/README.md)所述，
    Kubespray 提供了一些脚本来生成镜像和二进制文件列表（感谢 Kubespray！）

    得益于 Kubespray [生成二进制和镜像列表的脚本](https://github.com/kubernetes-sigs/kubespray/blob/master/contrib/offline/generate_list.sh)，
    我们可以使用 [manage-offline-files.sh](https://github.com/kubernetes-sigs/kubespray/tree/master/contrib/offline#manage-offline-files.sh)
    下载这些二进制文件和镜像。
    随后 Kubean 提供了 [offline-build.sh](https://github.com/kubean-io/kubean/blob/main/.github/workflows/call-offline-build.yaml)
    将所有这些融合于一起。

2. os-packages (deb/rpm)

    在 K8s 安装期间，有些文件包不会随二进制文件一起安装，因此我们必须通过 deb/rpm 来安装这些文件包。
    [os packages 列表](https://github.com/kubean-io/kubean/blob/main/build/os-packages/packages.yml)定义了所涉及的文件包。

    Github Action 将通过 `dnf/apt` 构建不同操作系统的文件包（例如 Qemu），便于下载和归档 RPM/DEB 包。

3. CI 流程

    离线文件包由 [Github Action 脚本](https://github.com/kubean-io/kubean/tree/main/.github/workflows)生成和管理。
