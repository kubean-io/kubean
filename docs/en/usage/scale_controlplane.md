# Cluster Control Plane Scaling

During Kubernetes cluster operations, we often need to scale control plane nodes to improve cluster high availability and performance. For clusters deployed with kubean, we can use a declarative approach to quickly scale control plane nodes.

In the `kubean/example/scale` directory that you cloned locally, templates for control plane node scaling are provided:

<details open>
<summary>The main configuration files in the scale directory and their purposes are as follows:</summary>

```yaml
    scale
    ├── 3.addControlPlane                         # Add control plane node template
    │   ├── ClusterOperation.yml                       # kubean version and task configuration
    │   └── HostsConfCM.yml                            # Current cluster node information configuration
    └── 4.delControlPlane                         # Delete control plane node template
        ├── ClusterOperation.yml                       # kubean version and task configuration
        ├── ClusterOperation2.yml                      # kubean version and task configuration
        └── HostsConfCM.yml                            # Current cluster node information configuration
```
</details>

The following demonstrates control plane node scaling operations using an existing single control plane node cluster as an example.
> Note: Before executing cluster control plane scaling, you must have already completed a cluster deployment using kubean.

## Scaling Up Control Plane Nodes

#### 1. Add new control plane node host parameters to HostsConfCM.yml

We want to add node configuration to the ConfigMap named `mini-hosts-conf` in the original single control plane node cluster, adding `node2` and `node3` control plane nodes based on the original `node1` control plane node.

Specifically, we can enter the `kubean/examples/scale/3.addControlPlane/` path, edit the prepared node configuration ConfigMap template `HostsConfCM.yml`, and replace the following parameters with your actual parameters:

  - `<IP2>`, `<IP3>`: Node IP addresses.
  - `<USERNAME>`: Username for logging into the node, recommended to use root or a user with root privileges.
  - `<PASSWORD>`: Password for logging into the node.

The template content of **`HostsConfCM.yml`** in the `kubean/examples/scale/3.addControlPlane/` path is as follows:

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
        node3:
          ip: <IP3>
          access_ip: <IP3>
          ansible_host: <IP3>
          ansible_connection: ssh
          ansible_user: <USERNAME>
          ansible_password: <PASSWORD>
      children:
        kube_control_plane:
          hosts:
            node1:
            node2:
            node3:
        kube_node:
          hosts:
            node1:
            node2:
            node3:
        etcd:
          hosts:
            node1:
            node2:
            node3:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}
```

**Important parameters:**
>* `all.hosts.node1`: Existing control plane node in the original cluster
>* `all.hosts.node2`, `all.hosts.node3`: New control plane nodes to be added during cluster scaling
>* `all.children.kube_control_plane.hosts`: Control plane node group in the cluster
>* `all.children.etcd.hosts`: etcd node group in the cluster, usually consistent with control plane nodes

!!! Example of adding control plane node host parameters

    === "Before adding nodes"

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
                  ip: 10.6.175.10 # Control plane node IP
                  access_ip: 10.6.175.10 # Control plane node IP
                  ansible_host: 10.6.175.10 # Control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
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

    === "After adding nodes"

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
                  ip: 10.6.175.10 # Control plane node IP
                  access_ip: 10.6.175.10 # Control plane node IP
                  ansible_host: 10.6.175.10 # Control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
                node2:
                  ip: 10.6.175.20 # New control plane node IP
                  access_ip: 10.6.175.20 # Worker node IP
                  ansible_host: 10.6.175.20 # Worker node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
                node3:
                  ip: 10.6.175.30 # New control plane node IP
                  access_ip: 10.6.175.30 # New control plane node IP
                  ansible_host: 10.6.175.30 # New control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
              children:
                kube_control_plane:
                  hosts:
                    node1:
                    node2:
                    node3:
                kube_node:
                  hosts:
                    node1:
                    node2:
                    node3:
                etcd:
                  hosts:
                    node1:
                    node2:
                    node3:
                k8s_cluster:
                  children:
                    kube_control_plane:
                    kube_node:
                calico_rr:
                  hosts: {}
        ```

#### 2. Add control plane scaling task through ClusterOperation.yml

Enter the `kubean/examples/scale/3.addControlPlane/` path, edit the template `ClusterOperation.yml`, and replace the following parameters with your actual parameters:

  - `<TAG>`: kubean image version, recommended to use the latest version, [refer to kubean version list](https://github.com/kubean-io/kubean/tags).

The template content of **`ClusterOperation.yml`** in the `kubean/examples/scale/3.addControlPlane/` path is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-acp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.26.4
  actionType: playbook
  action: cluster.yml
  extraArgs: >-
    --limit=etcd,kube_control_plane
    -e ignore_assert_errors=true
  postHook:
    - actionType: playbook
      action: upgrade-cluster.yml
      extraArgs: >-
        --limit=etcd,kube_control_plane
        -e ignore_assert_errors=true
```

**Important parameters:**
>* `spec.cluster`: Specifies the cluster name that needs control plane node scaling, the above specifies the cluster named `cluster-mini` as the scaling target.
>* `spec.action`: Specifies the Kubespray playbook for control plane node scaling, set to `cluster.yml` here.
>* `spec.extraArgs`: Specifies the node limitation for scaling, here the `--limit=` parameter limits scaling to `etcd`, `control-plane` node groups.
>* `spec.postHook.action`: Specifies the Kubespray playbook for control plane node scaling, set to `upgrade-cluster.yml` here, updating all Etcd configurations in the cluster.
>* `spec.postHook.extraArgs`: Specifies the node limitation for scaling, here the `--limit=` parameter limits scaling to `etcd`, `control-plane` node groups.

For example, the following shows a ClusterOperation.yml example:
<details>
<summary>ClusterOperation.yml example</summary>

```yaml
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-acp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.26.4
  actionType: playbook
  action: cluster.yml
  extraArgs: >-
    --limit=etcd,kube_control_plane
    -e ignore_assert_errors=true
  postHook:
    - actionType: playbook
      action: upgrade-cluster.yml
      extraArgs: >-
        --limit=etcd,kube_control_plane
        -e ignore_assert_errors=true
```
</details>

#### 3. Apply all configurations in the `scale/3.addControlPlane` directory

After completing the above steps and saving the HostsConfCM.yml and ClusterOperation.yml files, execute the following command:

```bash
$ kubectl apply -f examples/scale/3.addControlPlane/
```

#### 4. Restart kube-system/nginx-proxy

If the control plane and worker nodes are separated, you need to restart the nginx-proxy pod on all worker nodes. This pod is a local proxy for the API server. Kubean will update its static configuration, but it needs to be restarted to reload.

```bash
crictl ps | grep nginx-proxy | awk '{print $1}' | xargs crictl stop
```

At this point, you have completed the control plane node scaling for a cluster.
## Sc
aling Down Control Plane Nodes

> Note: Before scaling down control plane nodes, ensure that at least one control plane node remains in the cluster to ensure normal cluster operation.

#### 1. Add control plane scaling down task through ClusterOperation.yml

Enter the `kubean/examples/scale/4.delControlPlane/` path, edit the template `ClusterOperation.yml`, and replace the following parameters with your actual parameters:

  - `<TAG>`: kubean image version, recommended to use the latest version, [refer to kubean version list](https://github.com/kubean-io/kubean/tags).

The template content of **`ClusterOperation.yml`** in the `kubean/examples/scale/4.delControlPlane/` path is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.26.4
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2,node3 -e reset_nodes=false -e allow_ungraceful_removal=true
```

**Important parameters:**
>* `spec.cluster`: Specifies the cluster name that needs control plane node scaling down, the above specifies the cluster named cluster-mini as the scaling target.
>* `spec.action`: Specifies the kubespray playbook for node scaling down, set to remove-node.yml here.
>* `spec.extraArgs`: Specifies the nodes to be scaled down and related parameters, here specified through -e parameters:
>   * `node=node2,node3`: Specifies the control plane node names to be removed
>   * `reset_nodes=false`: Do not reset nodes (preserve data on nodes, can also be used when nodes are not accessible)
>   * `allow_ungraceful_removal=true`: Allow ungraceful removal (used when nodes are no longer accessible)

For example, the following shows a ClusterOperation.yml example:
<details>
<summary>ClusterOperation.yml example</summary>

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.26.4
  actionType: playbook
  action: remove-node.yml
  extraArgs: -e node=node2,node3
```
</details>

#### 2. Apply the ClusterOperation scaling down task manifest in the `scale/4.delControlPlane` directory

After completing the above steps and saving the ClusterOperation.yml file, execute the following command:

```bash
$ kubectl apply -f examples/scale/4.delControlPlane/ClusterOperation.yml
```

Enter the kubean-system namespace by default and check the scaling down task execution status:
```bash
$ kubectl -n kubean-system get pod | grep cluster-mini-dcp-ops
```
To understand the scaling down task execution progress, you can check the pod logs.

#### 3. Delete control plane node host parameters through HostsConfCM.yml

We have executed the scaling down task through the above two steps. After the scaling down task is completed, `node2` and `node3` will be permanently removed from the existing cluster. At this time, we also need to remove the node2 and node3 information from the node configuration related Configmap.

Enter the `kubean/examples/scale/4.delControlPlane/` path, edit the prepared node configuration template `HostsConfCM.yml`, and delete the control plane node configuration that needs to be removed.

**Parameters to delete:**

* Node3 access parameters under `all.hosts`.
* Host name node3 in `all.children.kube_control_plane.hosts`.
* Host name node3 in `all.children.kube_node.hosts` (if the node is also a worker node).
* Host name node3 in `all.children.etcd.hosts` (if the node is also an etcd node).

!!! Example of removing control plane node host parameters

    === "Before removing nodes"

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
                  ip: 10.6.175.10 # Control plane node IP
                  access_ip: 10.6.175.10 # Control plane node IP
                  ansible_host: 10.6.175.10 # Control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
                node2:
                  ip: 10.6.175.20 # Control plane node IP
                  access_ip: 10.6.175.20 # Control plane node IP
                  ansible_host: 10.6.175.20 # Control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
                node3:
                  ip: 10.6.175.30 # Control plane node IP
                  access_ip: 10.6.175.30 # Control plane node IP
                  ansible_host: 10.6.175.30 # Control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
              children:
                kube_control_plane:
                  hosts:
                    node1:
                    node2:
                    node3:
                kube_node:
                  hosts:
                    node1:
                    node2:
                    node3:
                etcd:
                  hosts:
                    node1:
                    node2:
                    node3:
                k8s_cluster:
                  children:
                    kube_control_plane:
                    kube_node:
                calico_rr:
                  hosts: {}
        ```

    === "After removing nodes"

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
                  ip: 10.6.175.10 # Control plane node IP
                  access_ip: 10.6.175.10 # Control plane node IP
                  ansible_host: 10.6.175.10 # Control plane node IP
                  ansible_connection: ssh
                  ansible_user: root # Username for logging into the node
                  ansible_password: password01 # Password for logging into the node
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
$ kubectl apply -f examples/scale/4.delControlPlane/HostsConfCM.yml
```

#### 4. Update Kubernetes and network configuration files

Run `cluster.yml` to regenerate configuration files on all remaining nodes.

Enter the kubean/examples/scale/4.delControlPlane/ path, edit the template ClusterOperation2.yml, and replace the following parameters with your actual parameters:

  - `<TAG>`: kubean image version, recommended to use the latest version, [refer to kubean version list](https://github.com/kubean-io/kubean/tags).

The template content of **`ClusterOperation2.yml`** in the `kubean/examples/scale/4.delControlPlane/` path is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops-2
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.26.4
  actionType: playbook
  action: cluster.yml
```

**Important parameters:**
>* `spec.cluster`: Specifies the cluster name that needs control plane node scaling down, the above specifies the cluster named cluster-mini as the scaling target.
>* `spec.action`: Specifies the kubespray playbook for node scaling down, set to cluster.yml here.

For example, the following shows a ClusterOperation.yml example:
<details>
<summary>ClusterOperation.yml example</summary>

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-dcp-ops-2
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:v0.26.4
  actionType: playbook
  action: cluster.yml
```
</details>

#### 5. Restart kube-system/nginx-proxy

If the control plane and worker nodes are separated, you need to restart the nginx-proxy pod on all worker nodes. This pod is a local proxy for the API server. Kubean will update its static configuration, but it needs to be restarted to reload.

```bash
crictl ps | grep nginx-proxy | awk '{print $1}' | xargs crictl stop
```

At this point, we have removed the node2 and node3 control plane nodes from the cluster, cleaned up the host information related to node2 and node3, and updated the cluster configuration. The entire control plane scaling down operation is now complete.

## Considerations

1. **High Availability Considerations**: In production environments, it is recommended to maintain at least 3 control plane nodes to ensure cluster high availability.

2. **etcd Cluster Scaling**: Control plane nodes are usually also etcd nodes. When scaling control plane nodes, special attention should be paid to ensuring that the number of etcd cluster nodes should be odd (1, 3, 5, etc.) to ensure etcd cluster high availability and consistency.

3. **Load Balancing**: When there are multiple control plane nodes, it is recommended to configure a load balancer to ensure API server high availability.

4. **Backup**: Before performing control plane node scaling operations, it is recommended to backup etcd data first to prevent data loss due to operation failure.