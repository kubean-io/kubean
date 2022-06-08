# 集群升级

### 升级时，KuBeanClusterOps 的关键参数：
```yaml
apiVersion: kubeanclusterops.kubean.io/v1alpha1
kind: KuBeanClusterOps
metadata:
  name: cluster1-ops-xxx
spec:
  ...
  actionType: playbook          # 执行 action 的类型
  action: upgrade-cluster.yml   # 指定 upgrade-cluster playbook
  ...
```

### kubespray 升级相关参数：
> 注： 参数位于 var-conf-cm.yml
```yaml
    # [必填] 指定要升级的版本
    kube_version: "v1.22.0"

    # [选填] 升级每个节点前，暂停60秒，60秒后将会恢复playbook执行
    # upgrade_node_pause_seconds: 60

    # [选填] 升级每个节点后，暂停60秒，60秒后将会恢复playbook执行
    # upgrade_node_post_upgrade_pause_seconds: 60
```

升级相关文档：
[Upgrading Kubernetes in Kubespray](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/upgrades.md)

### 如何查看 kubespray 每个版本所支持的 kubernetes 版本范围？

kubespray repo 先 checkout 到对应 tag，然后到 [roles/download/defaults/main.yml](https://github.com/kubernetes-sigs/kubespray/blob/master/roles/download/defaults/main.yml#L470) 可见,
比如：
```shell
kubeadm_checksums:
  arm:
    v1.24.1: 1c0b22c941badb40f4fb93e619b4a1c5e4bba7c1c7313f7c7e87d77150f35153
    v1.24.0: c463bf24981dea705f4ee6e547abd5cc3b3e499843f836aae1a04f5b80abf4c2
    v1.23.7: 18da04d52a05f2b1b8cd7163bc0f0515a4ee793bc0019d2cada4bbf3323d4044
    v1.23.6: da2221f593e63195736659e96103a20e4b7f2060c3030e8111a4134af0d37cfb
    v1.23.5: 9ea3e52cb236f446a33cf69e4ed6ac28a76103c1e351b2675cb9bfcb77222a61
    v1.23.4: 9ca72cf1e6bbbe91bf634a18571c84f3fc36ba5fcd0526b14432e87b7262a5ee
    v1.23.3: cb2513531111241bfb0f343cff18f7b504326252ae080bb69ad1ccf3e31a2753
    v1.23.2: 63a6ca7dca76475ddef84e4ff84ef058ee2003d0e453b85a52729094025d158e
    v1.23.1: 77baac1659f7f474ba066ef8ca67a86accc4e40d117e73c6c76a2e62689d8369
    v1.23.0: b59790cdce297ac0937cc9ce0599979c40bc03601642b467707014686998dbda
    v1.22.10: f1ab42fbadb0a66ba200392ee82c05b65e3d29a3d8f3e030b774cbc48915dedb
    v1.22.9: f68ca35fc71691e599d4913de58b6d77abcb2d27c324abc23388b4383b5299ea
    ...
```
