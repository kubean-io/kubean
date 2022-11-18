# :key: 使用 SSH 秘钥方式部署 K8S 集群

> [English](../en/sshkey_deploy_cluster.md) | 中文
## 内容

* ✓ [1. SSH 秘钥的生成与分发](#SSH秘钥的生成与分发)
* ✓ [2. 使用私钥制作 Secret](#使用私钥制作Secret)
* ✓ [3. 创建主机清单配置](#创建主机清单配置)
* ✓ [3. 制备部署集群的配置参数](#制备部署集群的配置参数)
* ✓ [4. 准备 KuBean 的自定义资源](#准备KuBean的自定义资源)
* ✓ [5. 开始部署集群](#开始部署集群)

## SSH秘钥的生成与分发

1. 通过 `ssh-keygen` 命令生成公私钥对，比如：
``` bash
$ ssh-keygen
Generating public/private rsa key pair.
Enter file in which to save the key (/root/.ssh/id_rsa):
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in /root/.ssh/id_rsa.
Your public key has been saved in /root/.ssh/id_rsa.pub.
The key fingerprint is:
SHA256:XBSD2HY1Lp8ZRfTC82cFEXzW/BRgEMd+SWiKzBNSUHN root@localhost.localdomain
The key's randomart image is:
+---[RSA 2048]----+
|       +B=E*XO*O.|
|      . =X =o=O.=|
|      .oo o oo++o|
|       + =   + .+|
|        S     . .|
|                 |
|                 |
|                 |
|                 |
+----[SHA256]-----+

$ ls /root/.ssh/id_rsa* -lh
-rw-------. 1 root root 1.7K Nov 10 03:47 /root/.ssh/id_rsa         # 私钥
-rw-r--r--. 1 root root  408 Nov 10 03:47 /root/.ssh/id_rsa.pub     # 公钥
```

2. 分发公钥到集群的各个节点：
``` bash
# 比如指定将公钥分发至 `192.168.10.11` `192.168.10.12` 两个节点
$ declare -a IPS=(192.168.10.11 192.168.10.12)

# 遍历节点 IP 分发公钥(/root/.ssh/id_rsa.pub)，假设用户名为: root, 密码为: kubean
$ for ip in ${IPS[@]}; do sshpass -p "kubean" ssh-copy-id -i /root/.ssh/id_rsa.pub -o StrictHostKeyChecking=no root@$ip; done
```

## 使用私钥制作Secret

1. 通过 kubectl 命令可以生成私钥的 Secret:
``` bash
$ kubectl -n kubean-system \                            # 指定命名空间 kubean-system
    create secret generic sample-ssh-auth \             # 指定 secret 名称为 sample-ssh-auth
    --type='kubernetes.io/ssh-auth' \                   # 指定 secret 类型为 kubernetes.io/ssh-auth
    --from-file=ssh-privatekey=/root/.ssh/id_rsa \      # 指定 ssh 私钥文件路径
    --dry-run=client -o yaml > ssh_auth_sec.yaml        # 指定 secret yaml 文件生成路径
```

2. 生成的 Secret YAML 内容大致如下所示：
``` yaml
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: sample-ssh-auth
  namespace: kubean-system
type: kubernetes.io/ssh-auth
data:
  ssh-privatekey: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBdWVDbC8rSng1b0RT...
```

## 创建主机清单配置

示例：主机清单 hosts_conf_cm.yaml 内容大致如下：
``` yaml
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
        worker:
          ip: 192.168.10.12
          access_ip: 192.168.10.12
          ansible_host: 192.168.10.12
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

> 注： 由于采用私钥登录，所以主机信息中不需要填写用户名密码(即: ansible_user、ansible_password)

## 制备部署集群的配置参数

集群配置参数 vars_conf_cm.yaml 的内容，可以参考: [demo vars conf](../../artifacts/demo/vars-conf-cm.yml).
``` yaml
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

## 准备KuBean的自定义资源

1. Cluster 自定义资源内容示例
``` yaml
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
  sshAuthRef:                   # 关键属性，指定集群部署期间的 ssh 私钥 secret
    namespace: kubean-system
    name: sample-ssh-auth
```

2. ClusterOperation 自定义资源内容示例
``` yaml
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

假设我们的所有 yaml 清单都存放在 create_cluster 目录
``` bash
$ tree create_cluster/
create_cluster
├── hosts_conf_cm.yml       # 主机清单
├── ssh_auth_sec.yml        # SSH私钥
├── vars_conf_cm.yml        # 集群参数
├── kubeanCluster.yml       # Cluster CR
└── kubeanClusterOps.yml    # ClusterOperation CR
```

通过 kubectl apply 开始部署集群:
``` bash
$ kubectl apply -f create_cluster/
```
