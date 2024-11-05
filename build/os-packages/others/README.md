# OS package support for other linux

Please make sure these tools are installed in your OS environment:
* curl
* sshpass
* tar

## 1. How to download the tool script

``` bash
$ cd /home
$ curl -Lo ./other_os_pkgs.sh https://raw.githubusercontent.com/kubean-io/kubean/main/build/os-packages/others/other_os_pkgs.sh && chmod +x other_os_pkgs.sh
```

## 2. How to build an OS package

Since the build process will download packages, you need to make sure that the network in other linux environments can be accessed properly.

``` bash
$ cd /home
$ ./other_os_pkgs.sh build
```

## 3. How to install the OS package

Prepare the OS package tarball file in advance.

``` bash
$ export PKGS_TAR_PATH=/home/os-pkgs.tar.gz
$ export HOST_IPS='192.168.10.11 192.168.10.12'

# username/password authentication
$ export SSH_USER=root
$ export SSH_CRED=dangerous
$ ./other_os_pkgs.sh install

# public/private key authentication
$ export SSH_MODE=KEY
$ export SSH_USER=root
$ ./other_os_pkgs.sh install

# public/private key authentication (specify the private key path)
$ export SSH_MODE=KEY
$ export SSH_USER=root
$ export SSH_CRED=/home/ssh/id_rsa
$ ./other_os_pkgs.sh install
```
