# 集群伸缩

> [English](../../en/LCM/scale.md) | 中文

我们以使用 [`minimal`](../../../examples/install/1.minimal/) 模板安装好的单节点集群为例，来进行集群节点的扩缩容操作；

## 新增一个工作节点

> 可以参考 [`scale/1.addWorkNode/`](../../../examples/scale/1.addWorkNode/) 样例模板；

### 1. 新增节点参数配置

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
          ...
        node2:
          ip: <IP2>
          access_ip: <IP2>
          ansible_host: <IP2>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
      children:
        ...
        kube_node:
          hosts:
            node1:
            node2:
        ...

```

重要参数设置：
* `all.hosts` 内配置新增的 node2 节点接入参数，这里采用用户名密码方式 (SSH私钥方式请见：[sshkey_deploy_cluster](../sshkey_deploy_cluster.md))
* `all.children.kube_node.hosts` 内新增 node2 主机名称；


### 2. 创建扩容节点操作任务

``` yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: sample-awn-ops
spec:
  ...
  cluster: cluster-mini
  action: scale.yml
  extraArgs: --limit=node2
  ...

```

重要参数设置：
* spec.cluster: 指定需要扩容节点的集群名称, 上述指定的是名为 `cluster-mini` 的集群为扩容目标
* spec.action: 指定扩容节点的 kubespray 剧本, 这里设置为 `scale.yml`
* spec.extraArgs: 指定扩容的节点限制，这里通过 `--limit=` 参数限定扩容 node2 节点


---

## 删除一个工作节点

> 可以参考 [`scale/2.delWorkNode/`](../../../examples/scale/2.delWorkNode/) 样例模板；

### 1. 创建缩容节点操作任务

``` yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: sample-dwn-ops
spec:
  cluster: cluster-mini
  action: remove-node.yml
  extraArgs: -e node=node2
```

重要参数设置：
* spec.cluster: 指定需要缩容节点的集群名称, 上述指定的是名为 `cluster-mini` 的集群为缩容目标
* spec.action: 指定缩容节点的 kubespray 剧本, 这里设置为 `remove-node.yml`
* spec.extraArgs: 指定缩容的节点，这里通过 `-e` 参数指定缩容 node2 节点

### 2. 删除节点参数配置

删除 [HostsConfCM](../../../examples/scale/2.delWorkNode/HostsConfCM.yml) 中 node2 的信息；

删除重要关注参数：
* `all.hosts` 内删除 node2 节点接入参数
* `all.children.kube_node.hosts` 内删除 node2 主机名称；

---
