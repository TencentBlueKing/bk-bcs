# bcs-terraform-bkprovider

bcs-terraform-bkprovider 是蓝鲸容器服务（BlueKing Container Service, BCS）生态系统中的一个微服务，充当 **蓝鲸后端系统的 API 代理**，同时提供 gRPC 和 HTTP REST 两种访问方式。

## 功能概述

本服务将蓝鲸后端系统的主机 Agent 管理和网络安全配置操作封装为统一的 gRPC/HTTP API，提供以下核心能力：

- **主机 Agent 管理**：通过蓝鲸节点管理（BK NodeMan）API 网关，代理 GSE Agent 的安装、卸载、重启等任务下发，以及主机信息和云区域的查询与管理
- **白名单管理**：通过腾讯云 VPC SDK 管理蓝鲸网关出口 IP 白名单的注册与查询

## API 接口

本服务通过 Protobuf 定义了以下 11 个 gRPC 接口，同时通过 gRPC-Gateway 提供 HTTP REST API：

| gRPC 方法 | HTTP 方法 | 路径 | 说明 |
|-----------|----------|------|------|
| `InstallJob` | POST | `/terraform-bkprovider/v1/install_job` | GSE Agent 安装/卸载/重启任务下发 |
| `ListHost` | POST | `/terraform-bkprovider/v1/list_host` | 查询 GSE Agent 主机信息 |
| `ListProxyHost` | GET | `/terraform-bkprovider/v1/list_proxy_host` | 查询 GSE Proxy 信息 |
| `CreateCloud` | POST | `/terraform-bkprovider/v1/cloud` | 创建云区域 |
| `UpdateCloud` | PUT | `/terraform-bkprovider/v1/cloud` | 更新云区域 |
| `ListCloud` | GET | `/terraform-bkprovider/v1/cloud` | 获取云区域列表 |
| `DeleteCloud` | DELETE | `/terraform-bkprovider/v1/cloud` | 删除云区域 |
| `GetJobDetail` | POST | `/terraform-bkprovider/v1/get_job_detail` | 获取任务详情 |
| `RegisterBkWhitelist` | POST | `/terraform-bkprovider/v1/register_bk_whitelist` | 注册出口 IP 到蓝鲸白名单 |
| `ListBkWhitelist` | GET | `/terraform-bkprovider/v1/list_bk_whitelist` | 查询蓝鲸白名单 |
| `GetBkOuterIP` | GET | `/terraform-bkprovider/v1/get_bk_outer_ip` | 获取蓝鲸出口 IP 列表 |

详细的 API 文档可通过 Swagger UI 查看：`/terraform-bkprovider/swagger/`

## 目录结构

```
bcs-terraform-bkprovider/
├── cmd/server/          # 服务启动与初始化（Cobra CLI、配置加载、gRPC + Gateway）
├── common/              # 公共模块（配置结构体、错误码、工具函数）
├── handler/             # gRPC Handler 业务逻辑层
├── middleware/           # 外部系统客户端
│   ├── xbknodeman/      #   蓝鲸节点管理 API 客户端（Cloud/Host/Job）
│   ├── xtencentcloud/   #   腾讯云 VPC 客户端（IP 白名单）
│   └── xrequests/       #   HTTP 请求工具库
├── pkg/middleware/auth/  # JWT 认证与授权中间件
├── proto/               # Protobuf 服务定义及生成代码
├── images/              # 容器镜像构建与配置模板
└── sdk/                 # 独立的 Go SDK 客户端库
```

## 技术栈

- **语言**：Go 1.23
- **微服务框架**：go-micro v4（gRPC 服务器/客户端）
- **服务注册**：etcd v3
- **HTTP 网关**：grpc-gateway（gRPC → HTTP REST 桥接）
- **认证**：JWT（BCS 统一认证体系）
- **CLI**：Cobra
- **外部集成**：蓝鲸节点管理 API 网关、腾讯云 VPC SDK
- **容器化**：Docker（CentOS 7 基础镜像）

## 快速开始

### 前置条件

- Go 1.23+
- etcd 集群（用于服务注册）
- 蓝鲸节点管理 API 网关访问权限
- 腾讯云 VPC 访问权限（如需白名单管理功能）

### 构建

```bash
# 编译二进制
make build

# 构建 Docker 镜像
make docker
```

### 代码生成

```bash
# 安装 protoc 代码生成工具链
make init

# 从 .proto 文件生成 Go/gRPC/Gateway/Swagger 代码
make proto
```

### 测试

```bash
# 运行所有测试
make test

# 运行指定包的测试
go test -v ./middleware/xbknodeman/... -cover
```

### 配置

服务通过 JSON 配置文件启动，默认路径为 `./bcs-terraform-bkprovider.json`。容器化部署时使用 `envsubst` 渲染配置模板，支持通过环境变量注入配置。

主要配置项包括：

- **Server**：gRPC/HTTP 监听地址、TLS 证书
- **Etcd**：服务注册中心地址、TLS 证书
- **Auth**：JWT 公钥/私钥配置
- **BkSystem**：蓝鲸应用认证（bk_app_code/bk_app_secret）、API 网关地址
- **TencentCloud**：腾讯云 API 密钥、VPC 配置

## SDK

项目附带独立的 Go SDK 客户端库，位于 `sdk/bcsprovider-sdk-go/`，提供集群管理、Helm 管理和项目管理等服务接口。使用示例参见 `sdk/bcsprovider-sdk-go/examples/`。
