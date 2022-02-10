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

| base_conf 参数                    | 必填  | 说明                                    | 默认           |
| ----------------------- | ----- | --------------------------------------- | -------------- |
| app_code                 | false |                        |        |
| app_secret                    | false |                        |            |
| time_zone       | false  | 时区                                |                | Asia/Shanghai
| langeuage_code              | false | 默认语言                          | zh-hans               |
| env              | false |                           |                | dev

| redis 参数                    | 必填  | 说明                                    | 默认           |
| ----------------------- | ----- | --------------------------------------- | -------------- |
| host                 | false | redis服务连接地址                       | 127.0.01       |
| port                    | false | http 服务监听端口                       | 6379           |
| password       | false  | redis服务连接密码                                |                |
| db              | false | Redis DB                          | 0               |

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
./bcs-webconsole --bcs-conf=./etc/config.yaml
```

Build a docker image

```
make docker
```
