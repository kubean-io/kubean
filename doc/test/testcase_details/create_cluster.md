#  Create cluster case

##  readme before all 
    1. all the cases in this file are destinated to check cluster setup 
    2. after cluster setup, check whether the kubean job status is "complete"
    3. after cluster setup, it will be inspected by sonobuoy


### create basic cluster
    1. deploy kubean in a k8s cluster
    2. prepare config files (include: hosts-conf-cm.yml, kubeanCluster.yml, kubeanClusterOps.yml, vars-conf-cm.yml)
    3. prepare one node to setup k8s cluster: centos 7.9, x86
    4. set k8s version :1.24.3
    5. set CRI: Containerd
    6. set CNI: calico
    7. set Calico tunnel mode: IPIP always
    8. set override_system_hostname: false
    9. set auto_renew_certificates: false
    10. set ssh method: username and password
    11. start create cluster

### create cluster with one master and one worker
    1. prepare the config file as basic cluster
    2. change cluster topology config： 1 master and 1 worker
    3. prepare 2 hosts
    3. start setup cluster

### create cluster with k8s version 1.22.0
    1. prepare the config file as basic cluster
    2. set k8s version :1.22.0
    3. start create cluster

### create cluster with k8s version 1.23.0
    1. prepare the config file as basic cluster
    2. set k8s version :1.23.0
    3. start create cluster

### create cluster with docker CRI
    1. prepare the config file as basic cluster
    2. set set CRI: docker
    3. start create cluster

### create cluster with secret key
    1. prepare the config file as basic cluster
    2. set ssh method: secret key
    3. prepare secret key file
    4. start create cluster

### create cluster with CNI: Cillium
    1. paas

### create cluster user set pods subnet
    1. prepare the config file as basic cluster
    2. set kube_pods_subnet
    3. start create cluster

### create cluster user set service address
    1. prepare the config file as basic cluster
    2. set kube_service_addresses
    3. start create cluster

### create cluster override hostname
    1. prepare the config file as basic cluster
    2. set override_system_hostname: true
    3. start create cluster

### reset cluster
    1. create cluster：topology 1 master and 1 worker
    2. reset cluster
    3. recreate cluster




