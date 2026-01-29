
.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go install github.com/go-micro/generator/cmd/protoc-gen-micro@latest
	go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3
	go install github.com/envoyproxy/protoc-gen-validate@v0.6.7

.PHONY: api
api:
	protoc -I./third_party/ --proto_path=. --openapi_out=./pkg/proto/ pkg/proto/bcs-storage.proto

.PHONY: proto
proto:
	protoc -I./third_party/ --proto_path=. --validate_out=paths=source_relative,lang=go:. --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:./pkg/proto/ --swagger_out=logtostderr=true:. --go-grpc_out=./pkg/proto/  --micro_out=./pkg/proto/ --go_out=:./pkg/proto/ ./pkg/proto/bcs-storage.proto
