#  Cluster operate case

### Cluster reset
    1. create cluster：topology 1 master and 1 worker
    2. reset cluster
    3. recreate cluster use the machines above

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