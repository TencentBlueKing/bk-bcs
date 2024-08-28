SERVER = bk-bscp-ui
ARGS   = -c ./etc/bscp-ui.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BIN) .
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo -e "\e[34;1mPackaging $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(PKGBIN) .
	@cp -f etc/bscp-ui.yaml ${PKGETC}/
	@echo -e "\e[34;1mPackage $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
