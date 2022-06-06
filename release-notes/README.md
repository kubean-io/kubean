# Kpanda Release Notes

此目录包含 Kpanda 相关的 release notes 的介绍。如果 PR 有任何关于用户的影响，都应该创建 release notes。


## 何时需要 release notes

当有如下更改时，应当添加 release notes，包括：
* CLI 更改
* API 更改
* 配置架构更改
* 非功能属性的变化，例如时间效率或可用性、新平台的可用性等更改
* 关于弃用的警告
* 修复了之前的已知问题
* 完成了一项新的功能

有如下情况时可以选择性提供 release notes，包括：
* 测试
* 构建基础设施
* 修复了尚未发布的错误

## 添加 release notes

要创建 release notes，请根据 [template](./template.yaml) 中的内容、在 [./notes](./notes) 路径下创建一个新的文件。
文件名无关紧要，但是文件的后缀需要是 yaml 文件（`.yaml`），请尽量让文件名具有描述性。关于每个字段的含义请看 [template](./template.yaml)

```yaml
kind: feature
area: work-api

issues:
  - 2100
  - https://gitlab.daocloud.cn/ndx/engineering/kubean/-/issues3

jiras:
 - https://jira.daocloud.io/browse/DCE-1252
 - 1253

releaseNotes:
- |
  **新增** `work-api` 中的 `v3/govern/namespaces` 接口。
- |
  **修复** `work-api` 中的 `v3/govern/namespaces` 接口。

securityNotes:
- |
  __[CVE-2020-15104](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-15104)__：
  在验证 TLS 证书时，Envoy 错误地允许通配符 DNS 主题备用名称应用于多个子域。例如，当 SAN 为 `*.example.com` 时，Envoy 错误地允许`nested.subdomain.example.com`，而它应该只允许`subdomain.example.com`。
    - CVSS 得分：6.6 [AV:N/AC:H/PR:H/UI:N/S:C/C:H/I:L/A:N/E:F/RL:O/RC:C] （https://nvd.nist.gov/vuln-metrics/cvss/v3-calculator?vector=AV:N/AC:H/PR:H/UI:N/S:C/C:H/I:L /A:N/E:F/RL:O/RC:C&version=3.1)
```

### Area

该字段描述了影响的 Kpanda 项目的领域，有效值包括:
* api
* api-service
* work-api
* infrastructure
* controller （跟 CRDs 相关）
* external (跟外部的一些组件相关，如insight)
* installation
* cli
* documentation

### Issue

虽然许多 PR 可能只修复了一个 Gitlab 问题，但有时候一个 PR 会修复多个问题，请把所有修复的问题都列在 issues 中。

### Jira

虽然许多 PR 可能只修复了一个 Jira 问题，但有时候一个 PR 会修复多个问题，请把所有修复的 Jira 问题都列在 jiras 中。

### Release Notes

这项内容详细说明了错误修复、功能添加、删除或其它对用户有影响的改动，第一个单词应该是一个动作，格式为 `**动作**`。可接受的动作为：
- `**新增**`
- `**弃用**`
- `**修复**`
- `**优化**`
- `**改进**`
- `**移除**`
- `**升级**`

## 向多个 Kpanda 版本添加 release notes

就像代码修复应该首先添加到 master 一样，release notes 也是一样。要将 notes 添加到多个版本，只需将它们 cherry-pick 到适当的版本中，release notes 生成工具就会将它们包含在其生成中。

### Release 步骤
详细步骤请参考：[Release Notes 生成流程](https://dwiki.daocloud.io/pages/viewpage.action?pageId=118569714)
