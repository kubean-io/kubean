# 测试移除workernode节点

## 准备环境

* 节点分布
    * node1 10.6.109.2
    * node2 10.6.109.3
* 角色分布
    * node1 (master etcd)
    * node2 (worker)
* 执行`kubectl apply -f 1prepareCluster`来初始化环境

## 执行移除worker节点

* hosts-conf主机清单无需改变(需要确保待移除的节点在主机清单之中)
* clusterOps中需要修改`action`为`remove-node.yml`和`extraArgs`为`-e node={nodeName}`
  * 对于移除离线节点，extraArgs存在稍许不同
* 执行`kubectl apply -f 2removeWorkerNode` 执行完毕后，查看nodes列表可以发现node2节点消失
