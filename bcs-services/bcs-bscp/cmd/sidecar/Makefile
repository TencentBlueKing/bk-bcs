SERVER = bk-bscp-sidecar
ARGS   = -c ./etc/sidecar.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) sidecar.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
