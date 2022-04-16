GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags || echo "unknown version")
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go build -trimpath
BIN_DIR=_output

.PHONY: init
init:
	@go get -u google.golang.org/protobuf/proto
	@go install github.com/golang/protobuf/protoc-gen-go@latest
	@go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@latest

.PHONY: proto
proto:
	@protoc --proto_path=. --micro_out=. --go_out=:. proto/bcs-unified-apiserver.proto

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: build
build:
	${GOBUILD} -o ${BIN_DIR}/bcs-unified-apiserver ./cmd/bcs-unified-apiserver

.PHONY: run
run:
	./${BIN_DIR}/bcs-unified-apiserver

.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: docker
docker:
	@docker build -t bcs-unified-apiserver:latest .
