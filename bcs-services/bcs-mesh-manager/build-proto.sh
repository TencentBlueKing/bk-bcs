#! /bin/bash

## 生成http gateway
protoc -I./third_party/ --proto_path=. --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. proto/meshmanagerv1/meshmanager.proto
## 生成符合micro的grpc
protoc -I./third_party/ --proto_path=. --micro_out=. --go_out=plugins=grpc:. proto/meshmanager/meshmanager.proto
## 生成http swagger
protoc -I./third_party/ --proto_path=. --swagger_out=logtostderr=true:. --go_out=plugins=grpc:. proto/meshmanagerv1/meshmanager.proto

