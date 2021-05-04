TARGET = libbkbscp-gseplugin.so

# build
GITTAG      = $(shell git rev-parse --short HEAD)
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

# product kind for build package, e.g oa/ee/ce
PRODUCT_KIND ?= oa

# e.g: 0.0.10 -> libbkbscp-gseplugin.so.0.0.10
PLUGIN_TAG ?= ${GITTAG}-$(shell date +%y.%m.%d)

# debug version
LDVersionFLAG ?= "-X bk-bscp/pkg/version.VERSION=DEBUG \
	-X bk-bscp/pkg/version.BUILDTIME=DEBUG \
	-X bk-bscp/pkg/version.GITHASH=DEBUG"

BINDIR = ${OUTPUT_DIR}/$(TARGET).$(PLUGIN_TAG)
LIBDIR = ${OUTPUT_DIR}/$(TARGET).$(PLUGIN_TAG)/lib/

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(TARGET).$(PLUGIN_TAG) ...\033[0m"
	go build -ldflags ${LDVersionFLAG} -buildmode=c-shared -o $(LIBDIR)$(TARGET).$(PLUGIN_TAG) main.go
	@cp -r etc $(BINDIR)
	@echo ${PRODUCT_KIND}-${PLUGIN_TAG} > ${BINDIR}/VERSION
	@echo -e "\e[34;1mBuild $(TARGET).$(PLUGIN_TAG) success!\n\033[0m"

clean:
	@rm -rf ${BINDIR} $(LOCALBUILD)
