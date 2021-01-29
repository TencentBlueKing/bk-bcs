TARGET  = libbkbscp-gseplugin.so

# debug build
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

# e.g: 0.0.10 -> libbkbscp-gseplugin.so.0.0.10
PLUGIN_TAG ?= debug

BINDIR = ${OUTPUT_DIR}/$(TARGET).$(PLUGIN_TAG)
LIBDIR = ${OUTPUT_DIR}/$(TARGET).$(PLUGIN_TAG)/lib/

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(TARGET).$(PLUGIN_TAG) ...\033[0m"
	go build -buildmode=c-shared -o $(LIBDIR)$(TARGET).$(PLUGIN_TAG) main.go
	@cp -r etc $(BINDIR)
	@echo -e "\e[34;1mBuild $(TARGET).$(PLUGIN_TAG) success!\n\033[0m"

clean:
	@rm -rf ${BINDIR} $(LOCALBUILD)
