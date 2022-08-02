# Offline Usage

The `kubean` project can be divided into three functions, `generating offline package`
, `importing offline package to minio and registry` and `installing k8s`.

Some bash scripts depend on `skopeo`, `minio` and `docker registry` or `harbor`, so we may need install skeopeo, minio
server and client, docker registry or harbor correctly first.

## generating offline package

We can execute `cd artifacts && bash generate_offline_package.sh all` to generate the offline package where we can reach
the internet easily.

Or we can perform these operations step by step equally like the following:

* `cd artifacts`
* `bash generate_offline_package.sh createtemplist`
    * this command will clone `kubespray` project to current folder and generate `images.list` and `files.list` in
      folder `kubespray/contrib/offline/temp`
    * We can optionally modify `images.list` and `files.list` with the specific version which we want
* `bash generate_offline_package.sh createfiles`
    * this command will generate file `offline-files.tar.gz` in current folder
* `bash generate_offline_package.sh createimages`
    * this command will generate file `offline-images.tar.gz` in current folder

## using offline package

Keep `offline-files.tar.gz`, `offline-images.tar.gz` and related bash scripts in the same folder, such as
folder `artifacts`.

### importing images to registry

The script `manage_images.sh` will unzip the file `offline-images.tar.gz` in the current folder and transfer the local
image data to the remote registry.

```bash
Dest_TLS_Verify=false Dest_User="" Dest_Password="" bash manage_images.sh import IP:PORT
```

* replace `IP:PORT` with the real registry IP and Port
* populate `Dest_User` and `Dest_Password` if the registry has password auth
* let `Dest_TLS_Verify` false if the registry is running on unsafe http

### importing files to minio server

The script `manage_files.sh` will unzip the file `offline-files.tar.gz` in the current folder and transfer the local
binary files to the minio server.

```bash
Minio_User=xxxxx Minio_Password=xxxxx bash manage_files.sh import http://IP:PORT
```

* replace `http://IP:PORT` with the real minio API Server Address.
* populate `Minio_User` and `Minio_Password` correctly.

## offline docker-ce linux repo

todo

## installing k8s

The files in `offlineDemo` are existing templates.

First we need replace `REGISTRY_HOST:REGISTRY_PORT` and `http://MINIO_API_HOST:MINIO_API_PORT` in
file `offlineDemo/vars-conf-cm.yml` with the target registry and the minio API server address.

And then we need update `offlineDemo/vars-conf-cm.yml` with adding the existing host ip where we want to install k8s
cluster.

Finally, perform `kubectl apply -f offlineDemo` to start the cluster.yml playbook to install k8s clusters. 
