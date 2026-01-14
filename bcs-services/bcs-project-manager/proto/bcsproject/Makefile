PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))
GWOBJ:=$(patsubst %.proto, %.pb.gw.go, $(PROTO))
SWAGGEROBJ:=$(patsubst %.proto, %.swagger.json, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

DIRECTORY := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
THIRD_PARTY := $(DIRECTORY)/../../third_party

.PHONY: all clean check-protoc

all: clean check-protoc $(OBJ) $(GWOBJ) $(SWAGGEROBJ) sync-to-bcsapi

$(OBJ):$(PROTO)

$(GWOBJ):$(PROTO)

$(SWAGGEROBJ):$(PROTO)

check-protoc:
	@which protoc > /dev/null || (echo "Error: protoc not found. Please install protobuf compiler"; exit 1)

%.pb.go: %.proto
	protoc -I$(THIRD_PARTY) --proto_path=. --go_out=plugins=grpc:. --validate_out=lang=go:. $<

%.pb.gw.go: %.proto
	protoc -I$(THIRD_PARTY) --proto_path=. --micro_out=. \
	--grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. $<

%.swagger.json: %.proto
	protoc -I$(THIRD_PARTY) --proto_path=. --swagger_out=allow_delete_body=true,logtostderr=true:. $<

clean:
	rm -f $(OBJ) $(GWOBJ) $(SWAGGEROBJ) *.pb.micro.go *.pb.validate.go

sync-to-bcsapi:
	cp -f *.pb.go *.pb.gw.go *.pb.micro.go *.pb.validate.go ../../../../bcs-common/pkg/bcsapi/bcsproject/
