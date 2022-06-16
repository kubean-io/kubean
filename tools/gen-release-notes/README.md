# 自动生成 ReleaseNotes

该文件夹包含了自动生成 ReleaseNotes 的相关功能。

## 指令介绍

可以通过运行 main.go 的方式，指定 oldBranch, newBranch, oldRelease, newBranch 的值来确定生成的 ReleaseNotes 的范围。 同时，需指定 notes
存放的位置，应在 `/release-notes/notes` 文件夹下。如需修改生成的模板，可在 `./templates` 中进行修改， 或生成完后自行修改。

样例指令如下：

```bash
go run . --oldRelease v0.2.0 --newRelease v0.2.1 --notes ../../ --outDir ../../changes
```

### 参数

* (optional) `--notes`  -- ReleaseNotes 存放的文件夹
* (optional) `--templates` -- 生成的 ReleaseNotes 的样例模板的存放位置。默认为: `./templates`
* (optional) `--validateOnly` -- 只验证但不生成 ReleaseNotes。
* `--oldRelease` -- 老的 release 的 tag
* `--newRelease` -- 新的 release 的名称

### gitlab CI说明

- 通过在 `.gitpab-ci.yml` 中添加 `Release Version` 步骤即可,祥看`.gitpab-ci.yml`文件内容

#### 如何通过gitlab创建一个release

- 进入项目repo的pipe地址：[kubean](https://gitlab.daocloud.cn/ndx/mspider/-/pipelines)
- 点击右侧`Run Pipe`新建一个流水线
- 分支选择`main`
- 在下方添加变量`PRE_VERSION`和`NEXT_VERSION`
  - `PRE_VERSION`: 之前存在的某个tag版本号，必须存在(表示发布的一个范围)
  - `NEXT_VERSION`: 下一个release的版本号
  - version.json 文件中的kubean版本号：表示此次release的版本号,
    在CI流水线完成之后，会根据此版本号打一个新的tag作为发布的版本，同时会将version.json中的kubean版本号更新为`NEXT_VERSION`
  - 版本关系为： `PRE_VERSION` < `version.json中版本号` < `NEXT_VERSION`
- 在CI完成之后在项目的/changes文件夹下生成ReleaseNotes
- 在需要发布时，进入项目的repo, 点击左侧`Deployment` -> `Release` -> 右上角创建发布
  - 发布的tag选择最新的（tag应该是在Release之前的version.json中的版本）
  - 填写发布的title
  - 发布内容选择 /changes目录下对应版本的md文件， 注意需要审核此文件内容，将一些不必要的内容给删除
  - 点击发布即可在Repo的 `Depolyment` -> `Release` 中看到发布的内容

### 实现原理

- 通过判断目录 `release-notes/notes` 目录下的yaml文件是否发生变更（例如新发一个版本可以建立一个Vx.y.z 的文件夹，将需要发布的内容写在这些yml文件中）
- 具体的yml文件格式即说明请看 `release-notes/template.yaml` 中的注释
- 判断出哪些yml文件变化，提取出变化的内容，替换到 `tools/gen-release-notes/templates`的模板中，将结果输出到`changes`目录中

### 目录说明

```
  /
    changes/               存放release 生成的md文档
    release-notes/   
      notes/               每个release发布的内容
    hack/               CI过程中使用的脚本
    tools/
      gen-release-notes/   生成release notes的源代码
    version.json           包含kubean版本号
```
