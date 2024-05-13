PREFIX                  ?= $(shell pwd)/bin/
GO                      ?= go
FIRST_GOPATH            := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOHOSTOS                ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH              ?= $(shell $(GO) env GOHOSTARCH)
BUILDTIME                = ${shell TZ=Asia/Shanghai date +%Y-%m-%dT%H:%M:%S%z}
GOBUILD                  = CGO_ENABLED=0 go build -trimpath

ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --tags --always --dirty="-dirty")
endif

GITHASH=$(shell git rev-parse HEAD)
VERSION=${GITTAG}-$(shell date +%y.%m.%d)
bcs_edition?=inner_edition

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

.PHONY: build_frontend
build_frontend:
	cd frontend; npm install; npm run build

.PHONY: build_ui
build_ui:
	${GOBUILD} ${LDFLAG} -o bcs-ui ./cmd/bcs-ui

.PHONY: docker
docker:
	docker build -f Dockerfile . -t bcs-ui
