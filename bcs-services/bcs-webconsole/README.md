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
