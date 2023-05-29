# Scaling Cluster Worker Nodes

In the process of software development and operation, business growth often requires adding worker nodes to a cluster to meet the demand. For clusters deployed using kubean, we can use a declarative approach to quickly scale the cluster's worker nodes.

In the `kubean/example/scale` directory cloned to your local machine, there is a sample template for scaling worker nodes:

<details open>
<summary> The main configuration files and purposes in the scale file are as follows: </summary>

```yaml
    scale
    ├── 1.addWorkNode                             # Template for adding worker nodes
    │   ├── ClusterOperation.yml                       # kubean version and task configuration
    │   └── HostsConfCM.yml                            # configuration of current cluster node information
    └── 2.delWorkNode                             # Template for deleting worker nodes
    │   ├── ClusterOperation.yml                       # kubean version and task configuration
    │   └── HostsConfCM.yml                             # configuration of current cluster node information
```
</details>

By observing the scaling configuration template in the `scale` file, it can be seen that scaling the cluster's worker nodes only requires executing two configuration files: `HostsConfCM.yml` and `ClusterOperation.yml`. You will need to replace the parameters such as the information of the new node with your actual parameters.

[Using the example of a single-node cluster deployed in all-in-one mode](./all-in-one-install.md) let's demonstrate how to scale the cluster's worker nodes.
> Note: Before scaling the cluster, you must have completed the deployment of a set of cluster using kubean.

## Scaling Worker Nodes

#### 1. Add New Node Host Parameters to HostsConfCM.yml

To add a new node configuration to the ConfigMap named `mini-hosts-conf` in the existing all-in-one mode, we will add a new worker node `node2` based on the original main node `node1`.

Specifically, we can go to the path `kubean/examples/scale/1.addWorkNode/`, edit the prepared node configuration ConfigMap template `HostsConfCM.yml`, and replace the following parameters with your actual parameters:

  - `<IP2>`: the IP address of the node.
  - `<USERNAME>`: the username to log in to the node. We recommend using either "root" or a user with root privileges.
  - `<PASSWORD>`: the password to log in to the node.

The template content of **`HostsConfCM.yml`**  in the `kubean/examples/scale/1.addWorkNode/` path is as follows:

<details>
<summary> HostsConfCM.yml 模板</summary>
```yaml
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
          ip: <IP1>
          access_ip: <IP1>
          ansible_host: <IP1>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
        node2:
          ip: <IP2>
          access_ip: <IP2>
          ansible_host: <IP2>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
      children:
        kube_control_plane:
          hosts:
            node1:
        kube_node:
          hosts:
            node1:
            node2:
        etcd:
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

**Important Parameters:**
>* `all.hosts.node1`: The original main node that already exists in the cluster.
>* `all.hosts.node2`: The worker node to be added to the cluster.
>* `all.children.kube_node.hosts`:  The group of worker nodes in the cluster.


!!! Example of Adding New Node Host Parameters

    === "Before Adding New Node"

        ``` yaml
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
                  ip: 10.6.175.10 # Your node's IP
                  access_ip: 10.6.175.10 # Your node's IP
                  ansible_host: 10.6.175.10 # Your node's IP
                  ansible_connection: ssh
                  ansible_user: root # The username to log in to the node
                  ansible_password: password01 # The password to log in to the node
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
        ```

    === "After Adding New Node"

        ``` yaml
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
                  ip: 10.6.175.10 # Your node's IP
                  access_ip: 10.6.175.10 # Your node's IP
                  ansible_host: 10.6.175.10 # Your node's IP
                  ansible_connection: ssh
                  ansible_user: root # The username to log in to the node
                  ansible_password: password01 # The password to log in to the node
                node2:
                  ip: 10.6.175.20 # Your node's IP
                  access_ip: 10.6.175.20 # Your node's IP
                  ansible_host: 10.6.175.20 # Your node's IP
                  ansible_connection: ssh
                  ansible_user: root # The username to log in to the node
                  ansible_password: password01 # The password to log in to the node
              children:
                kube_control_plane:
                  hosts:
                    node1:
                kube_node:
                  hosts:
                    node1:
                    node2:
                etcd:
                  hosts:
                    node1:
                k8s_cluster:
                  children:
                    kube_control_plane:
                    kube_node:
                calico_rr:
                  hosts: {}
        ```


#### 2. Add Scaling Task through ClusterOperation.yml

Go to the path `kubean/examples/scale/1.addWorkNode/` and edit the template `ClusterOperation.yml`,replacing the following parameter with your actual parameter:

  - `<TAG>`: the kubean image version. We recommend using the latest version. [Refer to the kubean version list](https://github.com/kubean-io/kubean/tags).

The template content of **`ClusterOperation.yml`** in the `kubean/examples/scale/1.addWorkNode/` path is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-awn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: scale.yml
  extraArgs: --limit=node2
```
**Important Parameters:**
>* `spec.cluster`: specifies the name of the cluster to be scaled. The above example specifies the cluster named `cluster-mini` as the scaling target.
>* `spec.action:`: specifies the kubespray script for scaling the node, which is set to `scale.yml` here.
>* `spec.extraArgs`: specifies the limit of the nodes to be scaled. Here, the `--limit=` parameter is used to limit the scaling to the node2.


For example, the following is an example of ClusterOperation.yml:
<details>
<summary> ClusterOperation.yml Example</summary>
```yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-awn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2
  backoffLimit: 0
  actionType: playbook
  action: scale.yml
  extraArgs: --limit=node2
```
</details>

#### 3.Apply all configurations under `scale/1.addWorkNode` folder

After completing the above steps and saving the HostsConfCM.yml and ClusterOperation.yml files, run the following command:

```bash
$ kubectl apply -f examples/install/scale/1.addWorkNode/
```

At this point, you have completed the scaling of a working node in a cluster.

## Shrink Working Nodes

#### 1. Add Scaling Task through ClusterOperation.yml

Go to the path `kubean/examples/scale/2.delWorkNode/` and edit the template `ClusterOperation.yml`, replacing the following parameter with your actual parameter:

  - `<TAG>`：the kubean image version. We recommend using the latest version. [Refer to the kubean version list](https://github.com/kubean-io/kubean/tags).

The template content of **`ClusterOperation.yml`** in the `kubean/examples/scale/2.delWoorkNode/` path is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dwn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2
```
**重要参数：**
>* `spec.cluster`: specifies the name of the cluster to be scaled. The above example specifies the cluster named cluster-mini as the scaling target.
>* `spec.action`: specifies the kubespray script for scaling the node, which is set to remove-node.yml here.
>* `spec.extraArgs`: specifies the nodes to be scaled down. Here, the -e parameter is used to specify the node2 to be scaled down.

For example, the following is an example of ClusterOperation.yml:
<details>
<summary> ClusterOperation.yml Example</summary>
```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dwn-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.5.2
  backoffLimit: 0
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2
```
</details>


#### 2.Apply the ClusterOperation scaling task list under the `scale/2.delWorkNode` directory

After completing the above steps and saving the ClusterOperation.yml file, run the following command:

```bash
$ kubectl apply -f examples/install/scale/2.delWorkNode/ClusterOperation.yml
```

By default, enter the kubean-system namespace and check the execution status of the scaling task:
``` bash
$ kubectl -n kubean-system get pod | grep cluster-mini-dwn-ops
```
To learn about the progress of the scaling task, you can view the logs of the pod.

#### 3. Delete the working node host parameters through HostsConfCM.yml

We have executed the scaling task through the above two steps. After the scaling task is completed, node2 will be permanently removed from the existing cluster. Then, we need to complete the final step, which is to remove the `node2` information from the node configuration related Configmap.

Go to the path `kubean/examples/scale/2.delWorkNode/` and edit the prepared node configuration template `HostsConfCM.yml` to remove the configuration of the working node that needs to be removed.

**The deleted parameters are as follows:**

* `all.hosts`: The node2 node access parameters.
* `all.children.kube_node.hosts`: The node name node2.

!!! Example of removing working node host parameters

    === "Before removing the node"

        ``` yaml
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
                  ip: 10.6.175.10 # Your node's IP
                  access_ip: 10.6.175.10 # Your node's IP
                  ansible_host: 10.6.175.10 # Your node's IP
                  ansible_connection: ssh
                  ansible_user: root # The username to log in to the node
                  ansible_password: password01 # The password to log in to the node
                node2:
                  ip: 10.6.175.20 # The IP address of node 2 is added
                  access_ip: 10.6.175.20 # The IP address of node 2 is added
                  ansible_host: 10.6.175.20 # The IP address of node 2 is added
                  ansible_connection: ssh
                  ansible_user: root # The username to log in to the node2
                  ansible_password: password01 # password to log in to the node2
              children:
                kube_control_plane:
                  hosts:
                    node1:
                kube_node:
                  hosts:
                    node1:
                    node2:
                etcd:
                  hosts:
                    node1:
                k8s_cluster:
                  children:
                    kube_control_plane:
                    kube_node:
                calico_rr:
                  hosts: {}
        ```

    === "After removing a node"

        ``` yaml
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
                  ip: 10.6.175.10 # Your node's IP
                  access_ip: 10.6.175.10 # Your node's IP
                  ansible_host: 10.6.175.10 # Your node's IP
                  ansible_connection: ssh
                  ansible_user: root # The username to log in to the node
                  ansible_password: password01 # The password to log in to the node
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
        ```

After completing the above steps and saving the HostsConfCM.yml file, execute the following command:

```bash
$ kubectl apply -f examples/install/scale/2.delWorkNode/HostsConfCM.yml
```

At this point, we have removed the node2 worker node from the cluster and cleaned up all the host information related to node2. The entire scaling down operation is now complete.

