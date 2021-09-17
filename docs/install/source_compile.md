# bk-bcs 编译指南

**GOPATH 是使用Golang编写项目的根目录，配置GOPATH的示例如下:**

``` shell
mkdir -p /data/workspace #为GOPATH新建一个目录
export GOPATH=/data/workspace   # 设置GOPATH地址
mkdir -p $GOPATH/src    #为GOPATH新建源代码存放路径
# 使用go mod进行依赖管理
export GO111MODULE=on
```

## 编译环境

- golang >= 1.11.2
- numactl-devel >= 2.0.9

```shell
sudo yum install numactl-devel -y
sudo yum install go
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
go mod tidy -v
go mod vendor
```

### 修改并初始化编译参数

$GOPATH/src/bk-bcs/scripts/env.sh中设置了zk，bcs相关服务的一些账号密码信息，可以自行修改

``` shell
source ./scripts/env.sh
```

### 编译

BCS包含大量可随时热插拔的方案组件，默认是全模块编译。

```shell
make -j
```

如果只想编译k8s相关模块：

``` shell
make bcs-k8s
```

### 编译产出物

编译结束后，在build 目录下会生成对应的产出物目录。例如k8s相关模块如下：

```text
.
|-- bcs-k8s-master
|   |-- bcs-k8s-driver
|   |   |-- bcs-k8s-driver
|   |   `-- start.sh
|   |-- bcs-k8s-watch
|   |   |-- bcs-k8s-watch
|   |   `-- bcs-k8s-watch.json.template
|   `-- bcs-kube-agent
|       |-- bcs-kube-agent
|       |-- kube-agent-secret.yml
|       `-- kube-agent.yaml
`-- bcs-services
    |-- bcs-api
    |   |-- bcs-api
    |   `-- bcs-api.json.template
    |-- bcs-client
    |   |-- bcs.conf.template
    |   `-- bcs-client
    `-- bcs-storage
        |-- bcs-storage
        |-- bcs-storage.json.template
        `-- storage-database.conf.template
```

可以对该目录进行打包进行部署

### 模块配置范例

针对所有模块，编译输出目录都有对应的配置模板，模板中预留了配置项的占位符。
请选择需要的模块，确认相关配置项，根据实际部署信息进行替换。

相关的部署配置选项可以参考 $GOPATH/src/bk-bcs/scripts/config.sh，完成选项
配置后可以在启动前进行相关渲染

```shell

source config.sh

for tpl in `find ./bcs-* -type f -name '*.template'`; do
    echo "${tpl} ==> ${tpl%.template}"
    cat $tpl | envsubst | tee ${tpl%.template}
done
```

### 镜像构建

针对k8s的集成插件driver、watch以及kube-agent，默认地已经在编译输出目录放置推荐的Dockerfile，可以参照
以下用例自行构建选择容器化方式使用。

```shell
cd ./build/bcs-*/bcs-k8s-master/bcs-k8s-driver/
docker build -t myk8sdriver:latest .
```

