
GOPATH:=$(shell go env GOPATH)

bcs_edition?=inner_edition
imagetag?=latest

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

ifndef BUILDTIME 
	BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
  	GITHASH=$(shell git rev-parse HEAD)
	VERSION?=${GITTAG}-$(shell date +%y.%m.%d)
endif

ifndef LDFLAG
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
endif

# build path config
OUTPUT=build/bcs.${VERSION}
ifdef SCENARIOSPACKAGE
	OUTPUT=${SCENARIOSPACKAGE}
endif


.PHONY: init
init:
	go get -u google.golang.org/protobuf/proto
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@v4.6.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3

.PHONY: proto
proto:
	protoc --proto_path=. \
    		--proto_path=../../bcs-common/pkg/third_party \
    		--go_out=. \
    		--micro_out=. \
    		--go-grpc_out=. \
    		--validate_out=paths=source_relative,lang=go:. \
    		--grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --swagger_out=logtostderr=true:. \
    		proto/bcs-gitops-manager.proto
	protoc --proto_path=. \
			--proto_path=../../bcs-common/pkg/third_party \
			--go_out=. \
			--micro_out=. \
			--go-grpc_out=. \
			--validate_out=paths=source_relative,lang=go:. \
			--grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --swagger_out=logtostderr=true:. \
			proto/bcs-gitops-webhook.proto

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: manager
manager: tidy
	mkdir -p ${OUTPUT}/bcs-gitops-manager
	cp -r images/manager/* ${OUTPUT}/bcs-gitops-manager/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-gitops-manager/bcs-gitops-manager main.go

.PHONY: proxy
proxy: tidy
	mkdir -p ${OUTPUT}/bcs-gitops-proxy
	cp -r images/proxy/* ${OUTPUT}/bcs-gitops-proxy/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-gitops-proxy/bcs-gitops-proxy cmd/proxy/main.go

.PHONY: webhook
webhook: tidy
	mkdir -p ${OUTPUT}/bcs-gitops-webhook
	cp -r images/proxy/* ${OUTPUT}/bcs-gitops-webhook/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-gitops-webhook/bcs-gitops-webhook cmd/webhook/main.go

.PHONY: vaultplugin
vaultplugin: tidy
	mkdir -p ${OUTPUT}/bcs-gitops-vaultplugin-server
	cp -r images/vaultplugin-server/* ${OUTPUT}/bcs-gitops-vaultplugin-server/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-gitops-vaultplugin-server/bcs-gitops-vaultplugin-server cmd/vaultplugin-server/main.go

.PHONY: docker
docker: manager proxy
	cd ${OUTPUT}/bcs-gitops-manager/ && docker build -t bcs-gitops-manager:${imagetag} . && cd -
	cd ${OUTPUT}/bcs-gitops-proxy/ && docker build -t bcs-gitops-proxy:${imagetag} . && cd -

.PHONY: build
build: manager proxy

.PHONY: clean
clean:
	rm -rf ./build
