SERVER = bk-bscp-templateserver
ARGS   = run

# debug build
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

# debug version
LDVersionFLAG ?= "-X bk-bscp/pkg/version.VERSION=DEBUG \
	-X bk-bscp/pkg/version.BUILDTIME=DEBUG \
	-X bk-bscp/pkg/version.GITHASH=DEBUG"

BINDIR = ${OUTPUT_DIR}/$(SERVER)
BIN    = $(BINDIR)/$(SERVER)

PLUGIN_GOTPL       = gotemplate
PLUGIN_MAKO        = makotemplate
PLUGIN_DIR         = ${BINDIR}/plugin
PLUGIN_GOTPL_BIN   = ${PLUGIN_DIR}/${PLUGIN_GOTPL}
PLUGIN_MAKETPL_BIN = ${PLUGIN_DIR}/${PLUGIN_MAKO}

GOENVPATH = $(shell go env GOPATH)
SCRIPTS   = $(GOENVPATH)/src/bk-bscp/scripts

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) main.go
	@echo -e "\e[34;1mBuilding ${PLUGIN_GOTPL} engine...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o ${PLUGIN_GOTPL_BIN} ./plugin/gotemplate/main.go
	@cp -r etc $(BINDIR)
	@sh $(SCRIPTS)/daemon-control/generator.sh '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
