.PHONY: proto build test docker

GOPATH:=$(shell go env GOPATH)
MODIFY=Mproto/imports/api.proto=github.com/micro/go-micro/v2/api/proto

proto:
	# --micro_out => pb.micro.go，--go_out => pb.go，--grpc-gateway_out => pb.gw.go --validate_out => pb.validate.go，--swagger_out => swagger.json
	protoc -I ./third_party/ --proto_path=. --micro_out=${MODIFY}:. --go_out=plugins=grpc:. --grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. --validate_out=lang=go:. --swagger_out=logtostderr=true:. proto/cluster-resources/cluster-resources.proto
	# move swagger.json to ./swagger/data for file server
	mv ./proto/cluster-resources/cluster-resources.swagger.json ./swagger/data/

build: proto
	go build -o cluster-resources-service *.go

# compile third_party/swagger-ui to swagger/datafile.go，execute only once!
swagger: proto
	go-bindata --nocompress -pkg swagger -o swagger/datafile.go third_party/swagger-ui/...

# run unittest in ./pkg recursively
test:
	go test -v ./pkg/... -cover

docker:
	docker build . -t cluster-resources-service:latest
