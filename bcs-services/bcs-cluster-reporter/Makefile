COMMIT_ID = $(shell  git rev-parse --short HEAD)
VERSION ?=  $(shell git describe --tags --match='v*' --dirty='.dirty')
REGISTRY = "mirrors.tencent.com"
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)-$(COMMIT_ID)
export REPO_ROOT := $(shell git rev-parse --show-toplevel)

default: build

build:
	go build -o ./bin/bcs-cluster-reporter ./cmd/reporter/main.go
	go build -o ./bin/bcs-nodeagent ./cmd/nodeagent/main.go

push:
	docker build -t $(REPO):$(TAG) .
	docker push $(REPO):$(TAG)

go-gen:
	@mkdir -p /tmp/bin
	@PATH="${PATH}":/tmp/bin go generate ./...