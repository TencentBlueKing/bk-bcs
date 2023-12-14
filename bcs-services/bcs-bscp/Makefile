# version
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH   = $(shell git rev-parse HEAD)
VERSION   = $(shell echo ${ENV_BK_BSCP_VERSION})
DEBUG     = $(shell echo ${ENV_BK_BSCP_ENABLE_DEBUG})
PREFIX   ?= $(shell pwd)
GOBIN     = ${PREFIX}/bin/proto
PATH     := ${PREFIX}/bin/proto:${PATH}
# protoc v4.22.0
export PROTOC_VERSION=25.1
GOBUILD=CGO_ENABLED=0 go build -trimpath


# output directory for release package and version for command line
ifeq ("$(VERSION)", "")
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-bscp
	export LDVersionFLAG = "-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.BUILDTIME=${BUILDTIME} \
    	-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.GITHASH=${GITHASH} \
		-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.DEBUG=${DEBUG}"
else
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-bscp-${VERSION}
	export LDVersionFLAG = "-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.VERSION=${VERSION} \
    	-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.BUILDTIME=${BUILDTIME} \
    	-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.GITHASH=${GITHASH} \
    	-X github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version.DEBUG=${DEBUG}"
endif

include ./scripts/makefile/uname.mk

default: all

.PHONY: init
init:
	@echo Download protoc
	@mkdir -p ${PREFIX}/bin/proto
	@cd ${PREFIX}/bin/proto && \
		rm -rf protoc-*.zip* && \
		wget -q https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
		unzip -o protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
		mv -f bin/protoc . && \
		rm -rf protoc-*.zip* readme.txt bin include
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.18.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.18.1
	@echo Download gotext
	go install golang.org/x/text/cmd/gotext@v0.14.0

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17

pre:
	@echo -e "\e[34;1mBuilding...\n\033[0m"
	mkdir -p ${OUTPUT_DIR}


install: pre
	@echo -e "\e[34;1mPackaging Install Tools...\033[0m"
	mkdir -p ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/install/start_all.sh ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/install/stop_all.sh ${OUTPUT_DIR}/install/
	@echo -e "\e[34;1mPackaging Install Tools Done\n\033[0m"

api: pre
	@echo -e "\e[34;1mPackaging API Docs...\033[0m"
	@mkdir -p ${OUTPUT_DIR}/api/
	@mkdir -p ${OUTPUT_DIR}/api/api-server
	@cp -f api/api-docs/api-server/api/bk_apigw_resources.yml ${OUTPUT_DIR}/api/api-server
	@tar -czf ${OUTPUT_DIR}/api/api-server/zh.tgz -C api/api-docs/api-server/docs zh
	@mkdir -p ${OUTPUT_DIR}/api/feed-server
	@cp -f api/api-docs/feed-server/api/bk_apigw_resources.yml ${OUTPUT_DIR}/api/feed-server
	@tar -czf ${OUTPUT_DIR}/api/feed-server/zh.tgz -C api/api-docs/feed-server/docs zh
	@echo -e "\e[34;1mPackaging API Docs Done\n\033[0m"

pb:
	@echo -e "\e[34;1mFormat proto...\033[0m"
	@find pkg/protocol -type f -name "*.proto"|xargs clang-format -i --style="{BasedOnStyle: Google, ColumnLimit: 120}"
	@echo -e "\e[34;1mMake Protocol...\033[0m"
	@cd pkg/protocol && make clean && make
	@echo -e "\e[34;1mMake Protocol Done\n\033[0m"

.PHONY: gen
gen:
	@go run scripts/gen/main.go

.PHONY: lint
lint:
	@golangci-lint run --issues-exit-code=0 --fix --skip-dirs=test --skip-dirs=pkg/logs/glog

test: pre
	@cd test/suite && make && cp -rf suite-test ${OUTPUT_DIR}/ && rm -rf suite-test
	@cd test/benchmark && make && cp -rf bench-test ${OUTPUT_DIR}/ && rm -rf bench-test

unit-test: pre
	@echo -e "\e[34;1mExec Unit Test...\033[0m"
	@mkdir -p ${OUTPUT_DIR}/unit-test
	@cd cmd/sidecar/scheduler && go test -o ${OUTPUT_DIR}/unit-test/scheduler.test -c
	@echo -e "\e[34;1mExec Unit Test Success\n\033[0m"

mock: pre
	@cd ${PRO_DIR}/test/mock/repo && make

all: pre validate pb install test unit-test api mock build_bscp
	@echo -e "\e[34;1mBuild All Success!\n\033[0m"

server: validate api
	@echo -e "\e[34;1mMaking Server...\n\033[0m"
	@echo "version: ${VERSION}" > ${OUTPUT_DIR}/VERSION
	@cp -rf ${PRO_DIR}/server-CHANGELOG.md ${OUTPUT_DIR}
	@mkdir -p ${OUTPUT_DIR}/install/
	@mkdir -p ${OUTPUT_DIR}
	@mkdir -p ${OUTPUT_DIR}/etc
	@cd ${PRO_DIR}/cmd && make server
	@echo -e "\e[34;1mMake Server All Success!\n\033[0m"

sidecar: pre validate
	@echo -e "\e[34;1mMaking Sidecar...\n\033[0m"
	@echo "version: ${VERSION}" > ${OUTPUT_DIR}/VERSION
	@cp -rf ${PRO_DIR}/sidecar-CHANGELOG.md ${OUTPUT_DIR}
	@cd ${PRO_DIR}/cmd/sidecar && make
	@echo -e "\e[34;1mMake Sidecar Success!\n\033[0m"

validate:
	@if [ "$(VERSION)" != "" ];then \
		if [[ "$(VERSION)" =~ ^(v[0-9]+.[0-9]+.[0-9]+){1}(-alpha[0-9]+)? ]];then \
			echo "version: "$(VERSION); \
		else \
			exit 1; \
		fi \
  	fi

clean:
	@cd ${PRO_DIR}/cmd && make clean
	@rm -rf ${PRO_DIR}/build


.PHONY: openapi
openapi:
	@swag init --outputTypes go,json -g pkg/web/web.go --exclude ./
	@swag fmt -g pkg/web/web.go --exclude ./


.PHONY: build_bscp
build_bscp:
	@cd ${PRO_DIR}/cmd && make all

.PHONY: build_frontend
build_frontend:
	cd ui; npm install; npm run build

.PHONY: build_ui
build_ui:
	${GOBUILD} -ldflags ${LDVersionFLAG} -o bscp-ui ./cmd/ui

.PHONY: docker
docker:
	@docker build -t bk-bscp-hyper:latest .

.PHONY: i18n
i18n:
	@go generate ./pkg/i18n/translations/translations.go
	@cp ./pkg/i18n/translations/locales/zh/out.gotext.json ./pkg/i18n/translations/locales/zh/messages.gotext.json