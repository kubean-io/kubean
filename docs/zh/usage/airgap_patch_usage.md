# 增量离线包的生成和使用

为了满足用户对于某些软件特定版本的需要，Kubean 提供脚本 `artifacts/airgap_patch.py` 根据配置文件 `manifest.yml`
来生成对应版本的离线包。

## 生成增量离线包

1. 在当前文件夹新建 `manifest.yml` 文件，内容举例如下：

    ```yaml
    image_arch:
      - "amd64"
      - "arm64"
    kube_version:
      - "1.24.6"
      - "1.24.4"
    calico_version:
      - "3.23.3"
    cni_version:
      - "1.1.1"
    containerd_version:
      - "1.6.8"
    cilium_version:
      - "1.12.1"
    ```

2. 在当前文件夹新建 `data` 文件夹

3. 使用镜像，等待运行退出后，在 `data` 文件夹中生成增量离线包

    ```bash
    docker run \
        -v $(pwd)/data:/data \
        -v $(pwd)/manifest.yml:/manifest.yml \
        -e ZONE=CN \
        ghcr.io/kubean-io/airgap-patch:v0.11.1
    ```

    | 环境变量 | 可选值描述 （:material-checkbox-marked-circle: :表示默认值） |
    | ----------- | ------------------------------------ |
    | ZONE | :material-checkbox-marked-circle: `DEFAULT`: 采用默认原始地址下载离线资源  |
    |      | :material-checkbox-blank-circle-outline: `CN`: 采用国内 DaoCloud 加速器地址下载离线资源 |
    | MODE | :material-checkbox-marked-circle: `INCR`: 仅构建配置中指定组件的离线资源（即：增量包）|
    |      | :material-checkbox-blank-circle-outline:  `FULL`: 将构建配置中指定的组件以及集群部署必要其他组件的离线资源（即：全量包）|

## 使用增量离线包

增量包的目录结构如下:

```
data
└── airgap_patch
    ├── amd64
    │   ├── files
    │   │   ├── import_files.sh
    │   │   └── offline-files.tar.gz
    │   └── images
    │       ├── import_images.sh
    │       └── offline-images.tar.gz
    ├── arm64
    │   ├── files
    │   │   ├── import_files.sh
    │   │   └── offline-files.tar.gz
    │   └── images
    │       ├── import_images.sh
    │       └── offline-images.tar.gz
    └── localartifactset.cr.yaml
```

1. 向 MinIO 中写入文件数据

    ```bash
    cd data/airgap_patch/amd64/files
   
    MINIO_USER=${username} MINIO_PASS=${password} ./import_files.sh ${minio_address}
    ```

    `minio_address` 是 `minio API Server` 地址，端口一般为 9000，比如 `http://1.2.3.4:9000`。

2. 向 Docker Registry（推荐使用 2.6.2 版本）或者 Harbor 写入镜像数据

    ```bash
    cd data/airgap_patch/amd64/images

    # 1. 免密模式
    REGISTRY_SCHEME=http REGISTRY_ADDR=${registry_address} ./import_images.sh

    # 2. 用户名口令模式
    REGISTRY_SCHEME=https REGISTRY_ADDR=${registry_address} REGISTRY_USER=${username} REGISTRY_PASS=${password} ./import_images.sh
    ```

    * `REGISTRY_ADDR` 是镜像仓库的地址，比如`1.2.3.4:5000`
    * 当镜像仓库存在用户名密码验证时，需要设置 `REGISTRY_USER` 和 `REGISTRY_PASS`

3. 将 `localartifactset.cr.yaml` 写入到 K8s 集群

    ```bash
    cd data/airgap_patch
    kubectl apply -f localartifactset.cr.yaml
    ```

    > 这一步是为了将新的可离线使用的软件版本信息告知 kubean-operator。
