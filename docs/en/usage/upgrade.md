# Cluster Version Upgrade

This section will introduce how to upgrade the Kubernetes version of a cluster using kubean. The `kubean/example/upgrade` directory that you cloned locally provides a sample template for cluster version upgrades:

<details open>
<summary> The main configuration files and their purposes in the upgrade directory are as follows:</summary>

```yaml
    upgrade
    ├── ClusterOperation.yml                  # Upgrade cluster tasks
    └── VarsConfCM.yml                        # Configuration parameters for cluster version upgrades
```
</details>

To demonstrate the process of upgrading a cluster version, we will use the example of [a single node cluster deployed using the all-in-one mode](./all-in-one-install.md).
> Note: that before upgrading the cluster version, you must have completed the deployment of a cluster using kubean.

#### 1. Add an upgrade task

Go to the `kubean/examples/upgrade/` directory and edit the `ClusterOperation.yml` template. Replace the following parameters with your actual parameters:

  - `<TAG>`：The version of the kubean image. We recommend using the latest version.[Refer to the kubean version list](https://github.com/kubean-io/kubean/tags).

The template for **`ClusterOperation.yml`** in the `kubean/examples/upgrade/` directory is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-upgrade-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: upgrade-cluster.yml
```
**Important Parameters:**
>* `spec.cluster`: Specifies the name of the cluster to be upgraded. In the above example, the cluster named `cluster-mini` is the upgrade target.
>* `spec.action:` Specifies the kubespray playbook related to the upgrade. Here it is set to `upgrade-cluster.yml`.

#### 2. Specify the upgraded version of the cluster

Go to the `kubean/examples/upgrade/` directory and edit the `VarsConfCM.yml` template. Specify the version of the cluster upgrade by configuring the `kube_version` parameter.

The template for `VarsConfCM.yml` in the `kubean/examples/upgrade/` directory is as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mini-vars-conf
  namespace: kubean-system
data:
  group_vars.yml: |
    kube_version: v1.25.8
    # upgrade_cluster_setup: true
    # upgrade_node_confirm: true
    # upgrade_node_pause_seconds: 60

    container_manager: containerd
    kube_network_plugin: calico
    etcd_deployment_type: kubeadm
```
**Important Parameters:**
>* `kube_version`: Specifies the version of the cluster to be upgraded. In the above example, it is set to upgrade to k8s v1.25.8.

!!! Example of upgrading the cluster version parameter
    === "Before upgrading the version"

        ```yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: mini-vars-conf
          namespace: kubean-system
        data:
          group_vars.yml: |
            kube_version: v1.25.0
            # upgrade_cluster_setup: true
            # upgrade_node_confirm: true
            # upgrade_node_pause_seconds: 60

            container_manager: containerd
            kube_network_plugin: calico
            etcd_deployment_type: kubeadm
        ```

    === "After upgrading the version"

        ```yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: mini-vars-conf
          namespace: kubean-system
        data:
          group_vars.yml: |
            kube_version: v1.25.8
            # upgrade_cluster_setup: true
            # upgrade_node_confirm: true
            # upgrade_node_pause_seconds: 60

            container_manager: containerd
            kube_network_plugin: calico
            etcd_deployment_type: kubeadm
        ```


Bonus: kubean cluster version support mechanism:

| kubean Version | Recommended Kubernetes Version | Supported Kubernetes Version Range                                   |
| ----------- | ---------------------- | ------------------------------------------------------------ |
| v0.5.2      | v1.25.4                | - "v1.27.2"<br/>        - "v1.26.5"<br/>        - "v1.26.4"<br/>        - "v1.26.3"<br/>        - "v1.26.2"<br/>        - "v1.26.1"<br/>        - "v1.26.0"<br/>        - "v1.25.10"<br/>        - "v1.25.9"<br/>        - "v1.25.8"<br/>        - "v1.25.7"<br/>        - "v1.25.6"<br/>        - "v1.25.5"<br/>        - "v1.25.4"<br/>        - "v1.25.3"<br/>        - "v1.25.2"<br/>        - "v1.25.1"<br/>        - "v1.25.0" |

For more detailed information about upgrading parameters, please refer to the kubespray documentation:[Updating Kubernetes with Kubespray](https://github.com/kubernetes-sigs/kubespray/blob/master/docs/upgrades.md).

#### 3.Apply all configurations under the `upgrade` directory

After completing the above steps andsaving the ClusterOperation.yml and VarsConfCM.yml files, run the following command:

```bash
$ kubectl apply -f examples/upgrade/
```

With this, you have completed the upgrade of the Kubernetes version for a cluster.