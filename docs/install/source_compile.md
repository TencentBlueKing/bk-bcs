# bk-bcs 编译指南

**GOPATH 是使用Golang编写项目的根目录，配置GOPATH的示例如下:**

``` shell
mkdir -p /data/workspace #为GOPATH新建一个目录
export GOPATH=/data/workspace   # 设置GOPATH地址
mkdir -p $GOPATH/src    #为GOPATH新建源代码存放路径
```

## 编译环境

- golang >= 1.11.2
- numactl-devel >= 2.0.9

```shell
sudo yum install numactl-devel -y
sudo yum install go
# 使用dep来管理Go依赖包
go get -u github.com/golang/dep/cmd/dep
export PATH=$PATH:$GOPATH/bin
```

## 源码下载

```shell
cd $GOPATH/src
git clone http://github.com/Tencent/bk-bcs.git
```

## 编译

### 进入源码根目录：

``` shell
cd $GOPATH/src/bk-bcs/
```

### 下载完整依赖：
``` shell
dep ensure -v
```

### 修改并初始化编译参数

$GOPATH/src/bk-bcs/scripts/env.sh中设置了zk，bcs相关服务的一些账号密码信息，可以自行修改

``` shell
source ./scripts/env.sh
```

### 编译
``` shell
make -j
```

### 编译产出物

编译结束后，在build 目录下会生成对应的产出物目录. 实例如下：

```tex
|-bin
|  |- bcs-api
|  |- bcs-check
|  |- bcs-client
|  |- bcs-container-executor
|  |- bcs-dns
|  |- bcs-health-master
|  |- bcs-health-slave
|  |- bcs-k8s-watch
|  |- bcs-loadbalance
|  |- bcs-mesos-driver
|  |- bcs-mesos-watch
|  |- bcs-scheduler
|  |- bcs-storage
|  `- ip-resource.so
|
`- conf
|- bcs-api
|	 |- config_file.json.template
...
...
```
