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

``` shell
sudo yum install numactl-devel -y
```

## 源码下载

``` shell
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
``` shell
source ./scripts/env.sh
```

### 编译
``` shell
make -j
```