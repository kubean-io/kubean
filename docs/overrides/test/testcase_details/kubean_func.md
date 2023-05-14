## kubean function 

###  Posthook cluster kubeconfig
    1. create a basic cluster
    2. use the kubeconfig in  configmap: {{cluster-name}}-kubeconf to query cluster
    3. reset the basic cluster

### Set multi kubean operator
    1. create a kind cluster
    2. deploy kubean
    3. use kubectl scale deployment to set kubean operator to 3
    4. check the kubean deployment, the ReadyReplicas should be 3
    5. deploy clusters with kubean