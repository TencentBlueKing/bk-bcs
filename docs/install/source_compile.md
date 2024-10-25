- [bk-bcs 编译指南](#bk-bcs-编译指南)
  - [编译环境](#编译环境)
    - [容器镜像编译环境](#容器镜像编译环境)
      - [官方编译镜像](#官方编译镜像)
      - [通过官方编译镜像 Dockerfile 完善自己的编译环境](#通过官方编译镜像-dockerfile-完善自己的编译环境)
    - [本地开发机编译环境](#本地开发机编译环境)
      - [编译环境准备](#编译环境准备)
  - [源码下载](#源码下载)
  - [编译步骤](#编译步骤)
    - [进入源码根目录](#进入源码根目录)
    - [下载完整依赖](#下载完整依赖)
    - [修改并初始化编译参数](#修改并初始化编译参数)
    - [编译](#编译)
    - [编译产出物](#编译产出物)
    - [模块配置范例](#模块配置范例)
    - [镜像构建](#镜像构建)


# bk-bcs 编译指南
> 说明：
> 1. 编译2024年10月23日之前的代码版本，建议使用 Golang 版本为 1.17～1.20.2。
> 2. 编译2024年10月23日之后的代码版本，建议使用 Golang 版本为 1.23.2 及以上。
> 3. 当前本编译构建方式在纯净的 CentOS 7、CentOS 8、Ubuntu环境进行测试验证；若您的环境为其他操作系统或发行版或CPU架构，欢迎您积极贡献。

## 编译环境
> 用docker镜像编译bk-bcs项目，解决了组件依赖及版本问题，相比本地编译会更加顺畅，推荐使用。

### 容器镜像编译环境
#### 官方编译镜像
```shell
hub.bktencent.com/blueking/centos8:go1.23.2_node20.18.2
hub.bktencent.com/blueking/centos7:go1.20.2_node16.20.2
```

另外，以下提供了本项目官方镜像的 Dockefile，便于用户可据此自行调整自己的编译镜像。

#### 通过官方编译镜像 Dockerfile 完善自己的编译环境
- CentOS 8 + Golang 1.23.2 参见 [CentOS 8 + Golang 1.23.2](./compile_env_dockerfile.md)
- CentOS 7 + Golang 1.20.2 参见 [CentOS 7 + Golang 1.20.2](./compile_env_dockerfile.md)
- Ubuntu + Golang 1.23.2 参见 [Ubuntu + Golang 1.23.2](./compile_env_dockerfile.md)

### 本地开发机编译环境

如需本地编译，请根据自己操作系统内核及发行版处理好依赖组件的安装问题（以下提供CentOS、Ubuntu示例），主要包括：
- golang
- nodejs
- numactl
- tongsuo项目的环境依赖（参见https://github.com/Tongsuo-Project/Tongsuo.git）

#### 编译环境准备
- 以 CentOS 7 为例：参见 [ CentOS 7 本地编译](./compile_env_local.md)
- 以 CentOS 8 为例：参见 [ CentOS 8 本地编译](./compile_env_local.md)
- 以 Ubuntu 为例：参见 [ Ubuntu 本地编译](./compile_env_local.md)

## 源码下载

```shell
cd $YOUR_SRC_PATH/
git clone https://github.com/TencentBlueKing/bk-bcs.git
```

## 编译步骤

### 进入源码根目录

``` shell
cd $YOUR_SRC_PATH/bk-bcs/
```

### 下载完整依赖
``` shell
go mod tidy -v
go mod vendor
```

### 修改并初始化编译参数

$GOPATH/src/bk-bcs/scripts/env.sh 中设置了 bcs 相关服务的一些账号密码信息，可以自行修改

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

编译结束后，在 build 目录下会生成对应的产出物目录。例如k8s相关模块如下：

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

可以对该目录进行打包进行部署。

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

针对k8s的集成插件 watch、kube-agent等，默认地已经在编译输出目录放置推荐的 Dockerfile，可以参照以下用例自行构建选择容器化方式使用。

```shell
cd ./build/bcs-*/bcs-k8s-master/bcs-k8s-driver/
docker build -t myk8sdriver:latest .
```