# CRD 概述

## Cluster

Kubean 允许通过 custom resource definitions (CRDs) 来声明（唯一标识）一个 Kubernetes 集群。所有对集群的操作都基于此 CRD 里声明的内容。

下面是一份示例，帮助理解下文的配置项说明：

```yaml
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
  name: cluster1-offline-demo
spec:
  hostsConfRef:
    namespace: kubean-system
    name: cluster1-offline-demo-hosts-conf
  varsConfRef:
    namespace: kubean-system
    name: cluster1-offline-demo-vars-conf
```

### 配置项

#### 元数据

- `name`：name 用于声明一个集群，全局唯一

#### 属性关联

- `hostConfRef`：hostConfRef 是一个 ConfigMap 资源，它的内容应满足 ansible inventory 的格式，包含集群节点信息、类型分组信息。内容可参考 [demo](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/HostsConfCM.yml)。
  - `name`：表示其引用的 ConfigMap 的名称
  - `namespace`：表示其引用的 ConfigMap 所在的命名空间
  
- `varsConfRef`：varsConfRef 是一个 ConfigMap 资源，用作初始化或覆盖 Kubespray 中声明的变量值。如果有离线需求，这将很有用。内容可参考 [demo](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/VarsConfCM.yml)。
  - `name`：表示其引用的 ConfigMap 的名称
  - `namespace`：表示其引用的 ConfigMap 所在的命名空间

- `sshAuthRef`：sshAuthRef 是一个 Secret 资源，仅在 SSH 私钥模式时使用。
  - `name`：表示其引用的 Secret 名称
  - `namespace`：表示其引用的 Secret 所在的命名空间

## ClusterOperation

Kubean 允许通过 custom resource definitions (CRDs) 来声明对一个 Kubernetes 集群的操作（部署、升级等），前提是正确关联一个已经定义的 Cluster CRD。完成操作所必要的信息从其关联的 Cluster CRD 中获取。

下面是一份示例，帮助理解下文的配置项说明：

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster1-demo-ops-1
spec:
  cluster: cluster1-demo
  image: ghcr.m.daocloud.io/kubean-io/spray-job:latest
  backoffLimit: 0
  actionType: playbook
  action: cluster.yml
  preHook:
    - actionType: playbook
      action: ping.yml
    - actionType: playbook
      action: disable-firewalld.yml
  postHook:
    - actionType: playbook
      action: kubeconfig.yml
    - actionType: playbook
      action: cluster-info.yml
```

### 配置项

#### 元数据

- `name`：name 唯一标识一个对所关联集群的操作
  
#### 操作定义

- `cluster`：与此操作关联的集群名称，其值为 Cluster CRD 中声明的名称
- `image`：kubespray 镜像地址；可以使用 Kubean 仓库构建的镜像，也可使用自行构建镜像
- `actionType`：操作类型，目前支持指定 [`playbook`](https://docs.ansible.com/ansible/latest/user_guide/playbooks_intro.html) 或 `shell`
- `action`：意图执行的操作，目前支持指定 playbook 文件的路径或 shell 命令
- `preHook`：前置执行操作，可以指定多个，例如可以测试节点连通性等
  - `actionType`：同上
  - `action`：同上
- `postHook`：后置执行操作，可以指定多个，例如可以获取集群状态等
  - `actionType`：同上
  - `action`：同上
- `backoffLimit`：操作执行失败后重试次数

## Manifest

Kubean 允许通过 custom resource definitions (CRDs) 来记录和维护当前版本的 Kubean 使用和兼容的组件、包及版本；使用者不用手动编写此资源，由 Kubean 自行维护。

下面是一份示例，帮助理解下文的 spec 说明：

```yaml
apiVersion: kubean.io/v1alpha1
kind: Manifest
metadata:
  name: kubeaninfomanifest-v0-4-0-rc2
spec:
  components:
  - defaultVersion: v1.1.1
    name: cni
    versionRange:
    - v1.0.1
    - v1.1.1
  - defaultVersion: 1.6.9
    name: containerd
    versionRange:
    .......
    - 1.6.7
    - 1.6.8
    - 1.6.9
  - defaultVersion: ""
    name: kube
    versionRange:
    - v1.25.3
    - v1.25.2
    - v1.25.1
    ........
  - defaultVersion: v3.23.3
    name: calico
    versionRange:
    - v3.23.3
    - v3.22.4
    - v3.21.6
  - defaultVersion: v1.12.1
    name: cilium
    versionRange: []
  - defaultVersion: "null"
    name: etcd
    versionRange:
    - v3.5.3
    - v3.5.4
    - v3.5.5
  docker:
  - defaultVersion: "20.10"
    os: redhat-7
    versionRange:
    - latest
    - "18.09"
    - "19.03"
    - "20.10"
    - stable
    - edge
  - defaultVersion: "20.10"
    os: debian
    versionRange:
    - latest
    - "18.09"
    - "19.03"
    - "20.10"
    - stable
    - edge
  - defaultVersion: "20.10"
    os: ubuntu
    versionRange:
    - latest
    - "18.09"
    - "19.03"
    - "20.10"
    - stable
    - edge
  kubeanVersion: v0.4.0-rc2
  kubesprayVersion: c788620
```

### spec 说明

- `components`：镜像或二进制文件的版本声明
  - `name`：组件名称
  - `defaultVersion`：使用的默认版本
  - `versionRange`：受支持的版本列表
- `docker`：Docker 的版本管理
  - `os`：受支持的操作系统
  - `defaultVersion`：使用的默认版本
  - `versionRange`：受支持的版本列表
- `kubeanVersion`：Kubean 版本号
- `kubesprayVersion`：当前 Kubean 依赖的 Kubespray 版本号

## LocalArtifact

Kubean 允许通过 custom resource definitions (CRDs) 来记录离线包支持的组件及版本信息；使用者不用手动编写此资源，由 Kubean 自行维护。

下面是一份示例，帮助理解下文的 spec 说明：

```yaml
apiVersion: kubean.io/v1alpha1
kind: LocalArtifactSet
metadata:
  name: offlineversion-20221101
spec:
  arch: ["x86_64"]
  kubespray: "c788620"
  docker:
    - os: "redhat-7"
      versionRange:
        - "18.09"
        - "19.03"
        - "20.10"
    - os: "debian"
      versionRange: []
    - os: "ubuntu"
      versionRange: []
  items:
    - name: "cni"
      versionRange:
        - v1.1.1
    - name: "containerd"
      versionRange:
        - 1.6.9
    - name: "kube"
      versionRange:
        - v1.24.7
    - name: "calico"
      versionRange:
        - v3.23.3
    - name: "cilium"
      versionRange:
        - v1.12.1
    - name: "etcd"
      versionRange:
        - v3.5.4
```

### spec 说明

- `arch`：受支持的 CPU 指令集架构列表
- `kubespray`：使用的 Kubespray 版本
- `docker`：Docker 版本管理
  - `os`：Docker 受支持的操作系统类型
  - `versionRange`：受支持的 Docker 版本列表
- `items`：其他组件版本管理
  - `name`：组件名称
  - `versionRange`：该组件受支持的版本列表
