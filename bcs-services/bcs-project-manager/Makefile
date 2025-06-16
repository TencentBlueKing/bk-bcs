REGISTRY = "mirrors.tencent.com"
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)

.PHONY: proto build-migration build-main build test tidy

BCSAPI_PKG_PATH=../../bcs-common/pkg/bcsapi/bcsproject

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

LDFLAGS="-X github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/version.Version=${VERSION} \
	-X github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/version.GitCommit=`git rev-parse HEAD` \
	-X github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/version.BuildTime=${shell date +%Y-%m-%dT%I:%M:%S}"

proto:
	$(MAKE) -C proto/bcsproject clean
	$(MAKE) -C proto/bcsproject
	find proto/bcsproject -type f -name "*.pb.*" -exec cp {} "$(BCSAPI_PKG_PATH)" \;

# build migration
build-migration: tidy
	@export CGO_ENABLED=0; GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o bin/bcs-project-migration ./script/migrations/project/migrate.go
	@export CGO_ENABLED=0; GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o bin/bcs-variable-migration ./script/migrations/variable/migrate.go

# build service
build-main: tidy
	@export CGO_ENABLED=0; GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o bin/bcs-project-manager ./main.go

build: build-migration build-main

# test
test:
	go test -v ./cmd/... ./internal/... -cover -count=1 -vet=off

tidy:
	go mod tidy

# build image
docker: build
	docker build -t ${REPO}/bcs-project-manager:${TAG} .

# push image
push: docker
	docker push ${REPO}/bcs-project-manager:${TAG}