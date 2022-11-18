# Generation and use of incremental offline packages

> English | [中文](../zh/airgap_patch_usage.md)

To meet users' needs for components of certain versions, Kubean provides the script `artifacts/offline_patch.py` to generate a corresponding version of offline packages based on the configuration file `manifest.yml`.

## Generate incremental offline packages

1. Create a new `manifest.yml` file in a folder, with the following example:

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

2. Create a new `data` folder in the same folder

3. Run the following command to generate an incremental offline package in the `data` folder

    ```bash
    docker run -v $(pwd)/manifest.yml:/manifest.yml -v $(pwd)/data:/data ghcr.io/hangscer8/airgap-patch:v0.2.0
    ```

## Use the incremental offline package:

The directory structure of the incremental package is as follows:

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

1. Write file data into MinIO

    ``` bash
    $ cd data/v_offline_patch/amd64/files

    $ MINIO_USER=${username} MINIO_PASS=${password} ./import_files.sh ${minio_address}
    ```

    * `minio_address` is the `minio API Server` address, typically on port 9000, for example: `http://1.2.3.4:9000`.

2. Write image data to the docker registry (version 2.6.2 recommended) or harbor

    ``` bash
    $ cd data/v_offline_patch/amd64/images 

    # 1. Non-secure password-free mode
    $ DEST_TLS_VERIFY=false ./import_images.sh ${registry_address}

    # 2. Username password mode
    $ DEST_USER=${username} DEST_PASS=${password} ./import_images.sh ${registry_address}
    ```

    * When `DEST_TLS_VERIFY=false`, the image is uploaded in non-secure HTTP mode.
    * `DEST_USER` and `DEST_PASS` need to be set when username and password authentication exists for the image registry.
    * `registry_address` is the address of the image registry, e.g. `1.2.3.4:5000`.

3. Write `kubeanofflineversion.cr.patch.yaml` to the k8s cluster

    ```bash
    $ cd data/v_offline_patch
    $ kubectl apply -f kubeanofflineversion.cr.patch.yaml 
    ```

    * This step is to inform the kubean-operator of the new software version available for offline use.
