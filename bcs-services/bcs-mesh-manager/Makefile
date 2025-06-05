REGISTRY = "mirrors.tencent.com"
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)

.PHONY: proto build test tidy lint fmt vendor fmttool push-image

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

proto:
	$(MAKE) -C proto/bcs-mesh-manager clean
	$(MAKE) -C proto/bcs-mesh-manager

build:
	go build -o bin/bcs-mesh-manager ./cmd/mesh-manager/main.go

# test
test:
	go test -v ./cmd/... ./internal/... -cover -count=1 -vet=off

# tidy
tidy:
	go mod tidy

# lint
lint:
	golangci-lint run --fix ./...

# vendor
vendor:
	go mod vendor

fmttool:
	go install -v github.com/incu6us/goimports-reviser/v3@latest

# format
fmt:fmttool
	goimports-reviser -rm-unused -project-name github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager -excludes ./third_party/,./proto/,./vendor  ./...

# use buildx build arm64/amd64 image and push
.PHONY: push-image
push-image: tidy fmt vendor
	docker buildx build --platform linux/amd64 --build-arg VERSION=$(VERSION) -f ./Dockerfile -t $(REPO)/bcs-mesh-manager:$(TAG) --push .
