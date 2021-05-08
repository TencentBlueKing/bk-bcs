# product kind for build package, e.g oa/ee/ce
PRODUCT_KIND ?= oa

# version
GITTAG    = $(shell git describe --tags --always)
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH   = $(shell git rev-parse HEAD)
VERSION   ?= ${GITTAG}-$(shell date +%y.%m.%d)

# output directory for release package
export OUTPUT_DIR = ${PRO_DIR}/build/bk-bscp.${VERSION}

# version for command line
export LDVersionFLAG = "-X bk-bscp/pkg/version.VERSION=${VERSION} \
	-X bk-bscp/pkg/version.BUILDTIME=${BUILDTIME} \
	-X bk-bscp/pkg/version.GITHASH=${GITHASH}"

export GO111MODULE=on

# NOLINTS is ignore golint check errors for protocols code, cgo export funcs.
# ISSUE: https://github.com/golang/lint/issues/186, there is no option to ignore it in golint cmd.
# bscp-gse-plugin includes some cgo funcs that have nonstandard comments.
NOLINTS = "protocol|bscp-gse-plugin"

# only support Linux/Darwin
UNAME := $(shell uname)

ifeq ($(UNAME),Linux)
	SEDCMD := sed
else ifeq ($(UNAME),Darwin)
	SEDCMD := gsed
else
$(error not support ${UNAME} build)
endif

default: all

pre:
	@echo -e "\e[34;1mBuilding...\n\033[0m"
	mkdir -p ${OUTPUT_DIR}

ver:
	@echo ${PRODUCT_KIND}-${VERSION} > ${OUTPUT_DIR}/VERSION
	@cp -rf ${PRO_DIR}/CHANGELOG.md ${OUTPUT_DIR}

ops:
	@echo -e "\e[34;1mPackaging Ops Tools...\033[0m"
	mkdir -p ${OUTPUT_DIR}/tools/
	@cd ${PRO_DIR}/tools/trace/ && make && mv bk-bscp-trace-tool ${OUTPUT_DIR}/tools/
	@cd ${PRO_DIR}/tools/healthz/ && make && mv bk-bscp-healthz-tool ${OUTPUT_DIR}/tools/
	@cd ${PRO_DIR}/tools/authorize/ && make && mv bk-bscp-authorize-tool ${OUTPUT_DIR}/tools/
	@cd ${PRO_DIR}/test/benchmark/ && make && mv bk-bscp-benchmark-tool ${OUTPUT_DIR}/tools/
	@cd ${PRO_DIR}/test/e2e/ && make && mv bk-bscp-e2e-testing ${OUTPUT_DIR}/tools/
	@echo -e "\e[34;1mPackaging Ops Tools Done\n\033[0m"

install:
	@echo -e "\e[34;1mPackaging Install Tools...\033[0m"
	mkdir -p ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/docs/fast_install.md ${OUTPUT_DIR}/install/README.md
	@cp -rf ${PRO_DIR}/scripts/install/* ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/sql/bscp.sql ${OUTPUT_DIR}/install/
	@${SEDCMD} -i 's/__PRODUCT_KIND__/${PRODUCT_KIND}/g' ${OUTPUT_DIR}/install/bscp.sql
	@echo -e "\e[34;1mPackaging Install Tools Done\n\033[0m"

support-files:
	@echo -e "\e[34;1mPackaging Support Files...\033[0m"
	mkdir -p ${OUTPUT_DIR}/support-files/sql
	mkdir -p ${OUTPUT_DIR}/support-files/templates
	@cp -rf ${PRO_DIR}/scripts/sql/patchs/*.sql ${OUTPUT_DIR}/support-files/sql/
	@sh ${PRO_DIR}/scripts/support-files/generate_templates.sh ${PRO_DIR}/cmd/ ${OUTPUT_DIR}/support-files/templates/
	@echo -e "\e[34;1mPackaging Support Files Done\n\033[0m"

lint:
	@echo -e "\e[34;1mCode Lint Checking...\033[0m"
	@find . -name '*.go' | grep -vE ${NOLINTS} | xargs -I {} golint {}
	@echo -e "\e[34;1mCode Lint Check Done\n\033[0m"

all: lint pre install support-files ops ver
	@cd ${PRO_DIR}/cmd && make
	@cd ${PRO_DIR}/client && make
	@echo -e "\e[34;1mBuild All Success!\n\033[0m"

server: lint pre install support-files ops ver
	@cd ${PRO_DIR}/cmd && make server
	@echo -e "\e[34;1mBuild Server Module Success!\n\033[0m"

sidecar: lint pre
	@cd ${PRO_DIR}/cmd/atomic-services/bscp-bcs-sidecar && make clean && make image
	@echo -e "\e[34;1mBuild BCS Sidecar Success!\n\033[0m"

plugin: lint pre
	@cd ${PRO_DIR}/cmd/atomic-services/bscp-gse-plugin && make clean && make
	@echo -e "\e[34;1mBuild GSE Plugin Success!\n\033[0m"

client: lint pre
	@cd ${PRO_DIR}/client && make clean && make
	@cp -rf ${PRO_DIR}/docs/client.md ${OUTPUT_DIR}/bk-bscp-client/README.md
	@echo -e "\e[34;1mBuild Command Line Client Success!\n\033[0m"

clean-ops:
	@cd ${PRO_DIR}/tools/trace/ && make clean
	@cd ${PRO_DIR}/tools/healthz/ && make clean
	@cd ${PRO_DIR}/tools/authorize/ && make clean
	@cd ${PRO_DIR}/test/benchmark/ && make clean
	@cd ${PRO_DIR}/test/e2e/ && make clean

clean: clean-ops
	@cd ${PRO_DIR}/cmd && make clean
	@cd ${PRO_DIR}/client && make clean
	@rm -rf ${PRO_DIR}/build
