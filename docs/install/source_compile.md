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

编译结束后，在 build 目录下会生成对应的产出物目录。

```text
.
|-- bcs-runtime
|   `-- bcs-k8s
|       |-- bcs-component
|       |   |-- bcs-apiserver-proxy
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-apiserver-proxy
|       |   |   |-- bcs-apiserver-proxy-tools
|       |   |   |-- bcs-apiserver-proxy.json.template
|       |   |   |-- bcs-apiserver-proxy.yaml
|       |   |   `-- container-start.sh
|       |   |-- bcs-cluster-autoscaler
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-cluster-autoscaler
|       |   |   `-- hyper
|       |   |       |-- bcs-cluster-autoscaler-1.16
|       |   |       `-- bcs-cluster-autoscaler-1.22
|       |   |-- bcs-external-privilege
|       |   |   |-- Dockerfile
|       |   |   `-- bcs-external-privilege
|       |   |-- bcs-general-pod-autoscaler
|       |   |   |-- Dockerfile
|       |   |   `-- bcs-general-pod-autoscaler
|       |   |-- bcs-image-loader
|       |   |   |-- Dockerfile
|       |   |   `-- bcs-image-loader
|       |   |-- bcs-k8s-custom-scheduler
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-k8s-custom-scheduler
|       |   |   |-- bcs-k8s-custom-scheduler-kubeconfig.yaml
|       |   |   |-- bcs-k8s-custom-scheduler.manifest.template
|       |   |   `-- policy-config.json
|       |   |-- bcs-netservice-controller
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-netservice-controller
|       |   |   |-- bcs-netservice-ipam
|       |   |   `-- bcs-underlay-cni
|       |   `-- bcs-webhook-server
|       |       |-- Dockerfile
|       |       |-- bcs-webhook-server
|       |       `-- container-start.sh
|       `-- bcs-network
|           `-- bcs-ingress-controller
|               |-- Dockerfile
|               |-- bcs-ingress-controller
|               `-- container-start.sh
|-- bcs-scenarios
|   |-- bcs-gitops-manager
|   |   |-- Dockerfile
|   |   |-- bcs-gitops-manager
|   |   |-- bcs-gitops-manager.json.template
|   |   `-- container-start.sh
|   |-- bcs-gitops-proxy
|   |   |-- Dockerfile
|   |   |-- bcs-gitops-proxy
|   |   |-- bcs-gitops-proxy.json.template
|   |   `-- container-start.sh
|   `-- bcs-powertrading
|       |-- Dockerfile
|       |-- bcs-powertrading.json.template
|       `-- container-start.sh
|-- bcs-services
|   |-- bcs-bkcmdb-synchronizer
|   |   |-- Dockerfile
|   |   |-- bcs-bkcmdb-synchronizer
|   |   |-- bcs-bkcmdb-synchronizer.json.template
|   |   `-- container-start.sh
|   |-- bcs-cluster-manager
|   |   |-- Dockerfile
|   |   |-- bcs-cluster-manager
|   |   |-- bcs-cluster-manager.json.template
|   |   |-- cloud.json.template
|   |   |-- container-start.sh
|   |   `-- swagger
|   |       `-- swagger-ui
|   |           |-- clustermanager.swagger.json
|   |           |-- favicon-16x16.png
|   |           |-- favicon-32x32.png
|   |           |-- index.html
|   |           |-- oauth2-redirect.html
|   |           |-- swagger-ui-bundle.js
|   |           |-- swagger-ui-bundle.js.map
|   |           |-- swagger-ui-standalone-preset.js
|   |           |-- swagger-ui-standalone-preset.js.map
|   |           |-- swagger-ui.css
|   |           |-- swagger-ui.css.map
|   |           |-- swagger-ui.js
|   |           `-- swagger-ui.js.map
|   |-- bcs-cluster-reporter
|   |   |-- Dockerfile
|   |   `-- bcs-cluster-reporter
|   |-- bcs-data-manager
|   |   |-- Dockerfile
|   |   |-- bcs-data-manager
|   |   |-- bcs-data-manager.json.template
|   |   |-- container-start.sh
|   |   `-- swagger
|   |       |-- bcs-data-manager.swagger.json
|   |       |-- favicon-16x16.png
|   |       |-- favicon-32x32.png
|   |       |-- index.html
|   |       |-- oauth2-redirect.html
|   |       |-- swagger-ui-bundle.js
|   |       |-- swagger-ui-bundle.js.map
|   |       |-- swagger-ui-standalone-preset.js
|   |       |-- swagger-ui-standalone-preset.js.map
|   |       |-- swagger-ui.css
|   |       |-- swagger-ui.css.map
|   |       |-- swagger-ui.js
|   |       `-- swagger-ui.js.map
|   |-- bcs-gateway-discovery
|   |   |-- Dockerfile.apisix
|   |   |-- Dockerfile.gateway
|   |   |-- Dockerfile.micro-gateway-apisix
|   |   |-- README.md
|   |   |-- apisix
|   |   |   |-- bcs-auth
|   |   |   |   |-- authentication.lua
|   |   |   |   |-- bklogin.lua
|   |   |   |   |-- jwt.lua
|   |   |   |   `-- mock-bklogin.lua
|   |   |   |-- bcs-auth.lua
|   |   |   |-- bcs-common
|   |   |   |   `-- upstreams.lua
|   |   |   |-- bcs-dynamic-route.lua
|   |   |   |-- bkbcs-auth
|   |   |   |   `-- bkbcs.lua
|   |   |   `-- bkbcs-auth.lua
|   |   |-- apisix-start.sh
|   |   |-- bcs-gateway-discovery
|   |   |-- bcs-gateway-discovery.json.template
|   |   |-- config.yaml.template
|   |   `-- container-start.sh
|   |-- bcs-helm-manager
|   |   |-- Dockerfile
|   |   |-- bcs-helm-manager
|   |   |-- bcs-helm-manager-migrator
|   |   |-- container-start.sh
|   |   |-- lc_msgs.yaml
|   |   `-- swagger
|   |       `-- swagger-ui
|   |           |-- bcs-helm-manager.swagger.json
|   |           |-- favicon-16x16.png
|   |           |-- favicon-32x32.png
|   |           |-- index.html
|   |           |-- oauth2-redirect.html
|   |           |-- swagger-ui-bundle.js
|   |           |-- swagger-ui-bundle.js.map
|   |           |-- swagger-ui-standalone-preset.js
|   |           |-- swagger-ui-standalone-preset.js.map
|   |           |-- swagger-ui.css
|   |           |-- swagger-ui.css.map
|   |           |-- swagger-ui.js
|   |           `-- swagger-ui.js.map
|   |-- bcs-k8s-watch
|   |   |-- Dockerfile
|   |   |-- bcs-k8s-watch
|   |   |-- bcs-k8s-watch.json.template
|   |   |-- bcs-k8s-watch.yaml.template
|   |   |-- container-start.sh
|   |   `-- filter.json
|   |-- bcs-kube-agent
|   |   |-- Dockerfile
|   |   |-- bcs-kube-agent
|   |   |-- kube-agent-secret.yml
|   |   `-- kube-agent.yaml
|   |-- bcs-nodegroup-manager
|   |   |-- Dockerfile
|   |   |-- bcs-nodegroup-manager
|   |   |-- bcs-nodegroup-manager.json.template
|   |   `-- container-start.sh
|   |-- bcs-project-manager
|   |   |-- Dockerfile
|   |   |-- bcs-project-manager
|   |   |-- bcs-project-migration
|   |   |-- bcs-variable-migration
|   |   `-- swagger
|   |       |-- bcsproject.swagger.json
|   |       `-- swagger-ui
|   |           |-- favicon-16x16.png
|   |           |-- favicon-32x32.png
|   |           |-- index.html
|   |           |-- oauth2-redirect.html
|   |           |-- swagger-ui-bundle.js
|   |           |-- swagger-ui-bundle.js.map
|   |           |-- swagger-ui-standalone-preset.js
|   |           |-- swagger-ui-standalone-preset.js.map
|   |           |-- swagger-ui.css
|   |           |-- swagger-ui.css.map
|   |           |-- swagger-ui.js
|   |           `-- swagger-ui.js.map
|   |-- bcs-storage
|   |   |-- Dockerfile
|   |   |-- bcs-storage
|   |   |-- bcs-storage.json.template
|   |   |-- container-start.sh
|   |   |-- queue.conf.template
|   |   `-- storage-database.conf.template
|   |-- bcs-user-manager
|   |   |-- Dockerfile
|   |   |-- bcs-user-manager
|   |   |-- bcs-user-manager.json.template
|   |   `-- container-start.sh
|   `-- cryptools
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