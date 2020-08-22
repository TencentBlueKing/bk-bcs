GITTAG    = $(shell git describe --tags)
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH   = $(shell git rev-parse HEAD)
VERSION   = ${GITTAG}-$(shell date +%y.%m.%d)

# OUTPUT directory for release
export OUTPUT_DIR = ${PRO_DIR}/build/bk-bscp.${VERSION}

ifndef IMAGE_TAG
	export IMAGE_TAG=dev
endif

ifndef PLUGIN_TAG
	export PLUGIN_TAG=dev
endif

ifndef CLIENT_TAG
	export CLIENT_TAG=dev
endif

# version for command line
export LDVersionFLAG = "-X bk-bscp/internal/version.VERSION=${VERSION} \
	-X bk-bscp/internal/version.BUILDTIME=${BUILDTIME} \
	-X bk-bscp/internal/version.GITHASH=${GITHASH}"

export GO111MODULE=on

default: all

pre:
	@echo -e "\e[34;1mBuilding...\033[0m"
	mkdir -p ${OUTPUT_DIR}
	go vet ./...

all: pre
	@cd cmd && make

clean:
	@cd cmd && make clean
	@rm -rf ${PRO_DIR}/build

images: pre
	@cd cmd && make images

bscpimages: pre
	@cd cmd && make bscpimages
	@cp -r scripts/helm ${OUTPUT_DIR}
