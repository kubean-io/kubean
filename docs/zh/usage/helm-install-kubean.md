# 使用 Helm 部署 kubean chart

## 前置条件

1. 您已拥有一个标准 kubernetes 集群或云厂商提供的集群。
2. 在您的集群控制节点或云终端上已完成 Helm 工具的安装。[如何安装 Helm 工具](https://helm.sh/docs/intro/install/)。

---

## 开始部署

1. 将 kubean Helm 仓库添加到本地 Helm 仓库，在现有集群控制节点或云终端上执行如下命令。

```bash
$ helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
```

完成上一步后检查 kubean 仓库是否已经正确添加至本地 Helm 仓库。

```bash
$ helm repo list

# 预期输出如下：
NAME          	URL
kubean-io     	https://kubean-io.github.io/kubean-helm-chart/
```

2. 检查 kubean Helm 仓库中可用的 Chart 及其版本，执行下面命令将列出 kubean Helm 仓库内所有的 Chart 列表。

```bash
helm search repo kubean

# 预期输出如下：
NAME            	CHART VERSION	APP VERSION	DESCRIPTION
kubean-io/kubean	v0.5.2       	v0.5.2     	A Helm chart for kubean
```

3. 完成上述步骤后，，执行如下命令安装 kubean。

```bash
$ helm install kubean kubean-io/kubean --create-namespace -n kubean-system
```

!!!note
    您还可以使用 --version 参数来指定 kubean 的版本。

4. 至此，您已经完成了 kubean helm chart 的部署，您可以执行如下命令查看 kubean-system 命名空间下的 helm release。

```bash
$ helm ls -n kubean-system

# 预期输出如下：
NAME  	NAMESPACE    	REVISION	UPDATED                                  	STATUS  	CHART            	APP VERSION
kubean	kubean-system	1       	2023-05-15 00:24:32.719770617 -0400 -0400	deployed	kubean-v0.4.9-rc1	v0.4.9-rc1

```