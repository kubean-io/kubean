# 命名规则

为了能够长时间在众多 API 中为开发者提供一致的体验，API 使用的所有名称都**应该具有以下特点**：

- 简单
- 直观
- 一致

这包括接口、方法和消息的名称。

## 1、软件包名称

API `.proto`文件中声明的软件包名称**应该**与产品名称和服务名称保持一致。软件包名称**应该**使用单数组件名称，以避免混合使用单数和复数组件名称。软件包名称**不能**使用下划线。进行版本控制的 API 的软件包名称**必须**以此版本结尾。例如：

```protobuf
// Kpanda Cluster API
package kpanda.io.api.cluster.v1alpha1;

option go_package = "kpanda.io/api/cluster/v1alpha1;v1alpha1";
```

## 2、接口名称

您可以将服务名称视为对一组 API 实际实现的引用，而接口名称则是 API 的抽象定义。

接口名称**应该**使用直观的名词，例如 Calendar 或 Blob。该名称**不得**与编程语言及其运行时库（如 File）中的成熟概念相冲突。

在极少数情况下，接口名称会与 API 中的其他名称相冲突，此时**应该**使用后缀（例如 `Api` 或 `Service`）来消除歧义。

```protobuf
// Library is the interface name.
service Library {
  rpc ListBooks(...) returns (...);
  rpc ...
}
```

## 3、方法名称

服务**可以**在其 IDL 规范中定义一个或多个远程过程调用 (RPC) 方法，这些方法需与集合和资源上的方法对应。方法名称**应**采用大驼峰式命名格式并遵循 `VerbNoun` 的命名惯例，其中 Noun（名词）通常是资源类型。

| 动词     | 名词   | 方法名称     | 请求消息            | 响应消息                |
| :------- | :----- | :----------- | :------------------ | :---------------------- |
| `List`   | `Book` | `ListBooks`  | `ListBooksRequest`  | `ListBooksResponse`     |
| `Get`    | `Book` | `GetBook`    | `GetBookRequest`    | `Book`                  |
| `Create` | `Book` | `CreateBook` | `CreateBookRequest` | `Book`                  |
| `Update` | `Book` | `UpdateBook` | `UpdateBookRequest` | `Book`                  |
| `Rename` | `Book` | `RenameBook` | `RenameBookRequest` | `RenameBookResponse`    |
| `Delete` | `Book` | `DeleteBook` | `DeleteBookRequest` | `google.protobuf.Empty` |

方法名称的动词部分**应该**使用用于要求或命令的[祈使语气](https://en.wikipedia.org/wiki/Imperative_mood#English)，而不是用于提问的陈述语气。

对于标准方法，方法名称的名词部分对于除 `List` 之外的所有方法**必须**采用单数形式，而对于 `List` **必须**采用复数形式。对于自定义方法，名词在适当情况下**可以**采用单数或复数形式。批处理方法**必须**采用复数名词形式。

**注意**：上面的情况指的是协议缓冲区中的 RPC 名称；HTTP/JSON URI 后缀使用 `:lowerCamelCase`。

如果关于 API 子资源的方法名称使用提问动词（经常使用陈述语气表示），则容易让人混淆。例如，要求 API 创建一本书，这显然是 `CreateBook`（祈使语气），但是询问 API 关于图书发行商的状态可能会使用陈述语气，例如 `IsBookPublisherApproved` 或 `NeedsPublisherApproval`。若要在此类情况下继续使用祈使语气，请使用“check”(`CheckBookPublisherApproved`) 和“validate”(`ValidateBookPublisher`) 等命令。

方法名称**不应**包含介词（例如“For”、“With”、“At”、“To”）。通常，带有介词的方法名称表示正在使用新方法，应将一个字段添加到现有方法中，或者该方法应使用不同的动词。

例如，如果 `CreateBook` 消息已存在且您正在考虑添加 `CreateBookFromDictation`，请考虑使用 `TranscribeBook` 方法。

## 4、自定义方法

自定义方法是指 5 个标准方法之外的 API 方法。这些方法**应该**仅用于标准方法不易表达的功能。通常情况下，API 设计者**应该**尽可能优先考虑使用标准方法，而不是自定义方法。标准方法具有大多数开发者熟悉的更简单且定义明确的语义，因此更易于使用且不易出错。另一项优势是 API 平台更加了解和支持标准方法，例如计费、错误处理、日志记录、监控。

自定义方法可以与资源、集合或服务关联。 它**可以**接受任意请求和返回任意响应，并且还支持流式请求和响应。

自定义方法名称**必须**遵循[方法命名惯例](https://cloud.google.com/apis/design/naming_convention#method_names)。

- ### 常用自定义方法

以下是常用或有用的自定义方法名称的精选列表。API 设计者在引入自己的名称之前**应该**考虑使用这些名称，以提高 API 之间的一致性。

| 方法名称   | 自定义动词  | HTTP 动词 | 备注                                                         |
| :--------- | :---------- | :-------- | :----------------------------------------------------------- |
| `Cancel`   | `:cancel`   | `POST`    | 取消一个未完成的操作，例如 [`operations.cancel`](https://github.com/googleapis/googleapis/blob/master/google/longrunning/operations.proto#L100)。 |
| `BatchGet` | `:batchGet` | `GET`     | 批量获取多个资源。如需了解详情，请参阅[列表描述](https://cloud.google.com/apis/design/standard_methods#list)。 |
| `Move`     | `:move`     | `POST`    | 将资源从一个父级移动到另一个父级，例如 [`folders.move`](https://cloud.google.com/resource-manager/reference/rest/v2/folders/move)。 |
| `Search`   | `:search`   | `GET`     | List 的替代方法，用于获取不符合 List 语义的数据，例如 [`services.search`](https://cloud.google.com/service-infrastructure/docs/service-consumer-management/reference/rest/v1/services/search)。 |
| `Undelete` | `:undelete` | `POST`    | 恢复之前删除的资源，例如 [`services.undelete`](https://cloud.google.com/service-infrastructure/docs/service-management/reference/rest/v1/services/undelete)。建议的保留期限为 30 天。 |

- ### 自定义命名规则：

使用 `:` 而不是 `/` 将自定义动词与资源名称分开以便支持任意路径。例如，恢复删除文件可以映射到 `POST /files/a/long/file/name:undelete`

选择 HTTP 映射时，**应**遵循以下准则：

- 自定义方法**应该**使用 HTTP `POST` 动词，因为该动词具有最灵活的语义，但作为替代 get 或 list 的方法（如有可能，**可以**使用 `GET`）除外。（详情请参阅第三条。）
- 自定义方法**不应该**使用 HTTP `PATCH`，但**可以**使用其他 HTTP 动词。在这种情况下，方法**必须**遵循该动词的标准 [HTTP 语义](https://tools.ietf.org/html/rfc2616#section-9)。
- 请注意，使用 HTTP `GET` 的自定义方法**必须**具有幂等性并且无负面影响。例如，在资源上实现特殊视图的自定义方法**应该**使用 HTTP `GET`。
- 接收与自定义方法关联的资源或集合的资源名称的请求消息字段**应该**映射到网址路径。
- 网址路径**必须**以包含冒号（后跟自定义动词）的后缀结尾。
- 如果用于自定义方法的 HTTP 动词允许 HTTP 请求正文（其适用于 `POST`、`PUT`、`PATCH` 或自定义 HTTP 动词），则此自定义方法的 HTTP 配置**必须**使用 `body: "*"` 子句，所有其他请求消息字段都**应**映射到 HTTP 请求正文。
- 如果用于自定义方法的 HTTP 动词不接受 HTTP 请求正文（`GET`、`DELETE`），则此方法的 HTTP 配置**不得**使用 `body` 子句，并且所有其他请求消息字段都**应**映射到网址查询参数。

路由设计举例：

```protobuf
// https://github.com/googleapis/googleapis/blob/master/google/api/serviceusage/v1/serviceusage.proto(参考文档地址)
// Enables multiple services on a project. The operation is atomic: if
// enabling any service fails, then the entire batch fails, and no state
// changes occur.
//
// Operation response type: `google.protobuf.Empty`
rpc BatchEnableServices(BatchEnableServicesRequest) returns (google.longrunning.Operation) {
  option (google.api.http) = {
    post: "/v1beta1/{parent=*/*}/services:batchEnable"
    body: "*"
  };
}

// https://github.com/googleapis/googleapis/blob/master/google/api/servicemanagement/v1/servicemanager.proto（参考文档地址）
rpc UndeleteService(UndeleteServiceRequest) returns (google.longrunning.Operation) {
  option (google.api.http) = {
  post: "/v1/services/{service_name}:undelete"
  };
}
```

## 5、proto中方法排序

proto service中如果含有标准方法和自定义方法和特殊方法，应该将标准方法排序在前面并且用换行符分割标准方法和自定义方法。

方法按照动作来排序（GET、POST、PUT、PATCH、DELETE等）

```protobuf
syntax = "proto3";
package kpanda.io.api.apps.v1alpha1;
option go_package = "kpanda.io/api/apps/v1alpha1;v1alpha1";

service Apps {
  rpc GetDeployment(GetDeploymentRequest) returns (Deployment); 
  rpc GetDeploymentJSON(GetDeploymentJSONRequest) returns (GetDeploymentJSONResponse); 
  rpc ListDeployments(ListDeploymentsRequest) returns (ListDeploymentsResponse); 
  
  rpc CreateDeployment(CreateDeploymentRequest) returns (Deployment); 
  rpc CreateDeploymentByJSON(CreateDeploymentByJSONRequest) returns (CreateDeploymentByJSONResponse); 
  rpc UpdateDeployment(UpdateDeploymentRequest) returns (Deployment); 
  rpc UpdateDeploymentByJSON(UpdateDeploymentRequest) returns (UpdateDeploymentByJSONResponse); 

  rpc DeleteDeployment(DeleteDeploymentRequest) returns (google.protobuf.Empty); 
  
  rpc RestartDeployment(RestartDeploymentRequest)returns (RestartDeploymentResponse);
  rpc PauseDeployment(PauseDeploymentRequest)returns (PauseDeploymentResponse);
  
  rpc ListReplicasetsByDeployment(ListReplicasetsByDeploymentRequest)returns (ListReplicasetsByDeploymentResponse);
  rpc ListEventsByDeployment(ListEventsByDeploymentRequest)returns (ListEventsByDeploymentResponse);
}
```

上面例子中，标准方法为GetDeployment、ListDeployments、CreateDeployment、UpdateDeployment、DeleteDeployment，自定义方法为RestartDeployment、PauseDeployment，以deployment找events和replicasets为特殊方法，写在最后。

## 6、消息名称

首先让我们看一个非常简单的例子。假设您想要定义一个搜索请求消息格式，其中每个搜索请求都有一个查询字符串、您感兴趣的结果的特定页面以及每个页面的大量结果。这里是用于定义消息类型的 proto 文件。

```protobuf
syntax = "proto3";

message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}
```

SearchRequest消息定义指定三个字段(名称/值对) ，每个字段表示希望包含在此类消息中的每一段数据。每个字段都有一个名称和一个类型。

消息名称应该准守UpperCamelCase （大写驼峰命名法）命名。

## 7、字段名称

在上面的示例中，所有字段都是标量类型: 两个整数(page _ number 和 result _ per _ page)和一个字符串(query)。但是，也可以为字段指定组合类型，包括枚举和其他消息类型。

`.proto` 文件中的字段定义**必须**使用 `lower_case_underscore_separated_names` (小写下划线分割命名法)格式。这些名称将映射到每种编程语言的生成代码中的原生命名惯例。

字段名称**不应**包含介词（例如“for”、“during”、“at”），例如：

- `reason_for_error` 应该改成 `error_reason`
- `cpu_usage_at_time_of_failure` 应该改成 `failure_time_cpu_usage`

字段名称**不应**使用后置形容词（名词后面的修饰符），例如：

- `items_collected` 应该改成 `collected_items`
- `objects_imported` 应该改成 `imported_objects`

API 中的重复字段**必须**使用正确的复数形式。这符合现有 Google API 的命名惯例和外部开发者的共同预期。

## 8、请求和响应消息

RPC 方法的请求和响应消息**应该**分别以带有后缀 `Request` 和 `Response` 的方法名称命名，除非方法请求或响应类型为以下类型：

- 一条空消息（使用 `google.protobuf.Empty`）、
- 一个资源类型，或
- 一个表示操作的资源

这通常适用于在标准方法 `Get`、`Create`、`Update` 或 `Delete` 中使用的请求或响应。

请求示例：

```protobuf
syntax = "proto3";
package kpanda.io.api.book.v1alpha1;
option go_package = "kpanda.io/api/book/v1alpha1;v1alpha1";
 
service BookService {
  rpc CreateBook(Book) returns (Book); // bad,不要这么做
  rpc CreateBook(CreateBookRequest) returns (CreateBookResponse); // bad,不要这么做

  rpc CreateBook(CreateBookRequest) returns (Book); // 正确
  rpc ListBooks(ListBooksRequest) returns (ListBooksResponse); // 这么做是可以的
  rpc GetBook(GetBookRequest) returns (Book); // 这么做是也是可以的
  rpc DeleteBook(DeleteBookRequest) returns (google.protobuf.Empty); // 这么做是也是可以的
}
```

## 9、添加注释

要向`.proto` 文件添加注释，请使用 c/c++-style `//`和`/* ... */ `语法。

以下是注释的范例：

```protobuf
// https://github.com/googleapis/googleapis/blob/master/google/api/serviceusage/v1/serviceusage.proto（参考文档地址）
// Request message for the `ListServices` method.
message ListServicesRequest {
  // Parent to search for services on.
  //
  // An example name would be:
  // `projects/123`
  // where `123` is the project number (not project ID).
  string parent = 1;

  // Requested size of the next page of data.
  // Requested page size cannot exceed 200.
  //  If not set, the default page size is 50.
  int32 page_size = 2;

  // Token identifying which result to start with, which is returned by a
  // previous list call.
  string page_token = 3;

  // Only list services that conform to the given filter.
  // The allowed filter strings are `state:ENABLED` and `state:DISABLED`.
  string filter = 4;
}
```

每一个注释后面请空出一行。

## 10、枚举

枚举类型**必须**使用`UpperCamelCase` 格式的名称。

枚举值**必须**使用 CAPITALIZED_NAMES_WITH_UNDERSCORES （全大写，用下划线分割）格式。每个枚举值**必须**以分号（而不是逗号）结尾。第一个值**应该**命名为 ENUM_TYPE_UNSPECIFIED，因为在枚举值未明确指定时系统会返回此值。

```protobuf
message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
  enum Corpus {
    CORPUS_UNSPECIFIED = 0;
    UNIVERSAL = 1;
    WEB = 2;
    IMAGES = 3;
    LOCAL = 4;
    NEWS = 5;
    PRODUCTS = 6;
    VIDEO = 7;
  }
  Corpus corpus = 4;
}
```

## 11、嵌套类型

如果某个消息里的字段可能会被别的消息引用，请去掉不必要的前缀，这两种是等价的:

```protobuf
message Cluster {
  enum Role {
    ROLE_UNSPECIFIED = 0;
    THIRD_PARTY = 1;
    SERVICE = 2;
  } 
}

enum ClusterRole {
  CLUSTER_ROLE_UNSPECIFIED = 0;
  THIRD_PARTY = 1;
  SERVICE = 2;
}
```

仅仅是使用者导入方式不同。

## 12、使用其他消息类型

```protobuf
message SearchResponse {
  message Result {
    string url = 1;
    string title = 2;
    repeated string snippets = 3;
  }
  repeated Result results = 1;
}
```

如果要在其父消息类型之外重用此消息类型，请将其称为_Parent_._Type_

```protobuf
message SomeOtherMessage {
  SearchResponse.Result result = 1;
}
```

## 13、驼峰式命名法

除**字段名称和枚举值**外，`.proto` 文件中的所有定义都**必须**使用由 [Google Java 样式](https://google.github.io/styleguide/javaguide.html#s5.3-camel-case)定义的 UpperCamelCase 格式的名称。

## 14、特殊字段

特殊缩写字段：例如`IP`,`IPs`,`CIDR` 等，无需改写成下划线形式。

原因：`podIP`如果改成了`pod_IP `,golang里面代码变量是`pod_IP`并且返回给前端也是`pod_IP`,`protobuf`针对这个下划线加大写不会自动把_给删除。后期需要针对这个找出解决办法。

## 15、其他参考文档

- https://developers.google.com/protocol-buffers/docs/proto3

- https://cloud.google.com/apis/design
- https://github.com/googleapis/googleapis
- https://github.com/etcd-io/etcd
- https://grpc-ecosystem.github.io/grpc-gateway
- https://github.com/dapr/dapr