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

bcs-webconsole 中设置了一些默认的配置项，当此配置项配置文件不存在时生效；
但仍有一些配置项是必填的

```
# 必填配置项
web-console-image 
```

使用配置文件启动
```
./bcs-webconsole -f ./conf/conf.json
```

也可以修改某一个配置项
```
# 把web端口设置为8081
./bcs-webconsole --port=8081
```


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
./bcs-webconsole
```

Build a docker image
```
make docker
```
