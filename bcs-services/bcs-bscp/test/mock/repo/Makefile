SERVER = bk-bscp-repomock
ARGS   = -c ./etc/repo_mock.yaml

LOCALBUILD  = $(PWD)/build
OUTPUT_DIR ?= $(LOCALBUILD)

BINDIR = ${OUTPUT_DIR}/$(SERVER)
BIN    = $(BINDIR)/$(SERVER)
SCRIPTS   = ../../../scripts

include ../../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -o $(BIN) repo_mock.go
	@mkdir -p ${BINDIR}/etc && cp -f etc/repo_mock.yaml ${BINDIR}/etc/
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
