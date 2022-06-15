## containerd

* 先采用在线安装方式(离线方式后续再做)
* 设置container_manager为containerd
* 设置containerd_version为1.6.4
  * 不是所有的版本都支持，支持的版本有1.6.4 1.4.12 1.4.9等等
* 设置etcd_deployment_type为host，官方文档有说明
* 执行 `kubectl apply -f containerd` ，基于文件夹containerd来创建job

* 待创建完毕后，使用`resetContainerdCluster`里的配置将由containerd新建而来的集群reset还原

## docker
* 对于kubespray可以使用两种docker相关技术: docker 和 cri-dockerd

### cri-dockerd

* todo
