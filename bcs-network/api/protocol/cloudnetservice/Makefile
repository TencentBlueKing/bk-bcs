PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))
GWOBJ:=$(patsubst %.proto, %.pb.gw.go, $(PROTO))
SWAGGEROBJ:=$(patsubst %.proto, %.swagger.json, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

.PHONY: all clean

all:$(OBJ) $(GWOBJ) $(SWAGGEROBJ)

$(OBJ):$(PROTO)

$(GWOBJ):$(PROTO)

$(SWAGGEROBJ):$(PROTO)

%.pb.go: %.proto
	@protoc -I../../../third_party --proto_path=. --proto_path=$(GOENVPATH)/src --go_out=plugins=grpc:. $<

%.pb.gw.go: %.proto
	@protoc -I../../../third_party --proto_path=. --proto_path=$(GOENVPATH)/src  --grpc-gateway_out=logtostderr=true:. $<

%.swagger.json: %.proto
	@protoc -I../../../third_party --proto_path=. --proto_path=$(GOENVPATH)/src  --swagger_out=logtostderr=true:. $<


clean:
	@rm -f $(OBJ) $(GWOBJ) $(SWAGGEROBJ)