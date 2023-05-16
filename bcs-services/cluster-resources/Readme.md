# BcsServices / ClusterResources

> 蓝鲸容器服务（Bcs）集群资源层，用于屏蔽底层集群类型，提供统一的 Restful 接口以供 SaaS / OpenAPI 使用

## 开发指南

### 依赖组件

```text
Go                    1.17.5
etcd                  3.5.0
protoc                3.20.3
micro                 v4
go-micro              v1.1.4
protoc-gen-go         1.5.2
protoc-gen-micro      v1.0.0
protoc-grpc-gateway   v1.16.0
protoc-gen-swagger    v1.16.0
grpc                  v1.42.0
```

### 环境准备

protoc
```bash
# 解压到 $PATH 任意目录
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.20.3/protoc-3.20.3-linux-x86_64.zip
```

```shell script
# 默认安装在 $GOPATH/bin 下
export GO111MODULE=on
# go-micro new service 等依赖
go install github.com/go-micro/cli/cmd/go-micro@v1.1.4
# proto 依赖
go install github.com/go-micro/generator/cmd/protoc-gen-micro@v1.0.0
go install github.com/golang/protobuf/protoc-gen-go@1.5.2
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0

# 编译 swagger-ui => datafile.go 用
go get github.com/go-bindata/go-bindata/...

go mod tidy
```

### 常用操作

#### 生成 pb.x.go, swagger.json 文件

```shell script
make proto
```

#### 生成可执行二进制

```shell script
make build
```

#### 启动服务

```shell script
# go run，若不指定 conf.yaml，则使用 ./conf/cr_conf.yaml
go run main.go --conf xxx.yaml

# 或 执行二进制文件
./bcs-cluster-resources --conf xxx.yaml
```

#### 采集国际化数据

```shell script
make gen-lc-msgs
```

执行以上命令后，完成 lc_msgs.yaml 中的 TODO 项

#### 验证服务

- 使用 micro call
```shell script
$ micro --registry=etcd --registry_address=127.0.0.1:2379 call clusterresources.bkbcs.tencent.com ClusterResources.Ping
{
        "ret": "pong"
}
```

- 使用 curl 或 Postman
```shell script
$ curl http://127.0.0.1:9091/clusterresources/v1/ping

{"ret":"pong"}
```

#### 查看 Swagger Api Doc

- 确认 conf 中 `swagger.enabled` 为 `true`，`swagger.dir` 不为空且该目录下存在 `x.swagger.json` 文件
- 访问 `http://127.0.0.1:9091/swagger-ui/`, 请求框中输入 `/swagger/x.swagger.json`

#### 目录说明

```text
.
├── cmd
│   ├── cr.go // 服务启动入口
│   └── init.go // 服务初始化相关
├── etc
│   └── conf.yaml // 默认服务配置
├── pkg
│   ├── action
│   │   ├── resource 资源操作相关封装
│   │   │   └── ...
|   │   └── util 工具类
│   │       └── ...
│   ├── cache // 缓存
│   │   ├── redis 缓存（redis）实现
│   │   │   └── ...
|   |   └── types.go 缓存相关类型
│   ├── cluster // 集群信息相关
│   │   └── ...
│   ├── common
│   │   ├── conf // 配置相关常量
│   │   │   └── ...
│   │   ├── ctxkey // ContextKey 常量
│   │   │   └── ...
│   │   ├── envs // 环境变量
│   │   │   └── ...
│   │   ├── errcode // 错误码常量
│   │   │   └── ...
│   │   ├── runmode // 运行模式常量
│   │   │   └── ...
│   │   ├── runtime // 运行时配置
│   │   │   └── ...
│   │   └── types // 通用类型
│   │       └── ...
│   ├── config // 服务配置
│   │   └── ...
│   ├── discovery // 服务发现
│   │   └── ...
│   ├── handler // 接口逻辑
│   │   ├── basic // 基础 API
│   │   │   └── ...
│   │   ├── config // 配置类资源 API
│   │   │   └── ...
│   │   ├── customresource // 自定义资源 API
│   │   │   └── ...
│   │   ├── hpa // hpa 资源 API
│   │   │   └── ...
│   │   ├── namespace // 命名空间资源 API
│   │   │   └── ...
│   │   ├── network // 网络类资源 API
│   │   │   └── ...
│   │   ├── rbac // 权限控制类资源 API
│   │   │   └── ...
│   │   ├── resource // K8S 资源相关 API
│   │   │   └── ...
│   │   ├── storage // 存储类资源 API
│   │   │   └── ...
│   │   └── workload // 工作负载类资源 API
│   │       └── ...
│   ├── i18n // 国际化相关
│   │   └── ...
│   ├── iam // 权限中心 V3 对接相关
│   │   ├── perm
│   │   │   ├── resource
│   │   │   │   ├── cluster // 集群（域）权限
│   │   │   │   │   └── ...
│   │   │   │   ├── namespace // 命名空间（域）权限
│   │   │   │   │   └── ...
│   │   │   │   └── project // 项目权限
│   │   │   │       └── ...
│   │   │   └── ... // 基础 client 实现等
│   │   └── ...
│   ├── logging // 日志组件
│   │   └── ...
│   ├── project // 项目信息相关
│   │   └── ...
│   ├── resource // client-go 相关封装
│   │   ├── client // Resource Client
│   │   │   └── ...
│   │   ├── example // 资源配置 Demo，参考文档等
│   │   │   ├── config // Demo 配置
│   │   │   │   └── ...
│   │   │   ├── manifest // Demo 示例
│   │   │   │   └── ...
│   │   │   └── reference // 参考文档
│   │   │       └── ...
│   │   ├── form // 表单化相关
│   │   │   ├── model // 表单模型
│   │   │   │   └── ...
│   │   │   ├── parser // k8s 资源解析器
│   │   │   │   └── ...
│   │   │   ├── renderer // manifest/schema 模板渲染器
│   │   │   │   └── ...
│   │   │   ├── tmpl // 表单化模版
│   │   │   │   ├── layout // schema 编排
│   │   │   │   │   └── ...
│   │   │   │   ├── manifest // k8s manifest
│   │   │   │   │   └── ...
│   │   │   │   └── schema // 表单 schema
│   │   │   │       └── ...
│   │   │   └── validator // 表单数据校验器
│   │   │       └── ...
│   │   ├── formatter // k8s 资源格式化方法
│   │   │   └── ...
│   │   ├── config.go // BCS Cluster Config
│   │   ├── constants.go // 集群资源等常量
│   │   └── discovery.go // Redis Discover 实现
│   ├── service // 服务层逻辑
│   │   ├── util
│   │   │   └── ...
│   │   └── ...
│   ├── util // 工具类
│   │   └── ...
│   ├── version // version 组件
│   │   └── ...
|   └── wrapper // 装饰器
|       └── ...
├── proto
│   └── cluster-resources
│       ├── ....pb.x.go // 由 .proto 生成，无须修改
│       └── cluster-resources.proto // RPC 接口定义
├── swagger
│   ├── data // 默认 swagger.json 文件存放目录，作文件服务
│   └── datafile.go // swagger-ui 编译结果
├── third_party // 第三方依赖（proto）
│   └── ...
├── Dockerfile
├── Dockerfile-Dev
├── generate.go
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── plugins.go
└── Readme.md
```

### 更多参考
[GoMicro 使用指引](https://github.com/Tencent/bk-bcs/blob/master/docs/specification/go-micro.md)