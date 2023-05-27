# 使用 kubean 扩缩容集群工作节点

在软件开发运维的过程中，业务的发展往往需要添加集群的工作节点以满足业务增长，对于使用 kubean 部署的集群，在 kubean 中我们可以使用声明式的方式，快速扩缩容集群工作节点。

在您克隆至本地的 `kubean/example/scale` 文件内，同样提供了工作节点扩缩容的样例模版：

<details open>
<summary> sacle 文件内主要的配置文件及用途如下：</summary>

```yaml
    scale
    ├── 1.addWorkNode                             # 增加工作节点模版
    │   ├── ClusterOperation.yml                       # kubean 版本及任务配置
    │   └── HostsConfCM.yml                            #当前集群的节点信息配置
    └── 2.delWorkNode                             # 删除工作节点模版
    │   ├── ClusterOperation.yml                       # kubean 版本及任务配置
    │   └── HostsConfCM.yml                             #当前集群的节点信息配置
```
</details>

观察伸缩配置模版 `scale` 文件可以发现，对集群工作节点进行扩缩容只需执行 `HostsConfCM.yml` 和 `ClusterOperation.yml` 两个配置文件，并将新增节点信息等参数改成替换成您的真实参数。

下面以[使用 all-in-one 模式部署的单节点集群](./all-in-one-install.md)为例，来演示集群节点的扩缩容操作。

## 扩容工作节点

#### 1. 配置主机配置参数 HostsConfCM.yml

进入 `kubean/examples/scale/1.addWorkNode/` 路径，编辑待建集群节点配置信息模版 `HostsConfCM.yml`，将下列参数替换为您的真实参数：

  - `<IP2>`：节点 IP。
  - `<USERNAME>`：登陆节点的用户名，建议使用 root 或具有 root 权限的用户登陆。
  - `<PASSWORD>`：登陆节点的密码。

`kubean/examples/scale/1.addWorkNode/` 路径下 **`HostsConfCM.yml`** 的模版内容如下：

```yaml
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
          ip: <IP1>
          access_ip: <IP1>
          ansible_host: <IP1>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
        node2:
          ip: <IP2>
          access_ip: <IP2>
          ansible_host: <IP2>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
      children:
        kube_control_plane:
          hosts:
            node1:
        kube_node:
          hosts:
            node1:
            node2:
        etcd:
          hosts:
            node1:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
```
**重要参数：**
>* `all.hosts.node1` 指的是集群中已存在的节点信息。
>* `all.hosts.node2` 指的是集群中待新增节点信息。
>* `all.children.kube_node.hosts` 集群内所有节点名称集合

例如，下面展示了一个 HostsConfCM.yml 示例：
<details>
<summary> HostsConfCM.yml 示例</summary>
``` yaml
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
          ip: 10.6.175.10 # 你的节点 IP
          access_ip: 10.6.175.10 # 你的节点 IP
          ansible_host: 10.6.175.10 # 你的节点 IP
          ansible_connection: ssh
          ansible_user: root # 登陆节点的用户名
          ansible_password: password01 # 登陆节点的密码
        node2:
          ip: 10.6.175.20 # 新增节点 2 的 IP
          access_ip: 10.6.175.20 # 新增节点 2 IP
          ansible_host: 10.6.175.20 # 新增节点的 2 IP
          ansible_connection: ssh
          ansible_user: root # 登陆节点 2 的用户名
          ansible_password: password01 # 登陆节点 2 的密码
      children:
        kube_control_plane:
          hosts:
            node1:
        kube_node:
          hosts:
            node1:
            node2:
        etcd:
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
$ vi kubean/examples/install/scale/1.addWorkNode/HostsConfCM.yml
```
#### 2. 配置扩容任务 ClusterOperation.yml 配置参数 

进入 `kubean/examples/scale/1.addWorkNode/` 路径，编辑模版 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/scale/1.addWorkNode/` 路径下 **`ClusterOperation.yml`** 的模版内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-awn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: scale.yml
  extraArgs: --limit=node2
```
**重要参数：**
>* `spec.cluster`: 指定需要扩容节点的集群名称，上述指定的是名为 `cluster-mini` 的集群为扩容目标。
>* `spec.action:` 指定扩容节点的 kubespray 剧本, 这里设置为 `scale.yml`.
>* `spec.extraArgs`: 指定扩容的节点限制，这里通过 `--limit=` 参数限定扩容 node2 节点


例如，下面展示了一个 ClusterOperation.yml 示例：
<details>
<summary> ClusterOperation.yml 示例</summary>
```yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-awn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2
  backoffLimit: 0
  actionType: playbook
  action: scale.yml
  extraArgs: --limit=node2
```
</details>

执行如下命令编辑 ClusterOperation.yml 配置模版：

```bash
$ vi kubean/examples/install/scale/1.addWorkNode/ClusterOperation.yml
```

#### 3.应用 `scale/1.addWorkNode` 文件下所有的配置

完成上述步骤并保存 HostsConfCM.yml 和 ClusterOperation.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/install/scale/1.addWorkNode/
```

至此，您已经使完成了一个集群的工作节点扩容。

## 缩容工作节点

#### 1. 配置主机配置参数 HostsConfCM.yml

进入 `kubean/examples/scale/2.delWorkNode/` 路径，编辑待建集群节点配置信息模版 `HostsConfCM.yml`，删除需要移除的节点及配置。

**删除参数如下：**

* `all.hosts` 下的 node2 节点接入参数。
* `all.children.kube_node.hosts` 内的主机名称 node2 。

例如，下面展示了一个 HostsConfCM.yml 示例：
<details>
<summary> HostsConfCM.yml 示例</summary>
``` yaml
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
          ip: 10.6.175.10 # 你的节点 IP
          access_ip: 10.6.175.10 # 你的节点 IP
          ansible_host: 10.6.175.10 # 你的节点 IP
          ansible_connection: ssh
          ansible_user: root # 登陆节点的用户名
          ansible_password: password01 # 登陆节点的密
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
```
</details>

执行如下命令编辑 HostsConfCM.yml 配置模版：

```bash
$ vi kubean/examples/install/scale/2.delWorkNode/HostsConfCM.yml
```
#### 2. 配置扩容任务 ClusterOperation.yml 配置参数 

进入 `kubean/examples/scale/2.delWorkNode/` 路径，编辑模版 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/scale/2.delWorkNode/` 路径下 **`ClusterOperation.yml`** 的模版内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dwn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2
```
**重要参数：**
>* `spec.cluster`: 指定需要缩容节点的集群名称, 上述指定的是名为 cluster-mini 的集群为缩容目标。
>* `spec.action`: 指定缩容节点的 kubespray 剧本, 这里设置为 remove-node.yml。
>* `spec.extraArgs`: 指定缩容的节点，这里通过 -e 参数指定缩容 node2 节点


例如，下面展示了一个 ClusterOperation.yml 示例：
<details>
<summary> ClusterOperation.yml 示例</summary>
```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dwn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2
  backoffLimit: 0
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2
```
</details>

执行如下命令编辑 ClusterOperation.yml 配置模版：

```bash
$ vi kubean/examples/install/scale/2.delWorkNode/ClusterOperation.yml
```

#### 3.应用 `scale/2.delWorkNode` 文件下所有的配置

完成上述步骤并保存 HostsConfCM.yml 和 ClusterOperation.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/install/scale/2.delWorkNode/
```

至此，您已经完成了一个集群工作节点的缩容操作。
