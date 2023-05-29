# Minimal Deployment

## Prerequisites

1. You have a standard Kubernetes cluster or a cluster provided by a cloud provider.
2. The [kubectl tool](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) has been installed on your cluster control node or cloud terminal.
3. The [kubean helm chart](helm-install-kubean.md) has been deployed on your cluster.
4. The [kubean project](https://github.com/kubean-io/kubean) has been cloned to your local machine. If you haven't cloned kubean yet, you can execute the following command to clone it:

```bash
$ git clone https://github.com/kubean-io/kubean.git
```

---

## Deployment

In this tutorial, we will use the `kubean/example` file cloned to your local machine as a template for demonstration purposes.

With the help of the example template, we can use kubean to complete the deployment of a single-node cluster in just two steps.

#### 1. Configure the AllInOne.yml parameters

Navigate to the `kubean/examples/install/1.minimal`  file path, edit the AllInOne.yml template for single-node mode deployment, and replace the following parameters with your actual parameters.

  - `<IP1>`: Node IP.
  - `<USERNAME>`: The username for logging into the node. It is recommended to use root or a user with root privileges to log in.
  - `<PASSWORD>`: The password for logging into the node.
  - `<TAG>`: kubean image version, it is recommended to use the latest version, [Refer to the kubean version list](https://github.com/kubean-io/kubean/tags).

For example, the following shows an example of AllInOne.yml:
<details>
<summary>Example of AllInOne.yml</summary>
```yaml
---
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
        ip: 10.6.175.10 # Your node IP
        access_ip: 10.6.175.10 # Your node IP
        ansible_host: 10.6.175.10 # Your node IP
        ansible_connection: ssh
        ansible_user: root # The username for logging into the node
        ansible_password: password01 # The password for logging into the node
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

---
apiVersion: v1
kind: ConfigMap
metadata:
name: mini-vars-conf
namespace: kubean-system
data:
group_vars.yml: |
  container_manager: containerd
  kube_network_plugin: calico
  etcd_deployment_type: kubeadm

---
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
name: cluster-mini
labels:
  clusterName: cluster-mini
spec:
hostsConfRef:
  namespace: kubean-system
  name: mini-hosts-conf
varsConfRef:
  namespace: kubean-system
  name:  mini-vars-conf

---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
name: cluster-mini-install-ops
spec:
cluster: cluster-mini
image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2 # kubean image version
backoffLimit: 0
actionType: playbook
action: cluster.yml
preHook:
  - actionType: playbook
    action: disable-firewalld.yml
postHook:
  - actionType: playbook
    action: cluster-info.yml
```
</details>

Execute the following command to edit the AllInOne.yml configuration template:

```bash
$ vi kubean/examples/install/1.minimal/AllInOne.yml
```

#### 2.Apply the AllInOne.yml configuration

After completing the above steps and saving the AllInOne.yml file, execute the following command:

```bash
$ kubectl apply -f examples/install/1.minimal/AllInOne.yml
```

At this point, you have completed the deployment of a simple single-node cluster.
