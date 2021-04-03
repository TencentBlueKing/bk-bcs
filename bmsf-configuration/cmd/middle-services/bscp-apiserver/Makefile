SERVER = bk-bscp-apiserver
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

GOENVPATH = $(shell go env GOPATH)
PROTOCOL  = $(GOENVPATH)/src/bk-bscp/internal/protocol
SCRIPTS   = $(GOENVPATH)/src/bk-bscp/scripts

SWAGGEDIR = $(BINDIR)/swagger
SWAGGERUI = $(GOENVPATH)/src/bk-bscp/third_party/swagger-ui
APIDOCCFG = $(PROTOCOL)/configserver/configserver.swagger.json
APIDOCTPL = $(PROTOCOL)/templateserver/templateserver.swagger.json

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) main.go
	@cp -r etc $(BINDIR)
	@mkdir -p $(SWAGGEDIR) && cp -r $(SWAGGERUI) $(SWAGGEDIR)
	@cp -r $(APIDOCCFG) $(SWAGGEDIR)/config.swagger.json
	@cp -r $(APIDOCTPL) $(SWAGGEDIR)/template.swagger.json
	@sh $(SCRIPTS)/daemon-control/generator.sh '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
