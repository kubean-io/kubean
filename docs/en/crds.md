# CRDs

> English | [中文](../zh/crds.md)

CustomResourceDefinition (CRD) is a Kubernetes built-in resource for creating custom resources to further extend the Kubernetes API. Kubean provides four built-in CRDs: Cluster, ClusterOperation, Manifest, and LocalArtifact.

## Cluster

In Kubean, you can declare (uniquely identify) a Kubernetes cluster with a `Cluster` CRD. Clusters will be deployed according to their `Cluster` CRDs.

Here's an example of the `Cluster` CRD:

```yaml
apiVersion: kubean.io/v1alpha1
kind: Cluster
metadata:
  name: cluster1-offline-demo
spec:
  hostsConfRef:
    namespace: kubean-system
    name: cluster1-offline-demo-hosts-conf
  varsConfRef:
    namespace: kubean-system
    name: cluster1-offline-demo-vars-conf
```

Each field in this CRD is explained as follows:

### Metadata Section

- `name`: declares a globally unique cluster.

### Spec Section

- `hostConfRef`: a ConfigMap resource in the format of ansible inventory. It includes information about nodes in a cluster, types, and groups. For further details, refer to [demo](../../artifacts/demo/hosts-conf-cm.yml).

  - `name`: name of the ConfigMap referenced by `hostConfRef`.
  - `namespace`: namespace of the ConfigMap referenced by `hostConfRef`.
  
- `varsConfRef`: a ConfigMap resource to initialize or override variable values declared in Kubespray. This is very useful if you need to execute actions offline. For its specific content, refer to [demo](../../artifacts/demo/vars-conf-cm.yml).

  - `name`: name of the ConfigMap referenced by `varsConfRef`.
  - `namespace`: namespace of the ConfigMap referenced by `varsConfRef`.

- `sshAuthRef`: a Secret resource used only in the SSH private key mode.
  
  - `name`: name of the Secret referenced by `sshAuthRef`.
  - `namespace`: namespace of the Secret referenced by `sshAuthRef`.

## ClusterOperation

In Kubean, you can declare actions (deployment, upgrade, etc.) against a Kubernetes cluster with a `ClusterOperation` CRD. This CRD must be correctly associated with the corresponding `Cluster` CRD, which provides necessary information for executing these actions.

Here's an example of the `ClusterOperation` CRD:

```yaml
apiVersion: kubean.io/v1alpha1
kind: ClusterOperation
metadata:
  name: cluster1-demo-ops-1
spec:
  cluster: cluster1-demo
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

Each field in this CRD is explained as follows:

### Metadata Section

- `name`: uniquely identifies an action against the associated cluster.
  
### Spec

- `cluster`: name of the cluster against which this action will be executed. It should be the same as the value declared in the Cluster CRD.
- `image`: address of the Kubespray image. You can use the image in the Kubean repo or your own image.
- `actionType`: type of the action. It currently can be set as either [`playbook`](https://docs.ansible.com/ansible/latest/user_guide/playbooks_intro.html) or `shell`.
- `action`: the action to be executed. It currently can be set as either a playbook file path or a shell command.
- `preHook`: what should be done before executing the `action`. Allow multiple values, such as test connectivity.
  - `actionType`: refer to the above `actionType`.
  - `action`: refer to the above `action`.
- `postHook`: what to do after executing the `action`. Allow multiple values, such as get the cluster status.
  - `actionType`: refer to the above `actionType`.
  - `action`: refer to the above `action`.
- `backoffLimit`: times of retry if the `action` fails.

## Manifest

<!--Kubean 允许通过 custom resource definitions (CRDs) 来记录和维护当前版本的 Kubean 使用和兼容的组件、包及版本；使用者不用手动编写此资源，由 Kubean 自行维护。-->
In Kubean, you can use a `Manifest` CRD to create and maintain a record of components, packages, and versions used by or compatible with the current version of Kubean. You don't need to do this job manually. Kubean will take care of it for you.

Here's an example of the `Manifest` CRD:

```yaml
apiVersion: kubean.io/v1alpha1
kind: Manifest
metadata:
  name: kubeaninfomanifest-v0-4-0-rc2
spec:
  components:
  - defaultVersion: v1.1.1
    name: cni
    versionRange:
    - v1.0.1
    - v1.1.1
  - defaultVersion: 1.6.9
    name: containerd
    versionRange:
    .......
    - 1.6.7
    - 1.6.8
    - 1.6.9
  - defaultVersion: ""
    name: kube
    versionRange:
    - v1.25.3
    - v1.25.2
    - v1.25.1
    ........
  - defaultVersion: v3.23.3
    name: calico
    versionRange:
    - v3.23.3
    - v3.22.4
    - v3.21.6
  - defaultVersion: v1.12.1
    name: cilium
    versionRange: []
  - defaultVersion: "null"
    name: etcd
    versionRange:
    - v3.5.3
    - v3.5.4
    - v3.5.5
  docker:
  - defaultVersion: "20.10"
    os: redhat-7
    versionRange:
    - latest
    - "18.09"
    - "19.03"
    - "20.10"
    - stable
    - edge
  - defaultVersion: "20.10"
    os: debian
    versionRange:
    - latest
    - "18.09"
    - "19.03"
    - "20.10"
    - stable
    - edge
  - defaultVersion: "20.10"
    os: ubuntu
    versionRange:
    - latest
    - "18.09"
    - "19.03"
    - "20.10"
    - stable
    - edge
  kubeanVersion: v0.4.0-rc2
  kubesprayVersion: c788620
```

Each field in this CRD is explained as follows:

- `components`: declares versions of images or binary files.
  - `name`: name of a component.
  - `defaultVersion`: default versions of the component.
  - `versionRange`: supported component versions.
- `docker`: manages Docker versions.
  - `os`: supported operating systems.
  - `defaultVersion`: the default version used.
  - `versionRange`: supported versions.
- `kubeanVersion`: version of Kubean.
- `kubesprayVersion`: version of the Kubespray used in Kubean.

## LocalArtifact

<!--Kubean 允许通过 custom resource definitions (CRDs) 来记录离线包支持的组件及版本信息；使用者不用手动编写此资源，由 Kubean 自行维护。-->

In Kubean, you can use a `LocalArtifact` CRD to record components and their versions supported by Kubean's offline package. You don't need to do this job manually. Kubean will take care of it for you.

Here's an example of the `LocalArtifact` CRD:

```yaml
apiVersion: kubean.io/v1alpha1
kind: LocalArtifactSet
metadata:
  name: offlineversion-20221101
spec:
  arch: ["x86_64"]
  kubespray: "c788620"
  docker:
    - os: "redhat-7"
      versionRange:
        - "18.09"
        - "19.03"
        - "20.10"
    - os: "debian"
      versionRange: []
    - os: "ubuntu"
      versionRange: []
  items:
    - name: "cni"
      versionRange:
        - v1.1.1
    - name: "containerd"
      versionRange:
        - 1.6.9
    - name: "kube"
      versionRange:
        - v1.24.7
    - name: "calico"
      versionRange:
        - v3.23.3
    - name: "cilium"
      versionRange:
        - v1.12.1
    - name: "etcd"
      versionRange:
        - v3.5.4
```

Each field in this CRD is explained as follows:

- `arch`: a list of supported CPU architectures.
- `kubespray`: Kubespray version used.
- `docker`: manages Docker versions.
  - `os`: operating systems supported by Docker
  - `versionRange`: a list of supported Docker versions.
- `items`: manages versions of other components.
  - `name`: name of a component.
  - `versionRange`: a list of supported versions of the component.
