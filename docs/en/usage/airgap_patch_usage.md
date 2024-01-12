# Generation and use of incremental offline packages

To meet users' needs for components of certain versions, Kubean provides the script `artifacts/airgap_patch.py` to generate a corresponding version of offline packages based on the configuration file `manifest.yml`.

## Generate an incremental offline package

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
    $ docker run \
          -v $(pwd)/data:/data \
          -v $(pwd)/manifest.yml:/manifest.yml \
          ghcr.io/kubean-io/airgap-patch:v0.11.1
    ```

    | Environment Variables | Optional Value Description （:material-checkbox-marked-circle: is default value） |
    | ----------- | ------------------------------------ |
    | ZONE | :material-checkbox-marked-circle: `DEFAULT`: Download offline resources using the default original address.  |
    |      | :material-checkbox-blank-circle-outline: `CN`: Download offline resources by using DaoCloud mirror address in China. |
    | MODE | :material-checkbox-marked-circle: `INCR`: Build only the offline resources for the components specified in the configuration (i.e.: incremental packages)|
    |      | :material-checkbox-blank-circle-outline:  `FULL`: Building offline resources includes the components specified in the configuration along with the components necessary for cluster deployment (i.e.: full packages)|


## Use the incremental offline package

The directory structure of the incremental package is as follows:

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

1. Write file data into MinIO

    ```bash
    $ cd data/airgap_patch/amd64/files

    $ MINIO_USER=${username} MINIO_PASS=${password} ./import_files.sh ${minio_address}
    ```

    `minio_address` is the `minio API Server` address, typically on port 9000, for example: `http://1.2.3.4:9000`.

2. Write image data to the docker registry (recommended version 2.6.2) or harbor

    ```bash
    $ cd data/airgap_patch/amd64/images 

    # 1. password-free mode
    $ REGISTRY_SCHEME=http REGISTRY_ADDR=${registry_address} ./import_images.sh

    # 2. Username password mode
    $ REGISTRY_SCHEME=https REGISTRY_ADDR=${registry_address} REGISTRY_USER=${username} REGISTRY_PASS=${password} ./import_images.sh
    ```

    * `REGISTRY_ADDR` is the address of the mirror repository, e.g. `1.2.3.4:5000`
    * `REGISTRY_USER` and `REGISTRY_PASS` need to be set when username and password authentication exists for the mirror repository

3. Write `localartifactset.cr.yaml` to the k8s cluster

    ```bash
    $ cd data/airgap_patch
    $ kubectl apply -f localartifactset.cr.yaml
    ```

    > This step is to inform the kubean-operator of the new software version available for offline use.
