.PHONY: proto build test docker

GOPATH:=$(shell go env GOPATH)
MODIFY=Mproto/imports/api.proto=github.com/micro/go-micro/v2/api/proto

proto:
	protoc -I ./third_party/ --proto_path=. --micro_out=${MODIFY}:. --go_out=plugins=grpc:. --grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. --validate_out=lang=go:. proto/cluster-resources/cluster-resources.proto

build: proto
	go build -o cluster-resources-service *.go

test:
	go test -v ./... -cover

docker:
	docker build . -t cluster-resources-service:latest
