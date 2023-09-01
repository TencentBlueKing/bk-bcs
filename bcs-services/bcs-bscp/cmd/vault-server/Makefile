SERVER = bk-bscp-vaultserver
ARGS   = -c ./etc/vault_server.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BIN) vault_server.go
	cd vault && CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BINDIR)/vault main.go
	cd vault-sidecar && CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BINDIR)/vault-sidecar *.go
	cd vault-plugins && CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(BINDIR)/vault-plugins/bk-bscp-secret main.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo -e "\e[34;1mPackaging $(SERVER)...\033[0m"
	CGO_ENABLED=0 go build -trimpath -ldflags ${LDVersionFLAG} -o $(PKGBIN) vault_server.go
	@cp -f etc/vault_server.yaml ${PKGETC}/
	@echo -e "\e[34;1mPackage $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
