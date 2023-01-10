## BCS-Monitor


> 蓝鲸容器服务（Bcs）监视器，提供资源的查询

## 开发指南

### 环境准备
```text
Go                    1.17.5
prometheus            2.41.0
```

```shell script
# 默认安装在 $GOPATH/bin 下
export GO111MODULE=on
go mod tidy
```

### 常用操作

#### 生成可执行二进制

```shell script
make build
```

#### 启动服务

```shell script
# go run，若不指定 conf.yaml，则使用 ./bcs-monitor.yml
go run ./cmd/bcs-monitor/main.go  query  --store="127.0.0.1:11901" --config="./etc/bcs-monitor.yml" 

# 或 执行二进制文件
./cmd/bcs-monitor\bcs-monitor.exe  query  --store="127.0.0.1:11901" --config="./etc/bcs-monitor.yml" 
# 启动prometheus服务
./prometheus --conf prometheus.yaml 
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

