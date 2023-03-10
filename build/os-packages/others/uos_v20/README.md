### How to build an os package based on unionTech server v20

1. Upload the following two files to the unionTech server v20 operating system:
* [`artifacts/import_ospkgs.sh`](https://github.com/kubean-io/kubean/blob/main/artifacts/import_ospkgs.sh)
* [`build/os-packages/others/uos_v20/build.sh`](https://github.com/kubean-io/kubean/blob/main/build/os-packages/others/uos_v20/build.sh)

2. Execute the following command to build the os package:

    ``` bash
    $ pwd
    /home/test

    $ ls
    build.sh  import_ospkgs.sh

    $ IMPORT_SH=/home/test/import_ospkgs.sh bash build.sh
    ...

    $ ls
    build.sh  import_ospkgs.sh  os-pkgs-uniontech-20.tar.gz
    ```
    Note:
    * The `IMPORT_SH` environment variable is used to specify the path to the `import_ospkgs.sh` file
    * After the build command is executed, the os package `os-pkgs-uniontech-20.tar.gz` will be generated in the current directory

### Resolve missing python dependencies in unionTech server v20 (1020a) minimization system

1. Execute the following command on the `unionTech server v20` OS to generate the rpm package for python3.6:
    
    ``` bash
    $ dnf install -y --downloadonly --downloaddir=rpm/ python36
    ...

    $ ls -lh rpm/
    total 204K
    -rw-r--r-- 1 root root  19K Mar 10 15:25 python3-pip-9.0.3-18.uelc20.01.noarch.rpm
    -rw-r--r-- 1 root root 162K Mar 10 15:25 python3-setuptools-39.2.0-7.uelc20.2.noarch.rpm
    -rw-r--r-- 1 root root  18K Mar 10 15:25 python36-3.6.8-2.module+uelc20+36+6174170c.x86_64.rpm
    ```

2. Upload and install the rpm package to a `unionTech server v20` OS node that is missing python3.6：
    
    ``` bash
    rpm -ivh python3-pip-9.0.3-18.uelc20.01.noarch.rpm python3-setuptools-39.2.0-7.uelc20.2.noarch.rpm python36-3.6.8-2.module+uelc20+36+6174170c.x86_64.rpm
    ```
