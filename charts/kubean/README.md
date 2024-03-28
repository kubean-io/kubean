# :seedling: kubean

[![helm workflow](https://github.com/kubean-io/kubean-helm-chart/actions/workflows/helm-release.yaml/badge.svg)](https://github.com/kubean-io/kubean-helm-chart/actions/workflows/helm-release.yaml)

## Introduction

kubean is a cluster lifecycle management tool based on [kubespray](https://github.com/kubernetes-sigs/kubespray).

## Features

The Kubean provides the following features.

* Based on the CRD cluster deployment method, all operations can be completed on the kubernetes API-server.

* Supports concurrent deployment of multiple clusters at the same time.

* Support air gap installation.(experimental)

* Support for both AMD64 and ARM64.

## Parameters

### kubean-operator parameters

| Name                                        | Description                                           | Value                       |
| ------------------------------------------- | ----------------------------------------------------- | --------------------------- |
| `kubeanOperator.replicaCount`               | Number of kubean-operator replicas to deploy          | `1`                         |
| `kubeanOperator.nameOverride`               | String to partially override kubean-operator.fullname | `""`                        |
| `kubeanOperator.fullnameOverride`           | String to fully override kubean-operator.fullname     | `""`                        |
| `kubeanOperator.operationsBackendLimit`     | Limit of operations backend                           | `5`                         |
| `kubeanOperator.podAnnotations`             | Annotations to add to the kubean-operator pods        | `{}`                        |
| `kubeanOperator.podSecurityContext`         | Security context for kubean-operator pods             | `{}`                        |
| `kubeanOperator.securityContext`            | Security context for kubean-operator containers       | `{}`                        |
| `kubeanOperator.serviceAccount.create`      | Specifies whether a service account should be created | `true`                      |
| `kubeanOperator.serviceAccount.annotations` | Annotations to add to the service account             | `{}`                        |
| `kubeanOperator.serviceAccount.name`        | The name of the service account to use.               | `""`                        |
| `kubeanOperator.image.registry`             | kubean-operator image registry                        | `ghcr.io`                   |
| `kubeanOperator.image.repository`           | kubean-operator image repository                      | `kubean-io/kubean-operator` |
| `kubeanOperator.image.tag`                  | kubean-operator image tag                             | `""`                        |
| `kubeanOperator.image.pullPolicy`           | kubean-operator image pull policy                     | `IfNotPresent`              |
| `kubeanOperator.image.pullSecrets`          | kubean-operator image pull secrets                    | `[]`                        |
| `kubeanOperator.service.type`               | kubean-operator service type                          | `ClusterIP`                 |
| `kubeanOperator.service.port`               | kubean-operator service port                          | `80`                        |
| `kubeanOperator.resources`                  | kubean-operator resources                             | `{}`                        |
| `kubeanOperator.nodeSelector`               | kubean-operator node selector                         | `{}`                        |
| `kubeanOperator.tolerations`                | kubean-operator tolerations                           | `[]`                        |

### kubean admission parameters

| Name                               | Description                                   | Value                        |
| ---------------------------------- | --------------------------------------------- | ---------------------------- |
| `kubeanAdmission.replicaCount`     | Number of kubean-admission replicas to deploy | `1`                          |
| `kubeanAdmission.image.registry`   | kubean-admission image registry               | `ghcr.io`                    |
| `kubeanAdmission.image.repository` | kubean-admission image repository             | `kubean-io/kubean-admission` |
| `kubeanAdmission.image.tag`        | kubean-admission image tag                    | `""`                         |

### sprayJob parameters

| Name                        | Description                | Value                 |
| --------------------------- | -------------------------- | --------------------- |
| `sprayJob.image.registry`   | spray-job image registry   | `ghcr.io`             |
| `sprayJob.image.repository` | spray-job image repository | `kubean-io/spray-job` |
| `sprayJob.image.tag`        | spray-job image tag        | `""`                  |


First, add the Kubean chart repo to your local repository.
``` bash 
$ helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
$ helm repo update
$ helm repo list
NAME          	URL
kubean-io     	https://kubean-io.github.io/kubean-helm-chart/
```

With the repo added, available charts and versions can be viewed.
``` bash
$ helm search repo kubean
```

You can run the following command to install kubean.
``` bash
$ helm install kubean kubean-io/kubean --create-namespace -n kubean-system
```

View cluster information.
``` bash
$ kubectl get clusters.kubean.io
```

View cluster operation jobs.
``` bash
$ kubectl get clusteroperations.kubean.io
```

## Uninstall

If kubean's related custom resources already exist, you need to clear.
``` bash
$ kubectl delete clusteroperations.kubean.io --all
$ kubectl delete clusters.kubean.io --all
$ kubectl delete manifests.kubean.io --all
$ kubectl delete localartifactsets.kubean.io --all
```

Uninstall kubean's components via helm.
``` bash
$ helm -n kubean-system uninstall kubean
$ kubectl delete crd clusteroperations.kubean.io
$ kubectl delete crd clusters.kubean.io
$ kubectl delete crd manifests.kubean.io
$ kubectl delete crd localartifactsets.kubean.io
```
