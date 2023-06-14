PREFIX                  ?= $(shell pwd)/bin/
GO                      ?= go
FIRST_GOPATH            := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOHOSTOS                ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH              ?= $(shell $(GO) env GOHOSTARCH)
GOBUILD=CGO_ENABLED=0 go build -trimpath

ifeq (arm, $(GOHOSTARCH))
	GOHOSTARM ?= $(shell GOARM= $(GO) env GOARM)
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)v$(GOHOSTARM)
else
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)
endif

PROMU_VERSION           ?= 0.12.0
PROMU_URL               := https://github.com/prometheus/promu/releases/download/v$(PROMU_VERSION)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM).tar.gz

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

.PHONY: promu
promu:
	$(eval PROMU_TMP := $(shell mktemp -d))
	curl -s -L $(PROMU_URL) | tar -xvzf - -C $(PROMU_TMP)
	mkdir -p $(FIRST_GOPATH)/bin
	cp $(PROMU_TMP)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM)/promu $(FIRST_GOPATH)/bin/promu
	rm -r $(PROMU_TMP)

.PHONY: build
build:
	${GOBUILD} ${LDFLAG} -o bin/bcs-monitor ./cmd/bcs-monitor

.PHONY: test
test:
	@echo ">> test not support"

.PHONY: openapi
openapi:
	@swag init --outputTypes go,json --parseDependency -g pkg/api/routes.go --exclude ./
	@swag fmt -g pkg/api/routes.go --exclude ./

.PHONY: build_benchmark
build_benchmark:
	@echo ">> build benchmark"
	cd test/benchmark && go build -o ../../bin/benchmark

.PHONY: docker
docker:
	@docker build -t bcs-monitor:latest .
