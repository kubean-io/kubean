# 集群升级

> English | [中文](../../zh/LCM/upgrade.md)

我们以使用 [`minimal`](../../../examples/install/1.minimal/) 模板安装好的单节点集群为例，来进行集群升级操作；

### 1. 创建升级任务的 ClusterOperation

假设当前名为 cluster-mini 的单节点集群已经部署完毕，我们要执行升级集群的操作，需要使用 `kubectl apply` 下发 `action` 为 `upgrade-cluster.yml` 的 CluterOperation 自定义资源；

ClusterOperation 的关键配置大体如下：

``` yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: sample-upgrade-ops
spec:
  ...
  cluster: cluster-mini
  action: upgrade-cluster.yml
  ...

```

关键配置：
* `spec.cluster`: 指定需要升级的集群名称, 上述指定的是名为 `cluster-mini` 的集群为升级目标
* `spec.action`: 指定升级相关的 kubespray 剧本, 这里设置为 `upgrade-cluster.yml`


更具体的配置可以参考 [`upgrade/ClusterOperation.yml`](../../../examples/upgrade/ClusterOperation.yml) 样例模板；

### 2. 更新集群参数配置 VarsConfCM.yml

```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mini-vars-conf
  namespace: kubean-system
data:
  group_vars.yml: |
    kube_version: v1.25.8
    # upgrade_cluster_setup: true
    ...

```

关键配置：
* `kube_version`: 指定需要升级的集群版本, 上述指定了要升级到 `k8s v1.25.8` 版本

升级操作参数的详细说明，请参考 kubespray 文档：[Upgrading Kubernetes in Kubespray](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/upgrades.md).

更具体的配置可以参考 [`upgrade/VarsConfCM.yml`](../../../examples/upgrade/VarsConfCM.yml) 样例模板；

---


### 如何查看当前 Kubean 支持的版本范围？

通过查看 kubean 的 manifest 自定义资源，即可知当前 kubean 支持的 k8s 等版本的支持范围： [Manifest](https://github.com/kubean-io/kubean-helm-chart/blob/main/charts/kubean/templates/manifest.cr.yaml#L83)

