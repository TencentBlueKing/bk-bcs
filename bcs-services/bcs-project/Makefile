.PHONY: build test 

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

LDFLAGS="-X github.com/Tencent/bk-bcs/bcs-services/bcs-project/pkg/version.Version=${VERSION} \
	-X github.com/Tencent/bk-bcs/bcs-services/bcs-project/pkg/version.GitCommit=`git rev-parse HEAD` \
	-X github.com/Tencent/bk-bcs/bcs-services/bcs-project/pkg/version.BuildTime=${shell date +%Y-%m-%dT%I:%M:%S}"

# build service
build:
	go build -ldflags ${LDFLAGS} -o bcs-project-service ./main.go

# test
test:
	go test -v ./... -cover -count=1
