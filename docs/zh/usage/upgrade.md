# 集群版本升级

本节将向您介绍如何使用 kubean 升级集群的 kubernnetes 版本。在您克隆至本地的 `kubean/example/upgrade` 文件内，同样提供了集群版本升级的样例模版：

<details open>
<summary> upgrade 文件内主要的配置文件及用途如下：</summary>

```yaml
    upgrade
    ├── ClusterOperation.yml                  # 升级集群任务
    └── VarsConfCM.yml                        # 集群升级版本等参数配置
```
</details>

下面以[使用 all-in-one 模式部署的单节点集群](./all-in-one-install.md)为例，来演示集群版本升级操作。
> 注意：执行集群版本升级前，您必须已经使用 kubean 完成了一套集群的部署。

#### 1. 新增升级任务

进入 `kubean/examples/upgrade/` 路径，编辑模版 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/upgrade/` 路径下 **`ClusterOperation.yml`** 的模版内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-upgrade-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: upgrade-cluster.yml
```
**重要参数：**
>* `spec.cluster`: 指定需要升级的集群名称，上述指定的是名为 `cluster-mini` 的集群为升级目标。
>* `spec.action:` 指定升级相关的 kubespray 剧本, 这里设置为 `upgrade-cluster.yml`。

#### 2. 指定集群升级版本

进入 `kubean/examples/upgrade/` 路径，编辑模版 `VarsConfCM.yml`，通过配置 `kube_version` 参数，指定集群升级的版本。

`kubean/examples/upgrade/` 路径下 **`VarsConfCM.yml`** 的模版内容如下：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mini-vars-conf
  namespace: kubean-system
data:
  group_vars.yml: |
    kube_version: v1.25.8
    # upgrade_cluster_setup: true
    # upgrade_node_confirm: true
    # upgrade_node_pause_seconds: 60

    container_manager: containerd
    kube_network_plugin: calico
    etcd_deployment_type: kubeadm
```
**重要参数：**
>* `kube_version`: 指定需要升级的集群版本, 上述指定了要升级到 k8s v1.25.8 版本。

!!! 移除工作节点主机参数的示例
    === "升级版本前"

        ```yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: mini-vars-conf
          namespace: kubean-system
        data:
          group_vars.yml: |
            kube_version: v1.25.0
            # upgrade_cluster_setup: true
            # upgrade_node_confirm: true
            # upgrade_node_pause_seconds: 60

            container_manager: containerd
            kube_network_plugin: calico
            etcd_deployment_type: kubeadm
        ```

    === "升级版本后"

        ```yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: mini-vars-conf
          namespace: kubean-system
        data:
          group_vars.yml: |
            kube_version: v1.25.8
            # upgrade_cluster_setup: true
            # upgrade_node_confirm: true
            # upgrade_node_pause_seconds: 60

            container_manager: containerd
            kube_network_plugin: calico
            etcd_deployment_type: kubeadm
        ```


附：kubean 集群版本支持机制：

| kubean 版本 | 推荐的 kubernetes 版本 | 支持的 kubernetes 版本范围                                   |
| ----------- | ---------------------- | ------------------------------------------------------------ |
| v0.5.2      | v1.25.4                | - "v1.27.2"<br/>        - "v1.26.5"<br/>        - "v1.26.4"<br/>        - "v1.26.3"<br/>        - "v1.26.2"<br/>        - "v1.26.1"<br/>        - "v1.26.0"<br/>        - "v1.25.10"<br/>        - "v1.25.9"<br/>        - "v1.25.8"<br/>        - "v1.25.7"<br/>        - "v1.25.6"<br/>        - "v1.25.5"<br/>        - "v1.25.4"<br/>        - "v1.25.3"<br/>        - "v1.25.2"<br/>        - "v1.25.1"<br/>        - "v1.25.0" |

更多升级操作参数的详细说明，请参考 kubespray 文档：[通过Kubespray 更新 kubernetes](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/upgrades.md)。

#### 3.应用 `upgrade` 文件下所有的配置

完成上述步骤并保存 ClusterOperation.yml 和 VarsConfCM.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/upgrade/
```

至此，您已经使完成了一个集群的 kuberntes 版本的升级。
