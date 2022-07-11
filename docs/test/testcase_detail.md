## 集群创建
### 创建单主节点DKE集群
    1. 创建一个1master +1 worker集群
    2. 检查集群能够正常工作

### 创建多主节点DKE集群
    1. 创建一个3master +2 worker集群
    2. 检查集群能够正常工作

### 安装过程中终止安装，环境可以reset
    （低优先级，待定）
    1. 部署一个3master + 2worker集群，在部署master的时候终止安装，能够reset
    1. 部署一个3master + 2worker集群，在部署worker的时候终止安装, 能够reset

### 节点时间同步
    1. 手动设置主机之间时间不同步
    2. 部署一个3master + 1 worker集群
    3. 完成后，检查节点间是否时间同步，集群是否正常

## 集群高可用

### 多主集群宕机第一台master节点
    1. 部署3master + 2worker的集群
    2. 宕机/断网最先部署的master节点，集群服务不中断
    3. 宕机/断网 剩余的任意一个master，集群服务不中断
    4. 恢复宕机/断网 的2个节点，节点正常工作，集群正常工作

### 多主集群第一台masterCPU占用99%
    1. 部署3master + 2worker集群
    2. 将最先部署的master的CPU占用99%，集群服务不中断
    3. 恢复节点的CPU到正常值，查看节点是否恢复正常使用

### 多主集群任意一台master节点内存占用99%
    1. 部署3master + 2worker集群
    2. 将一个master节点的Memory占用99%，集群服务不中断
    3. 恢复节点的Memory到正常值，查看节点是否恢复正常使用

### 多主集群任意一台master节点磁盘空间占满
    1. 部署3master + 2worker集群
    2. 将一个master节点的磁盘剩余空间占用满，集群服务不中断
    3. 恢复节点的磁盘空间，查看节点是否恢复正常使用

### 多主集群etcd的leader网络不稳定持续1分钟
    1. 部署3master + 2worker集群
    2. 模拟etcd的leader节点持续一段时间的网络不稳定，集群服务不中断
    3. 恢复节点，节点正常工作，集群正常工作

### 多主集群的etcd的leader节点宕机
    1. 部署3master + 2worker
    2. 断网/宕机etcd的leader节点，etcd选出新的leader节点，集群服务不中断。
    3. 恢复节点，该节点重新加入集群，成为etcd集群的follower节点

### 多主集群的etcd的follower节点宕机
    1. 部署3master + 2worker
    2. 断网/宕机etcd的follower节点，集群服务不中断。
    3. 恢复节点，该节点重新加入集群，成为etcd集群的follower节点

### work节点高可用
    1. 部署3master + 2worker集群
    2. 宕机/断网一个worker节点，集群服务不中断
    3. 恢复worker节点，节点重新加入集群，正常工作。

## 集群运维
### 同一个集群操作的并发限制
    1. 对集群下发升级请求，成功
    2. 操作1结束前，对DKE集群下发降级等请求，失败