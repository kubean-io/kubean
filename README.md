# :seedling: KuBean

<a href="https://trackgit.com">
<img src="https://us-central1-trackgit-analytics.cloudfunctions.net/token/ping/la6t1t81jgv27ys97ila" alt="trackgit-views" />
</a>

<div align="center">

  <p>

[<img src="docs/images/certified-kubernetes-color.png" height=120>](https://github.com/cncf/k8s-conformance/pull/2240)
[<img src="docs/images/kubean-logo.png" height=120>](https://kubean-io.github.io/website/)
<!--
Source: https://github.com/cncf/artwork/tree/master/projects/kubernetes/certified-kubernetes
-->

  </p>

  <p>

kubean is a cluster lifecycle management tool based on [kubespray](https://github.com/kubernetes-sigs/kubespray).

  </p>

  <p>

[![main workflow](https://github.com/kubean-io/kubean/actions/workflows/auto-main-ci.yaml/badge.svg)](https://github.com/kubean-io/kubean/actions/workflows/auto-main-ci.yaml)
[![codecov](https://codecov.io/gh/kubean-io/kubean/branch/main/graph/badge.svg?token=8FX807D3QQ)](https://codecov.io/gh/kubean-io/kubean)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6263/badge)](https://bestpractices.coreinfrastructure.org/projects/6263)
[![kubean coverage](https://raw.githubusercontent.com/dasu23/e2ecoverage/master/badges/kubean/kubeanCoverage.svg)](https://github.com/kubean-io/kubean/blob/main/docs/test/kubean_testcase.md)
[![kubean coverage](https://raw.githubusercontent.com/dasu23/e2ecoverage/master/badges/kubean/kubeanCoverage2.svg)](https://github.com/kubean-io/kubean/blob/main/docs/test/kubean_testcase.md)
[![license](https://img.shields.io/badge/license-AL%202.0-blue)](https://github.com/kubean-io/kubean/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubean-io/kubean)](https://goreportcard.com/report/github.com/kubean-io/kubean)

  </p>

</div>

---

## :anchor: Awesome features

- **Simplicity:** Deploying of Kubean and powerful lifecycle management of kubernetes cluster implementing by declarative API.
- **Offline Supported**: Offline packages(os-pkgs, images, binarys) are released with the release. You won't have to worry about how to gather all the resources you need.
- **Compatibility**: Multi-arch delivery Supporting. Such as AMD, ARM with common Linux distributions. Also include Kunpeng with Kylin.
- **Expandability**: Allowing custom actions be added to cluster without any changes for Kubespray. 

## :surfing_man: Quick Start

#### 1. Ensure that a Kubernetes Cluster exists and Helm installed

#### 2. Deploy Kubean-Operator

``` shell
$ helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
$ helm install kubean kubean-io/kubean --create-namespace -n kubean-system
```

Then check kubean-operator status by 
```shell 
$ kubectl get pods -n kubean-system | grep 'kubean'
```

#### 3. Start ClusterOperation for cluster.yml playbook

We cloud use the example in folder `artifacts/demo` which uses online resources to install k8s cluster.

  1. cd resources path
     ```shell
     $ cd artifacts/
     ```
  2. modify `demo/hosts-conf-cm.yml` by replacing `IP1`, `IP2`... with the real ip where we want to install k8s cluster
  3. start kubeanClusterOps which will start the kubespray job
     ```shell
     $ kubectl apply -f demo/
     ```
  4. check the kubespray job status
     ```shell
     $ kubectl get job -n kubean-system
     ```

[![quick_start_image](docs/images/quick_start.gif)](https://asciinema.org/a/511386)

## :ocean: Kubernetes compatibility

|               | Kubernetes 1.20 | Kubernetes 1.21 | Kubernetes 1.22 | Kubernetes 1.23 | Kubernetes 1.24 | Kubernetes 1.25 |
|---------------|:---------------:|:---------------:|:---------------:|:---------------:|:---------------:|:---------------:|
| Kubean v0.4.4 |        ✓        |        ✓        |        ✓        |        ✓        |        ✓        |        ✓        |

## :book: Documents
- [Architecture](docs/zh/architecture.md)
- [Kubean vs Kubespray](docs/zh/comparisons.md)
- [CRD Outline](docs/zh/crds.md)
- [Deploy cluster using SSH secret key method](docs/zh/sshkey_deploy_cluster.md)
- [Cluster deployment for air gap environments](docs/zh/offline.md)
- [Custom Action](docs/zh/custom_action.md)
