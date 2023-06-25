# Installing kubean Helm application

## Prerequisites

1. You have a standard Kubernetes cluster or a cluster provided by a cloud provider.
2. Helm tool has been installed on your cluster control node or cloud terminal. [How to install Helm tool](https://helm.sh/docs/intro/install/)

---

## Deployment

####  1. Add kubean Helm repository

Add the kubean Helm repository to your local Helm repository by executing the following command on your existing cluster control node or cloud terminal.

```
$ helm repo add kubean-io https://kubean-io.github.io/kubean-helm-chart/
```
After completing the previous step, check if the kubean repository has been added correctly to your local Helm repository.

```bash
$ helm repo list
# Expected output:
NAME          	URL
kubean-io     	https://kubean-io.github.io/kubean-helm-chart/
```

#### 2. Install kubean

Check the available charts and their versions in the kubean Helm repository by executing the following command, which will list all the charts available in the kubean Helm repository.

```bash
helm search repo kubean

# Expected output:
NAME            	CHART VERSION	APP VERSION	DESCRIPTION
kubean-io/kubean	v0.5.2       	v0.5.2     	A Helm chart for kubean
```

After completing the above steps, execute the following command to install kubean.

```bash
$ helm install kubean kubean-io/kubean --create-namespace -n kubean-system
```

!!!note
    You can also use the "--version" parameter to specify the version of kubean.

#### 3. View installed kubean release

You have now completed the deployment of the kubean Helm chart. You can execute the following command to view the helm release in the kubean-system namespace.

```bash
$ helm ls -n kubean-system

# Expected output:
NAME  	NAMESPACE    	REVISION	UPDATED                                  	STATUS  	CHART            	APP VERSION
kubean	kubean-system	1       	2023-05-15 00:24:32.719770617 -0400 -0400	deployed	kubean-v0.4.9-rc1	v0.4.9-rc1

```