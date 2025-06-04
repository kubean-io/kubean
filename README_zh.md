# :seedling: Kubean

> [English](./README.md)

<p align="center">
  <a href="https://github.com/cncf/k8s-conformance/pull/2240"><img src="docs/overrides/assets/images/certified_k8s.png" height=120 alt="k8s conformance"></a>
  <a href="https://kubean-io.github.io/kubean"><img src="docs/overrides/assets/images/kubean_logo.png" height=120 alt="kubean"></a>
</p>

<p align="center">
  Kubean 是一款准生产的集群生命周期管理工具，基于 <a href="https://github.com/kubernetes-sigs/kubespray">kubespray</a> 与其他集群 LCM 引擎。
</p>

<p align="center">
  <a href="https://github.com/kubean-io/kubean/actions/workflows/auto-main-ci.yaml"><img src="https://github.com/kubean-io/kubean/actions/workflows/auto-main-ci.yaml/badge.svg" alt="main workflow"></a>
  <a href="https://codecov.io/gh/kubean-io/kubean"><img src="https://codecov.io/gh/kubean-io/kubean/branch/main/graph/badge.svg?token=8FX807D3QQ" alt="codecov"></a>
  <a href="https://bestpractices.coreinfrastructure.org/projects/6263"><img src="https://bestpractices.coreinfrastructure.org/projects/6263/badge" alt="Best Practices"></a>
  <a href="https://github.com/kubean-io/kubean/blob/main/docs/overrides/test/kubean_testcase.md"><img src="https://raw.githubusercontent.com/dasu23/e2ecoverage/master/badges/kubean/kubeanCoverage.svg" alt="kubean coverage"></a>
  <a href="https://github.com/kubean-io/kubean/blob/main/docs/overrides/test/kubean_testcase.md"><img src="https://raw.githubusercontent.com/dasu23/e2ecoverage/master/badges/kubean/kubeanCoverage2.svg" alt="kubean coverage"></a>
  <a href="https://github.com/kubean-io/kubean/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-AL%202.0-blue" alt="license"></a>
  <a href="https://goreportcard.com/report/github.com/kubean-io/kubean"><img src="https://goreportcard.com/badge/github.com/kubean-io/kubean" alt="Go Report Card"></a>
  <a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fkubean-io%2Fkubean?ref=badge_large"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkubean-io%2Fkubean.svg?type=small" alt="FOSSA Status"></a>
</p>

---

<p align="center">
<img src="https://github.com/cncf/artwork/blob/main/other/illustrations/ashley-mcnamara/transparent/cncf-cloud-gophers-transparent.png" width="500" /><br/>
Kubean 是一个<a href="https://cncf.io/">云原生计算基金会(CNCF)全景图项目</a>.
</p>

## :anchor: 功能超赞

- **简单易用**：通过声明式 API 实现 Kubean 和 K8s 集群强劲生命周期管理的部署。
- **支持离线**：每个版本都会发布离线包（os-pkgs、镜像、二进制包）。你不必担心如何收集所需的资源。
- **兼容性**：支持多架构交付：AMD、ARM；常见的 Linux 发行版；以及基于鲲鹏构建的麒麟操作系统。
- **可扩展性**：允许使用原生 Kubespray 自定义集群。

## :surfing_man: 快速入门

### Killercoda

我们在 [killercoda](https://killercoda.com)（一个在线交互式技术学习平台）上创建了一个[项目](https://killercoda.com/kubean)，可以在上面进行试玩。

### 本地安装

1. 确保有一个 Kubernetes 集群且安装了 Helm

2. 部署 kubean-operator

   ``` shell
   helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
   helm install kubean kubean-io/kubean --create-namespace -n kubean-system
   ```

   检查 kubean-operator 状态：

   ```shell
   kubectl get pods -n kubean-system | grep 'kubean'
   ```

3. 在线模式部署最小化单节点集群

   1. 一个简单的方式是使用 [AllInOne.yml](./examples/install/1.minimal/)，
      替换 `<IP1>`、`<USERNAME>`... 等字符串为真实值。
   
   2. 启动 `kubeanClusterOps`，这将启动 kubespray job。

      ```shell
      kubectl apply -f examples/install/1.minimal
      ```

   3. 检查 kubespray job 状态。

      ```shell
      kubectl get job -n kubean-system
      ```

<p align="center">
  <a href="https://asciinema.org/a/jFTUi2IdU5yydv88kHkPYMni0"><img src="docs/overrides/assets/images/quick_start.gif" height=800 alt="quick start"></a>
</p>

## :ocean: Kubernetes 兼容性

| Kubean Version | Kubernetes Version Range | Kubernetes Default Version | kubespray SHA |
| :-----: | :-----------: | :-----: | :-----: |
| v0.25.2 | v1.30 ~ v1.32 | v1.31.6 | d0e9088 |
| v0.24.2 | v1.30 ~ v1.32 | v1.31.6 | 4ad9f9b |
| v0.23.9 | v1.30 ~ v1.32 | v1.31.6 | a4843ea |
| v0.22.5 | v1.29 ~ v1.31 | v1.30.5 | d173f1d |

要查看 Kubean 支持的 Kubernetes 版本列表，请参考 [Kubernetes 版本列表](./docs/zh/usage/support_k8s_version.md)。

## :book: 开发路线图

有关功能特性，请参阅 [roadmap](docs/en/develop/roadmap.md)。

## :book: 参考文档

请浏览我们的网站 [kubean-io.github.io/kubean/](https://kubean-io.github.io/kubean/)。

## :envelope: 联系我们

你可以通过以下渠道与我们联系：

- Slack：通过请求 CNCF Slack 的[邀请](https://slack.cncf.io/)加入 CNCF Slack 的
  [#Kubean](https://cloud-native.slack.com/messages/kubean) 频道。一旦您可以访问 CNCF Slack，您就可以加入 Kubean 频道。
- 电子邮件: 请参阅 [MAINTAINERS.md](./MAINTAINERS.md) 查找所有维护人员的电子邮件地址。
  随时通过电子邮件与他们联系，报告任何问题或提出问题。

## :thumbsup: 贡献者
<p align="center">
  <a href="https://github.com/kubean-io/kubean/graphs/contributors">
    <img src="https://contrib.rocks/image?repo=kubean-io/kubean" width="700" />
  </a>
</p>

---

<div align="center">
  <p>
    Copyright The Kubean Authors<br/>
    We are a <a href="https://www.cncf.io/">Cloud Native Computing Foundation sandbox project</a>.<br/>
    The Linux Foundation® (TLF) has registered trademarks and uses trademarks. <br/>
    For a list of TLF trademarks, see <a href="https://www.linuxfoundation.org/legal/trademark-usage">Trademark Usage</a>.
  <p>
  <p>
    <img src="https://landscape.cncf.io/images/cncf-landscape-horizontal-color.svg" width="180"/><br/>
    Kubean 位列 <a href="https://landscape.cncf.io/?selected=kubean">CNCF 云原生全景图</a>
  </p>
</div>

