# 增量离线包的生成和使用

> [English](../en/airgap_patch_usage.md) | 中文

为了满足用户对于某些软件特定版本的需要，kubean提供脚本 `artifacts/offline_patch.py` 用来根据配置文件 `manifest.yml`
来生成对应版本的离线包。

## 生成增量离线包

1. 在当前文件夹新建 `manifest.yml` 文件，内容举例如下:

    ```yaml
    image_arch:
      - "amd64"
      - "arm64"
    kube_version:
      - "v1.24.6"
      - "v1.24.4"
    calico_version:
      - "v3.23.3"
    cni_version:
      - "v1.1.1"
    containerd_version:
      - "1.6.8"
    cilium_version:
      - "v1.12.1"
    etcd_version:
      - "v3.5.3"
    ```

2. 在当前文件夹新建 `data` 文件夹

3. 使用镜像，等待运行退出后，`data` 文件夹中生成增量离线包

    ```bash
    docker run -v $(pwd)/manifest.yml:/manifest.yml -v $(pwd)/data:/data ghcr.io/kubean-io/airgap-patch:v0.4.0-rc5 
    ```

## 使用增量离线包

增量包的目录结构如下:

```
data
└── v_offline_patch
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
    └── kubeanofflineversion.cr.patch.yaml
```

1. 向 MinIO 中写入文件数据

    ``` bash
    $ cd data/v_offline_patch/amd64/files

    $ MINIO_USER=${username} MINIO_PASS=${password} ./import_files.sh ${minio_address}
    ```

    * `minio_address` 是 `minio API Server` 地址，端口一般为9000，比如 `http://1.2.3.4:9000`。

2. 向 docker registry (版本推荐使用2.6.2) 或者 harbor 写入镜像数据

    ``` bash
    $ cd data/v_offline_patch/amd64/images 

    # 1. 非安全免密模式
    $ DEST_TLS_VERIFY=false ./import_images.sh ${registry_address}

    # 2. 用户名口令模式
    $ DEST_USER=${username} DEST_PASS=${password} ./import_images.sh ${registry_address}
    ```

    * 当 `DEST_TLS_VERIFY=false`, 此时采用非安全 HTTP 模式上传镜像。
    * 当镜像仓库存在用户名密码验证时，需要设置 `DEST_USER` 和 `DEST_PASS`。
    * `registry_address` 是镜像仓库的地址，比如`1.2.3.4:5000`。

3. 将 `kubeanofflineversion.cr.patch.yaml` 写入到 k8s 集群

    ```bash
    $ cd data/v_offline_patch
    $ kubectl apply -f kubeanofflineversion.cr.patch.yaml 
    ```

    * 该步骤是为了将新的可离线使用的软件版本信息告知 kubean-operator。
