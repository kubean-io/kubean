# Kubean Infrastructure

The overall architecture of Kubean is shown below:

[![kubean-architecture]][kubean-architecture]

Kubean needs to run on an existing Kubernetes cluster. It controls and manages cluster lifecycle (install, uninstall, upgrade, scale up & down, etc.) by applying the standard CRDs provided by Kubean and Kubernetes built-in resources. Kubean uses Kubespray as the underlying technology. On the one hand, it simplifies the operation process of cluster deployment and lowers the threshold of use. On the other hand, many new features such as cluster operation records and offline version records have been added on the basis of Kubespray's capabilities.

<br/>

[![kubean-components]][kubean-components]

Kubean runs several controllers to track changes of Kubean CRDs and communicates with the underlying cluster's API server to create Kubernetes native resources. It consists of four components:

  1. Cluster Controller: monitors 'Cluster Objects'. It uniquely identifies a cluster, has the access information, type information, and deployment parameter information of the cluster node, and is associated with all operations on the cluster ('ClusterOperation Objects');
  2. ClusterOperation Controller: monitors `ClusterOperation Objects`. When a `ClusterOperation Object` is created, the controller assembles a [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) to perform the operations defined in the CRD object;
  3. Manifest Controller: monitors `Manifest Objects`. It records and maintains components, packages and versions that are used by or compatible with the current version of Kubean;
  4. LocalArtifactSet Controller: monitors `LocalArtifactSet Objects`. It records information about the components and versions supported by the offline package.

  [kubean-architecture]: /kubean/en/assets/images/kubean-architecture.png
  [kubean-components]: /kubean/en/assets/images/kubean-components.png