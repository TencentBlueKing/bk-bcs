
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
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/go-micro/generator/cmd/protoc-gen-micro@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3


.PHONY: api
api:
	protoc -I ../../bcs-services/bcs-nodegroup-manager/third_party/ --openapi_out=./proto/ --proto_path=. proto/bcs-gitops-pre-check.proto

.PHONY: proto
proto:
	protoc -I ../../bcs-services/bcs-nodegroup-manager/third_party/ --proto_path=. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --go-grpc_out=.  --micro_out=. --go_out=:. proto/bcs-gitops-pre-check.proto


.PHONY: build
build:
	go build ${LDFLAG} -o bcs-gitops-pre-check *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t bcs-gitops-pre-check:latest

.PHONY: client
client:
	GOOS=linux GOARCH=amd64 go build -o kubectl-bcs-gitops-pre-check cmd/client/client.go