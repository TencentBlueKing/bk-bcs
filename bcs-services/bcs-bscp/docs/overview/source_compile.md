# BSCP编译文档

## 1. 编译环境

- golang >= 1.16
- protoc 3.20.0
- protoc-gen-go v1.28.0
- protoc-gen-go-grpc v1.2
- protoc-gen-grpc-gateway v1.16.0

**注：** BSCP源码文件中的 <u>pkg/protocol/README.md</u> 包含了 protoc 相关依赖的安装教程。

**将go mod设置为on**

```shell
go env -w GO111MODULE="on"
```



## 2. 源码下载

```shell
cd $GOPATH/src
git clone https://github.com/Tencent/bk-bscp.git github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp
```



## 3. 编译

**进入源码根目录：**

```shell
cd $GOPATH/src/github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp
```

### 3.1 编译共有三种模式

#### 3.1.1 模式一：全部编译 （Protoc文件、安装脚本、集成测试、API-Gateway网关相关文档、后端服务）

```shell
make
```

**进入编译完的文件夹下**

```shell
cd build/bk-bscp.${GITTAG}-${DATE}

# 如：
cd build/bk-bscp.release-v1.0.0_alpha1-2-g42xxxx5-22.04.25
```

**编译后的文件说明：**

```shell
├── CHANGELOG.md 													# 变更文档
├── VERSION																# 当前编译包版本说明
├── api																		# API-Gateway相关接口文档，会在部署文档进行介绍
│   ├── api-server
│   └── feed-server
├── bk-bscp-apiserver											# BSCP服务目录，后续服务同理不进行重复介绍
│   ├── bk-bscp-apiserver									# 编译完的apiserver服务二进制
│   ├── bk-bscp-apiserver.sh							# 服务启动脚本，包括启动、停止、检测服务是否启动命令
│   └── etc																# 服务配置文件存放目录
├── ...
├── install																# 安装脚本存放目录
│   ├── start_all.sh											# 启动BSCP服务的脚本
│   └── stop_all.sh												# 停止BSCP服务的脚本
└── suite-test														# BSCP集成测试
    ├── README.md													# README文档
    ├── application.test									# BSCP application资源相关基础测试用例
    ├── start.sh													# BSCP 集成测试允许脚本
    └── tools.sh													# 对集成测试结果进行统计分析的工具
```

#### 3.1.2 模式二：仅编译后端服务（Protoc文件、后端服务）

```shell
make server
```

编译后的文件夹下，仅有各服务目录，不包含安装脚本等其他目录。

#### 3.1.3 模式三：打包编译

```shell
make package
```

此编译方式用于正式环境部署使用，里面仅有各服务二进制、配置文件，以及安装脚本。本地开发和初次部署勿用！

**编译后的文件说明：**

```shell
├── CHANGELOG.md 													# 变更文档
├── VERSION																# 当前编译包版本说明
├── api																		# API-Gateway相关接口文档，会在部署文档进行介绍
│   ├── api-server
│   └── feed-server
├── bin																		# BSCP各服务二进制文件
│   ├── bk-bscp-apiserver
│   ├── bk-bscp-authserver
│   ├── bk-bscp-cacheservice
│   ├── bk-bscp-configserver
│   ├── bk-bscp-dataservice
│   └── bk-bscp-feedserver
├── etc																		# BSCP各服务配置文件
│   ├── api_server.yaml
│   ├── apiserver_api_gw_public.key				# BSCP apiserver API-Gateway网关JWT解析public key
│   ├── auth_server.yaml
│   ├── cache_service.yaml
│   ├── config_server.yaml
│   ├── data_service.yaml
│   ├── feed_server.yaml
│   └── feedserver_api_gw_public.key			# BSCP feedserver API-Gateway网关JWT解析public key
└── install																# 安装脚本存放目录
```
