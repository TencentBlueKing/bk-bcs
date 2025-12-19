VERSION ?=  $(shell git describe --tags --match='v*' --dirty='.dirty')
REGISTRY ?= ""
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)

.PHONY: push
push:
	docker build -t $(REPO)/bcs-bk-apisix-gateway:$(TAG) . --push