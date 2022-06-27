# Kubean TestCase


| 子模块         | 功能点                      | 用例名称                                         | 备注     | 基线  | 是否完成 | 代码链接 | 详细说明                                      |
| ---------------- | ----------------------------- | -------------------------------------------------- | ---------- | ------- | ---------- | ---------- | ----------------------------------------------- |
| 集群创建       | 自动创建集群                | 创建单主节点DKE集群                              |          | C-001 | <ul><li>[x] </li></ul> | [代码链接](../../test/e2e/kubean_cluster_install_test.go)          |                                               |
|                | 自动创建集群                | 创建多主节点DKE集群                              |          | C-001 |          |          |                                               |
|                | 自定义创建集群              | 使用containerd创建DKE集群                        |          | C-013 |          |          |                                               |
|                | 自定义创建集群              | 使用runc创建DKE集群                              |          | C-042 |          |          |                                               |
|                | 自定义创建集群              | 使用gvisor创建DKE集群                            |          | C-042 |          |          |                                               |
|                | 自定义创建集群              | 使用K8S V1.20创建DKE集群                         |          | C-012 |          |          |                                               |
|                | 秘钥模式创建集群            | 私钥模式创建集群                                 |          | null  |          |          |                                               |
|                | 密码模式创建集群            | 密码模式创建集群                                 |          | null  |          |          |                                               |
|                | 使用不同的CNI创建集群       | 使用Cilium创建集群                               |          | C-001 |          |          |                                               |
|                | 使用不同的CNI创建集群       | 使用multus + Calico创建集群                      |          | C-001 |          |          |                                               |
|                | 非root用户的集群安装        | 非root用户安装DKE集群                            | 低优先级 | null  |          |          |                                               |
|                | 创建过程中终止安装          | 安装过程中终止安装，环境可以reset                | 低优先级 | null  |          |          |                                               |
|                | 物理机部署集群              | 物理机部署DKE集群                                | 低优先级 | C-048 |          |          |                                               |
|                | 虚拟机部署集群              | 虚拟机部署DKE集群                                |          | C-048 |          |          |                                               |
|                | 私有云部署集群              | 私有云部署DKE集群                                | 低优先级 | C-048 |          |          |                                               |
|                | 公有云部署集群              | 公有云部署DKE集群                                | 低优先级 | C-048 |          |          |                                               |
|                | 时间同步                    | 在部署DKE集群之前实现节点的时间同步              |          |       |          |          |                                               |
|                | 安装参数设置                | override_system_hostname                         |          |       |          |          |                                               |
|                | 安装参数设置                | kube_proxy_mode 为iptables和ipvs                 |          |       |          |          |                                               |
|                | 安装参数设置                | enable_nodelocaldns                              |          |       |          |          |                                               |
|                | 安装参数设置                | etcd_deployment_type                             |          |       |          |          |                                               |
|                | 安装参数设置                | metrics_server_enabled                           |          |       |          |          |                                               |
|                | 安装参数设置                | local_path_provisioner_enabled                   |          |       |          |          |                                               |
|                | 安装参数设置                | download_run_once                                |          |       |          |          |                                               |
|                | 安装参数设置                | download_container                               |          |       |          |          |                                               |
|                | 安装参数设置                | download_force_cache                             |          |       |          |          |                                               |
|                | 安装参数设置                | download_localhost                               |          |       |          |          |                                               |
|                | 安装信息记录                | 安装完成后，详细记录安装信息                     |          |       |          |          |                                               |
|                | 审计日志开关                | 审计日志开关                                     |          |       |          |          |                                               |
| 硬件与操作系统 | 混合部署                    | arm,x86混合部署                                  |          | H-001 |          |          |                                               |
|                | 支持主流操作系统            | Suse、Centos,RedHat,Ubuntu,OracleLinux           |          | H-004 |          |          |                                               |
| 集群升级       | 升级K8S版本                 | 在线/离线升级K8S版本，服务不中断                 |          | C-003 |          |          |                                               |
|                | 升级网络方案                | 升级网络方案服务不中断                           | 待确认   | C-003 |          |          |                                               |
|                | 升级DKE版本                 | 升级DKE版本服务不中断                            |          | C-003 |          |          |                                               |
|                | 降级DKE版本                 | 降级DKE版本服务不中断                            | 待确认   | C-003 |          |          |                                               |
|                | 新增补丁                    | 新增不定，服务不中断                             |          |       |          |          |                                               |
|                | 卸载补丁                    | 卸载补丁，服务不中断                             |          |       |          |          |                                               |
|                | 卸载集群                    | 集群能够卸载后重装成功                           |          |       |          |          |                                               |
| 节点运维       | 新增master节点              | 安装后新增一个master节点                         |          | C-004 |          |          |                                               |
|                | 新增worker节点              | 安装后新增一个worker节点                         |          | C-004 |          |          |                                               |
|                | 卸载master节点              | 安装后卸载一个master节点                         |          | C-004 |          |          |                                               |
|                | 卸载worker节点              | 安装后卸载一个worker节点                         |          | C-004 |          |          |                                               |
|                | master节点替换              | 多master集群，可以轮流替换掉现有的所有master节点 | 往后放   |       |          |          |                                               |
| 高可用         | master高可用-宕机           | 多主集群宕机第一台master节点                     |          | L-006 |          |          |                                               |
|                | master高可用-CPU            | 多主集群第一台masterCPU占用99%                   |          | L-006 |          |          |                                               |
|                | master高可用-Mem            | 多主集群etcd的leader节点内存占用99%              |          |       |          |          |                                               |
|                | master-高可用-磁盘          | 多主集群etcd的leader节点磁盘空间占满             |          |       |          |          |                                               |
|                | master高可用-网络           | 多主集群单个master的网络不稳定持续1分钟          |          |       |          |          |                                               |
|                | master的etcd高可用-leader   | 多主集群的etcd的follower节点宕机                 |          |       |          |          |                                               |
|                | master的etcd高可用-follower | 多主集群的etcd的leader节点宕机                   |          |       |          |          |                                               |
|                | worker节点高可用            | 多worker集群worker节点宕机                       |          |       |          |          | [detail](./testcase_detail.md#work节点高可用) |
| 网络           | ip_forward生效              | 集群部署后各个节点ip_forward生效                 |          |       |          |          |                                               |
|                | 硬件加速器                  | SR-IOV 硬件加速                                  | 往后放   | N-001 |          |          |                                               |
| ClusterOps     | ClusterOps的自动清除        | ClusterOps按照时间逆序清楚                       |          |       |          |          |                                               |
| 集群运维       | 同一个集群操作的并发限制    | 同一个DKE集群不能同时执行2个job                  |          |       |          |          |                                               |
