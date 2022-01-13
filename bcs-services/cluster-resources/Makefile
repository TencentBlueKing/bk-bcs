.PHONY: proto build test docker

GOPATH:=$(shell go env GOPATH)
MODIFY=Mproto/imports/api.proto=github.com/micro/go-micro/v2/api/proto
# TODO Generate VERSION base on git commit id and build time
VERSION=latest
BUILD_DIR=../build/bcs-cluster-resources.${VERSION}

proto:
	# --micro_out => pb.micro.go，--go_out => pb.go，--grpc-gateway_out => pb.gw.go --validate_out => pb.validate.go，--swagger_out => swagger.json
	protoc -I ./third_party/ --proto_path=. --micro_out=${MODIFY}:. --go_out=plugins=grpc:. --grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. --validate_out=lang=go:. --swagger_out=logtostderr=true:. proto/cluster-resources/cluster-resources.proto
	# move swagger.json to ./swagger/data for file server
	mv ./proto/cluster-resources/cluster-resources.swagger.json ./swagger/data/

build:
	go build -o cluster-resources-service *.go

# compile third_party/swagger-ui to swagger/datafile.go，execute only once!
swagger: proto
	go-bindata --nocompress -pkg swagger -o swagger/datafile.go third_party/swagger-ui/...

# run unittest in ./pkg recursively
test:
	go test -v ./pkg/... -cover -count=1

docker:
	# build executable binary for linux, amd64
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}/cluster-resources-service *.go

	# move swagger dependency file to build dir
	mkdir -p ${BUILD_DIR}/swagger/data
	cp -R ./third_party/swagger-ui/* ${BUILD_DIR}/swagger/
	cp ./swagger/data/cluster-resources.swagger.json ${BUILD_DIR}/swagger/data/cluster-resources.swagger.json

	# move dockerfile and default config file to build dir
	cp ./Dockerfile ${BUILD_DIR}/Dockerfile
	cp ./conf.yaml ${BUILD_DIR}/conf.yaml

	# build docker image
	cd ${BUILD_DIR} && docker build . -t cluster-resources:${VERSION}
