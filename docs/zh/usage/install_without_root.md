# 以非 root 用户部署集群

## 内容

- ✓ [1. sudo 权限校验](#sudo权限校验)
- ✓ [2. 创建主机清单配置](#创建主机清单配置)
- ✓ [3. 制备部署集群的配置参数](#制备部署集群的配置参数)
- ✓ [4. 准备 Kubean 的自定义资源](#准备Kubean的自定义资源)
- ✓ [5. 开始部署集群](#开始部署集群)

## sudo 权限校验

  安装过程中涉及系统特权操作，故用户需要具备 sudo 权限，可进行如下检查：

  1. 使用非 root 用户登录到目标节点

  2. 检查是否存在 sudo 命令，不存在则通过系统包管理器进行安装

     `which sudo`

  3. 在终端执行 `echo | sudo -S -v`

      若结果输出 `xxx is not in the sudoers file.  This incident will be reported` 或 `User xxx do not have sudo privilege` 等类似信息，即说明当前用户不具备 sudo 权限，反之说明当前用户具有 sudo 权限。

## 配置主机清单
   
  示例：主机清单 `HostsConfCM.yml` 内容大致如下，将下方<USERNAME> 和 <PASSWORD> 替换为实际的用户名和密码：

  ```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: sample-hosts-conf
    namespace: kubean-system
  data:
    hosts.yml: |
      all:
        hosts:
          master:
            ip: 192.168.10.11
            access_ip: 192.168.10.11
            ansible_host: 192.168.10.11
            ansible_connection: ssh
            ansible_user: <USERNAME>
            ansible_password: <PASSWORD>
            ansible_become_password: <PASSWORD>
          worker:
            ip: 192.168.10.12
            access_ip: 192.168.10.12
            ansible_host: 192.168.10.12
            ansible_connection: ssh
            ansible_user: <USERNAME>
            ansible_password: <PASSWORD>
            ansible_become_password: <PASSWORD>
        children:
          kube_control_plane:
            hosts:
              master:
          kube_node:
            hosts:
              master:
              worker:
          etcd:
            hosts:
              master:
          k8s_cluster:
            children:
              kube_control_plane:
              kube_node:
          calico_rr:
            hosts: {}
  ```
  > 注：如果在 /etc/sudoers 文件内该用户配置为 NOPASSWD（即无密码提权），可将 `ansible_become_password` 所在行注释

## 制备部署集群的配置参数

集群配置参数 `VarsConfCM.yml `的内容，可以参考
[demo vars conf](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/VarsConfCM.yml)。

```yaml
# VarsConfCM.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sample-vars-conf
  namespace: kubean-system
data:
  group_vars.yml: |
    container_manager: containerd
    kube_network_plugin: calico
    kube_network_plugin_multus: false
    kube_proxy_mode: iptables
    enable_nodelocaldns: false
    etcd_deployment_type: kubeadm
    ntp_enabled: true
    ...
```

## 准备 Kubean 的自定义资源

- Cluster 自定义资源内容示例

    ```yaml
    # Cluster.yml
    apiVersion: kubean.io/v1alpha1
    kind: Cluster
    metadata:
      name: sample
    spec:
      hostsConfRef:
        namespace: kubean-system
        name: sample-hosts-conf
      varsConfRef:
        namespace: kubean-system
        name: sample-vars-conf
      sshAuthRef: # 关键属性，指定集群部署期间的 ssh 私钥 secret
        namespace: kubean-system
        name: sample-ssh-auth
    ```

- ClusterOperation 自定义资源内容示例

    ```yaml
    # ClusterOperation.yml
    apiVersion: kubean.io/v1alpha1
    kind: ClusterOperation
    metadata:
      name: sample-create-cluster
    spec:
      cluster: sample
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

## 开始部署集群

假设所有 YAML 清单都存放在 `create_cluster` 目录：

```bash
$ tree create_cluster/
create_cluster
├── HostsConfCM.yml       # 主机清单
├── SSHAuthSec.yml        # SSH私钥
├── VarsConfCM.yml        # 集群参数
├── Cluster.yml           # Cluster CR
└── ClusterOperation.yml  # ClusterOperation CR
```

通过 `kubectl apply` 开始部署集群:

```bash
kubectl apply -f create_cluster/
```
