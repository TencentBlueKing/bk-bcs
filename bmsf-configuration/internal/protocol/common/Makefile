PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

.PHONY: all clean

all:$(OBJ)

$(OBJ):$(PROTO)

%.pb.go: %.proto
	@protoc --proto_path=. --proto_path=$(GOENVPATH)/src/ --proto_path=$(GOENVPATH)/src/bk-bscp/vendor/github.com/grpc-ecosystem/grpc-gateway --go_out=plugins=grpc:. $<

clean:
	@rm -f $(OBJ)
