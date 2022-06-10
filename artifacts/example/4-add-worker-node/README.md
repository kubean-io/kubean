# 工作节点添加

### 首先，主机清单要添加新节点信息：

添加 node2 节点：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster1-hosts-conf
  namespace: kubean-system
data:
  hosts.yml: |
    all:
      hosts:
        ...
        node2:      # 设置 node2 的访问参数
          ip: 10.6.170.23
          access_ip: 10.6.170.23
          ansible_host: 10.6.170.23
          ansible_connection: ssh
          ansible_user: root
          ansible_ssh_pass: daocloud
      children:
        kube_control_plane:
          hosts:
            ...
        kube_node:
          hosts:
            ...
            node2:  # 将 node2 添加为 worker 节点
        etcd:
          hosts:
            ...
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
```

### 新增节点时，KuBeanClusterOps 的关键参数：
```yaml
apiVersion: kubeanclusterops.kubean.io/v1alpha1
kind: KuBeanClusterOps
metadata:
  name: cluster1-ops-xxx
spec:
  ...
  actionType: playbook
  action: scale.yml         # 工作节点添加要运行 scale.yml 的 playbook
  extraArgs: --limit=node2  # 使用 limit 指定要添加的节点名称
  ...
```

### kubespray 新增节点相关参数：

> vars-conf-cm 中无需修改任何参数，保持与 install cluster 一致的参数即可；

节点相关文档: [Adding/replacing a node](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/nodes.md)