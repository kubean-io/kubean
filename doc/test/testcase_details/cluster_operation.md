#  Cluster operate case

### Cluster reset
    1. create a cluster with topology: 1master +0 worker
    2. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    3. start reset cluster
    4. after reset, check the reset job pod status is "Succeeded"
    5. login node, check node reset:
        5.1 kubectl check: execute "kubectl", output should contain "command not found"
        5.2 CRI check: execute "systemctl status containerd.service"(if cri is docker, check docker service), output should contain "inactive" and "dead"
        5.3 CNI check1: execute "ls -al /opt", the output should not contain "cni"
        5.4 CNI check2: execute "ls -al /etc",the output should not contain "cni"
        5.5 k8s config file check: execute "ls -al /root", the output should not contain "\\.kube"
        5.6 kubelet check: execute "ls -al /usr/local/bin", the output should not contain "kubelet"
    6. start a new cluster creation
    7. after the second create job finished, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy

### Add worker
    1. create a cluster: 1 master + 1worker
    2. check cluster status
    3. add a worker to cluster
    4. check cluster status
    
### Remove online worker
    1. create a cluster: 1master + 2worker
    2. check cluster status
    3. remove a worker
    4. check cluster status

### Remove offline worker
    1. create a cluster: 1master +2worker
    2. check cluster status 
    3. shutdown a worker
    4. remove the powered down worker
    5. check cluster status

### Online worker down in remove procedure
    1. create a cluster: 1master + 2worker
    2. check cluster status 
    3. start remove a worker
    4. while in remove procedure: shutdown the worker
    5. after remove job fail, start a new remove operation to remove the offline worker
    6. after the second remove job completed, check the cluster status

### Readd a worker to cluster
    1. create a cluster: 1master + 2worker
    2. check cluster status 
    3. remove a worker
    4. check cluster status
    5. change the hostname and ip of the removed worker
    6. readd the removed worker to cluster
    7. check cluster status

### Ntp func while cluster in use
    1. prepare the config file as basic cluster, and set parameter: ntp_enabled=true
    2. change cluster topology config： 3 master and 1 worker
    3. start create cluster
    4. after cluster created, change the system time of one master node
    5. check cluster status after the cluster run at least 10 minutes

### Hot upgrade k8s Y version: online
    1. prepare the config file as basic cluster and set kube_version: {{X.Y.Z}}
    2. set cluster topology: 1master + 1worker
    3. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    4. start upgrade k8s version from  {{X.Y.Z}} to {{X.Y+1.*}}
    5. after creation, check the job-related pod status is "Succeeded", and check cluster status by sonobuoy
    6. check k8s version by cmd: kubectl version  --short, the "Server Version" should be: {{X.Y+1.*}}
    7. check node version: kubectl get node -o wide, the version should be {{X.Y+1.*}}


### Retry 0 times when job fail
    1. prepare the config file, and set backoffLimit=0
    2. start a bound to fail jod
    3. check the job-related pod status is "Error"
    4. wait 100s, check the job-related pod total count is 1