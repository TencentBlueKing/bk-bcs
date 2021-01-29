# go-micro使用指引

解决问题：

* 优化服务发现，保障体系基础能力
* 引入服务订阅能力，弥补Service服务能力短板，强化数据体系
* 规范各模块开发，强化各模块SDK能力
* 持续优化service层容器化

待细化问题：

* bcs-gateway-discovery两套服务发现机制兼容问题
* bcs-api-gateway服务映射兼容问题
* 服务发现体系切换问题

## 框架使用原则

* go-micro service实现需要使用grpc
* 对接grpc-gateway实现http转发
* 日志插件不限定，推荐BCS原有日志组件或者logrus
* 服务发现基于etcd，复用BCS Service层etcd集群
* 事件订阅限定：rabbitMQ、go-nats

## go-micro框架

* registry：服务发现抽象
  * etcd，zookeeper，dns，consul，kubenetes、gossip
* selector：负载均衡抽象
* store：kv存储抽象
  * consul，etcd，memcached，mysql，redis
* config：配置动态加载与监听
  * 支持文件，etcd，consul，vault，configmap
* broker：消息中间件
  * google、aws、kafka、rabbitMQ，redis
* tracer：链路追踪
  * zipkin，jaeger

## 使用时注意问题

* 使用版本：推荐2.9.x
* 版本问题：etcd与grpc版本相互影响

## 基础环境构建

统一版本：

* protoc： 3.12.3
* micro： v2.9.3
* go-micro：v2.9.1
* protoc-gen-go： v1.3.2
* protoc-gen-micro： v2.9.1
* protoc-grpc-gateway： v1.14.6
* protoc-gen-swagger： v1.14.6
* grpc: v1.26.0

**注意**，因为grpc与etcd版本存在冲突，protoc-gen-go不能超过1.3.2

grpc链接：[https://github.com/grpc/grpc-go](https://github.com/grpc/grpc-go)

安装micro命令，[下载链接](https://github.com/micro/micro/releases/download/v2.9.3/micro-v2.9.3-linux-amd64.tar.gz)

安装protobuf，[下载链接](https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip)

建议：

* protoc安装在/usr/local/bin
* google定义建议安装在项目的third_party目录

安装protoc-gen-go, protoc-gen-micro

```shell
#默认安装在$GOPATH/bin下
export GO111MODULE=on
go get -v github.com/micro/micro/v2/cmd/protoc-gen-micro@master
go get -v github.com/golang/protobuf/protoc-gen-go@v1.3.2
go get -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.14.6
go get -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.14.6

export PATH=$PATH:$GOPATH/bin
```

## 服务信息要求

* 服务名称注册注册 ${moduleName}.bkbcs.tencent.com，例如datamanager.bkbcs.tencent.com
* 默认以grpc方式暴露服务，模块需要集成grpc-gateway，gateway的端口为grpc端口+1
  * 除兼容老模块http接口部分，新增模块需要以grpc方式提供接口，兼容http接口部分并入grpc-gateway模块中
  * grpc服务以模块名称命名，例如datamanaer.DataManager
  * grpc-gateway暴露的http规则需要以模块名称为前缀，例如ip:port/datamanager/v1
* 基于默认命名规则，bcs-api-gateway提供转发规则
  * http转发：以/bcsapi/v4/${moduleName}/ => /$moduleName/
  * grpc转发为透传，例如/datamanager.DataManager/默认转发至服务发现datamanager.bkbcs.tencent.com的服务实例
* 未依赖go-micro开发模块，集成etcd服务发现时采用bcs-common/pkg/registry

## 项目初始化示例

```shell
micro new bcs-data-manager
Creating service go.micro.service.bcs-data-manager in bcs-data-manager

.
├── main.go
├── generate.go
├── plugin.go
├── handler
│   └── bcs-data-manager.go
├── subscriber
│   └── bcs-data-manager.go
├── proto
│   └── bcs-data-manager
│       └── bcs-data-manager.proto
├── Dockerfile
├── Makefile
├── README.md
├── .gitignore
└── go.mod
```

* proto：放置service协议定义
* subscriber：消息订阅相关实现，依赖proto定义，如无需使用直接删除
* handler：service对外提供服务接口实现，消息定义依赖proto定义

## 服务定义

基础定义调整：

* proto/bcs-data-manager下proto调整，package调整为bcsdatamanager
* 增加go_package信息定义：实际go package的名称为bcsdatamanager

```protoc
syntax = "proto3";

package datamanager;

option go_package = "proto/bcs-data-manager;datamanager";
```

定义数据与grpc服务

```protoc
service DataManager {
    rpc Call(Request) returns (Response) {}
    rpc Stream(StreamingRequest) returns (stream StreamingResponse) {}
    rpc PingPong(stream Ping) returns (stream Pong) {}
}
```

## http服务定义（可选）

如果需要对外提供http API接口，并集成至bcs-api-gateway，需要使用grpc-gateway屏蔽grpc接口，
对外提供http API接口。

grpc-gateway需要依赖google定义的[annotation](https://github.com/googleapis/googleapis)，
需要提前下载，建议放置在项目下的third_party目录下，生成gateway实现时需要引用。

```protoc
//grpc-gateway requirement
import "google/api/annotations.proto";

service DataManager {
    rpc Call(Request) returns (Response) {
        option (google.api.http) = {
            post: "/v1/hello"
            body: "*"
        };
    }
    rpc Stream(StreamingRequest) returns (stream StreamingResponse) {
        option (google.api.http) = {
            post: "/v1/stream"
            body: "*"
        };
    }
    rpc PingPong(stream Ping) returns (stream Pong) {
        option (google.api.http) = {
            post: "/v1/pingpong"
            body: "*"
        };
    }
}
```

gateway实现代码生成，建议合入Makefile中

```shell
protoc -I./third_party/ --proto_path=. --grpc-gateway_out=logtostderr=true:. --micro_out=. --go_out=plugins=grpc:. proto/bcs-data-manager/bcs-data-manager.proto
```

相对于非gateway版本，grpc-gateway了以下部分代码：

* bcs-data-manager.pb.go中增加了grpc原生client与Server的定义
* 增加了文件bcs-data-manager.pb.gw.go，用于实现http至grpc server的动态映射

## 命令行参数/配置文件

默认集成bcs-common/conf

## 日志使用

默认集成bcs-common/blog
