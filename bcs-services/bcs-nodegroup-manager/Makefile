
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
	go install go-micro.dev/v4/cmd/protoc-gen-micro@v4.7.0
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3


.PHONY: api
api:
	protoc -I./third_party/ --openapi_out=./proto/ --proto_path=. proto/bcs-nodegroup-manager.proto

.PHONY: proto
proto:
	protoc -I./third_party/ --proto_path=. --validate_out=paths=source_relative,lang=go:. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --swagger_out=logtostderr=true:. --go-grpc_out=.  --micro_out=. --go_out=:. proto/bcs-nodegroup-manager.proto

.PHONY: rs-proto
rs-proto:
	protoc -I./third_party/ --proto_path=./pkg/resourcemgr/proto --validate_out=paths=source_relative,lang=go:./pkg/resourcemgr/proto --grpc-gateway_out=allow_delete_body=true,register_func_suffix=Gw:./pkg/resourcemgr/proto --go-grpc_out=./pkg/resourcemgr/proto  --micro_out=./pkg/resourcemgr/proto --go_out=:./pkg/resourcemgr/proto pkg/resourcemgr/proto/bcs-resource-manager.proto

.PHONY: build
build:
	go build ${LDFLAG} -o bcs-nodegroup-manager *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t bcs-nodegroup-manager:latest

.PHONY: client
client:
	GOOS=linux GOARCH=amd64 go build -o kubectl-bcs-nodegroup-manager cmd/client/client.go


.PHONY: cm-proto
cm-proto:
	protoc -I./third_party/ --proto_path=./pkg/cluster/proto --validate_out=paths=source_relative,lang=go:./pkg/cluster/proto --grpc-gateway_out=allow_delete_body=true,register_func_suffix=Gw:./pkg/cluster/proto --go-grpc_out=./pkg/cluster/proto  --micro_out=./pkg/cluster/proto --go_out=:./pkg/cluster/proto pkg/cluster/proto/bcs-cluster-manager.proto