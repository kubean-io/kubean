# Kubean vs Kubespray

> English | [中文](../zh/comparisons.md)

<!--Kubespray 使用 Ansible 作为底层来配置和编排，可以运行在裸金属机、虚拟机、大多数云环境等。它支持众多 Kubernetes 版本和插件，可以完成集群从 0 到 1 的搭建和配置，也包含集群生命周期的维护，使用方式非常灵活。-->
Kubespray uses Ansible as the underlying layer to configure and orchestrate clusters. It can run on bare metal machines, virtual machines, and most kinds of cloud environment. It supports a wide range of Kubernetes versions and various plugins. With Kubespray, you can flexibly build and configure clusters from 0 to 1, and maintain you clusters through their lifecycles.

<!--Kubean 基于 Kubespray，拥有 Kubespray 所有优势。并且 Kubean 引用 Operator 概念以实现完全云原生化，原生以容器方式运行，提供 Helm Chart 包进行快速部署。-->
Kubean is based on Kubespray and boasts all the advantages of Kubespray. Moreover, Kubean introduces the concept of Operator to fully implement the philosophy of cloud native. Kubean is designed to run as a container, and can be easily installed with a Helm chart.

<!--Kubespray 仅在参数级别上支持离线，并没有包含一个完成构建离线安装包的过程，所以对于有离线场景需求的使用者来说，直接使用 Kubespray 会变得非常繁琐，这通常会让他们失去耐心。-->
Kubespray only supports offline at the parameter level and provides no process for building an offline install package, making it very troublesome for users who need to use it offline. They may gradually lose patience with Kubespray.

<!--Kubean 不仅有一套完善的制作离线包的工作流，还适配国产信创环境，简化 Kubespray 的复杂配置，能够对集群生命周期以云原生的方式去管理。-->
Kubean not only has a mature workflow for making offline packages, it also simplifies Kubespray's configuration, allowing users to manage cluster life cycle in a cloud-native way.
