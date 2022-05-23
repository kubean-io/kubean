# KUBEAN（应用工作台）架构设计说明

```
Author: Xiao Yang<yang.xiao@daocloud.com>
Date: 2022-5-12
Version: v1.0
```

## 数据流图

[](../artifacts/dataflow.png)

应用工作台主要包含以下主要模块:

- apiserver，主要负责接收和分发请求，鉴权、审计等通用性的工作也是由apiserver来处理；
- app-lifecycle，主要负责应用的生命周期管理，包括引导式创建应用、应用概览和详情、应用可观测性等；
- biz，包括pipline（流水线）、rollout（灰度发布）、gitops这些功能模块。这些模块一方面会被app-lifecycle调用以提供应用的部分能力、另一方面会调用一些外部服务，例如pipeline依赖的jenkins、gitops依赖的argocd等
- support，提供诸如workspace、凭据管理等支撑性的功能；
- ACL，提供对外访问的一层抽象，诸如缓存、服务抽象等；
- providers，外部服务，通过公共的addon组件拉起来。值得注意的是，prometheus、istio这种多个子系统都会用到的基础组件可以和其他系统公用一个；

## 分层架构设计

[](../artifacts/layerd-arch.png)

Amamba 在架构上主要分为四层：

### APISERVER 

apiserver 提供http和grpc server，做一些通用的处理:

- 使用grpc-gateway完成请求从http到grpc的转换；
- 通过中间件对外部的http请求进行处理，包括鉴权、头信息的处理、拦截等；
- 通过inceptor（拦截器）对内部的grpc请求进行鉴权，同时也包括统一的审计等；

### APP

应用层主要负责grpc服务接口的实现，它向apiserver注册，按照.proto文件的定义的格式接受请求和响应。

并将数据从基于protobuf生成的对象转换成内部的基于业务表示的对象。


### BIZ

业务层是业务逻辑的抽象和实现。这一层主要负责基于业务逻辑抽象出实体和方法。这些实体最核心，不依赖任何其他模块。

repository也在这一层，定义了抽象于具体数据存储方式的接口；

对于调用外部服务也需要在这一层抽象成基于实体的方法，避免直接调用外部函数；

### INFRA

基础设施层规定了外部服务和相关软件，例如ORM和数据库，SDK和其他服务。他们实现了在biz层的接口。

## 代码布局

TODO


