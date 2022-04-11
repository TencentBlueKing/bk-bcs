export GO111MODULE=on

go mod vendor -e

go install k8s.io/code-generator/cmd/go-to-protobuf
go install k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo
go install github.com/golang/protobuf/protoc-gen-go@v1.3.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0
go install go-micro.dev/v4/cmd/protoc-gen-micro@v4
go install sigs.k8s.io/controller-tools/cmd/controller-gen