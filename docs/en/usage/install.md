# Cluster Installation

Preconditions: Install [kubean charts](https://github.com/kubean-io/kubean-helm-chart) via helm.

---

## Install in a cluster with a single node

> Refer to [`minimal`](https://github.com/kubean-io/kubean/blob/main/examples/install/1.minimal/) sample template

Referring to the template, we will create an all-in-one single-node cluster:

#### 1. Update placeholders in [`AllInOne.yml`](https://github.com/kubean-io/kubean/blob/main/examples/install/1.minimal/AllInOne.yml) to real values

* `<IP1>`
* `<USERNAME>`
* `<PASSWORD>`
* `<TAG>`

#### 2. Apply [`AllInOne.yml`](https://github.com/kubean-io/kubean/blob/main/examples/install/1.minimal/AllInOne.yml)

``` bash
$ kubectl apply -f examples/install/1.minimal/
```

---

## Accelerator mode deployment

> Refer to [`mirror`](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/) sample template

#### 1. Update placeholders for yaml manifests in [`2.mirror`](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/) directory to real values

* `<IP1>` / `<IP2>` ...
* `<USERNAME>`
* `<PASSWORD>`
* `<TAG>`

#### 2. Apply the yaml manifest in [`2.mirror`](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/)

``` bash
$ kubectl apply -f examples/install/2.mirror/
```

#### 3. Please refer to [`VarsConfCM`](https://github.com/kubean-io/kubean/blob/main/examples/install/2.mirror/VarsConfCM.yml) for accelerator mirror settings

Accelerators used in this example:
* Binary acceleration: [public binary files mirror](https://github.com/DaoCloud/public-binary-files-mirror)
* Mirror acceleration: [public image mirror](https://github.com/DaoCloud/public-image-mirror)

---

## Offline installation

> Refer to [`airgap`](https://github.com/kubean-io/kubean/blob/main/examples/install/3.airgap/) sample template

For details, please refer to [Use of Offline Scenarios](./airgap.md)

---

## SSH key mode installation

For details, please refer to [Use SSH key to deploy K8S cluster](./sshkey_deploy_cluster.md)

---
