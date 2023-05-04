# 集群部署

> [English](../../en/LCM/install.md) | 中文


前置条件：通过 helm 安装 [kubean charts](https://github.com/kubean-io/kubean-helm-chart).

---

## 单节点集群部署

> 参考 [`minimal`](../../../examples/install/1.minimal/) 样例模板

参照模板，我们将创建一个多合一的单节点集群：

#### 1. 更新 [`AllInOne.yml`](../../../examples/install/1.minimal/AllInOne.yml) 中的占位符为真实值
* `<IP1>`
* `<USERNAME>`
* `<PASSWORD>`
* `<TAG>`

#### 2. 应用 [`AllInOne.yml`](../../../examples/install/1.minimal/AllInOne.yml) 

``` bash
$ kubectl apply -f examples/install/1.minimal/
```

---

## 加速器模式部署

> 参考 [`mirror`](../../../examples/install/2.mirror/) 样例模板

#### 1. 更新 [`2.mirror`](../../../examples/install/2.mirror/) 目录中 yaml 清单的占位符为真实值
* `<IP1>` / `<IP2>` ...
* `<USERNAME>`
* `<PASSWORD>`
* `<TAG>`

#### 2. 应用 [`2.mirror`](../../../examples/install/2.mirror/) 中的 yaml 清单

``` bash
$ kubectl apply -f examples/install/2.mirror/
```

#### 3. 加速器镜像设置请见 [`VarsConfCM`](../../../examples/install/2.mirror/VarsConfCM.yml)

本例中使用到的加速器:
* 二进制加速：[public binary files mirror](https://github.com/DaoCloud/public-binary-files-mirror)
* 镜像加速：[public image mirror](https://github.com/DaoCloud/public-image-mirror)

---

## 纯离线模式部署

> 参考 [`airgap`](../../../examples/install/3.airgap/) 样例模板


详细请浏览 [离线场景的使用](../offline.md)


---

## SSH秘钥模式部署

详细请浏览 [使用 SSH 秘钥方式部署 K8S 集群](../sshkey_deploy_cluster.md)

---