# Background

The following is an unofficial introduction, purely personal understanding:

e2e testing: End-to-End scenario testing from user point of view, which is different from unit testing for a single function or method with fixed input and output. E2E testing often targets a set of functions or methods combined to achieve a specific function, and it may involve some pre-preparation and creation work, as well as deletion and clean-up recovery work after the test is completed.

# Tool Selection

Kind: Quickly build a kubernetes cluster for testing by running containers to make tests repeatable. Recommended reading: https://kind.sigs.k8s.io/

Ginkgo: A popular e2e testing framework in the industry. Recommended reading: https://ke-chain.github.io/ginkgodoc/

Gomega: The assertion library that best matches ginkgo.

# HowTo
## Environment setup

````
hack/local-up-kindcluster.sh
````
This script does the following:

1. Determine whether the docker, go, helm and go versions installed in the environment meet the requirements, otherwise exit.
2. If kind and kubectl are not available in the current environment, install them.
3. Import the kubean image into the host cluster.
4. Deploy kubean on the host cluster.

The output at the end of the run is as follows:
````
Local kubean is running.
To start using your kubean, run:
  export KUBECONFIG=/home/actions-runner/.kube/kubean-latest-32323-host.config
Please use 'kubectl config use-context kubean-host' to switch the host and control plane cluster.
````
You can switch the cluster context as prompted:
````
export KUBECONFIG=/home/actions-runner/.kube/kubean-latest-32323-host.config
````
Check the kubean running status:
````
NAME READY STATUS RESTARTS AGE
pod/kubean-88ccdcf4b-qzqfb 1/1 Running 0 8m59s
pod/kubean-e2e-cluster1-install-job-l9hqx 1/1 Running 0 7m2s

NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
service/kubean ClusterIP 10.96.94.43 <none> 80/TCP 8m59s

NAME READY UP-TO-DATE AVAILABLE AGE
deployment.apps/kubean 1/1 1 1 8m59s

NAME DESIRED CURRENT READY AGE
replicaset.apps/kubean-88ccdcf4b 1 1 1 8m59s

NAME COMPLETIONS DURATION AGE
job.batch/kubean-e2e-cluster1-install-job 0/1 7m2s 7m3s
````
Run successfully.

## E2E Test
````
hack/e2e.sh
````
This script mainly installs ginkgo and runs all e2e tests under test/ with the ginkgo command.

### Testing Framework
test/kubean_deploy_e2e/test_suite_test.go, with which you can run all e2e tests under test/kubean_deploy_e2e/ with ginkgo command.

### Testing Code
Take test/kubean_deploy_e2e/kubean_deploy_test.go as an example, it is used to test whether the kubean operator is installed successfully.

Import method: avoid dot-import, dot-import cannot pass the syntax check of kubean CI.

Entry: var _ = ginkgo.Describe >> ginkgo.Describe >> ginkgo.Context >> ginkgo.It

* ginkgo.BeforeEach: performed before ginkgo.It, here is to create a CR in the host cluster and manage the member1 cluster.
* ginkgo.AfterEach: After ginkgo.It, here is to delete the CR in the host cluster and remove the member1 cluster.
* ginkgo.Describe: scene description.
* ginkgo.Context: Test scenario entry, for example, you can write a positive example Context and a negative example Context. For example "When ...".
* ginkgo.It: Under the input described by Context, what should be the output, or what the result should be. For example "Should...".
* gomega.Expect: result assertion, such as true or false, whether it is empty, whether it is equal.
The above and more are to organize our tests for different use cases in different scenarios in a more orderly manner. You can read more ginkgo official documents and refer to excellent open source projects to learn.

## Environment cleanup After Testing
````
hack/delete-cluster.sh
````
This script is responsible for cleaning up the environment, i.e. deleting the kind cluster.

**Please note to developers: each PR will run the e2e test; if the test fails, developer need to check whether the commits has problems or the commits does not make consistent corresponding changes in the e2e test**