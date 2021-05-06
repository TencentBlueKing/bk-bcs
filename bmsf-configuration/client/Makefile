TARGET = bk-bscp-client

# debug build
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

# debug version
LDVersionFLAG ?= "-X bk-bscp/pkg/version.VERSION=DEBUG \
	-X bk-bscp/pkg/version.BUILDTIME=DEBUG \
	-X bk-bscp/pkg/version.GITHASH=DEBUG"

BINDIR = ${OUTPUT_DIR}/$(TARGET)
BIN    = $(BINDIR)/$(TARGET)

GOENVPATH = $(shell go env GOPATH)

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(TARGET)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) main.go
	@cp -r etc $(BINDIR)
	@echo -e "\e[34;1mBuild $(TARGET) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
