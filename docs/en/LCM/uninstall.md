# 集群卸载

> English | [中文](../../zh/LCM/uninstall.md)


我们以使用 [`minimal`](../../../examples/install/1.minimal/) 模板安装好的单节点集群为例，来进行集群卸载操作；

假设当前名为 cluster-mini 的单节点集群已经部署完毕，我们要执行卸载集群的操作，需要使用 `kubectl apply` 下发 `action` 为 `reset.yml` 的 CluterOperation 自定义资源；

ClusterOperation 的关键配置大体如下：

``` yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: sample-uninstall-ops
spec:
  ...
  cluster: cluster-mini
  action: reset.yml
  ...

```

关键配置：
* `spec.cluster`: 指定需要卸载的集群名称, 上述指定的是名为 `cluster-mini` 的集群为卸载目标
* `spec.action`: 指定卸载相关的 kubespray 剧本, 这里设置为 `reset.yml`


更具体的配置可以参考 [`uninstall/ClusterOperation.yml`](../../../examples/uninstall/ClusterOperation.yml) 样例模板；
