SERVER = bk-bscp-client

# debug build
PWD         = $(shell pwd)
LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

# V1.0.1 -> bk-bscp-client-V1.0.1
CLIENT_TAG ?=debug

# debug version
LDVersionFLAG ?= "-X bk-bscp/internal/version.VERSION=DEBUG \
	-X bk-bscp/internal/version.BUILDTIME=DEBUG \
	-X bk-bscp/internal/version.GITHASH=DEBUG"

BINDIR = ${OUTPUT_DIR}/$(SERVER)-$(CLIENT_TAG)
BIN    = $(BINDIR)/$(SERVER)-$(CLIENT_TAG)

GOENVPATH = $(shell go env GOPATH)
PROTOCOL  = $(GOENVPATH)/src/bk-bscp/internal/protocol
SCRIPTS   = $(GOENVPATH)/src/bk-bscp/scripts
DOCS      = $(GOENVPATH)/src/bk-bscp/docs

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(SERVER)-$(CLIENT_TAG)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -gcflags "-N -l" -o $(BIN) main.go
	@cp -r etc/install.sh $(BINDIR)
	@cp -r etc $(BINDIR)
	@mkdir -p $(BINDIR)/docs
	@cp -r $(DOCS)/admin_book.md $(BINDIR)/docs
	@cp -r $(DOCS)/client_book.md $(BINDIR)/docs
	@rm $(BINDIR)/etc/install.sh
	@echo -e "\e[34;1mBuild $(SERVER)-$(CLIENT_TAG) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
