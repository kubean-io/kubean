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

## Install

First, add the Kubean chart repo to your local repository.
``` bash 
$ helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/

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
$ kubectl get kubeancluster
```

View cluster operation jobs.
``` bash
$ kubectl get kubeanclusterops
```

## Uninstall

If kubean's related custom resources already exist, you need to clear.
``` bash
$ kubectl delete kubeanclusterops --all
$ kubectl delete kubeancluster --all
```

Uninstall kubean's components via helm.
``` bash
$ helm -n kubean-system uninstall kubean
```
