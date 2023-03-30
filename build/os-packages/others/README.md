# OS package support for other linux

Please make sure these tools are installed in your OS environment:
* curl
* sshpass

## 1. How to download the tool script

``` bash
$ cd /home
$ curl -Lo ./pkgs.yml https://raw.githubusercontent.com/kubean-io/kubean/main/build/os-packages/others/pkgs.yml
$ curl -Lo ./other_os_pkgs.sh https://raw.githubusercontent.com/kubean-io/kubean/main/build/os-packages/others/other_os_pkgs.sh && chmod +x other_os_pkgs.sh
```

## 2. How to build an OS package

Since the build process will download packages, you need to make sure that the network in other linux environments can be accessed properly.

``` bash
$ cd /home
$ export PKGS_YML_PATH=/home/pkgs.yml
$ ./other_os_pkgs.sh build
```

## 3. How to install the OS package

Prepare the OS package tarball file in advance.

``` bash
$ export PKGS_YML_PATH=/home/pkgs.yml
$ export PKGS_TAR_PATH=/home/os-pkgs.tar.gz
$ export SSH_USER=root
$ export SSH_PASS=dangerous
$ export HOST_IPS='192.168.10.11 192.168.10.12'
$ ./other_os_pkgs.sh install
```
