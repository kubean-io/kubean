# Kubean Roadmap

The current Roadmap is tentative, and the exact schedule depends on the needs of the community.

> **For features not mentioned in the Roadmap, we can discuss them in the [issues](https://github.com/kubean-io/kubean/issues).**



## Q3 2022
- [x] Design Kubean project architecture process like [architecture.md](https://github.com/kubean-io/kubean/blob/main/docs/en/architecture.md)
- [x] Verify Kubean's cluster lifecycle management operations
- [x] Add OS packages to build CI https://github.com/kubean-io/kubean/pull/62
- [x] Provide Kubean API https://github.com/kubean-io/kubean/pull/128


## Q4 2022
- [x] Implement E2E tests like [kubean test case](https://github.com/kubean-io/kubean/blob/main/docs/test/kubean_testcase.md)
- [x] Ensure k8s images and binary packages support the ARM architecture https://github.com/kubean-io/kubean/pull/200
- [x] Support for upgrade package builds https://github.com/kubean-io/kubean/pull/289
- [x] Adapt the deployment for offline scenario RHEL8.4 https://github.com/kubean-io/kubean/pull/325
- [x] Support for restore package manager configuration https://github.com/kubean-io/kubean/pull/298
- [x] Support for restoring Kubeconfig after the cluster deployment https://github.com/kubean-io/kubean/pull/192
- [x] Add SSH Key authentication deployment method https://github.com/kubean-io/kubean/pull/302


## Q1 2023
- [x] Support for apt package manager configuration https://github.com/kubean-io/kubean/pull/459
- [x] Support for custom actions for Cluster Operation CRD https://github.com/kubean-io/kubean/issues/361
- [x] Kubean chart supports charts-syncer https://github.com/kubean-io/kubean/pull/468
- [x] Add pre-testing before deployment https://github.com/kubean-io/kubean/pull/555
- [x] Adapt the Uniontech V20 1020a linux https://github.com/kubean-io/kubean/pull/583


## Q2 2023
- [x] Support for clustered deployments based on OpenEuler offline scenarios https://github.com/kubean-io/kubean/pull/628
- [x] Support for Other Linux to automatically build offline scenario-dependent system packages via scripts https://github.com/kubean-io/kubean/pull/627
- [x] Update the kubean documentation site with mkdocs https://github.com/kubean-io/kubean/pull/728
- [x] Optimize release CI https://github.com/kubean-io/kubean/pull/863
- [x] Add ansible script for certificate renewal https://github.com/kubean-io/kubean/pull/884
- [x] Update the release process https://github.com/kubean-io/kubean/pull/869

## Q3 2023
- [ ] Optimize Kubean documentation website
- [ ] Support for deployment in more OS offline scenarios
- [ ] Provide a convenient way to generate customized resource templates
- [ ] Optimization of the use of resources such as images and binaries related to offline scenarios