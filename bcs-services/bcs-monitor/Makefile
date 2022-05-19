PREFIX                  ?= $(shell pwd)/bin/
GO                      ?= go
FIRST_GOPATH            := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOHOSTOS                ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH              ?= $(shell $(GO) env GOHOSTARCH)

ifeq (arm, $(GOHOSTARCH))
	GOHOSTARM ?= $(shell GOARM= $(GO) env GOARM)
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)v$(GOHOSTARM)
else
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)
endif

PROMU_VERSION           ?= 0.12.0
PROMU_URL               := https://github.com/prometheus/promu/releases/download/v$(PROMU_VERSION)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM).tar.gz

.PHONY: promu
promu:
	$(eval PROMU_TMP := $(shell mktemp -d))
	curl -s -L $(PROMU_URL) | tar -xvzf - -C $(PROMU_TMP)
	mkdir -p $(FIRST_GOPATH)/bin
	cp $(PROMU_TMP)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM)/promu $(FIRST_GOPATH)/bin/promu
	rm -r $(PROMU_TMP)

.PHONY: build
build:
ifeq (, $(shell which promu))
	@echo ">> installing promu"
	GO111MODULE=on GOOS= GOARCH= go install github.com/prometheus/promu
endif
	@echo ">> rebuilding binaries using promu"
	GO111MODULE=on promu build --prefix $(PREFIX)

.PHONY: test
test:
	@echo ">> test not support"

.PHONY: openapi
openapi:
	@swag init --outputTypes go,json -g pkg/api/routes.go --exclude ./
	@swag fmt -g pkg/api/routes.go --exclude ./

.PHONY: build_benchmark
build_benchmark:
	@echo ">> build benchmark"
	cd test/benchmark && go build -o ../../bin/benchmark
