# debug build
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

# version for command line
LDVersionFLAG ?= "-X github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version.BUILDTIME=${BUILDTIME}"

BINDIR = ${OUTPUT_DIR}/$(SERVER)
BIN    = $(BINDIR)/$(SERVER)
PKGBINDIR = ${OUTPUT_DIR}/bin
PKGBIN = ${OUTPUT_DIR}/bin/$(SERVER)
PKGETC = ${OUTPUT_DIR}/etc
PKGINSTALL = ${OUTPUT_DIR}/install
SCRIPTS   = ../../scripts

export GO111MODULE=on
