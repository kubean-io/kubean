# API设计规范

```text
Author: Xiao Wu.Zhu <xiaowu.zhu@daocloud.io>
Date:   2022-05-11
Version: v1.0.0
```

## Table of Contents

本文档描述了`Amamba`的API设计规范，面向希望更深入了解`Amamba`API结构的用户以及希望扩展`Amamba`的API功能的开发人员。

<!-- toc -->
- [URL格式定义](#URL格式定义)
- [协议](#协议)
- [域名](#域名)
- [path](#path)
  - [API version](#api-version)
    - [API弃用策略](#API弃用策略)
    - [Alpha-> Beta->GA的要求](#Alpha -> Beta -> GA 的要求)
      - [Alpha](#Alpha)
      - [Beta](#Beta)
      - [GA](#GA)
  - [路径](#路径)
  - [过滤信息](#Query)
  - [API示例](#API示例)
- [HTTP动词](#HTTP动词)
- [状态码](#状态码)
- [错误处理](#错误处理)
- [返回结果](#返回结果)
- [Hypermedia API](#Hypermedia API)
- [授权](#授权)
- [性能要求](#性能要求)
- [SDK](#SDK)
- [其他](#其他)
- [Ref](#Ref)
<!-- /toc -->

Amamba的 API的样式是RESTful的，客户端通过标准的HTTP协议来访问API。通过POST、PUT、DELETE、GET来创建、更新、删除、查询数。这些API接收并返回JSON格式的数据。

# URL格式定义
URL 组成部分 `protocol://username:password@host:port/path?query`

# 协议

API与用户的通信协议，总是使用HTTPs协议，端口采用443端口。除非调试接口，否则严禁使用HTTP协议。`Amamba` 使用 [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) 同时支持HTTP和gRPC协议。

# 域名

应该尽量将API部署在专用域名之下。如`https://api.example.com`如果确定API很简单，不会有进一步扩展，可以考虑放在主域名下。`https://example.org/api/`

# Path
path部分以api开头，后跟微服务名、版本及模块名或者资源地址。例如`https://host:port/api/kubean.io/v1/notification`。

API 将一组一起公开的资源、版本分组为"Group/Version/Resource",例如 `kubean.io/v1alpha/workspaces`。`kubean.io` 为组名，`v1alpha` 为版本号, `workspaces` 为资源名。
- 在选择组名时，应该使用组织拥有的子域，必须小写且为有效的DNS子域名，请避免使用预定义的组名，例如`*.k8s.io`。
- 版本字符串应该与格式匹配
- 资源集合应该全部为小写和复数形式。

## API version
应该将API的版本号放入URL。`https://api.example.com/v1`。
另一种做法是，将版本号放在HTTP头信息中，但不如放入URL方便和直观。Github采用这种做法。

参考 OpenAPI 子系统，API 的版本，也同样存在一个 `alpha → beta → GA` 的演进。对于 “比较复杂” 的软件，可以区分各个 API 的版本成l熟度。alpha 若干次 版本后，才允许晋级到 beta。

### API弃用策略

- 只有当有替代方案的时候，API才可以被弃用
- 对于 beta 需要 9个月的弃用时间，ga 的软件需要 12个月的弃用时间，确保API的兼容性

### Alpha -> Beta -> GA 的要求

#### Alpha

- Implemented the feature
- Unit test
- Investigate how e2e tests

#### Beta

- Addresses feedback from alpha testers
- Sufficient E2E and unit testing

#### GA

- Addresses feedback from beta
- Sufficient number of users using the feature
- Confident that no further API / kubelet config configuration options changes are needed
- Close on any remaining open issues & bugs

参考kubernetes的一些实践：[kubernetes 版本弃用策略](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/2000-graceful-node-shutdown)

## 路径

路径又称"终点"（endpoint），表示API的具体网址。在RESTFul架构中，每个网址代表一种资源（resource），所以网址中不能有动词，只能有名词，而且所用的名词往往与数据库的表名对应。一般来说，数据库中的表都是同种记录的"集合"（collection），所以API中的名词也应该使用复数。例如：

- https://api.example.com/v1/zoos

在路径设计中需要遵守下列约定
- 资源命名全部小写且易读，可使用连字符（-）或下划线（_）进行分隔
- 资源名不可以使用(.) 和(..)，会被浏览器识别为相对路径。
- 资源的命名需要符合RESTful风格，应该采用名词。
- 路径部分使用斜杠分隔符（/）来表达层次结构。`/parent/child`

## Query
如果记录数量很多，服务器不可能都将它们返回给用户。API应该提供参数，过滤返回结果。下面是一些常见的参数。
- ?limit=10：指定返回记录的数量
- ?offset=10：指定返回记录的开始位置。
- ?pageNum=2&pageSize=100：指定第几页，以及每页的记录数。
- ?orderBy=name&orderType=asc：指定返回结果按照哪个属性排序，以及排序顺序。
- ?animalTypeId=1：指定筛选条件

参数的设计允许存在冗余，即允许API路径和URL参数偶尔有重复。比如，GET /zoo/ID/animals 与GET /animals?zoo_id=ID的含义是相同的

参数设计需要遵循下列约定：
- 参数名使用下划线分隔小写字母的方式命名
- 采用OAS 3.0描述的API，必须遵循查询参数序列化规则。https://swagger.io/docs/specification/serialization/

### API示例
`Amamba` 的API根据 `Amamba` 的[架构](../artifacts/kubean-architecture.png)模块来做划分不同的组,例如
- gitops
- rollout
不同组的API各自维护和发版，它们的组名分别为为`pipeline.kubean.io`,`gitops.kubean.io`。示例如下：

`https://host:port/api/pipeline.kubean.io/v1alpha1/jenkinspipelines/{pipelineID}`

- `https://host:port/api/` 这部分是通用的
- `pipeline.kubean.io` 表示组名
- `v1alpha1` 表示api的版本号
- `jenkinspipelines`：表示在`pipeline.kubean.io`组下，版本为`v1alpha1`的一种资源。而`pipeline.kubean.io`这个组下可能还有其他的一些资源，例如`gitlabpipeline`等
- `{pipelineID}`: 表示`jenkinspipeline`对象的唯一标识符

从上述的API描述中可以看出`Amamba`的api path是follow k8s的GVR风格的，由Group + Version + Resource来表示一个资源。
而对于resource的具体操作则是通过[HTTP的动词](#HTTP动词)呈现的。常见的几个示例如下

- HTTP GET `https://host:port/api/pipeline.kubean.io/v1alpha1/workspaces/{ws}/jenkinspipelines/{pipelineID}` 表示根据`pipelineID`获取pipeline对象的详情
- HTTP GET `https://host:port/api/pipeline.kubean.io/v1alpha1/jenkinspipelines?pipelineName="test"`, 表示根据查询条件获取pipeline对象列表
- HTTP DELETE `https://host:port/api/pipeline.kubean.io/v1alpha1/workspaces/{ws}/jenkinspipelines/{pipelineID}` 表示根据`pipelineID` 删除pipe对象
- HTTP POST `https://host:port/api/pipeline.kubean.io/v1alpha1/workspaces/{ws}/jenkinspipelines` 表示新建一个pipe资源，数据通过body传递
- HTTP PUT `https://host:port/api/pipeline.kubean.io/v1alpha1/workspaces/{ws}/jenkinspipelines/{pipelineID}` 表示根据`pipelineID` 更新pipe对象

而对于某个对象的一些操作，可能无法使用名词来清晰的描述，那么可以使用subPath的方式，例如`/workspaces/{ws}/jenkinspipelines/{pipelineID}/retry`,常见的一些action如下
- retry
- download
- copy
- stop

从上述的API示例中也能看到，有的path中存在`workspaces/{ws}/`,说明`jenkinspipelines` 资源是与`workspace`无关的，`workspace`是一个scope的概念，它可以作为`jenkinspipelines`资源的一个属性，那么实际上url中包含`workspaces/{ws}/jenkinspipelines` 与 `jenkinspipelines?workspace={ws}`本质上是一样的，只是在url中表示出查询`jenkinspipelines`的范围，而将ws写到path中的方式具有很高的可读性，并且对于编程实现比较友好（routerHandleFunc,Auth）

# HTTP动词

对于资源的具体操作类型，由HTTP动词表示,常用的HTTP动词有下面几个:

- GET（SELECT）：从服务器获取一个或者多个资源
- HEAD: 与GET类似，但是响应的时候只返回首部
- POST（CREATE）：在服务器新建一个资源
- PUT（UPDATE）：在服务器更新资源（全量更新）
- PATCH（UPDATE）：在服务器更新资源（增量更新）， 因为PATCH存在部分浏览器的兼容性问题，使用PUT实现PATCH的效果
- DELETE（DELETE）：从服务器删除资源

# 状态码

服务器向用户返回的状态码和提示信息，常见的有以下一些（方括号中是该状态码对应的HTTP动词）。
- 200 OK -  [GET]：服务器成功返回用户请求的数据，该操作是幂等的。
- 201 CREATED -  [POST/PUT/PATCH]：用户新建或修改数据成功（例如上传新文件）,Location报头指向新创建资源的地址。
- 202 Accepted -  [*]：表示一个请求已经进入后台排队（异步任务），服务器还未对其执行任何动作。
- 204 NO CONTENT -  [DELETE]：用户删除数据成功。表示服务器已成功满足请求，并且响应有效内容正文中没有其他要发送的内容。
- 301 Moved Permanetly - [*]: 重定向状态码, 通过响应头的Location 字段访问资源现在的地址。
- 400 INVALID REQUEST -  [POST/PUT/PATCH]：用户发出的请求有错误，服务器没有进行新建或修改数据的操作，该操作是幂等的。
- 401 Unauthorized -  [*]：表示用户没有权限（令牌、用户名、密码错误）。
- 403 Forbidden -  [*] 表示用户得到授权（与401错误相对），但是访问是被禁止的。
- 404 NOT FOUND -  [*]：用户发出的请求针对的是不存在的记录，服务器没有进行操作，该操作是幂等的。
- 405 Method Not Allowed - [*]: 目标资源不知道对应的请求方法。
- 406 Not Acceptable -  [GET]：用户请求的格式不可得（比如用户请求JSON格式，但是只有XML格式）。
- 410 Gone - [GET]：用户请求的资源被永久删除，且不会再得到的。
- 422 Unprocesable entity -  [POST/PUT/PATCH] 当创建一个对象时，发生一个验证错误。
- 500 INTERNAL SERVER ERROR -  [*]：服务器发生错误，用户将无法判断发出的请求是否成功。
- 503 Service Unavailable - [*]: 由于服务器暂时不可用而导致API请求失败，返回此状态码。

# 错误处理
需要统一 `Amamba` 接口返回的错误风格，与开源对齐，兼顾 restful 与 grpc 。

## 前置背景
google 有一套 grpc 返回码与 http 状态码的对应关系，如下表所示:

|HTTP   |gRPC     |说明    |
|-----------|--------|------------ |
200|        OK|         无错误。
400|        INVALID_ARGUMENT|       客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。
400|        FAILED_PRECONDITION|        请求无法在当前系统状态下执行，例如删除非空目录。
400|        OUT_OF_RANGE|       客户端指定了无效范围。
401|        UNAUTHENTICATED|        由于 OAuth 令牌丢失、无效或过期，请求未通过身份验证。
403|        PERMISSION_DENIED|          客户端权限不足。这可能是因为 OAuth 令牌没有正确的范围、客户端没有权限或者 API 尚未启用。
404|        NOT_FOUND|          未找到指定的资源。
409|        ABORTED|        并发冲突，例如读取 / 修改 / 写入冲突。
409|        ALREADY_EXISTS|         客户端尝试创建的资源已存在。
429|        RESOURCE_EXHAUSTED|         资源配额不足或达到速率限制。如需了解详情，客户端应该查找 google.rpc.QuotaFailure 错误详细信息。
499|        CANCELLED|          请求被客户端取消。
500|        DATA_LOSS|          出现不可恢复的数据丢失或数据损坏。客户端应该向用户报告错误。
500|        UNKNOWN|        出现未知的服务器错误。通常是服务器错误。
500|        INTERNAL|       出现内部服务器错误。通常是服务器错误。
501|        NOT_IMPLEMENTED|        API 方法未通过服务器实现。
502|        不适用|         到达服务器前发生网络错误。通常是网络中断或配置错误。
503|        UNAVAILABLE|        服务不可用。通常是服务器已关闭。
504|        DEADLINE_EXCEEDED|          超出请求时限。仅当调用者设置的时限比方法的默认时限短（即请求的时限不足以让服务器处理请求）并且请求未在时限范围内完成时，才会发生这种情况。

[原始连接](https://cloud.google.com/apis/design/errors#handling_errors)

## User Stories (Optional)

### Story 1

前端通过 Amamba restful api 进行访问的时候，http status code 以及返回内容需要尽可能 follow restful 相关规范(20x,30x,40x,50x)

### Story 2

其它后端模块(例如: insight)通过 Amamba grpc service 进行访问的时候，相关返回值以及返回状态码需要尽可能 follow grpc 相关的状态码以及返回值规范

### Design Details

grpc-gateway 内置有异常处理的中间件，使用方式非常简单，在 grpc 初始化的时候开启即可,如下所示:

```bash
mux := runtime.NewServeMux(
  runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
)
```

`grpc-gateway` 内置错误处理能够同时处理 grpc 以及 http 的访问异常处理，内部维护一个 grpc 返回码 与 http 状态码的 mapping 关系(即上表 google 定义的转换关系)。

回到` Amamba `本身的异常错误设计，个人觉得参考 grpc-gateway 异常设计即可，没必要重复设计一套特立独行的例子。我们在使用的时候需要符合 `grpc-gateway` 默认设计。需要在 API 请求层进行错误处理，对应到 `Amamba` 即在 BFF 层显示的进行异常错误处理并返回。

### 示例
在 grpc 内部服务返回如下异常

```bash
return nil, status.Errorf(codes.Internal, "%v", errors.New("hello the world"))
```

服务内部 500 发生时， grpc 调用方式错误返回信息
```bash
2022/03/28 15:44:22 rpc error: code = Internal desc = hello the world
```

服务内部 500 发生时， http restful 调用方式错误返回信息
```bash
{
    "code": 500,
    "message": "hello the world",
    "details": []
}
```
此时，http status code 会 grpc-gateway 默认的异常处理中间件自动标记为 500

# 返回结果

针对不同操作，服务器向用户返回的结果应该符合以下规范。
- GET /collection：返回资源对象的列表（数组）
- GET /collection/resource：返回单个资源对象
- POST /collection：返回新生成的资源对象
- PUT /collection/resource：返回完整的资源对象
- PATCH /collection/resource：返回完整的资源对象
- DELETE /collection/resource：返回一个空文档

对于HTTP方法的使用，一定要遵循相应方法的安全性和幂等性

|   方法   |  安全性    |  幂等性    |
|-----------|--------|------------ |
|   GET   | YES | YES |
| HEAD | YES | YES |
| OPTIONS | YES | YES |
| PUT | NO | YES |
| DELETE | NO | YES |
| POST | NO | NO |

# Hypermedia API

RESTFul API最好做到Hypermedia，即返回结果中提供链接，连向其他API方法，使得用户不查文档，也知道下一步应该做什么。
比如，当用户向api.example.com的根目录发出请求，会得到这样一个文档。

```json
{
    "link": {
        "rel": "collection https://www.example.com/zoos",
        "href": "https://api.example.com/zoos",
        "title": "List of    zoos",
        "type": "application/vnd.yourformat+json"
    }
}
```

上面代码表示，文档中有一个link属性，用户读取这个属性就知道下一步该调用什么API了。rel表示这个API与当前网址的关系（collection关系，并给出该collection的网址），href表示API的路径，title表示API的标题，type表示返回类型。

Hypermedia API的设计被称为HATEOAS。Github的API就是这种设计，访问api.github.com会得到一个所有可用API的网址列表。

```json
{
    "current_user_url": "https://api.github.com/user",
    "authorizations_url": "https://api.github.com/authorizations"
  // ....
}

```

从上面可以看到，如果想获取当前用户的信息，应该去访问api.github.com/user，然后就得到了下面结果

```json
{
    "message": "Requires authentication",
    "documentation_url": "https://developer.github.com/v3"
}

```

上面代码表示，服务器给出了提示信息，以及文档的网址。

# 授权
使用OAuth 2.0 的方式实现API授权，先请求授权，获取ACCESS_TOKEN，然后再通过该ACCESS_TOKEN来调用需要授权的API。调用需要授权的API时，必须将ACCESS_TOKEN放在HTTP header中："Authorization: Bearer ACCESS_TOKEN"。

# 性能要求
[DCE5.0的API性能要求](https://dwiki.daocloud.io/pages/viewpage.action?pageId=95578845#heading-13API%E3%80%81%E9%A1%B5%E9%9D%A2%E7%9A%84%E6%80%A7%E8%83%BD%E8%A6%81%E6%B1%82)：
- 正确率： 100%
- 延迟： 99% 的请求在100ms完成（不包含浏览器到服务器的网络延迟），平均20ms最佳
- QPS： 1000以上
- 可扩展性： API性能可以随着实例数线性扩展

# SDK
- API Spec 文档，通过 swagger 内置在源代码中， 并且应该由源代码自动生成而不是手动维护

# 其他
- 请求body、响应body中的数据，即是资源表述，描述资源地址所指向数据的样子，格式为json，只允许使用object和array，不允许直接请求或响应string、number等。
- 所有请求和响应的"Content-Type"必须和实际的数据格式保持一致，保证调试工具(如：Postman)、SDK生成工具、文档生成工具能够正确识别接口。

# Ref
- https://dwiki.daocloud.io/pages/viewpage.action?pageId=103964476
- https://dwiki.daocloud.io/pages/viewpage.action?pageId=25559704&preview=%2F25559704%2F25559709%2FREST+API%E8%AE%BE%E8%AE%A1%E8%A7%84%E8%8C%83.docx
- https://dwiki.daocloud.io/pages/viewpage.action?pageId=60911406&preview=%2F60911406%2F73347363%2FRESTful+API%E8%AE%BE%E8%AE%A1%E8%A7%84%E8%8C%83.pdf
- https://dwiki.daocloud.io/pages/viewpage.action?pageId=95578845#heading-13API%E3%80%81%E9%A1%B5%E9%9D%A2%E7%9A%84%E6%80%A7%E8%83%BD%E8%A6%81%E6%B1%82
