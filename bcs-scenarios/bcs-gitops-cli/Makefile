MAKE:=make
bcs_edition?=inner_edition

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

VERSION=${GITTAG}-$(shell date +%y.%m.%d)
BUILDTIME = $(shell date '+%Y-%m-%d %T')
GITCOMMIT = $(shell git rev-parse HEAD)
GOVERSION = $(shell go version)

LDFLAG=-ldflags "-s -w -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITCOMMIT}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.GoVersion=${GOVERSION}'"

.PHONY: bin
bin:
	mkdir -p ./bin
	go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o bin/bcs-gitops ./main.go

.PHONY: clean
clean:
	rm -rf ./bin
