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
- [x] Add certificate renewal script: https://github.com/kubean-io/kubean/pull/884
- [x] Implement daily deployment verification for the latest version of upstream kubespray: https://github.com/kubean-io/kubean/pull/870
- [x] Ensure cascading deletion of Cluster resources: https://github.com/kubean-io/kubean/pull/918
- [x] Add cleanup weight for ClusterOperation records: https://github.com/kubean-io/kubean/pull/983

## Q4 2023
- [x] Optimize offline image package to OCI format: https://github.com/kubean-io/kubean/pull/996
- [x] Improve logging input for Operators: https://github.com/kubean-io/kubean/pull/1032
- [x] Enhance query efficiency of Manifest resources: https://github.com/kubean-io/kubean/pull/1036
- [x] Refactor image import script to support multi-architecture import: https://github.com/kubean-io/kubean/pull/1040

## Q1 2024
- [x] Improve execution efficiency of precheck script: https://github.com/kubean-io/kubean/pull/1076
- [x] Optimize tuning performance of ClusterOperation: https://github.com/kubean-io/kubean/pull/1082
- [x] Refactor logic of custom resource generation script: https://github.com/kubean-io/kubean/pull/1152
- [x] Fix offline package version issue for Ubuntu 18.04: https://github.com/kubean-io/kubean/pull/1158
- [x] Automate pre-steps for limiting disk usage per container in Docker: https://github.com/kubean-io/kubean/pull/1179

## Q2 2024
- [ ] Provide a client command-line tool and convenient method for generating custom resource modules
- [ ] Capacity planning for cluster deployment on different node scales
- [ ] Provide a complete offline resource management solution
- [ ] Support multiple lifecycle management engines, such as kubespray and kubekey.
- [ ] Enable cluster operation rollback based on ostree.