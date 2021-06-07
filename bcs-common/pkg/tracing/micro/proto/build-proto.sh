#! /bin/bash

## 生成符合micro的grpc
protoc --proto_path=. --micro_out=. --go_out=plugins=grpc:. ./greeter.proto