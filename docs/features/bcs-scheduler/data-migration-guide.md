# bcs-scheduler存储数据迁移指引

## 工具介绍

### 二进制名称

bcs-migrate-data

### 使用方法

#### 示例脚本

```shell
。/bcs-migrate-data --bcs_zookeeper 127.0.0.1:2181 \
    --kubeconfig ./kubeconfigfile \
    --log_dir ./logs
```

#### 参数说明

* **bcs_zookeeper**：原本bcs-scheduler使用的zk的地址
* **kubeconfig**：kube-apiserver（即etcd代理）的访问配置文件
* **log_dir**：日志文件存储目录
* **alsologtostderr**：同时将日志文件输出至标准错误

## 迁移步骤说明

1. 停止所有bcs-scheduler，bcs-mesos-watch
2. 运行bcs-migrate-data，确定无报错信息
3. 修改bcs-scheduler存储类型为etcd，修改bcs-mesos-watch存储类型为etcd
4. 重启bcs-scheduler
