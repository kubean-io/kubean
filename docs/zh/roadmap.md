# Kubean Roadmap

当前的 Roadmap 是暂时的，具体的时间表以社区需要而定。

> **在 Roadmap 中未提及的功能特性, 我们可以在 [issues](https://github.com/kubean-io/kubean/issues) 中讨论.**

## Q3 2022
* Kubean 项目架构流程设计 [architecture.md](https://github.com/kubean-io/kubean/blob/main/docs/en/architecture.md)
* 验证 Kubean 的集群生命周期管理操作
* 添加系统包构建 CI https://github.com/kubean-io/kubean/pull/62
* 提供 Kubean API https://github.com/kubean-io/kubean/pull/128


## Q4 2022
* E2E tests [kubean test case](https://github.com/kubean-io/kubean/blob/main/docs/test/kubean_testcase.md)
* k8s 镜像及二进制包支持 arm 架构 https://github.com/kubean-io/kubean/pull/200
* 支持升级包构建 https://github.com/kubean-io/kubean/pull/289
* 离线场景 RHEL8.4 部署适配 https://github.com/kubean-io/kubean/pull/325
* 支持还原系统包管理配置 https://github.com/kubean-io/kubean/pull/298
* 支持集群部署完后回传 Kubeconfig https://github.com/kubean-io/kubean/pull/192
* 增加 SSH Key 认证部署方式 https://github.com/kubean-io/kubean/pull/302


## Q1 2023
* 支持 apt 系统包管理配置 https://github.com/kubean-io/kubean/pull/459
* Cluster Operation CRD 支持自定义 Action https://github.com/kubean-io/kubean/issues/361
* Kubean Chart 支持 Charts Syncer https://github.com/kubean-io/kubean/pull/468
* 支持 Pre check 部署前的预检测 https://github.com/kubean-io/kubean/pull/555
* 统信 UOS 1020a 系统包适配 https://github.com/kubean-io/kubean/pull/583


## Q2 2023
* 支持基于 OpenEuler 的集群部署 https://github.com/kubean-io/kubean/pull/628
* 支持 Other Linux 不同包管理器下系统包构建
* E2E CI 优化
* 为 release 中输出的制品 (Assets) 增加 checksum 文件
* 通过 Join 异构节点，支持混合架构部署
