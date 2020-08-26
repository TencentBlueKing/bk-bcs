#! /bin/bash

protoc -I./third_party/ --proto_path=. --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. proto/meshmanager/meshmanager.proto