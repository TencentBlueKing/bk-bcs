VERSION ?=  $(shell git describe --tags --match='v*' --dirty='.dirty')
REGISTRY = "mirrors.tencent.com"
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)
GITTAG ?= $(shell git describe --tags --match='v*' --dirty='.dirty')
BUILDTIME ?= $(shell date +%Y-%m-%dT%T%z)
GITHASH ?= $(shell git rev-parse HEAD)

.PHONY: build test push-image vendor tidy

# 构建变量
BINARY_NAME=bcs-mesh-proxy
DOCKER_IMAGE=bcs-mesh-proxy
DOCKER_TAG=latest

# 构建二进制文件
build:
	go build -o bin/$(BINARY_NAME) ./cmd/mesh-proxy

# 运行测试
test:
	go test ./...

# 检查依赖
deps:
	go mod download
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

tidy:
	go mod tidy

# vendor
vendor:
	go mod vendor

# use buildx build arm64/amd64 image and push
push-image: tidy fmt vendor
	docker buildx build --platform linux/amd64 --build-arg VERSION=$(VERSION) --build-arg LDFLAG=$(LDFLAG) -f ./Dockerfile -t $(REPO)/bcs-mesh-proxy:$(TAG) --push .
