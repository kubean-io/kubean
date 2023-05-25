# 使用 all-in-one 模式部署单节点集群

## 前置条件

1. 您已拥有一个标准 kubernetes 集群或云厂商提供的集群。
2. 集群控制节点或云终端已将安装了 [kubectl 工具](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)。
3. [kubean helm chart](helm-install-kubean.md) 已在您的集群上部署。
4. [kubean 项目](https://github.com/kubean-io/kubean)已经克隆至您本地，如果您还未克隆 kubean，可以执行执行如下命令进行克隆：

```bash
$ git clone https://github.com/kubean-io/kubean.git
```

---

## 开始部署

本教程将使用您克隆至本地的 `kubean/example` 文件作为范例模版，进行集群部署演示。

借助案例模版，仅需两步我们就能使用 kubean 完成一个单节点集群的部署。

#### 1. 配置 AllInOne.yml 参数

进入 `kubean/examples/install/1.minimal` 文件路径下，编辑单节点模式部署模版 AllInOne.yml，将下列参数替换为您的真实参数。

  - `<IP1>`：节点 IP。
  - `<USERNAME>`：登陆节点的用户名，建议使用 root 或具有 root 权限的用户登陆。
  - `<PASSWORD>`：登陆节点的密码。
  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

例如，下面展示了一个 AllInOne.yml 示例：
<details>
<summary>AllInOne.yml 示例</summary>
```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
name: mini-hosts-conf
namespace: kubean-system
data:
hosts.yml: |
  all:
    hosts:
      node1:
        ip: 10.6.175.10 # 你的节点 IP
        access_ip: 10.6.175.10 # 你的节点 IP
        ansible_host: 10.6.175.10 # 你的节点 IP
        ansible_connection: ssh
        ansible_user: root # 登陆节点的用户名
        ansible_password: password01 # 登陆节点的密码
    children:
      kube_control_plane:
        hosts:
          node1:
      kube_node:
        hosts:
          node1:
      etcd:
        hosts:
          node1:
      k8s_cluster:
        children:
          kube_control_plane:
          kube_node:
      calico_rr:
        hosts: {}

---
apiVersion: v1
kind: ConfigMap
metadata:
name: mini-vars-conf
namespace: kubean-system
data:
group_vars.yml: |
  container_manager: containerd
  kube_network_plugin: calico
  etcd_deployment_type: kubeadm

---
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
name: cluster-mini
labels:
  clusterName: cluster-mini
spec:
hostsConfRef:
  namespace: kubean-system
  name: mini-hosts-conf
varsConfRef:
  namespace: kubean-system
  name:  mini-vars-conf

---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
name: cluster-mini-install-ops
spec:
cluster: cluster-mini
image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2 # kubean 镜像版本
backoffLimit: 0
actionType: playbook
action: cluster.yml
preHook:
  - actionType: playbook
    action: disable-firewalld.yml
postHook:
  - actionType: playbook
    action: cluster-info.yml
```
</details>

执行如下命令编辑 AllInOne.yml 配置模版：

```bash
$ vi kubean/examples/install/1.minimal/AllInOne.yml
```

#### 2.应用 AllInOne.yml 配置

完成上述步骤并保存 AllInOne.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/install/1.minimal/AllInOne.yml
```

至此，您已经完成了一个简单的单节点集群的部署。