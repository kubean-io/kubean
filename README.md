# :seedling: KuBean

<a href="https://trackgit.com">
<img src="https://us-central1-trackgit-analytics.cloudfunctions.net/token/ping/la6t1t81jgv27ys97ila" alt="trackgit-views" />
</a>

<div align="center">

  <p>

[<img src="doc/images/certified-kubernetes-color.png" height=150>](https://github.com/cncf/k8s-conformance/pull/2240)
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
[![kubean coverage](https://raw.githubusercontent.com/dasu23/e2ecoverage/master/badges/kubean/kubeanCoverage.svg)](https://github.com/kubean-io/kubean/blob/main/doc/test/kubean_testcase.md)
[![kubean coverage](https://raw.githubusercontent.com/dasu23/e2ecoverage/master/badges/kubean/kubeanCoverage2.svg)](https://github.com/kubean-io/kubean/blob/main/doc/test/kubean_testcase.md)
[![license](https://img.shields.io/badge/license-AL%202.0-blue)](https://github.com/kubean-io/kubean/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubean-io/kubean)](https://goreportcard.com/report/github.com/kubean-io/kubean)

  </p>

</div>

---

## Quick Start

#### 1. Deploy Kubean-Operator

``` shell
helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
helm install kubean kubean-io/kubean --create-namespace -n kubean-system
```

Then check kubean-operator status by `kubectl get pods -n kubean-system | grep 'kubean'`.

#### 2. Start ClusterOperation for cluster.yml playbook

We cloud use the example in folder `artifacts/demo` which uses online resources to install k8s cluster.

  1. `cd artifacts/`
  2. modify `demo/hosts-conf-cm.yml` by replacing `IP1`, `IP2`... with the real ip where we want to install k8s cluster
  3. `kubectl apply -f demo/` to start kubeanClusterOps which will start the kubespray job
  4. `kubectl get job -n kubean-system` to check the kubespray job status

[![quick_start_image](doc/images/quick_start.gif)](https://asciinema.org/a/511386)

## Offline Usage

[offline](doc/offline.md)
