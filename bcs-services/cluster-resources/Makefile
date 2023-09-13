.PHONY: proto tidy build test swagger docker

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

ifdef GITCOMMIT
    GITCOMMIT=${GITCOMMIT}
else
    GITCOMMIT=$(shell git rev-parse HEAD)
endif

GOPATH=$(shell go env GOPATH)
BUILDTIME=${shell date +%Y-%m-%dT%I:%M:%S}

LDFLAGS="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.Version=${VERSION} \
	-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.GitCommit=${GITCOMMIT} \
	-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.BuildTime=${BUILDTIME}"

LGFLAGS4TEST="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime.RunMode=UnitTest"

LGFLAGS4DEV="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime.RunMode=Dev"

proto:
	$(MAKE) -C proto/cluster-resources

tidy:
	go mod tidy -compat=1.17

# build executable binary locally for develop
build: tidy
	go build -ldflags ${LDFLAGS} -o bcs-cluster-resources *.go

# build dev
dev: tidy
	go build -ldflags ${LGFLAGS4DEV} -o bcs-cluster-resources *.go

# run unittest in ./pkg recursively
test: tidy
	go test -ldflags ${LGFLAGS4TEST} -v ./pkg/... -cover -count=1

# run unittest and generate coverage report (html)
testcov: tidy
	go test -ldflags ${LGFLAGS4TEST} -v ./pkg/... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# build docker image
docker:
	docker build -f ./Dockerfile \
	--build-arg VERSION=${VERSION} \
	--build-arg GITCOMMIT=${GITCOMMIT} \
	-t cluster-resources:${VERSION} \
	--no-cache .

# build docker image for develop / test (include soruce code)
dockerdev:
	docker build -f ./Dockerfile-Dev \
	--build-arg VERSION=${VERSION} \
	--build-arg GITCOMMIT=${GITCOMMIT} \
	-t cluster-resources:${VERSION} \
	--no-cache .

# collect，merge and re-generate pkg/i18n/locale/lc_msg.yaml
gen-lc-msgs:
	python3 ./scripts/gen_lc_msgs.py

# auto translate pkg/i18n/locale/lc_msg.yaml
trans-lc-msgs:
	python3 ./scripts/translate_lc_msgs.py