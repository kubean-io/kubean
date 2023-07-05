#  Create cluster case

### Create basic cluster
    1. deploy kubean in a k8s cluster
    2. prepare config files (include: hosts-conf-cm.yml, kubeanCluster.yml, kubeanClusterOps.yml, vars-conf-cm.yml)
    3. prepare one machine to setup k8s cluster: centos 7.9, x86
    4. set k8s version :1.24.3
    5. set CRI: containerD
    6. set CNI: calico
    7. set Calico tunnel mode: IPIP always
    8. set override_system_hostname: false
    9. set auto_renew_certificates: false
    10. set ssh method: username and password
    11. start create cluster

### Support CNI: Calico
    1. prepare the config file and set CNI: Calico
    2. set cluster topology to：1 master + 1 worker
    3. start create cluster
    3. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    4. check calico (calico-node and calico-kube-controller)pod status: pod status should be "Running"
    5. check folder /opt/cni/bin contains  file "calico" and "calico-ipam" are exist in both master and worker node
    6. check pod connection:
        6.1. create a deployment of nginx1 on master, on namespace ns1: set replicaset to 1(here call the pod as pod1)
        6.2. create a deployment of nginx2 on worker, on namespace ns2: set replicaset to 1(here call the pod as pod2)
        6.3  login master node, ping the pod2 ip, should be success
        6.4  login pod2, ping pod1 ip, should be success

### Support CRI: ContainerD
    1. prepare the config file and set CRI: ContainerD
    2. set cluster topology to：1 master
    3. start create cluster
    3. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    4. check containerD with cmd: systemctl status containerD, the output should contain 'active' and 'running'

### Not overwrite hostname
    1. prepare the config file as basic cluster
    2. set cluster topology to：1 master + 0 worker
    3. start create cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. check hostname of node: the hostname is not changed by kubean

### Create cluster with one master and one worker
    1. prepare the config file as basic cluster
    2. set cluster topology config: 1 master and 1 worker
    3. prepare 2 hosts
    4. start setup cluster
    5. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    6. check the cluster topology is: 1 master and 1 worker

### Support CRI: docker
    1. prepare the config file and set CRI: ContainerD
    2. set cluster topology to：1 master + 0 worker
    3. start create cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. check docker status with cmd: systemctl status docker, the output should contain 'active' and 'running'

### SSH authorization: private key
    1. prepare the config file, and set cluster topology to: 1 master + 0 worker
    2. generate secret key on kubean node with cmd: ssh-keygen
    3. scp public key file to every cluster node use cmd: ssh-copy-id -i ~/.ssh/id_rsa.pub root@xx.xx.xx.xx
    4. write base64 code of /root/.ssh/id_rsa to the config file ssh-auth-secret.yml
    5. start create cluster
    6. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy

### Support CRI: Cilium
    1. prepare the config file and set CNI: Cilium
    2. set cluster topology to：1 master + 1 worker
    3. start create cluster
    3. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    4. check calico pod status: pod status be "Running"
    5. login node, check the file "cilium" and "cilium-ipam" are in folder  /opt/cni/bin
    6. create 2 pods, check the connection between pod to pod and between node to pod:
        6.1 login node, then ping the pod ip
        6.2 login pod, ping the other pod ip

### Support kube_pods_subnet
    1. prepare the config file and set kube_pods_subnet: 192.168.128.0/20
    2. set cluster topology to：1 master + 0 worker
    3. start create cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. create a nginx deployment, check whether the pod ip is within the subnet

### Support kube_service_addresses
    1. prepare the config file and set kube_service_addresses: 10.96.0.0/12
    2. set cluster topology to：1 master + 0 worker
    3. start create cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. create a nginx service, check whether the service ip is within the service subnet

### Support cluster ca auto_renew
    1. prepare the config file, and set config of auto_renew_certificates: true
    2. set cluster topology to：1 master + 0 worker
    3. set config of update frequency, refer to: auto_renew_certificates_systemd_calendar: "*-*-* 15:{{ groups['kube_control_plane'].index(inventory_hostname) }}0:00"
    4. start create cluster
    5. after creation, list timer on master node with cmd: systemctl list-timers
    6. check timer whether the list include k8s-cert-renew.timer, and the cert-renew timer's frequency is right
    7. after timer excused, check whether the k8s ca info is updated with cmd: kubeadm certs check-expiration

### Support overwrite hostname
    1. prepare the config file and set override_system_hostname: true
    2. set cluster topology to：1 master + 0 worker
    3. start create cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. check the hostname of cluster node: the hostname is overwrited by kuean

### Support iptables mode
    1. prepare the config file and set kube_proxy_mode: iptables
    2. set cluster topology to：1 master + 0 worker
    3. start create cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. check the prox mode of cluster node: the prox mode is iptables

### Support Readhat8 OS
    1. prepare the config file
    2. set cluster topology config in hosts-config-cm.yml： 1 master + 0 worker
    3. prepare 2 redhat8 machines and enable the repository using subscription-manager in RHEL
    4. start install cluster
    5. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy

### Support Centos8 OS
    1. prepare the config file
    2. set cluster topology: 1 master + 0 worker
    3. prepare 2 Centos8 machines and change the Yum source
    4. start install cluster
    5. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy

### Ntp func when create cluster
    1. prepare the config file and set parameter: ntp_enabled=true
    2. set cluster topology: 1 master + 1 worker
    3. change the system time of one master node with cmd: date -s "**:**:**"
    4. start create cluster
    5. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy

### Create cluster topology ：3 master and 2 worker
    1. prepare the config file as basic cluster
    2. set cluster topology: 3 master + 2 worker
    3. prepare 5 machines
    4. start setup cluster
    5. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy

### Create cluster set all the nodes with same name
    1. prepare the config file, and set all the node name with same name in host-conf-cm.yml(ex set both nodes name as "master1")
    2. start create cluster
    3. after creation, check the job-related pod status is "Error"

### Create cluster with calico IPIP_Always tuning mode
    1. prepare the config file as basic cluster
    2. set network card mode to any one of the (FIRST FOUND, KUBERNETES INTERNAL_IP, INTERFACE REGEX)
    3. set calico tunning mode to IPIP_Always:
            calico_ipip_mode: Always
            calico_vxlan_mode: Never
            calico_network_backend: bird
    4. set cluster topology: 1 master + 1 worker
    5. start setup cluster
    6. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    7. check tunning mode:
            7.1 get ippool inf cmd: "calicoctl get ippools", and get the poolName
            7.2 query pool details: calicoctl get ippools {{poolName}} -o yaml
            7.3 check point:
                    spec.ipipMode=Always;
                    spec.vxlanMode=Never
    8. create pod1 on master, on namespace1
    9. create pod2 on worker, on namespace2
    10. login master node, ping pod2; login worker, ping pod1
    11. login pod2, ping pod1


### Create cluster with calico IPIP_CrossSubnet tuning mode
    1. prepare the config file as basic cluster
    2. set network card mode to any one of the (FIRST FOUND, KUBERNETES INTERNAL_IP, INTERFACE REGEX)
    3. set calico tunning mode to IPIP_Always:
        calico_ipip_mode: CrossSubnet
        calico_vxlan_mode: Never
        calico_network_backend: bird
    4. set cluster topology: 1 master + 1 worker
    5. start setup cluster
    6. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    7. check tunning mode:
        7.1 get ippool inf cmd: "calicoctl get ippools", and get the poolName
        7.2 query pool details: calicoctl get ippools {{poolName}} -o yaml
        7.3 check point:
                spec.ipipMode=CrossSubnet
                spec.vxlanMode=Never
    8. create pod1 on master, on namespace1
    9. create pod2 on worker, on namespace2
    10. login master node, ping pod2; login worker, ping pod1
    11. login pod2, ping pod1

### Create cluster with calico Vxlan_Always tuning mode
    1. prepare the config file as basic cluster
    2. set network card mode to any one of the (FIRST FOUND, KUBERNETES INTERNAL_IP, INTERFACE REGEX)
    3. set calico tunning mode to IPIP_CrossSubnet:
        calico_ipip_mode: Never
        calico_vxlan_mode: Always
    4. set cluster topology: 1 master + 1 worker
    5. start setup cluster
    6. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    7. check tunning mode:
            7.1 get ippool inf cmd: "calicoctl get ippools", and get the poolName
            7.2 query pool details: calicoctl get ippools {{poolName}} -o yaml
            7.3 check point:
                    spec.ipipMode=Never
                    spec.vxlanMode=Always
    8. create pod1 on master, on namespace1
    9. create pod2 on worker, on namespace2
    10. login master node, ping pod2; login worker, ping pod1
    11. login pod2, ping pod1

### Create cluster with calico Vxlan_CrossSubnet tuning mode
    1. prepare the config file as basic cluster
    2. set network card mode to any one of the (FIRST FOUND, KUBERNETES INTERNAL_IP, INTERFACE REGEX)
    3. set calico tunning mode to IPIP_CrossSubnet:
        calico_ipip_mode: Never
        calico_vxlan_mode: CrossSubnet
    4. set cluster topology: 1 master + 1 worker
    5. start setup cluster
    6. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    7. check tunning mode:
        7.1 get ippool inf cmd: "calicoctl get ippools", and get the poolName
        7.2 query pool details: calicoctl get ippools {{poolName}} -o yaml
        7.3 check point:
                spec.ipipMode=Never
                spec.vxlanMode=CrossSubnet
    8. create pod1 on master, on namespace1
    9. create pod2 on worker, on namespace2
    10. login master node, ping pod2; login worker, ping pod1
    11. login pod2, ping pod1

### Redhat84 OS Compitable
    1. prepare the config file as basic cluster
    2. set cluster topology: 1 master + 1 worker
    3. start setup cluster
    4. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    5. create pod1 on master, on namespace1
    6. create pod2 on worker, on namespace2
    7. login master node, ping pod2; login worker, ping pod1
    8. login pod2, ping pod1

### Create cluster support  calico dual stack v4: Vxlan_Always(v4)_Vxlan_Always(v6)
    1. prepare the config file as basic cluster
    2. enable the dual stack switch
         enable_dual_stack_networks: true
    3. set network card mode to any one of the:
        calico_ip_auto_method: first-found
        calico_ip6_auto_method: first-found
    4. set calico tunning mode:
        calico_ipip_mode: Never
        calico_vxlan_mode: Always
        calico_ipip_mode_ipv6: Never
        calico_vxlan_mode_ipv6: Always
    5. set cluster topology: 1 master + 1 worker
    6. start setup cluster
    7. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    8. check calico pods are Running
    9. check tunning mode of ipv4
    10. check tunning mode of ipv6
    11. create pod1 on master, on namespace1
    12. create pod2 on worker, on namespace2
    13. check pod1 and pod2 both have 2 ips
    14. login master node, ping pod2; login worker, ping pod1
    15. login pod2, ping pod1

### Create cluster support  calico dual stack v4: Vxlan_CrossSubnet(v4)_Vxlan_Always(v6)
    1. prepare the config file as basic cluster
    2. enable the dual stack switch
         enable_dual_stack_networks: true
    3. set network card mode to any one of the:
        calico_ip_auto_method: first-found
        calico_ip6_auto_method: first-found
    4. set calico tunning mode:
        calico_ipip_mode: Never
        calico_vxlan_mode: CrossSubnet
        calico_ipip_mode_ipv6: Never
        calico_vxlan_mode_ipv6: Always
    5. set cluster topology: 1 master + 1 worker
    6. start setup cluster
    7. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    8. check calico pods are Running
    9. check tunning mode of ipv4
    10. check tunning mode of ipv6
    11. create pod1 on master, on namespace1
    12. create pod2 on worker, on namespace2
    13. check pod1 and pod2 both have 2 ips
    14. login master node, ping pod2; login worker, ping pod1
    15. login pod2, ping pod1

### Create cluster support  calico dual stack v4: IPIP_Always(v4)_Vxlan_CrossSubnet(v6)
    1. prepare the config file as basic cluster
    2. enable the dual stack switch
        enable_dual_stack_networks: true
    3. set network card mode to any one of the:
        calico_ip_auto_method: first-found
        calico_ip6_auto_method: first-found
    4. set calico tunning mode:
        calico_ipip_mode: Always
        calico_vxlan_mode: Never
        calico_network_backend: bird
        calico_ipip_mode_ipv6: Never
        calico_vxlan_mode_ipv6: CrossSubnet
    5. set cluster topology: 1 master + 1 worker
    6. start setup cluster
    7. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    8. check calico pods are Running
    9. check tunning mode of ipv4
    10. check tunning mode of ipv6
    11. create pod1 on master, on namespace1
    12. create pod2 on worker, on namespace2
    13. check pod1 and pod2 both have 2 ips
    14. login master node, ping pod2; login worker, ping pod1
    15. login pod2, ping pod1

### Create cluster support  calico dual stack v4: IPIP_CrossSubnet(v4)_Vxlan_CrossSubnet(v6)
    1. prepare the config file as basic cluster
    2. enable the dual stack switch
        enable_dual_stack_networks: true
    3. set network card mode to any one of the:
        calico_ip_auto_method: first-found
        calico_ip6_auto_method: first-found
    4. set calico tunning mode:
        calico_ipip_mode: CrossSubnet
        calico_vxlan_mode: Never
        calico_network_backend: bird
        calico_ipip_mode_ipv6: Never
        calico_vxlan_mode_ipv6: CrossSubnet
    5. set cluster topology: 1 master + 1 worker
    6. start setup cluster
    7. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    8. check calico pods are Running
    9. check tunning mode of ipv4
    10. check tunning mode of ipv6
    11. create pod1 on master, on namespace1
    12. create pod2 on worker, on namespace2
    13. check pod1 and pod2 both have 2 ips
    14. login master node, ping pod2; login worker, ping pod1
    15. login pod2, ping pod1

### Create cilium cluster
    1. prepare one master and one worker as basic cluster, the os should be redhat8.
    2. set cni to cilium
        kube_network_plugin: cilium
    3. set kube_service_addresses and kube_pods_subnet in vars-conf-cm.yml, for example:
        kube_service_addresses: 10.88.0.0/16
        kube_pods_subnet: 192.88.128.0/20
    4. set cluster topology: 1 master + 1 worker, change config of hosts
    5. start setup cluster, use kubectl apply -f ***
    6. check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    7. check there are four pods named 'cilium-***'
    8. create pod1 and pod2 on master, pod3 on worker
    9. check ip, it should in the range of kube_pods_subnet
    10. create a service, and check the ip of the service, it should in the range of kube_service_addresses
    11. login master node, ping pod2 and pod3; login worker, ping pod1
    12. login pod1, ping pod2 and pod3

### Create 3master cluster (vip)
    1 Prepare three virtual machines as the nodes of the created work cluster. The OS is centos7
    2. yml files required for configuring Kubean:
    Configure hosts-config-cm.yml file to write the information of the three nodes
    Configure kubeanClusterOps using action: cluster.yml
    Configure parameters in vars-conf-cm.yml that support kube-vip
    kube_vip_arp_enabled: true
    kube_proxy_strict_arp: false
    kube_vip_enabled: true
    kube_vip_controlplane_enabled: true
    kube_vip_address: <ip address> # Select an ip address that is the same network segment as the host and cannot be pinged
    3. Create a cluster using kubectl apply - f ***
    4. Check that the job related pod status is "Succeeded" and check the cluster status by sonobuoy
    5. Traverse and login to each master node to check which vip is in effect
    6. shutdown the node where vip is located
    7. Login with kube-vip ip address and check that vip has drifted to a node that is not powered off

### cluster master node scale
1 Create a working cluster of 1 master node
2 Add in a master node
Check that:
    a. node is installed successfully 
    b. k8s components are running correctly (including etcd)
    c. Create nginx application running correctly
3 Add another master node. i.e. 1+1+1 3master node cluster.
Check that:
    a. node is installed successfully 
    b. k8s components are running correctly (including etcd)
    c. Create nginx application running correctly 
    d. after k8s component restart, check component can restart successfully and application pod running normally
4 Cluster reverts to one master node
5 Add in a master node
Check that:
    a. node is installed successfully
    b. k8s components are running correctly (including etcd)
    c. Create nginx application running correctly
6 Similarly, add two master nodes at once. The final cluster will be 1+1+2 master nodes
Check that:
    a. node is installed successfully
    b. k8s components are running correctly (including etcd)
    c. Create nginx application running correctly

### non root users create a work cluster
1 Create user aaa for vm, set password and give non-root privileges
2 Modify the username and password information in host-conf-cm.yml
3 kubean creates the job cluster
4 Expectation: not successful (Missing sudo password output in job logs)

### Configure spec.activeDeadlineSeconds and verify if the job terminaled after timeout
1 Add activeDeadlineSeconds: 30 to the kubeanClusterOps configuration file
2 Start a job that creates a single-node job cluster
3 Disable the worker node network at this point (such as: nmcli c down ens192)
4 Verify that the worker node network is unreachable
5 Check that the job is stopped (complete), and expects to change state within the timeout set by spec.activeDeadlineSeconds
