# 集群控制平面扩缩容

在 Kubernetes 集群运维过程中，为了提高集群的高可用性和性能，我们常常需要对控制平面节点进行扩缩容操作。对于使用 kubean 部署的集群，我们可以使用声明式的方式，快速扩缩容集群控制平面节点。

在您克隆至本地的 `kubean/example/scale` 文件内，提供了控制平面节点扩缩容的样例模板：

<details open>
<summary> scale 文件内主要的配置文件及用途如下：</summary>

```yaml
    scale
    ├── 3.addControlPlane                         # 增加控制平面节点模板
    │   ├── ClusterOperation.yml                       # kubean 版本及任务配置
    │   └── HostsConfCM.yml                            # 当前集群的节点信息配置
    └── 4.delControlPlane                         # 删除控制平面节点模板
        ├── ClusterOperation.yml                       # kubean 版本及任务配置
        ├── ClusterOperation2.yml                      # kubean 版本及任务配置
        └── HostsConfCM.yml                            # 当前集群的节点信息配置
```
</details>

下面以已有的单控制平面节点集群为例，来演示集群控制平面节点的扩缩容操作。
> 注意：执行集群控制平面扩缩容前，您必须已经使用 kubean 完成了一套集群的部署。

## 扩容控制平面节点

#### 1. 向 HostsConfCM.yml 增加新控制平面节点主机参数

我们要在原有的单控制平面节点集群中，对名为 `mini-hosts-conf` 的 ConfigMap 进行新增节点配置，在原来 `node1` 控制平面节点的基础上, 新增 `node2`, `node3` 控制平面节点；

具体地，我们可以进入 `kubean/examples/scale/3.addControlPlane/` 路径，编辑已准备好的节点配置 ConfigMap 模板 `HostsConfCM.yml`，将下列参数替换为您的真实参数：

  - `<IP2>`, `<IP3>`：节点 IP。
  - `<USERNAME>`：登陆节点的用户名，建议使用 root 或具有 root 权限的用户登陆。
  - `<PASSWORD>`：登陆节点的密码。

`kubean/examples/scale/3.addControlPlane/` 路径下 **`HostsConfCM.yml`** 的模板内容如下：


```yaml
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
          ip: <IP1>
          access_ip: <IP1>
          ansible_host: <IP1>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
        node2:
          ip: <IP2>
          access_ip: <IP2>
          ansible_host: <IP2>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
        node3:
          ip: <IP3>
          access_ip: <IP3>
          ansible_host: <IP3>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
      children:
        kube_control_plane:
          hosts:
            node1:
            node2:
            node3:
        kube_node:
          hosts:
            node1:
            node2:
            node3:
        etcd:
          hosts:
            node1:
            node2:
            node3:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
```
</details>

**重要参数：**
>* `all.hosts.node1`: 原集群中已存在的控制平面节点
>* `all.hosts.node2`, `all.hosts.node3`: 集群扩容待新增的控制平面节点
>* `all.children.kube_control_plane.hosts`: 集群中的控制平面节点组
>* `all.children.etcd.hosts`: 集群中的 etcd 节点组，通常与控制平面节点保持一致

!!! 新增控制平面节点主机参数的示例

    === "新增节点前"

        ``` yaml
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
                  ip: 10.6.175.10 # 控制平面节点 IP
                  access_ip: 10.6.175.10 # 控制平面节点 IP
                  ansible_host: 10.6.175.10 # 控制平面节点 IP
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
        ```

    === "新增节点后"

        ``` yaml
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
                  ip: 10.6.175.10 # 控制平面节点 IP
                  access_ip: 10.6.175.10 # 控制平面节点 IP
                  ansible_host: 10.6.175.10 # 控制平面节点 IP
                  ansible_connection: ssh
                  ansible_user: root # 登陆节点的用户名
                  ansible_password: password01 # 登陆节点的密码
                node2:
                  ip: 10.6.175.20 # 新增控制平面节点 IP
                  access_ip: 10.6.175.20 # 工作节点 IP
                  ansible_host: 10.6.175.20 # 工作节点 IP
                  ansible_connection: ssh
                  ansible_user: root # 登陆节点的用户名
                  ansible_password: password01 # 登陆节点的密码
                node3:
                  ip: 10.6.175.30 # 新增控制平面节点 IP
                  access_ip: 10.6.175.30 # 新增控制平面节点 IP
                  ansible_host: 10.6.175.30 # 新增控制平面节点 IP
                  ansible_connection: ssh
                  ansible_user: root # 登陆节点的用户名
                  ansible_password: password01 # 登陆节点的密码
              children:
                kube_control_plane:
                  hosts:
                    node1:
                    node2:
                    node3:
                kube_node:
                  hosts:
                    node1:
                    node2:
                    node3:
                etcd:
                  hosts:
                    node1:
                    node2:
                    node3:
                k8s_cluster:
                  children:
                    kube_control_plane:
                    kube_node:
                calico_rr:
                  hosts: {}
        ```


#### 2. 通过 ClusterOperation.yml 新增控制平面扩容任务  

进入 `kubean/examples/scale/3.addControlPlane/` 路径，编辑模板 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/scale/3.addControlPlane/` 路径下 **`ClusterOperation.yml`** 的模板内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-acp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.26.4
  actionType: playbook
  action: cluster.yml
  extraArgs: >-
    --limit=etcd,kube_control_plane
    -e ignore_assert_errors=true
  postHook:
    - actionType: playbook
      action: upgrade-cluster.yml
      extraArgs: >-
        --limit=etcd,kube_control_plane
        -e ignore_assert_errors=true
```
**重要参数：**
>* `spec.cluster`: 指定需要扩容控制平面节点的集群名称，上述指定的是名为 `cluster-mini` 的集群为扩容目标。
>* `spec.action` 指定扩容控制平面节点的 Kubespray 剧本, 这里设置为 `cluster.yml`。
>* `spec.extraArgs`: 指定扩容的节点限制，这里通过 `--limit=` 参数限定扩容 `etcd`, `control-plane` 节点组。
>* `spec.postHook.action` 指定扩容控制平面节点的 Kubespray 剧本, 这里设置为 `upgrade-cluster.yml`，更新集群中所有的 Etcd 配置。
>* `spec.postHook.extraArgs`: 指定扩容的节点限制，这里通过 `--limit=` 参数限定扩容 `etcd`, `control-plane` 节点组。

例如，下面展示了一个 ClusterOperation.yml 示例：
<details>
<summary> ClusterOperation.yml 示例</summary>
```yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-acp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.26.4
  actionType: playbook
  action: cluster.yml
  extraArgs: >-
    --limit=etcd,kube_control_plane
    -e ignore_assert_errors=true
  postHook:
    - actionType: playbook
      action: upgrade-cluster.yml
      extraArgs: >-
        --limit=etcd,kube_control_plane
        -e ignore_assert_errors=true
```
</details>

#### 3.应用 `scale/3.addControlPlane` 文件下所有的配置

完成上述步骤并保存 HostsConfCM.yml 和 ClusterOperation.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/scale/3.addControlPlane/
```

#### 4. 重启 kube-system/nginx-proxy

如果控制平面和普通节点分离，需要在所有普通节点上重启 nginx-proxy pod，这个 pod 是 api server 的本地代理。Kubean 将更新它的静态配置，但是需要重新启动它才能重新加载。

```bash
crictl ps | grep nginx-proxy | awk '{print $1}' | xargs crictl stop
```

至此，您已经完成了一个集群的控制平面节点扩容。

## 缩容控制平面节点

> 注意：在缩容控制平面节点前，请确保集群中至少保留一个控制平面节点，以保证集群的正常运行。

#### 1. 通过 ClusterOperation.yml 新增控制平面缩容任务 

进入 `kubean/examples/scale/4.delControlPlane/` 路径，编辑模板 `ClusterOperation.yml`，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/scale/4.delControlPlane/` 路径下 **`ClusterOperation.yml`** 的模板内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.26.4
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2,node3 -e reset_nodes=false -e allow_ungraceful_removal=true
```
**重要参数：**
>* `spec.cluster`: 指定需要缩容控制平面节点的集群名称, 上述指定的是名为 cluster-mini 的集群为缩容目标。
>* `spec.action`: 指定缩容节点的 kubespray 剧本, 这里设置为 remove-node.yml。
>* `spec.extraArgs`: 指定缩容的节点及相关参数，这里通过 -e 参数指定：
>   * `node=node2,node3`: 指定要移除的控制平面节点名称
>   * `reset_nodes=false`: 不重置节点（保留节点上的数据，也可当节点不可访问时使用）
>   * `allow_ungraceful_removal=true`: 允许非优雅移除（当节点已不可访问时使用）

例如，下面展示了一个 ClusterOperation.yml 示例：
<details>
<summary> ClusterOperation.yml 示例</summary>
```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.26.4
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2,node3
```
</details>


#### 2.应用 `scale/4.delControlPlane` 目录下的 ClusterOperation 缩容任务清单

完成上述步骤并保存 ClusterOperation.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/scale/4.delControlPlane/ClusterOperation.yml
```

默认进入 kubean-system 命名空间，查看缩容任务执行状态：
``` bash
$ kubectl -n kubean-system get pod | grep cluster-mini-dcp-ops
```
了解缩容任务执行进度，可查看该 pod 日志；

#### 3. 通过 HostsConfCM.yml 删除控制平面节点主机参数

我们已经通过如上两步操作执行了缩容任务，待缩容任务执行完成后，`node2`, `node3` 将从现有集群中永久移除，则此时我们还需要将节点配置相关 Configmap 中的 node2, node3 信息移除;

进入 `kubean/examples/scale/4.delControlPlane/` 路径，编辑已准备好的节点配置模板 `HostsConfCM.yml`，删除需要移除的控制平面节点配置。

**删除参数如下：**

* `all.hosts` 下的 node3 节点接入参数。
* `all.children.kube_control_plane.hosts` 内的主机名称 node3。
* `all.children.kube_node.hosts` 内的主机名称 node3（如果该节点同时也是工作节点）。
* `all.children.etcd.hosts` 内的主机名称 node3（如果该节点同时也是 etcd 节点）。

!!! 移除控制平面节点主机参数的示例

    === "移除节点前"

        ``` yaml
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
                  ip: 10.6.175.10 # 控制平面节点 IP
                  access_ip: 10.6.175.10 # 控制平面节点 IP
                  ansible_host: 10.6.175.10 # 控制平面节点 IP
                  ansible_connection: ssh
                  ansible_user: root # 登陆节点的用户名
                  ansible_password: password01 # 登陆节点的密码
                node2:
                  ip: 10.6.175.20 # 控制平面节点 IP
                  access_ip: 10.6.175.20 # 控制平面节点 IP
                  ansible_host: 10.6.175.20 # 控制平面节点 IP
                  ansible_connection: ssh
                  ansible_user: root # 登陆节点的用户名
                  ansible_password: password01 # 登陆节点的密码
                node3:
                  ip: 10.6.175.30 # 控制平面节点 IP
                  access_ip: 10.6.175.30 # 控制平面节点 IP
                  ansible_host: 10.6.175.30 # 控制平面节点 IP
                  ansible_connection: ssh
                  ansible_user: root # 登陆节点的用户名
                  ansible_password: password01 # 登陆节点的密码
              children:
                kube_control_plane:
                  hosts:
                    node1:
                    node2:
                    node3:
                kube_node:
                  hosts:
                    node1:
                    node2:
                    node3:
                etcd:
                  hosts:
                    node1:
                    node2:
                    node3:
                k8s_cluster:
                  children:
                    kube_control_plane:
                    kube_node:
                calico_rr:
                  hosts: {}
        ```

    === "移除节点后"

        ``` yaml
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
                  ip: 10.6.175.10 # 控制平面节点 IP
                  access_ip: 10.6.175.10 # 控制平面节点 IP
                  ansible_host: 10.6.175.10 # 控制平面节点 IP
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
        ```

完成上述步骤并保存 HostsConfCM.yml 文件后，执行如下命令：

```bash
$ kubectl apply -f examples/scale/4.delControlPlane/HostsConfCM.yml
```

#### 4. 更新 Kubernetes 和网络配置文件

运行`cluster.yml`以在所有其余节点上重新生成配置文件。

进入 kubean/examples/scale/4.delControlPlane/ 路径，编辑模板 ClusterOperation2.yml，将下列参数替换为您的真实参数：

  - `<TAG>`：kubean 镜像版本，推荐使用最新版本，[参阅 kubean 版本列表](https://github.com/kubean-io/kubean/tags)。

`kubean/examples/scale/4.delControlPlane/` 路径下 **`ClusterOperation2.yml`** 的模板内容如下：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops-2
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.26.4
  actionType: playbook
  action: cluster.yml
```
**重要参数：**
>* `spec.cluster`: 指定需要缩容控制平面节点的集群名称, 上述指定的是名为 cluster-mini 的集群为缩容目标。
>* `spec.action`: 指定缩容节点的 kubespray 剧本, 这里设置为 cluster.yml。

例如，下面展示了一个 ClusterOperation.yml 示例：
<details>
<summary> ClusterOperation.yml 示例</summary>
```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops-2
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.26.4
  actionType: playbook
  action: cluster.yml
```
</details>

#### 5. 重启 kube-system/nginx-proxy

如果控制平面和普通节点分离，需要在所有普通节点上重启 nginx-proxy pod，这个 pod 是 api server 的本地代理。Kubean 将更新它的静态配置，但是需要重新启动它才能重新加载。

```bash
crictl ps | grep nginx-proxy | awk '{print $1}' | xargs crictl stop
```

此时，我们已将 node2, node3 控制平面节点从集群中移除，并且清理掉了有关 node2, node3 的主机信息，更新了集群配置，整个控制平面缩容操作就此结束。

## 注意事项

1. **高可用性考虑**：在生产环境中，建议至少保留 3 个控制平面节点，以确保集群的高可用性。

2. **etcd 集群扩缩容**：控制平面节点通常也是 etcd 节点，扩缩容控制平面节点时，需要特别注意 etcd 集群的节点数量应该是奇数（1、3、5等），以确保 etcd 集群的高可用性和一致性。

3. **负载均衡**：当有多个控制平面节点时，建议配置负载均衡器，以确保 API 服务器的高可用性。

4. **备份**：在进行控制平面节点扩缩容操作前，建议先备份 etcd 数据，以防操作失败导致数据丢失。