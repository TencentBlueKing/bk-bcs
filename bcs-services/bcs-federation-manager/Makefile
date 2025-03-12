
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

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientPwd=${bcs_zk_client_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.EncryptionKey=${bcs_encryption_key} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.ServerCertPwd=${bcs_server_cert_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.ClientCertPwd=${bcs_client_cert_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.LicenseServerClientCertPwd=${bcs_license_server_client_cert_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.BcsDefaultUser=${bcs_registry_default_user} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.BcsDefaultPasswd=${bcs_registry_default_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

# build path config
INNER_PACKAGEPATH=build/bcs.${VERSION}

.PHONY: init
init:
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install go-micro.dev/v4/cmd/protoc-gen-micro@v4.9.0
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3


.PHONY: api
api:
	protoc -I./third_party/ --openapi_out=./proto/ --proto_path=. proto/bcs-federation-manager.proto

.PHONY: proto
proto:
	protoc -I./third_party/ --experimental_allow_proto3_optional --proto_path=. --validate_out=paths=source_relative,lang=go:. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --openapiv2_out=logtostderr=true:.  --micro_out=. --go-grpc_out=. --go_out=. proto/bcs-federation-manager/bcs-federation-manager.proto

.PHONY: rs-proto
rs-proto:
	protoc -I./third_party/ --proto_path=./proto/bcs-federation-manager --validate_out=paths=source_relative,lang=go:./proto/bcs-federation-manager --grpc-gateway_out=allow_delete_body=true,register_func_suffix=Gw:./proto/bcs-federation-manager --go-grpc_out=./proto/bcs-federation-manager  --micro_out=./proto/bcs-federation-manager --go_out=:./proto/bcs-federation-manager ./proto/bcs-federation-manager/bcs-federation-manager.proto

.PHONY: build
build:
	go build ${LDFLAG} -o bcs-federation-manager *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t bcs-federation-manager:latest

.PHONY: client
client:
	GOOS=linux GOARCH=amd64 go build -o kubectl-bcs-federation-manager cmd/client/client.go