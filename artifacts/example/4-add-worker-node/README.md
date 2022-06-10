# 测试新增workernode节点

## 准备环境

* 准备单节点的k8s
* 节点分布 node1 10.6.109.2 (master etcd worker都在node1上)
* 执行`kubectl apply -f 1prepareCluster`则会拉取一个单节点的集群(基于用户密码)

## 执行新增节点

* 新增节点 node2 10.6.109.3
* 在hosts-conf-cm里修改`all.hosts` 和 `all.children.kube_node.hosts`字段
* clusterOps里修改`action`为scale.yml和`extraArgs`(格式为`--limit=${new_node_name}`(new_node_name用逗号分隔主机名称))

