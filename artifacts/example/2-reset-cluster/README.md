# 集群重置

注：集群重置的含义，即直接移除集群，并非移除后再安装！

### 重置时，KuBeanClusterOps 的关键参数：
```yaml
apiVersion: kubeanclusterops.kubean.io/v1alpha1
kind: KuBeanClusterOps
metadata:
  name: cluster1-ops-xxx
spec:
  ...
  actionType: playbook # 执行 action 的类型
  action: reset.yml    # 指定 reset playbook
  ...
```

### kubespray 相关参数：

> vars-conf-cm 中没有重置的相关参数，保持与 install cluster 一致的参数即可；
