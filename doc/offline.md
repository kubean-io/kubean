# Offline Usage

The `kubean` project can be divided into three functions, `generating offline package`
, `importing offline package to minio and registry` and `installing k8s`.

Some bash scripts depend on skopeo, minio and docker registry or harbor, so we may need install skeopeo, minio server
and client, docker registry or harbor correctly first.

## generating offline package

We can execute `cd artifacts && bash generate_offline_package.sh all` to generate the offline package where we can reach
the internet easily.

Or we can perform these operations step by step equally like the following:

* `cd artifacts`
* `bash generate_offline_package.sh createtemplist`
    * this command will clone `kubespray` project to current folder and generate `images.list` and `files.list` in
      folder `kubespray/contrib/offline/temp`.
    * We can modify `images.list` and `files.list` with the specific version which we want optionally.
* `bash generate_offline_package.sh createfiles`
* `bash generate_offline_package.sh createimages`
* `bash generate_offline_package.sh createtar`

## importing offline package to minio and registry

todo
```
Dest_TLS_Verify=false Dest_User="" Dest_Password="" bash manage_images.sh import IP:PORT 

Minio_User=admin Minio_Password=password bash manage_files.sh import http://IP:PORT

```

## installing k8s

todo
``