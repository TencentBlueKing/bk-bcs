## BCS-Monitor


> 蓝鲸容器服务（Bcs）监视器，提供资源的查询

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
prometheus            2.41.0
redis                 7.0
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
# go run，若不指定 conf.yaml，则使用 ./bcs-monitor.yml
go run main.go --config xxx.yaml

# 或 执行二进制文件
./bcs-monitor --config xxx.yaml

# 启动prometheus服务
./prometheus --conf prometheus.yaml
# 启动redis
./redis-server redis.conf
# 启动etcd
nohup etcd --config-file=/data/etcd/etcd.conf.yaml 
```

#### 验证服务

- 使用bcs-monitor query指令
```shell script
$ ./bcs-monitor  query  --store="127.0.0.1:11901"
```

- 使用thanos加载sidecar及query
```shell script
$  ./thanos  query --grpc-address=localhost:12901 --http-address=localhost:12902 --store=127.0.0.1:11901
$  ./thanos sidecar --prometheus.url=http://localhost:9090/ --tsdb.path=/prometheus --grpc-address=localhost:11901  --http-address=localhost:11902
```

#### 查看 thanos query

- 访问 `http://127.0.0.1:12902/stores`

#### 结果说明

访问 `http://127.0.0.1:12902/stores` 显示sidecar

### 更多参考
Thanos使用 ：https://github.com/thanos-io/thanos

