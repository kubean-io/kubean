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
- [ ] kubean 文档站优化
- [ ] 更多的操作系统离线化支持
- [ ] 提供便捷的自定义资源模板生成方式
- [ ] 离线场景相关镜像及二进制等资源的使用优化
