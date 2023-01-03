SERVER = bk-bscp-apiserver
ARGS   = -c ./etc/api_server.yaml --public-key ./etc/api_gw_public.key

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BIN) api_server.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo -e "\e[34;1mPackaging $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(PKGBIN) api_server.go
	@cp -f etc/api_server.yaml ${PKGETC}/
	@cp -f etc/api_gw_public.key ${PKGETC}/apiserver_api_gw_public.key
	@echo -e "\e[34;1mPackage $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
