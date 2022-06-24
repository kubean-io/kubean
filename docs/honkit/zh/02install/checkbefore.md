# 集群创建前准备

本页介绍集群安装前准备事项。

# 部署前要求

- 可安装集群的 Kubernetes 最低版本为 v1.22

- 集群节点上必须安装如下工具：Ansible v2.11+, Jinja 2.11+ and python-netaddr 

- 如在线安装，请保障网络的联通性；如离线安装请参考 [离线安装要求]()

- 目标安装节点请 开启 **IPv4 forwarding**

- 如需给 Pods 或 Service 配置 IPv6 地址，请开启 **IPv6 forwarding**

- 如开启防火墙规则，请自行管理防火墙规则， 为避免安装过程中，防火墙规则对集群部署造成干扰，建议关闭防火墙

- 如果需要使用 non-root 账号进行部署，请在目标节点上提升权限。

- 如使用 用户名/密码部署集群，请确保所有目标节点上的用户名/密码保持一致。

- 如需部署超大规模集群，请参考：[部署大规模集群](https://kubernetes.io/docs/setup/cluster-large/#size-of-master-and-master-components) 

  

# 节点系统及硬件要求

**控制器节点**

| 硬件要求       | 要求                                                         | 生产推荐 | 备注                                                         |
| -------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| 系统版本       | 1. **Debian** Bullseye, Buster, Jessie, Stretch <br />2.**Ubuntu** 16.04, 18.04, 20.04, 22.04 <br />3. **CentOS/RHEL** 7, [8](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/centos.md#centos-8)<br />4. **Fedora** 34, 35<br />5. **Fedora CoreOS** (see [fcos Note](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/fcos.md)) <br />6. **openSUSE** Leap 15.x/Tumbleweed <br />7.**Oracle Linux** 7, [8](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/centos.md#centos-8) <br /> |          |                                                              |
| CPU 架构       | X86 架构<br/>ARM 架构                                        |          |                                                              |
| CPU 要求       | CPU ≥ 4 Core                                                 | 8 Core   |                                                              |
| 内存要求<br /> | 内存≥ 1.5 G                                                  | 4G       |                                                              |
| 磁盘           | 至少两块磁盘：<br />一块系统盘<br />一块数据盘（用于 overlay2 模式，可用容量至少 100 GB） |          | 用于 Docker Storage 的 overlay2 模式。<br />CentOS/Redhat 生产环境必须具备，Ubuntu 不需要。 |
| 网卡           | 至少 1 张网卡                                                |          |                                                              |

**工作节点**

| 硬件要求       | 要求                                                         | 生产推荐 | 备注                                                         |
| -------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| 系统版本       | 1. **Debian** Bullseye, Buster, Jessie, Stretch <br />2.**Ubuntu** 16.04, 18.04, 20.04, 22.04 <br />3. **CentOS/RHEL** 7, [8](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/centos.md#centos-8)<br />4. **Fedora** 34, 35<br />5. **Fedora CoreOS** (see [fcos Note](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/fcos.md)) <br />6. **openSUSE** Leap 15.x/Tumbleweed <br />7.**Oracle Linux** 7, [8](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/centos.md#centos-8) <br /> |          |                                                              |
| CPU 架构       | X86 架构<br/>ARM 架构                                        |          |                                                              |
| CPU 要求       | CPU ≥ 4 Core                                                 | 8 Core   |                                                              |
| 内存要求<br /> | 内存≥ 1.5 G                                                  | 4G       |                                                              |
| 磁盘           | 至少两块磁盘：<br />一块系统盘<br />一块数据盘（用于 overlay2 模式，可用容量至少 100 GB） |          | 用于 Docker Storage 的 overlay2 模式。<br />CentOS/Redhat 生产环境必须具备，Ubuntu 不需要。 |
| 网卡           | 至少 1 张网卡                                                |          |                                                              |

