GOPATH:=$(shell go env GOPATH)

VERSION?=$(shell git describe --tags --always --dirty)
OUTPUT=$(shell pwd)/build/bcs.${VERSION}

# 依赖安装目录
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

PROTOC=$(LOCALBIN)/protoc
# protoc 版本锁定
PROTOC_VERSION:=25.2
PROTOC_GEN_GO_VERSION:=1.33.0
PROTOC_GEN_GO_GRPC_VERSION:=1.3.0

# protoc 平台检测
OS_NAME ?= $(shell uname -s)
ARCH_NAME ?= $(shell uname -m)
PROTOC_OS ?= linux
ifeq ($(OS_NAME), Darwin)
	PROTOC_OS = osx
endif
PROTOC_ARCH ?= x86_64
ifeq ($(ARCH_NAME), arm64)
	PROTOC_ARCH = aarch_64
endif
PROTOC_ZIP_NAME:=protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip

.PHONY: init
init: install-protoc install-protoc-plugins

.PHONY: install-protoc
install-protoc: $(LOCALBIN)
	test -s $(PROTOC) && $(PROTOC) --version | grep -q $(PROTOC_VERSION) || ( \
		GOBIN=$(LOCALBIN) wget -q https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP_NAME) && \
 		unzip -qo $(PROTOC_ZIP_NAME) -d ./protoc && \
 		mv -f ./protoc/bin/protoc $(PROTOC) && \
 		rm -rf ./protoc && \
 		rm -f $(PROTOC_ZIP_NAME) \
 	)

.PHONY: install-protoc-plugins
install-protoc-plugins: $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VERSION)
	GOBIN=$(LOCALBIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)
	GOBIN=$(LOCALBIN) go install go-micro.dev/v4/cmd/protoc-gen-micro@v4.6.0
	GOBIN=$(LOCALBIN) go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0
	GOBIN=$(LOCALBIN) go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
	GOBIN=$(LOCALBIN) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4

.PHONY: proto
proto:
	PATH="$(LOCALBIN):$$PATH" $(PROTOC) --proto_path=. \
    		--proto_path=../../bcs-services/bcs-cluster-manager/third_party \
    		--go_out=. \
    		--micro_out=. \
    		--go-grpc_out=. \
    		--validate_out=paths=source_relative,lang=go:. \
    		--grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --swagger_out=logtostderr=true:. \
    		proto/bcs-terraform-bkprovider.proto

.PHONY: update
update:
	go get -u

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build: tidy
	mkdir -p ${OUTPUT}/bcs-terraform-bkprovider
	cp -r images/* ${OUTPUT}/bcs-terraform-bkprovider/
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ${OUTPUT}/bcs-terraform-bkprovider/bcs-terraform-bkprovider main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker: build
	docker build -t bcs-terraform-bkprovider:${VERSION} ${OUTPUT}/bcs-terraform-bkprovider