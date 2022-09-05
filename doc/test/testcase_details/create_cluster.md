#  Create cluster case

##  readme before all 
    1. all the cases in this file are destinated to check cluster setup 
    2. after cluster setup, check whether the kubean job status is "complete"
    3. after cluster setup, it will be inspected by sonobuoy

### Create basic cluster
    1. deploy kubean in a k8s cluster
    2. prepare config files (include: hosts-conf-cm.yml, kubeanCluster.yml, kubeanClusterOps.yml, vars-conf-cm.yml)
    3. prepare one machine to setup k8s cluster: centos 7.9, x86
    4. set k8s version :1.24.3
    5. set CRI: Containerd
    6. set CNI: calico
    7. set Calico tunnel mode: IPIP always
    8. set override_system_hostname: false
    9. set auto_renew_certificates: false
    10. set ssh method: username and password
    11. start create cluster

### Create cluster with one master and one worker
    1. prepare the config file as basic cluster
    2. change cluster topology config：1 master and 1 worker
    3. prepare 2 hosts
    3. start setup cluster

### Support k8s: 1.22.0
    1. prepare the config file as basic cluster
    2. set k8s version: 1.22.0
    3. start create cluster

### Support k8s: 1.23.0
    1. prepare the config file as basic cluster
    2. set k8s version: 1.23.0
    3. start create cluster

### Support CRI: docker
    1. prepare the config file as basic cluster
    2. set CRI: docker
    3. start create cluster

### SSH authorization: private key
    1. prepare the config file as basic cluster
    2. generate secret key on kubean node with cmd: ssh-keygen
    3. scp public key file to every cluster node use cmd: ssh-copy-id -i ~/.ssh/id_rsa.pub root@xx.xx.xx.xx
    4. write base64 code of /root/.ssh/id_rsa to the config file ssh-auth-secret.yml
    5. start create cluster

### Support CRI: Cilium
    1. prepare the config file as basic cluster
    2. set CNI: Cilium
    3. prepare machine: the kernel should be not lower than centos8
    4. start create cluster

### Support kube_pods_subnet
    1. prepare the config file as basic cluster
    2. set kube_pods_subnet
    3. start create cluster

### Support kube_service_addresses
    1. prepare the config file as basic cluster
    2. set kube_service_addresses
    3. start create cluster

### Support cluster ca auto_renew
    1. prepare the config file as basic cluster
    2. set config of auto_renew_certificates: true
    3. set config of update frequency, refer to: auto_renew_certificates_systemd_calendar: "*-*-* 15:{{ groups['kube_control_plane'].index(inventory_hostname) }}0:00"
    4. start create cluster
    5. after creation, list timer on master node with cmd：systemctl list-timers
    6. check timer whether the list include k8s-cert-renew.timer, and the cert-renew timer's frequency is right
    6. after timer excused, check whether the k8s ca info is updated with cmd: kubeadm certs check-expiration

### Support overwrite hostname
    1. prepare the config file as basic cluster
    2. set override_system_hostname: true
    3. start create cluster
    4. check the hostname of every cluster node

### Support Readhat8 OS
    1. prepare the config file as basic cluster
    2. change cluster topology config in hosts-config-cm.yml： 1 master and 1 worker
    3. prepare 2 redhat8 machines and enable the repository using subscription-manager in RHEL
    4. start install cluster

### Support Centos8 OS
    1. prepare the config file as basic cluster
    2. change cluster topology config in hosts-config-cm.yml: 1 master and 1 worker
    3. prepare 2 Centos8 machines and change the Yum source
    4. start install cluster

### Ntp func when create cluster
    1. prepare the config file as basic cluster, and set parameter: ntp_enabled=true
    2. change cluster topology config in hosts-config-cm.yml： 1 master and 1 worker
    3. change the system time of one master node with cmd: date -s "**:**:**"
    4. start create cluster

### Create cluster topology ：3 master and 2 worker
    1. prepare the config file as basic cluster
    2. change cluster topology config： 3 master + 2 worker
    3. prepare 5 machines
    3. start setup cluster



