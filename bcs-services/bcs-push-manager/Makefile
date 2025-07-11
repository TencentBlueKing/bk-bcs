GOPATH:=$(shell go env GOPATH)

bcs_edition?=inner_edition

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse HEAD)
VERSION?=${GITTAG}-$(shell date +%y.%m.%d)
WORKSPACE=$(shell pwd)

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

# build path config
INNER_PACKAGEPATH = build/bcs-push-manager.${VERSION}

.PHONY: init
init:
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install go-micro.dev/v4/cmd/protoc-gen-micro@v4.7.0
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3

.PHONY: proto
proto:
	protoc -I./third_party/ --proto_path=. \
	 --validate_out=paths=source_relative,lang=go:. \
	 --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. \
	 --openapiv2_out=logtostderr=true:. \
	 --go-grpc_out=. \
	 --micro_out=. \
	 --go_out=:. proto/bcs-push-manager.proto

.PHONY: build
build:
	go build ${LDFLAG} -o bcs-thirdparty-service *.go

.PHONY: docker
docker:
	docker build . -t bcs-thirdparty-service:latest

.PHONY: client
client:
	GOOS=linux GOARCH=amd64 go build -o bcs-thirdparty-service main.go