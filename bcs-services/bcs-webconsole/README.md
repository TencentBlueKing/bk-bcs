



# BcsWebconsole Service

This is the Bcs Webconsole service

Generated with

```
go install go-micro.dev/v4/cmd/micro@v4.5.0
micro new service bcs-webconsole
```

## Getting Started

- [Configuration](#configuration)
- [Dependencies](#dependencies)
- [Usage](#usage)

## Configuration

| 参数                    | 必填  | 说明                                    | 默认           |
| ----------------------- | ----- | --------------------------------------- | -------------- |
| address                 | false | http 服务注册地址                       | 127.0.01       |
| port                    | false | http 服务监听端口                       | 8080           |
| web-console-image       | true  | 镜像地址                                |                |
| kubeconfig              | false | .kube 配置路径                          |                |
| redis-address           | false | redis服务连接地址                       | 127.0.0.1:6379 |
| redis-password          | false | redis服务连接密码                       |                |
| redis-database          | false | Redis DB                                | 0              |
| redis-master-name       | false | redis master 名称， redis主从配置时生效 |                |
| redis-sentinel-password | false | redis master 名称，redis主从配置时生效  |                |
| redis-poolSize          | false | redis 连接池容量                        | 3000           |


## Dependencies

Micro services depend on service discovery. The default is multicast DNS, a zeroconf system.

In the event you need a resilient multi-host setup we recommend etcd.

```
# install etcd
brew install etcd

# run etcd
etcd
```

## Usage

A Makefile is included for convenience

Build the binary

```
make build
```

Run the service
```
./bcs-webconsole --bcs-conf=./etc/config.yaml -f ./conf/conf.json
```

Build a docker image
```
make docker
```
