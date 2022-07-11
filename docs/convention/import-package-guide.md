## 介绍

go 的代码被组织在包中，包的客户端要使用包的函数、常量和变量，都必须进行显式申明。包的名称为其内容提供了上下文，使客户更容易理解包的用途以及如何使用它。命名良好的包名可以让客户更容易找到所需的代码，也可以让客户更好的使用代码。

本文档记录了 `kpanda` 项目的 import 包的一些最佳实践，它遵循了 `google`, [uber](https://github.com/xxjwxc/uber_go_guide_cn) 以及 [golang](https://go.dev/blog/package-names) 提出的规范。开发人员在编写代码时，请遵守此规范。

## 包名

包名的设计在开发的过程中时非常重要的，它在一定程度上减轻了开发的复杂性。所以我们在创建一个新的包的时候请遵循下面的规范：

* 全部小写，没有大写或下划线，它也并不遵循`驼峰规则`。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
"k8s.io/clusterpedia-io/client-go/client_cluster"
```

</td><td>

```go
"k8s.io/clusterpedia-io/client-go/client"
```

</td></tr>
</tbody></table>


* 包名尽量不要跟广泛使用的常量名重复。e.g: 缓冲 I/O 包被称为bufio，而不是buf，因为buf 它是一个很好的缓冲区变量名称。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
"github.com/daocloud/kpanda/pkg/buf"
```

</td><td>

```go
"github.com/daocloud/kpanda/pkg/bufio"
```

</td></tr>
</tbody></table>


* 尽量不要使用 "common"，"util"，"shared" 或 "lib" 信息量不足的名称。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
"github.com/kpanda/common"
```

</td><td>

```go
"github.com/kpanda/pkg/util/common"
```

</td></tr>
</tbody></table>

* 尽量不要使用复数。e.g: net/url，而不是net/urls。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
"k8s.io/clusterpedia-io/client-go/util/certs"
```

</td><td>

```go
"k8s.io/clusterpedia-io/client-go/util/cert"
```

</td></tr>
</tbody></table>

* 不要把所有的 api 放在一个包，可以根据不同的功能设计不同的包。

* 明智的缩写，程序员都熟悉的名称可以缩写。广泛使用的包通常具有压缩名称：
  - strconv（字符串转换）
  - syscall（系统调用）
  - fmt（格式化的 I/O）


## 别名

如果程序包名称与导入路径的最后一个元素不匹配，则必须使用导入别名。虽然我们遵循了一定的规范，但是为了避免在同一个项目中一个包出现多种别名，我们应该在先参考其他的相同包的别名，再去定义。

```go
import (
  "net/http"

  client "example.com/client-go"
  trace "example.com/trace/v2"
)
```

当命名包时，请遵循下面的规范：

* 在所有其他情况下，除非导入之间有直接冲突，否则应避免导入别名。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
import (
  "fmt"
  "os"

  nettrace "golang.net/x/trace"
)
```

</td><td>

```go
import (
  "fmt"
  "os"
  "runtime/trace"

  nettrace "golang.net/x/trace"
)
```

</td></tr>
</tbody></table>


* 别名应该全部小写，没有大写或下划线。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
coreV1 "k8s.io/api/core/v1"
meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
```

</td><td>

```go
corev1 "k8s.io/api/core/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
```

</td></tr>
</tbody></table>

* 尽量简短而清晰。另一方面，如果缩写包名称会使其模棱两可或不清楚的时候，请不要这样做。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
kubernetesfakeclientset "k8s.io/client-go/kubernetes/fake"
pdynamic "github.com/clusterpedia-io/client-go/dynamic"
```

</td><td>

```go
fakeClient "k8s.io/client-go/kubernetes/fake"
pediadynamic "github.com/clusterpedia-io/client-go/dynamic"
```

</td></tr>
</tbody></table>

## 分组

所有的代码都必须使用 `gofmt` 进行格式化，可选的工具还包括 `goimports` (可看作gofmt的超集), 对 `imports` 进行分类排序，`goimports` 可以帮你完成这个操作。除此之外，还应该遵下下面的规则：

* 如果有超过1个包，必须使用数组的方式进行申明。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
import "a"
import "b"
```

</td><td>

```go
import (
  "a"
  "b"
)
```

</td></tr>
</tbody></table>


* 对 import 分组，导入应该按照 `标准库`, `项目中的库`, `第三方库` 分为三组。在 kpanda 的 ci 中会对代码进行静态检查，所以在提交代码之前，请先执行 `make test-staticcheck` 进行预检查。

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
import (
  "fmt"
  "os"
  "go.uber.org/atomic"
  "golang.org/x/sync/errgroup"
)

---

import (
	"os"
	"os/user"
	"path"

	"github.com/daocloud/kpanda/pkg/util/reflectutils"

	"k8s.io/client-go/util/homedir"

	"github.com/spf13/pflag"
)
```

</td><td>

```go
import (
  "fmt"
  "os"

  "go.uber.org/atomic"
  "golang.org/x/sync/errgroup"
)

---

import (
   "os"
   "os/user"
   "path"
   
   "github.com/daocloud/kpanda/pkg/util/reflectutils"

   "github.com/spf13/pflag"
   "k8s.io/client-go/util/homedir"
)
```

</td></tr>
</tbody></table>


## 参考

下面是 `golang` 和 `uber` 编码规范文档，不仅仅只是关于引包的规范，还有很多编码方面的文档可以查阅：

![](https://go.dev/blog/package-names)
[](https://github.com/xxjwxc/uber_go_guide_cn)
