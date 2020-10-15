SERVER = bk-bscp-bcs-sidecar
ARGS   = run --configfile ./etc/sidecar.yaml

# debug build
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

# 1.0.4 -> bcs/bscp-bcs-sidecar:1.0.4
IMAGE_TAG ?=debug

# debug version
LDVersionFLAG ?= "-X bk-bscp/internal/version.VERSION=DEBUG \
	-X bk-bscp/internal/version.BUILDTIME=DEBUG \
	-X bk-bscp/internal/version.GITHASH=DEBUG"

BINDIR = ${OUTPUT_DIR}/$(SERVER)
BIN    = $(BINDIR)/$(SERVER)

GOENVPATH = $(shell go env GOPATH)
PROTOCOL  = $(GOENVPATH)/src/bk-bscp/internal/protocol
SCRIPTS   = $(GOENVPATH)/src/bk-bscp/scripts

SWAGGEDIR = $(BINDIR)/swagger
SWAGGERUI = $(GOENVPATH)/src/bk-bscp/third_party/swagger-ui
APIDOCS   = $(PROTOCOL)/sidecar/sidecar.swagger.json

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -gcflags "-N -l" -o $(BIN) main.go
	@cp -r etc $(BINDIR)
	@mkdir -p $(SWAGGEDIR) && cp -r $(SWAGGERUI) $(SWAGGEDIR) && cp -r $(APIDOCS) $(SWAGGEDIR)/api.swagger.json
	@sh $(SCRIPTS)/daemon-control/generator.sh '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)

image: default
	cd $(BINDIR) && docker build -t bcs/bscp-bcs-sidecar:$(IMAGE_TAG) -f etc/image/Dockerfile . && docker save -o bscp-bcs-sidecar-$(IMAGE_TAG).tar bcs/bscp-bcs-sidecar:$(IMAGE_TAG)
