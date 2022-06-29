## 简单排错

```
Failed to connect to the host via ssh: Shared connection to ***.***.***.178 closed
```

* ssh连接会话超时，修改`/etc/ssh/sshd_config`，将`ClientAliveInterval`修改为合适时间值 ，并重启ssh服务
* 也可能是目标集群机器性能不足，将虚拟机迁移搭配资源充足的资源池中

```
Failed to connect to the host via ssh
```

* 说明ssh连接操作失败了，也许是know_hosts机制在起作用或者密码 key等错误

## 文档说明和使用

使用步骤在 https://dwiki.daocloud.io/pages/viewpage.action?pageId=126747049

### no-first-master节点的移除和新增

* 执行`kubectl apply -f 1prepare_three_puls_one_master_etcd_Cluster`来准备3+1集群，其中etcd节点在master1，且master1为first-master
* 执行`kubectl apply -f 1remove_master2`从controller nodes中移除master2
* 执行`kubectl apply -f 2add_master2_again`将master2再添加到controller nodes列表中来
* 重启nginx-proxy
* 执行`kubectl apply -f 1resetPrepare_three_puls_one_master_etcd_Cluster`将集群还原

### first-master节点的移除

* 执行`kubectl apply -f 3prepare_three_puls_one_master_etcd_Cluster`来准备3+1集群，其中etcd节点在master2，其中master1为first-master
* 执行`kubectl apply -f 3change_master1_order`将master1节点调整位置顺序到末尾
* 执行`kubectl apply -f 3remove_master1`将master1移除
* 执行`kubectl apply -f 3reset_three_puls_one_master_etcd_Cluster`将集群还原

### first-master节点的新增
* 执行`kubectl apply -f 4prepare_two_puls_one_master_etcd_Cluster`来准备2+1集群,其中controller nodes分别为master2和master3
* 执行`kubectl apply -f 4add_master1_end`来将master1添加到controller nodes的末尾