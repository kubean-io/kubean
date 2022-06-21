# cilium

* cilium是什么
    * 功能类似kube-proxy，能够取代kube-proxy
* cilium怎么安装
    * cilium通过helm charts安装，需要确认镜像仓库(github_image_repo)里有该版本的cilium镜像
    * `http://10.6.170.10:5000/v2/cilium/tags/list`
    * 或者其他代理镜像仓库
* 查看cilium状态
    * `kubectl exec -n kube-system -t ds/cilium -- cilium status`

* 如何使用
    * 修改kube_network_plugin为cilium
    * cilium_version有默认值(kubespray2.19版本默认值为v1.11.3)
    * 执行`kubectl apply -f 1CiliumCluster`即可测试安装
    * 执行`kubectl apply -f 1resetCiliumCluster`将集群卸载还原

# calico + multus

# calico

* 参数说明
  * enable_dual_stack_networks: false 不启用双栈
    kube_pods_subnet: 10.244.0.0/16  pod所用子网
    calico_vxlan_mode: Always 使用vxlan模式
    calico_ipip_mode: Never 不使用ipip模式(vxlan和ipip模式只能二选一)
    calico_iptables_backend: "NFT"
    calico_pool_name: "default-pool-ipv4" ipv4pool名称
  
* 执行`kubectl apply -f 3Calico_Cluster`新建k8s集群
* 执行`kubectl apply -f 3resetCalico_Cluster`环境集群

