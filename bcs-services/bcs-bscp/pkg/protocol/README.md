## 1. Protobuf 依赖版本说明

| 依赖 | 版本 |
| ------ | ------ |
| protoc | 3.20.0 |
| protoc-gen-go | v1.28.0 |
| protoc-gen-go-grpc | v1.2 |
| protoc-gen-grpc-gateway | v1.16.0 |

## 2. 依赖安装说明文档：
### 2.1 Protoc 安装
1. 去 [Protocol Buffers 官网](https://github.com/protocolbuffers/protobuf/releases) 找到 Protocol Buffers v3.20.0 版本，根据设备类型下载对应的 protoc 安装包
2. 解压安装包
3. cp bin/protoc $GOPATH/bin

如果没有版本变动的话，以上3步已经完成 Protoc 安装。不需要去处理 include 文件夹中的文件，因为 pkg/thirdparty/google 已经存放了 bscp 需要的 Protoc 依赖的包。

### 2.2 其他依赖二进制下载
```shell
go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
```

## 3. 版本升级说明
### 3.1 protoc 版本升级
1. 需改此 README.md 文件中与 protoc 相关的版本信息
2. 用安装包中的 include/google/protobuf 文件覆盖 pkg/thirdparty/protobuf/google/protobuf 文件
3. 将安装包中的readme.txt文件复制到 pkg/thirdparty/protobuf/google/protobuf 文件下

### 3.2 protoc-gen-grpc-gateway 版本升级
1. 需改此 README.md 文件中与 protoc-gen-grpc-gateway 相关的版本信息
2. 用 $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/google/api 文件覆盖 pkg/thirdparty/protobuf/google/api 文件
3. 将 $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/google/README.grpc-gateway 复制到 pkg/thirdparty/protobuf/google/api 文件下
