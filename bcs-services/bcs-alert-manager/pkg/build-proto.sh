#! /bin/bash

## 生成http gateway
protoc -I./third_party/ --proto_path=. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --go_out=plugins=grpc:. --validate_out=lang=go:. ./proto/alertmanager/alertmanager.proto
## 生成符合micro的grpc
protoc -I./third_party/ --proto_path=. --micro_out=. --go_out=plugins=grpc:. --validate_out=lang=go:. ./proto/alertmanager/alertmanager.proto
## 生成http swagger
protoc -I./third_party/ --proto_path=. --swagger_out=logtostderr=true:. --go_out=plugins=grpc:. --validate_out=lang=go:. ./proto/alertmanager/alertmanager.proto

