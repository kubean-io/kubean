# Deploy Kubernetes clusters with SSH

> English | [中文](../zh/sshkey_deploy_cluster.md)

**Contents**

* ✓ 1. [Generate and distribute an SSH private key](#generate-and-distribute-an-ssh-private-key)
* ✓ 2. [Make a Secret with private key](#make-a-secret-with-private-key)
* ✓ 3. [Create a host configuration file](#create-a-host-configuration-file)
* ✓ 4. [Provision parameters for cluster deployment](#provision-parameters-for-cluster-deployment)
* ✓ 5. [Prepare KuBean's CRDs](#prepare-kubeans-crds)
* ✓ 6. [Deploy a cluster](#deploy-a-cluster)

## Generate and distribute an SSH private key

1. Generate a pair of public-private keys with `ssh-keygen` command:

    ``` bash
    $ ssh-keygen -t rsa -b 4096 -C "your_email@example.com" -f $HOME/.ssh/id_rsa
    Generating public/private rsa key pair.
    Created directory '/root/.ssh'.
    Enter passphrase (empty for no passphrase):
    Enter same passphrase again:
    Your identification has been saved in /root/.ssh/id_rsa.
    Your public key has been saved in /root/.ssh/id_rsa.pub.
    The key fingerprint is:
    SHA256:oMqlhL8wLuYycOkUNXyiDso62C+ryNYc9k3LMDltQZs your_email@example.com
    The keys randomart image is:
    +---[RSA 4096]----+
    |   .             |
    |    = ..         |
    |   o +o o        |
    |..o  . E         |
    |+o.oo o S        |
    |o==* = +         |
    |*=O o O .        |
    |@=++ . +         |
    |OBo+.            |
    +----[SHA256]-----+

    $ ls /root/.ssh/id_rsa* -lh
    -rw-------. 1 root root 1.7K Nov 10 03:47 /root/.ssh/id_rsa         # 私钥
    -rw-r--r--. 1 root root  408 Nov 10 03:47 /root/.ssh/id_rsa.pub     # 公钥
    ```

2. Distribute the key pair to nodes of the cluster to be deployed:

    ``` bash
    # for example, specify to distribute the public key to nodes `192.168.10.11` and `192.168.10.12`.
    $ declare -a IPS=(192.168.10.11 192.168.10.12)

    # traverse node IPs to distribute the public key (/root/.ssh/id_rsa.pub) with the presumptive account/password: root/kubean
    $ for ip in ${IPS[@]}; do sshpass -p "kubean" ssh-copy-id -i /root/.ssh/id_rsa.pub -o StrictHostKeyChecking=no root@$ip; done
    ```

## Make a Secret with private key

1. Generate a Secret for the private key with the following command:

    ``` bash
    $ kubectl -n kubean-system \                            # specify namespace: kubean-system
        create secret generic sample-ssh-auth \             # specify the name of Secret: sample-ssh-auth
        --type='kubernetes.io/ssh-auth' \                   # specify the type of Secret: kubernetes.io/ssh-auth
        --from-file=ssh-privatekey=/root/.ssh/id_rsa \      # specify the filepath of the ssh private key
        --dry-run=client -o yaml > SSHAuthSec.yml           # specify the target path of the new Secret YAML
    ```

The expected `SSHAuthSec.yml` looks like:

    ``` yaml
    # SSHAuthSec.yml
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

## Create a host configuration file

The `HostsConfCM.yml` file looks like:

``` yaml
# HostsConfCM.yml
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

> Note: No need to include the account and password (`ansible_user` and `ansible_password`) because of logging in with the private key.

## Provision parameters for cluster deployment

For contents of `VarsConfCM.yaml`, refer to [demo vars conf](../../examples/install/2.mirror/VarsConfCM.yml).

``` yaml
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

## Prepare KuBean's CRDs

* Example of a `Cluster` CR:

    ``` yaml
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
      sshAuthRef:                   # key field: specifies the Secret of the ssh private key for cluster deployment
        namespace: kubean-system
        name: sample-ssh-auth
    ```

* Example of a `ClusterOperation` CR:

    ``` yaml
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

## Deploy a cluster

Suppose all our YAML manifests are stored in the `create_cluster` directory:

``` bash
$ tree create_cluster/
create_cluster
├── HostsConfCM.yml       # 主机清单
├── SSHAuthSec.yml        # SSH私钥
├── VarsConfCM.yml        # 集群参数
├── Cluster.yml           # Cluster CR
└── ClusterOperation.yml  # ClusterOperation CR
```

Deploy a cluster with `kubectl apply`:

``` bash
kubectl apply -f create_cluster/
```
