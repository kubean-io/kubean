# 使用加速模式部署集群

## 前置条件

1. 您已拥有一个标准 kubernetes 集群或云厂商提供的集群。
2. 集群控制节点或云终端已将安装了 [kubectl 工具](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)。
3. [kubean helm chart](helm-install-kubean.md) 已在您的集群上部署。
4. [kubean 项目](https://github.com/kubean-io/kubean)已经克隆至您本地，如果您还未克隆 kubean，可以执行执行如下命令进行克隆：

```bash
$ git clone https://github.com/kubean-io/kubean.git
```

---

## 开始部署

本教程将使用您克隆至本地的 `kubean/example/2.mirror` 文件作为范例模版，进行集群加速部署演示。

在加速部署模版 `2.mirror` 内已经内置了加速参数配置，您只需要将 `/2.mirror`文件路径下的 **`HostsConfCM.yml` ** 和 **`ClusterOperation.yml`** 两个配置模版文件内的主机等信息改成您的真实参数。

<details open>
<summary> 2.mirror` 文件内主要的配置文件及用途如下：</summary>

```yaml
    .2.mirror
    ├── Cluster.yml                        # 待建集群信息的抽象
    ├── ClusterOperation.yml        # kubean 版本及任务配置
    ├── HostsConfCM.yml              # 待建集群的节点信息配置
    └── VarsConfCM.yml                # 加速等它特性配置
```
</details>

#### 1. 配置主机配置参数 HostsConfCM.yml
进入 `kubean/examples/install/2.mirror/` 路径，编辑待建集群节点配置信息模版 `HostsConfCM.yml`，将下列参数替换为您的真实参数：

  - `<IP1>`：节点 IP。
  - `<USERNAME>`：登陆节点的用户名，建议使用 root 或具有 root 权限的用户登陆。
  - `<PASSWORD>`：登陆节点的密码。

例如，下面展示了一个 HostsConfCM.yml 示例：
<details>
<summary> HostsConfCM.yml 示例</summary>
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: online-hosts-conf
  namespace: kubean-system
data:
  hosts.yml: |
    all:
      hosts:
        node1:
          ip: 10.6.175.10 # 你的节点 IP
          access_ip: 10.6.175.10 # 你的节点 IP
          ansible_host: 10.6.175.10 # 你的节点 IP
          ansible_connection: ssh
          ansible_user: root # 登陆节点的用户名
          ansible_password: password01 # 登陆节点的密码
        node2:
          ip: 10.6.175.20 # 节点 2 的 IP
          access_ip: 10.6.175.20 # 节点 2 IP
          ansible_host: 10.6.175.20 # 节点的 2 IP
          ansible_connection: ssh
          ansible_user: root # 登陆节点 2 的用户名
          ansible_password: password01 # 登陆节点 2 的密码
      children:
        kube_control_plane: # 配置集群控制节点
          hosts:
            node1:
        kube_node: # 配置集群工作节点
          hosts:
            node1:
            node2:
        etcd: # 配置集群 ETCD 节点
          hosts:
            node1:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
```
</details>

执行如下命令编辑 HostsConfCM.yml 配置模版：

```bash
$ vi kubean/examples/install/2.mirror/HostsConfCM.yml
```

#### 2. 配置 kubean 任务配置参数 ClusterOperation.yml

进入 `kubean/examples/install/2.mirror/` 路径，编辑待建集群节点配置信息模版 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

例如，下面展示了一个 ClusterOperation.yml 示例：
<details>
<summary> ClusterOperation.yml 示例</summary>
```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster1-online-install-ops
spec:
  cluster: cluster1-online
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2 # kubean 镜像版本
  backoffLimit: 0
  actionType: playbook
  action: cluster.yml
  preHook:
    - actionType: playbook
      action: ping.yml
    - actionType: playbook
      action: disable-firewalld.yml
  postHook:
    - actionType: playbook
      action: kubeconfig.yml
    - actionType: playbook
      action: cluster-info.yml
```
</details>

执行如下命令编辑 ClusterOperation.yml 配置模版：

```bash
$ vi kubean/examples/install/2.mirror/ClusterOperation.yml
```

#### 3.应用 2.mirror 文件下所有的配置

完成上述步骤并保存 HostsConfCM.yml 和 ClusterOperation.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/install/2.mirror
```

至此，您已经使用加速模式完成了一个集群的部署。