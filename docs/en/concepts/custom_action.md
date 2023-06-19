# Custom Actions

## Motivation

For users, the products of Kubean and Kubespray are OCI images, Helm charts, and K8s manifests.
If you want to customize some operations after obtaining these products, it is possible but can
be complicated and requires a lot of manual configuration modifications. We hope to simplify this process.

## Goal

Provide a convenient way for users to use customized actions to view, modify, and control the status of cluster nodes.

## CRD Design

1. Add the ActionSource field to declare the source of the action, whose value currently supports:

   - builtin (default)

        Indicates the use of Kubean's built-in Ansible playbook or shell script in the manifest.

   - configmap

        Indicates that the required Ansible playbook or shell script is obtained by referencing a K8s ConfigMap.

2. Add the ActionSourceRef field to declare the resource object referenced
   when ActionSource is configmap. This field only takes effect when ActionSource is configmap,
   and its format is:

    ```yaml
    actionSourceRef:
      name: <configmap name>
      namespace: <namespace of configmap>
    ```

Configuration example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster1-demo-myaction
  namespace: kubean-system
data:
  myplaybook.yml: |
    - hosts: k8s_cluster
      gather_facts: false
      become: yes
      any_errors_fatal: "{{ any_errors_fatal | default(true) }}"
      tasks:
        - name: Print inventory hostname
          debug:
            msg: "inventory_hostname is {{ inventory_hostname }}"
  hello.sh: |
    echo "hello world!"
---
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster1-demo-ops-1
spec:
  cluster: cluster1-demo
  image: ghcr.io/kubean-io/spray-job:latest
  backoffLimit: 0
  actionType: playbook
  action: myplaybook.yml
  actionSource: configmap
  actionSourceRef:
    name: cluster1-demo-myaction
    namespace: kubean-system
  preHook:
    - actionType: shell
      action: hello.sh
      actionSource: configmap
      actionSourceRef:
        name: cluster1-demo-myaction
        namespace: kubean-system
```
