.PHONY: proto build test docker

ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always)
endif

GOPATH=$(shell go env GOPATH)
MODIFY=Mproto/imports/api.proto=github.com/micro/go-micro/v2/api/proto
BUILDTIME=${shell date +%Y-%m-%dT%I:%M:%S}
BUILDDIR=../build/bcs-cluster-resources.${VERSION}.${BUILDTIME}

LDFLAGS="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.Version=${VERSION} \
	-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.GitCommit=`git rev-parse HEAD` \
	-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.BuildTime=${BUILDTIME}"

LGFLAGS4TEST="-X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common.RunMode=UnitTest"

proto:
	# --micro_out => pb.micro.go，--go_out => pb.go，--grpc-gateway_out => pb.gw.go --validate_out => pb.validate.go，--swagger_out => swagger.json
	protoc -I ./third_party/ --proto_path=. --micro_out=${MODIFY}:. --go_out=plugins=grpc:. \
		--grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. \
		--validate_out=lang=go:. --swagger_out=logtostderr=true:. proto/cluster-resources/cluster-resources.proto
	# move swagger.json to ./swagger/data for file server
	mv ./proto/cluster-resources/cluster-resources.swagger.json ./swagger/data/

# build executable binary locally for develop
build:
	go build -ldflags ${LDFLAGS} -o cluster-resources-service *.go

# compile third_party/swagger-ui to swagger/datafile.go，execute only once!
swagger: proto
	go-bindata --nocompress -pkg swagger -o swagger/datafile.go third_party/swagger-ui/...

# run unittest in ./pkg recursively
test:
	go test -ldflags ${LGFLAGS4TEST} -v ./pkg/... -cover -count=1

docker:
	# build executable binary for linux, amd64
	GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ${BUILDDIR}/cluster-resources-service *.go

	# move swagger dependency file to build dir
	mkdir -p ${BUILDDIR}/swagger/data
	cp -R ./third_party/swagger-ui/* ${BUILDDIR}/swagger/
	cp ./swagger/data/cluster-resources.swagger.json ${BUILDDIR}/swagger/data/cluster-resources.swagger.json

	# move dockerfile and default config file to build dir
	cp ./Dockerfile ${BUILDDIR}/Dockerfile
	cp ./conf.yaml ${BUILDDIR}/conf.yaml

	# build docker image
	cd ${BUILDDIR} && docker build . -t cluster-resources:${VERSION}
