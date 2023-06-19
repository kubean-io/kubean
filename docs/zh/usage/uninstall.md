# 集群卸载

本节将向您介绍如何使用 kubean 卸载集群。在您克隆至本地的 `kubean/example/uninstall` 文件内，同样提供了卸载集群的样例模板：

<details open>
<summary> uninsatall 文件内主要的配置文件及用途如下：</summary>

```yaml
    uninstall
    ├── ClusterOperation.yml                # 卸载集群任务
```
</details>

下面以[使用 all-in-one 模式部署的单节点集群](./all-in-one-install.md)为例，来演示集群版本升级操作。
> 注意：执行集群卸载前，您必须已经使用 kubean 完成了一套集群的部署。

#### 1. 新增卸载任务
进入 `kubean/examples/uninstall/` 路径，编辑模板 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/uninstall/` 路径下 **`ClusterOperation.yml`** 的模板内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-uninstall-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: reset.yml
```
**重要参数：**
>* `spec.cluster`：指定需要卸载的集群名称, 上述指定的是名为 `cluster-mini` 的集群为卸载目标。
>* `spec.action:`：指定卸载相关的 kubespray 剧本, 这里设置为 `reset.yml`。

#### 2.应用 `uninstall` 文件下的配置

完成上述步骤并保存 ClusterOperation.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/uninstall/
```

至此，您已经使完成了一个集群的卸载。