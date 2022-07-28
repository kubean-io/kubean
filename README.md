# :seedling: KuBean
[![example workflow](https://github.com/kubean-io/kubean/actions/workflows/main.yaml/badge.svg)](https://github.com/kubean-io/kubean/actions/workflows/main.yaml) [![codecov](https://codecov.io/gh/kubean-io/kubean/branch/main/graph/badge.svg?token=8FX807D3QQ)](https://codecov.io/gh/kubean-io/kubean) [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6263/badge)](https://bestpractices.coreinfrastructure.org/projects/6263)
# Introduction
kubean is a cluster lifecycle management tool based on kubespray.

# Quick Start

## Deploy Kubean-Operator

```
helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
helm install kubean kubean-io/kubean --create-namespace -n kubean-system
```

Then check kubean-operator status by `kubectl get pods -n kubean-system | grep 'kubean'`.

## Start KuBeanClusterOps for cluster.yml playbook

We cloud use the example in folder `artifacts/demo` which uses online resources to install k8s cluster.

1. `cd artifacts`
2. modify `demo/hosts-conf-cm.yml` by replacing `IP1`, `IP2`... with the real ip where we want to install k8s cluster
3. `kubectl apply -f demo` to start kubeanClusterOps which will start the kubespray job
4. `kubectl get job -n kubean-system` to check the kubespray job status


[![quick_start_image](doc/images/quick_start.gif)](https://asciinema.org/a/511386)

# Offline Usage

[offline](doc/offline.md)