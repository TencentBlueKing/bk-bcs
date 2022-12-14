# version
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
VERSION   = $(shell echo ${ENV_BK_BSCP_VERSION})
DEBUG     = $(shell echo ${ENV_BK_BSCP_ENABLE_DEBUG})
GOBUILD=CGO_ENABLED=0 go build -trimpath

# output directory for release package and version for command line
ifeq ("$(VERSION)", "")
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-bscp
	export LDVersionFLAG = "-X bscp.io/pkg/version.BUILDTIME=${BUILDTIME} \
		-X bscp.io/pkg/version.DEBUG=${DEBUG}"
else
	GITHASH   = $(shell git rev-parse HEAD)
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-bscp-${VERSION}
	export LDVersionFLAG = "-X bscp.io/pkg/version.VERSION=${VERSION} \
    	-X bscp.io/pkg/version.BUILDTIME=${BUILDTIME} \
    	-X bscp.io/pkg/version.GITHASH=${GITHASH} \
    	-X bscp.io/pkg/version.DEBUG=${DEBUG}"
endif

include ./scripts/makefile/uname.mk

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

default: all

pre:
	@echo -e "\e[34;1mBuilding...\n\033[0m"
	mkdir -p ${OUTPUT_DIR}


install: pre
	@echo -e "\e[34;1mPackaging Install Tools...\033[0m"
	mkdir -p ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/install/start_all.sh ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/install/stop_all.sh ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/sql/bscp.sql ${OUTPUT_DIR}/install/
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
	@echo -e "\e[34;1mMake Protocol...\033[0m"
	@cd pkg/protocol && make clean && make
	@echo -e "\e[34;1mMake Protocol Done\n\033[0m"

test: pre
	@cd test/suite && make && cp -rf suite-test ${OUTPUT_DIR}/ && rm -rf suite-test
	@cd test/benchmark && make && cp -rf bench-test ${OUTPUT_DIR}/ && rm -rf bench-test

unit-test: pre
	@echo -e "\e[34;1mExec Unit Test...\033[0m"
	@mkdir -p ${OUTPUT_DIR}/unit-test
	@cd cmd/sidecar/scheduler && go test -o ${OUTPUT_DIR}/unit-test/scheduler.test -c
	@echo -e "\e[34;1mExec Unit Test Success\n\033[0m"

all: pre validate pb install test unit-test api
	@cd ${PRO_DIR}/cmd && make
	@cd ${PRO_DIR}/test/mock/repo && make
	@echo -e "\e[34;1mBuild All Success!\n\033[0m"

server: validate api
	@echo -e "\e[34;1mMaking Server...\n\033[0m"
	@echo "version: ${VERSION}" > ${OUTPUT_DIR}/VERSION
	@cp -rf ${PRO_DIR}/server-CHANGELOG.md ${OUTPUT_DIR}
	@mkdir -p ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/sql/bscp.sql ${OUTPUT_DIR}/install/
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

.PHONY: build_bscp
build_bscp:
	@cd ${PRO_DIR}/cmd && make all

.PHONY: build_frontend
build_frontend:
	cd ui; npm install; npm run build

.PHONY: build_ui
build_ui:
	${GOBUILD} ${LDFLAG} -o bscp-ui ./cmd/ui

.PHONY: docker
docker:
	@docker build -t bcs-bscp:latest .