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
BUILDDIR=../build/bcs-cluster-resources.${VERSION}

LDFLAGS="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.Version=${VERSION} \
	-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.GitCommit=${GITCOMMIT} \
	-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.BuildTime=${BUILDTIME}"

LGFLAGS4TEST="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime.RunMode=UnitTest"

proto:
	# --micro_out => pb.micro.go，--go_out => pb.go，--grpc-gateway_out => pb.gw.go --validate_out => pb.validate.go，--swagger_out => swagger.json
	protoc -I ./third_party/ --proto_path=. --micro_out=. --go_out=plugins=grpc:. \
		--grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. \
		--validate_out=lang=go:. --swagger_out=logtostderr=true:. proto/cluster-resources/cluster-resources.proto
	# move swagger.json to ./swagger/data for file server
	mv ./proto/cluster-resources/cluster-resources.swagger.json ./swagger/data/

tidy:
	go mod tidy

# build executable binary locally for develop
build: tidy
	go build -ldflags ${LDFLAGS} -o cluster-resources-service *.go

# compile third_party/swagger-ui to swagger/datafile.go，execute only once!
swagger: proto
	go-bindata --nocompress -pkg swagger -o swagger/datafile.go third_party/swagger-ui/...

# run unittest in ./pkg recursively
test: tidy
	go test -ldflags ${LGFLAGS4TEST} -v ./pkg/... -cover -count=1

# run unittest and generate coverage report (html)
testcov: tidy
	go test -ldflags ${LGFLAGS4TEST} -v ./pkg/... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# build docker image
docker:
	docker build \
	--build-arg VERSION=${VERSION} \
	--build-arg GITCOMMIT=${GITCOMMIT} \
	-t cluster-resources:${VERSION} \
	--no-cache .
