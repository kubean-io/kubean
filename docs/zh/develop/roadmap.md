# Kubean 路线图

当前的 Roadmap 是暂时的，具体的时间表以社区需要而定。

> **在 Roadmap 中未提及的功能特性, 我们可以在 [issues](https://github.com/kubean-io/kubean/issues) 中讨论.**

## Q3 2022
- [x] 设计 Kubean 项目架构流程 [architecture.md](https://github.com/kubean-io/kubean/blob/main/docs/en/architecture.md)
- [x] 验证 Kubean 的集群生命周期管理操作
- [x] 添加系统包构建 CI https://github.com/kubean-io/kubean/pull/62
- [x] 提供 Kubean API https://github.com/kubean-io/kubean/pull/128


## Q4 2022
- [x] E2E tests [kubean test case](https://github.com/kubean-io/kubean/blob/main/docs/test/kubean_testcase.md)
- [x] k8s 镜像及二进制包支持 arm 架构 https://github.com/kubean-io/kubean/pull/200
- [x] 支持升级包构建 https://github.com/kubean-io/kubean/pull/289
- [x] 离线场景 RHEL8.4 部署适配 https://github.com/kubean-io/kubean/pull/325
- [x] 支持还原系统包管理配置 https://github.com/kubean-io/kubean/pull/298
- [x] 支持集群部署完后回传 Kubeconfig https://github.com/kubean-io/kubean/pull/192
- [x] 增加 SSH Key 认证部署方式 https://github.com/kubean-io/kubean/pull/302


## Q1 2023
- [x] 支持 apt 系统包管理配置 https://github.com/kubean-io/kubean/pull/459
- [x] Cluster Operation CRD 支持自定义 Action https://github.com/kubean-io/kubean/issues/361
- [x] Kubean Chart 支持 Charts Syncer https://github.com/kubean-io/kubean/pull/468
- [x] 支持 Pre check 部署前的预检测 https://github.com/kubean-io/kubean/pull/555
- [x] 统信 UOS 1020a 系统包适配 https://github.com/kubean-io/kubean/pull/583


## Q2 2023
- [x] 支持基于 OpenEuler 离线场景的集群部署 https://github.com/kubean-io/kubean/pull/628
- [x] 支持 Other Linux 通过脚本自主构建离线场景依赖的系统包 https://github.com/kubean-io/kubean/pull/627
- [x] 使用 mkdocs 更新 kubean 文档站 https://github.com/kubean-io/kubean/pull/728
- [x] 优化 release 发版 CI https://github.com/kubean-io/kubean/pull/863
- [x] 新增关于证书更新的 ansible 剧本 https://github.com/kubean-io/kubean/pull/884
- [x] 更新 release 发版流程 https://github.com/kubean-io/kubean/pull/869

## Q3 2023
- [x] 添加证书更新剧本 https://github.com/kubean-io/kubean/pull/884
- [x] 新增流水线对上游 kubespray 最新版本的每日部署验证 https://github.com/kubean-io/kubean/pull/870
- [x] 确保 Cluster 资源的级联删除 https://github.com/kubean-io/kubean/pull/918
- [x] 为 ClusterOperation 记录添加清除权重 https://github.com/kubean-io/kubean/pull/983

## Q4 2023
- [x] 优化镜像离线包为 OCI 格式 https://github.com/kubean-io/kubean/pull/996
- [x] 优化 Operator 的日志输入 https://github.com/kubean-io/kubean/pull/1032
- [x] 提高 Manifest 资源的查询效率 https://github.com/kubean-io/kubean/pull/1036
- [x] 重构镜像导入脚本，使其支持多架构导入 https://github.com/kubean-io/kubean/pull/1040

## Q1 2024
- [x] 提高 precheck 剧本的执行效率 https://github.com/kubean-io/kubean/pull/1076
- [x] 优化 ClusterOperation 的调谐性能 https://github.com/kubean-io/kubean/pull/1082
- [x] 重构自定义资源生成脚本逻辑 https://github.com/kubean-io/kubean/pull/1152
- [x] 修复 ubuntu18.04 离线包版本问题 https://github.com/kubean-io/kubean/pull/1158
- [x] 自动化 docker 限制单容器磁盘占用的前置步骤 https://github.com/kubean-io/kubean/pull/1179

## Q2 2024
- [ ] 提供客户端命令行工具，及便捷的自定义资源模块生成方式
- [ ] 不同节点规模集群部署的容量规划
- [ ] 提供完整的离线资源管理方案
- [ ] 支持多种生命周期管理引擎，比如kubespray、kubekey
- [ ] 支持基于 ostree 的集群操作回滚
