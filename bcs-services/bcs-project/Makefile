.PHONY: proto build test tidy

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

LDFLAGS="-X github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/version.Version=${VERSION} \
	-X github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/version.GitCommit=`git rev-parse HEAD` \
	-X github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/version.BuildTime=${shell date +%Y-%m-%dT%I:%M:%S}"

proto:
	$(MAKE) -C proto/bcsproject clean
	$(MAKE) -C proto/bcsproject

# build service
build: tidy
	go build -ldflags ${LDFLAGS} -o bcs-project-service ./main.go

# test
test:
	go test -v ./cmd/... ./internal/... -cover -count=1 -vet=off

tidy:
	go mod tidy

# build image
docker:
	go mod tidy
