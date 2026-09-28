# Python 版本与镜像选择

> **先选镜像，再检查目标节点 Python。** 更新 `spray-job` 镜像不会升级目标节点 Python；Python 版本符合要求也不等于完整集群生命周期已经验证。

## 两套镜像的要求

控制端是运行 Ansible 的 **`spray-job` 容器**；目标端是被管理的 **Kubernetes 节点，包括 control-plane 和 worker**。

以下对应当前源码的构建配置，镜像仓库为 `ghcr.io/kubean-io/spray-job`，`<version>` 表示 Kubean 版本。历史镜像以其实际依赖为准。

> 表中的 `2.21` 及对应 Ansible 版本会随锁定的 Kubespray 升级而变化，**以镜像实际内容为准**，本页可能滞后。用页末命令自查。

| 镜像 tag | Ansible / ansible-core | 控制端 Python 支持范围 | 目标端 Python 支持范围 |
| --- | --- | --- | --- |
| `<version>-ansible-core2.16`，别名 `<version>` | 9.13.0 / 2.16 | 3.10–3.12 | 2.7 或 3.6–3.12 |
| `<version>-ansible-core2.21`，别名 `<version>-modern` | 当前为 14.4.0 / 2.21 | 3.12–3.14 | 3.9–3.14 |

- 范围包含两端，是 ansible-core 的支持范围，不是镜像内实际安装的 Python 版本。
- core 版本 tag 与对应别名指向同一次构建的镜像。core 版本从镜像读取；未来升级时，准确 tag 会随之变化，`<version>-modern` 保持简短入口。
- `modern` 跟随锁定的 Kubespray；`latest` 也指向现代镜像，并非普通版本 tag 的别名。生产环境请选择明确版本或 digest。
- Ansible 9 / core 2.16 已 EOL；旧版引擎不能保证兼容新 Kubespray 的所有操作，例如使用 `mount_facts` 的 reset。

自查镜像内实际版本：`docker run --rm --entrypoint ansible <image> --version`。

## 主要系统的默认 Python 参考

以下为实机记录及补充信息，整理于 2026-09-28。补丁版本和命令别名可能随系统镜像、更新及配置变化，请以目标节点实际解释器为准；本表不是完整 OS 支持认证。

| 系统 | `python` | `python3` | 该 Python 3 是否满足 core 2.21 范围 |
| --- | --- | --- | --- |
| Ubuntu 22.04.4 LTS | 未找到命令 | 3.10.12 | 是 |
| Ubuntu 24.04 LTS | 未找到命令 | 3.12.3 | 是 |
| Rocky Linux 9.2 | 3.9.16 | 3.9.16 | 是 |
| Kylin V10（配置为 SP3） | 2.7.18 | 3.7.9 | **否** |
| Kylin V11（Swan25） | 未找到命令 | 3.11.6 | 是 |
| RHEL 9.2（Plow） | 3.9.16 | 3.9.16 | 是 |
| RHEL 8 | 未记录 | 3.6.8（默认版本补充） | **否** |

上述 Python 3 版本均在 core 2.16 的目标端支持范围内。`python` 命令缺失不代表无法使用 Ansible；关键是实际选中的解释器。

## 如何选择

- 目标节点已有 Python 3.9–3.14：可复用该解释器评估现代镜像，**不必强制安装 Python 3.12**。
- RHEL 8 / 上表 Kylin V10：使用现代镜像前需额外准备受支持的 Python，并通过 `ansible_python_interpreter` 指定实际路径；保留系统 Python。
- 新 Python 必须在执行依赖 Python 的预检查及 hooks 前准备好。离线环境还需提供匹配系统和架构的软件包及完整依赖；不能只更换控制端镜像。

依据：[Ansible 官方支持矩阵](https://docs.ansible.com/projects/ansible-core/devel/reference_appendices/release_and_maintenance.html#ansible-core-support-matrix)、[当前锁定的 Kubespray requirements](https://github.com/kubernetes-sigs/kubespray/blob/b33bb4786450395df3c499faa2b7880d4130c693/requirements.txt)。
