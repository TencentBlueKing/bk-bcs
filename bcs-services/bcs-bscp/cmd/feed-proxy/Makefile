SERVER = bk-bscp-proxy
ARGS   = -c ./etc/feed_proxy.yaml --public-key ./etc/api_gw_public.key

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BIN) feed_proxy.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo -e "\e[34;1mPackaging $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(PKGBIN) feed_proxy.go
	@cp -f etc/feed_proxy.yaml ${PKGETC}/
	@echo -e "\e[34;1mPackage $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
