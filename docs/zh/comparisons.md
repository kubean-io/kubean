# 优劣对比

> [English](../en/comparisons.md) | 中文

## Kubean vs Kubespray

Kubespray 使用 Ansible 作为底层来配置和编排，可以运行在裸金属机、虚拟机、大多数云环境等。它支持众多 Kubernetes 版本和插件，可以完成集群从 0 到 1 的搭建和配置，也包含集群生命周期的维护，使用方式非常灵活。

Kubean 基于 Kubespray，拥有 Kubespray 所有优势。并且 Kubean 引用 Operator 概念以实现完全云原生化，原生以容器方式运行，提供 Helm Chart 包进行快速部署。

Kubespray 仅在参数级别上支持离线，并没有包含一个完成构建离线安装包的过程，所以对于有离线场景需求的使用者来说，直接使用 Kubespray 会变得非常繁琐，这通常会让他们失去耐心。

Kubean 不仅有一套完善的制作离线包的工作流，还适配国产信创环境，简化 Kubespray 的复杂配置，能够对集群生命周期以云原生的方式去管理。
