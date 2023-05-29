# Cluster Uninstallation

This section will show you how to use kubean to uninstall a cluster. In the `kubean/example/uninstall` directory that you cloned to your local machine, there is a sample template for uninstalling a cluster:

<details open>
<summary> The main configuration files and their purposes in the uninstall directory are as follows:</summary>

```yaml
    uninstall
    ├── ClusterOperation.yml                # Uninstall cluster task
```
</details>

In the following example, we will [use a single-node cluster deployed in all-in-one mode](./all-in-one-install.md) to demonstrate the cluster upgrade operation.
> Note: Before performing a cluster uninstallation, you must have completed the deployment of a cluster using kubean.

#### 1. Add an uninstallation task

Go to the `kubean/examples/uninstall/` directory and edit the template `ClusterOperation.yml`, replacing the following parameters with your actual parameters:

  - `<TAG>`：The kubean image version. It is recommended to use the latest version.[Refer to the kubean version list](https://github.com/kubean-io/kubean/tags).

The template content of `kubean/examples/uninstall/`  **`ClusterOperation.yml`** path is as follows:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster-mini-uninstall-ops
spec:
  cluster: cluster-mini
  image: ghcr.m.daocloud.io/kubean-io/spray-job:<TAG> # Please replace <TAG> with the specified version, such as v0.4.9
  backoffLimit: 0
  actionType: playbook
  action: reset.yml
```
**Important Parameters:**
>* `spec.cluster`: Specifies the name of the cluster to be uninstalled. In the example above, the cluster named `cluster-mini` is the target for uninstallation.
>* `spec.action:`：: Specifies the Kubespray playbook for uninstallation. Here it is set to `reset.yml`.

#### 2.Apply the Configuration in the uninstall Directory

After completing the above steps and saving the ClusterOperation.yml file, execute the following command:

```bash
$ kubectl apply -f examples/uninstall/
```

At this point, you have successfully uninstalled a cluster.