
# 背景

以下内容非官方介绍，纯个人理解：

e2e 测试：end to end，端到端场景测试，区别于有固定输入和输出的针对单个函数或方法的单元测试，e2e 测试往往针对的是一组函数或方法组合起来以实现的特定功能，也可能涉及到一些前置的准备和创建工作以及测试完成后的删除和清理还原工作。

# 工具选型

kind：以运行容器的方式，快速搭建一个测试用的 kubernetes 集群，让测试可重复。推荐学习阅读：https://kind.sigs.k8s.io/

ginkgo：业界流行的e2e测试框架。推荐学习阅读：https://ke-chain.github.io/ginkgodoc/

gomega：与ginkgo最为匹配的断言库。

# 实现
## 环境搭建

```
hack/local-up-kpanda.sh
```
这个脚本做了以下事情：

1. 判断环境是否安装 docker、go、helm 及 go 版本是否符合要求，否则退出。
2. 如果当前环境没有 kind 和 kubectl，则安装。
3. 按照 artifacts/kindClusterConfig/kpanda-host.yaml && member1.yaml 创建网络连通的 host 集群和 member1 集群（kind 集群），并检查集群就绪情况。
4. 把 kpanda 的镜像导入 host 集群。**开发者请注意：请自行保障 kpanda 各组件在 release.daocloud.io 中的 latest 版本为最新版本，如果新增组件，在这里也需相应地新增代码**
5. 在 host 集群上部署 kpanda。
6. kpanda 的就绪检查，依赖 pod label。**开发者请注意：请自行保障 kpanda 各组件的 pod label 与此处代码保持一致**

运行结束的输出如下：
```
Local kubean is running.

To start using your kpanda, run:
  export KUBECONFIG=/root/.kube/kpanda.config
Please use 'kubectl config use-context kpanda-host' to switch the host and control plane cluster.

To manage your member clusters, run:
  export KUBECONFIG=/root/.kube/member1.config
Please use 'kubectl config use-context member1' to switch to the member cluster.
```
即可根据提示切换集群上下文：
```
export KUBECONFIG=/root/.kube/kpanda.config
```
查看 kubean 运行状态：
```
[root@node1 ~]# kubectl get all -n kpanda-system
NAME                                             READY   STATUS    RESTARTS   AGE
pod/kpanda-apiserver-7d69cb6857-qclq6            1/1     Running   0          18h
pod/kpanda-apiserver-7d69cb6857-tvq5t            1/1     Running   0          18h
pod/kpanda-controller-manager-5c7d5db7df-fjhjk   1/1     Running   0          18h
pod/kpanda-controller-manager-5c7d5db7df-hqpcv   1/1     Running   0          18h

NAME                       TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
service/kpanda-apiserver   NodePort   10.96.174.187   <none>        80:30000/TCP   18h

NAME                                        READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/kpanda-apiserver            2/2     2            2           18h
deployment.apps/kpanda-controller-manager   2/2     2            2           18h

NAME                                                   DESIRED   CURRENT   READY   AGE
replicaset.apps/kpanda-apiserver-7d69cb6857            2         2         2       18h
replicaset.apps/kpanda-controller-manager-5c7d5db7df   2         2         2       18h

[root@node1 ~]# kubectl get crd |grep kpanda
clusters.cluster.kpanda.io   2022-01-05T07:55:46Z
```
运行成功。

## e2e 测试
```
hack/run-e2e.sh
```
这个脚本主要是安装 ginkgo 并用 ginkgo 命令跑 test/e2e/ 下所有的 e2e 测试。

### 脚手架
test/e2e/test_suite_test.go，有了它就可以用 ginkgo 命令跑 test/e2e/ 下所有的 e2e 测试。

### 测试代码
以 test/e2e/test_cluster_cr.go 为例，它是用来测试 host 集群纳管 member 集群，即 host 集群上创建的 cluster CR 是否符合预期，如 node 数和 pod 数。

导入方式：避免 dot-import ，dot-import 无法通过 kubean CI 的语法检查。

入口：var _ = ginkgo.Describe >> ginkgo.Describe >> ginkgo.Context >> ginkgo.It

* ginkgo.BeforeEach：在 ginkgo.It 之前进行，此处是在 host 集群创建 CR ，纳管 member1 集群。
* ginkgo.AfterEach：在 ginkgo.It 之后进行，此处是在 host 集群删除 CR ，移除 member1 集群。
* ginkgo.Describe：场景描述。
* ginkgo.Context：测试场景入口，比如可以写一个正例 Context 和一个反例 Context。比如 "When ..."。
* ginkgo.It：在 Context 描述的输入下，输出应该是什么，或者说结果应该是什么。比如 "Should ..."。
* gomega.Expect：结果断言，比如真假、是否为空、是否相等。
以上以及更多都是为了更加井然有序地组织我们针对不同场景的不同用例的测试，可多阅读 ginkgo 官方文档并参考优秀开源项目学习。

## 环境清理
```
hack/delete-cluster.sh
```
这个脚本负责清理环境，即删除 kind 集群。

# 与CI集成

CI 选用的是 GitLab Runner，其流程定义在 [.gitlab-ci.yml](https://docs.gitlab.com/ee/ci/yaml/gitlab_ci_yaml.html)，e2e 测试在 stage: e2e-test 中的 e2e_test 中定义。e2e 测试 Job 用 image: release.daocloud.io/kpanda/e2e:latest 启动一个临时容器 A，以 docker in docker 的方式创建 kind 集群将在同一虚机上的创建容器 B。看似 docker in docker，其实容器 A 和 B 都跑在同一虚机上。

release.daocloud.io/kpanda/e2e:<tag> 从哪里来？需要手动构建和推送，安装以上过程必须的 docker、go、helm、kind、kubectl 等命令，Dockerfile build/images/kpanda-e2e/Dockerfile。这个镜像的作用主要是提供一些可执行命令，一般不需要再次构建，直接使用即可。

**开发者请注意：每一个 PR 都会跑 e2e 测试，如果测试失败，需自行检查是否改动的代码有问题或改动代码没有在 e2e 测试中做一致的相应改动**
