PREFIX                  ?= $(shell pwd)/bin/
GO                      ?= go
FIRST_GOPATH            := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOBUILD=CGO_ENABLED=0 go build -trimpath

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

BUILDTIME=$(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse HEAD)
VERSION=${GITTAG}-$(shell date +%y.%m.%d)
bcs_edition?=inner_edition

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

.PHONY: build
build:
	${GOBUILD} ${LDFLAG} -o bin/bcs-platform-manager .

.PHONY: docker
docker:
	@docker build -t bcs-platform-manager:latest .
