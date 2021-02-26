PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))
GWOBJ:=$(patsubst %.proto, %.pb.gw.go, $(PROTO))
SWAGGEROBJ:=$(patsubst %.proto, %.swagger.json, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

GWVER ?= $(shell grep -rn "grpc-gateway" ../../../go.mod | awk -F ' ' '{print $$3}')

.PHONY: all clean

all:$(OBJ) $(GWOBJ) $(SWAGGEROBJ)

$(OBJ):$(PROTO)

$(GWOBJ):$(PROTO)

$(SWAGGEROBJ):$(PROTO)

%.pb.go: %.proto
	@protoc --proto_path=. --proto_path=../../ --proto_path=$(GOENVPATH)/src --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/third_party/googleapis --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/ --go_out=plugins=grpc:. $<

%.pb.gw.go: %.proto
	@protoc --proto_path=. --proto_path=../../ --proto_path=$(GOENVPATH)/src --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/third_party/googleapis --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/ --grpc-gateway_out=logtostderr=true:. $<

%.swagger.json: %.proto
	@protoc --proto_path=. --proto_path=../../ --proto_path=$(GOENVPATH)/src --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/third_party/googleapis/ --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/ --swagger_out=logtostderr=true:. $<
	@cp -rf ./templateserver.swagger.json ../../../api/swagger/template.swagger.json

clean:
	@rm -f $(OBJ) $(GWOBJ) $(SWAGGEROBJ)
