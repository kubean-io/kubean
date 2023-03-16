
# Theory of Air-gapped packages


Kubean CI builds offline assets, to install Kubernetes in air-gapped environment.

NOTE: About usage for the offline assets, please refer to [offline.md(../en/offline.md)

This chapter explain the theory how the air-gapped packages come from.


- Assets :   [Github Releases](https://github.com/kubean-io/kubean/releases)
- Major contents:

  |  asset | description  |
  |  -----------------------  | :---------------------  |
  | files-*.tar.gz  | the binaries required in kubespray installation: example : kubeadm, runc |
  | images-*.tar.gz  | the k8s cluster images as well as CNI images  |
  | os-pkgs-${linux_distribution}-${tag}.tar.gz | deb/rpm required during k8s installion  |


## How to build the assets


(1) images & binaries

As state in Kubespray [offline deployment guide](https://github.com/kubernetes-sigs/kubespray/blob/master/contrib/offline/README.md),  Kubespray already provides scripts  to generate images and binaries list. (Thanks to great Kubespray! )

With help of kubespray [script to generate binaries & images list](https://github.com/kubernetes-sigs/kubespray/blob/master/contrib/offline/generate_list.sh), 
 then we can use [manage-offline-files.sh](https://github.com/kubernetes-sigs/kubespray/tree/master/contrib/offline#manage-offline-files.sh) to download those binaries and images.
 At last, Kubean provides an [offline-build.sh](https://github.com/kubean-io/kubean/blob/main/.github/workflows/call-offline-build.yaml) to make them together.
 
(2) os-packages (deb/rpm)

During the k8s installation, a few packages could not be installed as binaries, so we have to install them via deb/rpm. The [os packages list](https://github.com/kubean-io/kubean/blob/main/build/os-packages/packages.yml) defines what packages will be involved. 

Github Action will build packages in different OS (as Qemu) , via `dnf/apt` to download and archieve RPM/DEB packages.

(3) CI process

The offline assets are generated/managed by [Github Action scripts](https://github.com/kubean-io/kubean/tree/main/.github/workflows)

