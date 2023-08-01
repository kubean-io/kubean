# Deploy the cluster as a non-root user

## Contents

- ✓ [1. sudo permission check](#sudo-permission-check)
- ✓ [2. Create host inventory configuration](#create-host-inventory-configuration)
- ✓ [3. Prepare the configuration parameters of the deployment cluster](#prepare-the-configuration-parameters-of-the-deployment-cluster)
- ✓ [4. Prepare Kubean CRs](#prepare-kubean-crs)
- ✓ [5. Start deploying the cluster](#start-deploying-the-cluster)

## sudo permission check

  The installation process involves system privileged operations,
  so users need to have sudo privileges, and the following checks can be performed:

  1. Log in to target node as a non-root user

  2. Check for the existence of the sudo command, and install it through
     the system package manager if it does not exist:

     `which sudo`

  3. Execute `echo | sudo -S -v` in the terminal
  
      If the result outputs `xxx is not in the sudoers file. This incident will be reported`
      or `User xxx do not have sudo privilege` and other similar information, it means that the
      current user does not have sudo privileges, otherwise it means that the current user has sudo privileges.

## Configure host list
   

  Example: The content of the host list `HostsConfCM.yml` is roughly as follows, replace
  <USERNAME> and <PASSWORD> below with the actual username and password:

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
  > Note: If the user is configured as NOPASSWD (no password escalation) in the /etc/sudoers file, you can comment the line where `ansible_become_password` is located

## Prepare parameters of the deployment cluster

For the content of the cluster configuration parameter `VarsConfCM.yml`, please refer to
[demo vars conf](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/VarsConfCM.yml).

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

## Prepare Kubean CRs

- Example of Cluster CR

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
      sshAuthRef: # Key attribute, specifying the ssh private key secret during cluster deployment
        namespace: kubean-system
        name: sample-ssh-auth
    ```

- Example of ClusterOperation CR

    ```yaml
    # ClusterOperation.yml
    apiVersion: kubean.io/v1alpha1
    kind: ClusterOperation
    metadata:
      name: sample-create-cluster
    spec:
      cluster: sample
      image: ghcr.m.daocloud.io/kubean-io/spray-job:latest
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

## Start deploying the cluster

Assuming all YAML manifests are stored in the `create_cluster` directory:

```bash
$ tree create_cluster/
create_cluster
├── HostsConfCM.yml       # host list
├── SSHAuthSec.yml        # SSH private key
├── VarsConfCM.yml        # cluster parameters
├── Cluster.yml           # Cluster CR
└── ClusterOperation.yml  # ClusterOperation CR
```

Start deploying the cluster with `kubectl apply`:

```bash
kubectl apply -f create_cluster/
```
