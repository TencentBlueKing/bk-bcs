
.PHONY: init
init:
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install go install go-micro.dev/v4/cmd/protoc-gen-micro@v4.7.0
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3

.PHONY: api
api:
	protoc -I../../bcs-common/pkg/third_party/ --proto_path=. --openapi_out=./pkg/proto/ pkg/proto/bcs-storage.proto

.PHONY: proto
proto:
	protoc -I../../bcs-common/pkg/third_party/ --proto_path=. --validate_out=paths=source_relative,lang=go:. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:./pkg/proto/ --swagger_out=logtostderr=true:. --go-grpc_out=./pkg/proto/  --micro_out=./pkg/proto/ --go_out=:./pkg/proto/ ./pkg/proto/bcs-storage.proto
