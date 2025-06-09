REGISTRY = "mirrors.tencent.com"
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)

.PHONY: proto build test tidy lint fmt vendor fmttool proto-fmt push-image

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

# install fmt tool
fmttool:
	go install github.com/incu6us/goimports-reviser/v3@latest

# format
fmt:fmttool
	goimports-reviser -rm-unused -project-name github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager -excludes ./third_party/,./proto/,./vendor  ./...

# generate proto
proto:
	$(MAKE) -C proto/bcs-mesh-manager clean
	$(MAKE) -C proto/bcs-mesh-manager

# format proto
proto-fmt:
	clang-format -i proto/bcs-mesh-manager/bcs-mesh-manager.proto

# build
build:fmt
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


# use buildx build arm64/amd64 image and push
.PHONY: push-image
push-image: tidy fmt vendor
	docker buildx build --platform linux/amd64 --build-arg VERSION=$(VERSION) -f ./Dockerfile -t $(REPO)/bcs-mesh-manager:$(TAG) --push .
