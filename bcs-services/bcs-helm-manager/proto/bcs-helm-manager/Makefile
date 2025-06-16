PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))
GWOBJ:=$(patsubst %.proto, %.pb.gw.go, $(PROTO))
SWAGGEROBJ:=$(patsubst %.proto, %.swagger.json, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

DIRECTORY := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
PROTO_IMAGE := "proto"

.PHONY: all clean

all:docker $(OBJ) $(GWOBJ) $(SWAGGEROBJ) sync-to-bcsapi

$(OBJ):$(PROTO)

$(GWOBJ):$(PROTO)

$(SWAGGEROBJ):$(PROTO)

%.pb.go: %.proto
	docker run --rm \
		-v $(DIRECTORY)/../../third_party:/third_party \
		-v $(DIRECTORY):/proto \
		${PROTO_IMAGE} \
		protoc -I/third_party --proto_path=/proto --go_out=plugins=grpc:/proto --validate_out=lang=go:/proto $<

%.pb.gw.go: %.proto
	docker run --rm \
		-v $(DIRECTORY)/../../third_party:/third_party \
		-v $(DIRECTORY):/proto \
		${PROTO_IMAGE} \
		protoc -I/third_party --proto_path=/proto --micro_out=/proto \
	--grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:/proto $<

%.swagger.json: %.proto
	docker run --rm \
		-v $(DIRECTORY)/../../third_party:/third_party \
		-v $(DIRECTORY):/proto \
		${PROTO_IMAGE} \
		protoc -I/third_party --proto_path=/proto --swagger_out=allow_delete_body=true,logtostderr=true:/proto $<

clean:
	rm -f $(OBJ) $(GWOBJ) $(SWAGGEROBJ) *.pb.micro.go *.pb.validate.go

docker:
	@docker build -t ${PROTO_IMAGE} .

sync-to-bcsapi:
	cp -f *.pb.go *.pb.gw.go *.pb.micro.go *.pb.validate.go ../../../../bcs-common/pkg/bcsapi/helmmanager/