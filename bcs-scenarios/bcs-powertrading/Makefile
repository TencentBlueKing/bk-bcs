
GOPATH:=$(shell go env GOPATH)
OUTPUT=build/bcs.${VERSION}

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
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3
	go install github.com/go-micro/generator/cmd/protoc-gen-micro@latest


.PHONY: api
api:
	protoc -I./third_party/ --openapi_out=./proto/ --proto_path=. proto/bcs-powertrading.proto

.PHONY: proto
proto:
	protoc -I../../bcs-services/bcs-nodegroup-manager/third_party/ --proto_path=. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --go-grpc_out=.  --micro_out=. --go_out=:. proto/bcs-powertrading.proto

.PHONY: rs-proto
rs-proto:
	protoc -I../../bcs-services/bcs-nodegroup-manager/third_party/ --proto_path=./pkg/apis/resourcemgr/resourcemanagerv4 --validate_out=paths=source_relative,lang=go:./pkg/apis/resourcemgr/resourcemanagerv4 --grpc-gateway_out=allow_delete_body=true,register_func_suffix=Gw:./pkg/apis/resourcemgr/resourcemanagerv4 --go-grpc_out=./pkg/apis/resourcemgr/resourcemanagerv4  --micro_out=./pkg/apis/resourcemgr/resourcemanagerv4 --go_out=:./pkg/apis/resourcemgr/resourcemanagerv4 pkg/apis/resourcemgr/resourcemanagerv4/resourcemanager.proto


.PHONY: cm-proto
cm-proto:
	protoc -I../../bcs-services/bcs-nodegroup-manager/third_party/ --proto_path=./pkg/apis/clustermgr/clustermanagerv4 --validate_out=paths=source_relative,lang=go:./pkg/apis/clustermgr/clustermanagerv4 --grpc-gateway_out=allow_delete_body=true,register_func_suffix=Gw:./pkg/apis/clustermgr/clustermanagerv4 --go-grpc_out=./pkg/apis/clustermgr/clustermanagerv4  --micro_out=./pkg/apis/clustermgr/clustermanagerv4 --go_out=:./pkg/apis/clustermgr/clustermanagerv4 pkg/apis/clustermgr/clustermanagerv4/clustermanager.proto

.PHONY: build
build:
	mkdir -p ${OUTPUT}/bcs-powertrading/
	cp -r ./image/* ${OUTPUT}/bcs-powertrading/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-powertrading/bcs-powertrading *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t bcs-powertrading:latest

.PHONY: client
client:
	GOOS=linux GOARCH=amd64 go build -o kubectl-bcs-powertrading cmd/client/client.go