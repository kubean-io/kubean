# 自定义 Action

## 动机

对于使用者来讲，Kubean 和 Kubesprary 的产物都是 OCI 镜像、Helm Chart 及 K8s manifest。
在已拿到这些产物的情况下要自定义一些操作，也能做到，但是会比较复杂，需要手动修改不少的配置。希望能够简化这一过程。

## 目标

提供一种便捷的方式能够让使用者使用一些自定义的操作来查看、修改和控制集群节点的状态。

## CRD 设计

1. 增加 ActionSource 字段以声明 Action 来源，其值目前支持：

    - builtin（缺省值）

        表明使用 kubean 内建 ansible playbook 或在 manifest 内联的 shell 脚本

    - configmap

        表明需要的 ansible playbook 或 shell 脚本通过 引用 K8s configmap 来获取

2. 增加 ActionSourceRef 字段以声明当 ActionSource 值为 configmap 时所引用的资源对象，且仅当 ActionSource 为 configmap 时此字段才生效，其格式为：

    ```yaml
    actionSourceRef:
      name: <configmap name>
      namespace: <namespace of configmap>
    ```

配置示例：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster1-demo-myaction
  namespace: kubean-system
data:
  myplaybook.yml: |
    - hosts: k8s_cluster
      gather_facts: false
      become: yes
      any_errors_fatal: "{{ any_errors_fatal | default(true) }}"
      tasks:
        - name: Print inventory hostname
          debug:
            msg: "inventory_hostname is {{ inventory_hostname }}"
  hello.sh: |
    echo "hello world!"
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster1-demo-ops-1
spec:
  cluster: cluster1-demo
  image: ghcr.io/kubean-io/spray-job:latest
  backoffLimit: 0
  actionType: playbook
  action: myplaybook.yml
  actionSource: configmap
  actionSourceRef:
    name: cluster1-demo-myaction
    namespace: kubean-system
  preHook:
    - actionType: shell
      action: hello.sh
      actionSource: configmap
      actionSourceRef:
        name: cluster1-demo-myaction
        namespace: kubean-system
```
