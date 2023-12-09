MAKE:=make
bcs_edition?=inner_edition

VERSION = v1.28.0
BUILDTIME = $(shell date '+%Y-%m-%d %T')
GITCOMMIT = $(shell git rev-parse HEAD)
GOVERSION = $(shell go version)

LDFLAG=-ldflags "-X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/version.BuildVersion=${VERSION}' \
 -X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/version.BcsBuildTime=${BUILDTIME}' \
 -X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/version.BcsGitHash=${GITCOMMIT}' \
 -X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/version.GoVersion=${GOVERSION}'"

.PHONY: bin
bin:
	mkdir -p ./bin
	go mod tidy && CGO_ENABLED=0 go build ${LDFLAG} -o bin/kubectl-bcs-data-manager ./main.go

.PHONY: clean
clean:
	rm -rf ./bin
	