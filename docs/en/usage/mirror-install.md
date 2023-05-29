# Deploying a Cluster using Accelerated Mode

## Prerequisites

1. You already have a standard Kubernetes cluster or a cluster provided by a cloud provider.
2. The control node or cloud terminal for the cluster has the [kubect](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) tool installed.。
3. The [kubean helm chart](helm-install-kubean.md) has been deployed on your cluster.
4. The [kubean 项目](https://github.com/kubean-io/kubean)has been cloned to your local machine. If you haven't cloned kubean yet, you can do so by executing the following command:

```bash
$ git clone https://github.com/kubean-io/kubean.git
```

---

## Getting Started

This tutorial will use the `kubean/example/2.mirror` file that you cloned to your local machine as an example template for demonstrating cluster deployment using accelerated mode.

The `2.mirror` accelerated deployment template already contains built-in acceleration parameter configurations. 
You only need to modify the host information and other relevant information in the two configuration template files, **`HostsConfCM.yml` and `ClusterOperation.yml`**, located in the /2.mirror file path.

<details open>
<summary>The main configuration files and purposes inside the `2.mirror` file are as follows:</summary>

```yaml
    .2.mirror
    ├── Cluster.yml                        # The main configuration files and their purposes in the `2.mirror` file are as follows:
    ├── ClusterOperation.yml        # kubean version and task configuration
    ├── HostsConfCM.yml              # Node information configuration for the cluster to be built
    └── VarsConfCM.yml                # Configuration for acceleration and other features
```
</details>

#### 1.Configure Host Parameters in HostsConfCM.yml

Navigate to the `kubean/examples/install/2.mirror/` path and edit the `HostsConfCM.yml`template for the node configuration information of the cluster to be built. Replace the following parameters with your actual parameters:

  - `<IP1>`：Node IP.
  - `<USERNAME>`： Username for logging in to the node. We recommend using root or a user with root privileges to log in.
  - `<PASSWORD>`：Password for logging in to the node.

For example, the following is an example HostsConfCM.yml file:
<details>
<summary> HostsConfCM.yml Example</summary>
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: online-hosts-conf
  namespace: kubean-system
data:
  hosts.yml: |
    all:
      hosts:
        node1:
          ip: 10.6.175.10 # Your node IP
          access_ip: 10.6.175.10 # Your node IP
          ansible_host: 10.6.175.10 # Your node IP
          ansible_connection: ssh
          ansible_user: root # The username for logging into the node
          ansible_password: password01 # The password for logging into the node
        node2:
          ip: 10.6.175.20 # Your node2 IP
          access_ip: 10.6.175.20 # Your node2 IP
          ansible_host: 10.6.175.20 # Your node2 IP
          ansible_connection: ssh
          ansible_user: root # The username for logging into the node2
          ansible_password: password02 # The password for logging into the node2
      children:
        kube_control_plane: # Configuring the control node
          hosts:
            node1:
        kube_node: # Configuring the working nodes of the cluster
          hosts:
            node1:
            node2:
        etcd: # Configuring the ETCD nodes of the cluster
          hosts:
            node1:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
```
</details>

Execute the following command to edit the HostsConfCM.yml configuration template:

```bash
$ vi kubean/examples/install/2.mirror/HostsConfCM.yml
```

#### 2. Configure kubean Task Parameters in ClusterOperation.yml

Navigate to the `kubean/examples/install/2.mirror/` path and edit the `ClusterOperation.yml` template for the configuration information of the cluster to be built. Replace the following parameters with your actual parameters:

  - `<TAG>`: kubean image version. We recommend using the latest version.[Refer to the kubean version list](https://github.com/kubean-io/kubean/tags)

For example, the following is an example `ClusterOperation.yml` file:
<details>
<summary> ClusterOperation.yml Example</summary>
```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster1-online-install-ops
spec:
  cluster: cluster1-online
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2 # kubean image version
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
</details>

To edit the ClusterOperation.yml configuration template, run the following command:

```bash
$ vi kubean/examples/install/2.mirror/ClusterOperation.yml
```

#### 3.Apply all configurations under the 2.mirror directory

After completing the above steps and saving the HostsConfCM.yml and ClusterOperation.yml files, run the following command:

```bash
$ kubectl apply -f examples/install/2.mirror
```

With this, you have completed the deployment of a cluster using the acceleration mode.