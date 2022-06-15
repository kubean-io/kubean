## containerd

* 先采用在线安装方式(离线方式后续再做)
* 设置container_manager为containerd
* 设置containerd_version为1.6.4 1.6.3 1.6.2等等都可以(github有的版本都行)
* 设置etcd_deployment_type为host，官方文档有说明
* 执行 `kubectl apply -f containerd` ，基于文件夹containerd来创建job

* 待创建完毕后，使用`resetContainerdCluster`里的配置将由containerd新建而来的集群reset还原

## docker

* docker 是 docker engine
* cri-dockerd是什么
    * cri-dockerd是dockershim的代替品
    * k8s1.24版本移除了dockershim，所以可以使用cri-dockerd作为代替
    * k8s通过内置dockershim组件来操作docker engine
    * 容器引擎和k8s之间通过CRI(容器运行时接口)来沟通，但docker engine不兼容CRI，所以先前用k8s基于dockershim来操纵docker engine
    * k8s1.20建议移除dockershim

* 如果需要离线搭建部署环境需要注意的点
    * 各个离线资源包，比如kubelet kubeproxy
    * 针对于docker软件，需要对于各个linux发行版搭建repo仓库

* 文件夹docker_cri_dockerd用于演示部署基于docker的k8s
    * container_manager为docker
    * etcd_deployment_type不再只能为host
    * cri_dockerd_enabled为 true(对于低于1.24的k8s不需要)
    * 指定cri_dockerd_version版本(对于低于1.24的k8s不需要)
    * docker是采用linux repo形式安装docker-ce软件(apt yum等等)
* 执行`kubectl apply -f docker_cri_dockerd`来创建job以执行kubespray任务
* 执行`kubectl apply -f resetDocker_cri_dockerd`来还原由docker_cri_dockerd创建而来的集群
