# Offline Usage

> English | [中文](zh/offline.md)

The `kubean` project can be divided into three functions, `generating offline package`
, `importing offline package to minio and registry` and `installing k8s`.

Some bash scripts depend on `skopeo`, `minio` and `docker registry` or `harbor`, so we may need install skeopeo, minio
server and client, docker registry or harbor correctly first.

## generating offline package

We can execute `cd artifacts && bash generate_offline_package.sh all` to generate the offline package where we can reach
the internet easily.

Or we can perform these operations step by step equally like the following:

* `cd artifacts`
* `bash generate_offline_package.sh list`
    * this command will clone `kubespray` project to current folder and generate `images.list` and `files.list` in
      folder `kubespray/contrib/offline/temp`
    * We can optionally modify `images.list` and `files.list` with the specific version which we want
* `bash generate_offline_package.sh files`
    * this command will generate file `offline-files.tar.gz` in current folder
* `bash generate_offline_package.sh images`
    * this command will generate file `offline-images.tar.gz` in current folder

## using offline package

Keep `offline-files.tar.gz`, `offline-images.tar.gz` and related bash scripts in the same folder, such as
folder `artifacts`.

### importing images to registry

The script `import_images.sh` will unzip the file `offline-images.tar.gz` in the current folder and transfer the local
image data to the remote registry.

```bash
DEST_TLS_VERIFY=false DEST_USER="" DEST_PASS="" bash import_images.sh IP:PORT
```

* replace `IP:PORT` with the real registry IP and Port
* populate `DEST_USER` and `DEST_PASS` if the registry has password auth
* let `DEST_TLS_VERIFY` false if the registry is running on unsafe http

### importing files to minio server

The script `import_files.sh` will unzip the file `offline-files.tar.gz` in the current folder and transfer the local
binary files to the minio server.

```bash
MINIO_USER=xxxxx MINIO_PASS=xxxxx bash import_files.sh http://IP:PORT
```

* replace `http://IP:PORT` with the real minio API Server Address.
* populate `MINIO_USER` and `MINIO_PASS` correctly.

## offline docker-ce linux repo

todo

## installing k8s

The files in `offlineDemo` are existing templates.

First we need replace `REGISTRY_HOST:REGISTRY_PORT` and `http://MINIO_API_HOST:MINIO_API_PORT` in
file `offlineDemo/vars-conf-cm.yml` with the target registry and the minio API server address.

And then we need update `offlineDemo/vars-conf-cm.yml` with adding the existing host ip where we want to install k8s
cluster.

Finally, perform `kubectl apply -f offlineDemo` to start the cluster.yml playbook to install k8s clusters. 
