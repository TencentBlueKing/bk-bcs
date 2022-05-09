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
	protoc -I../../third_party --proto_path=. --go_out=plugins=grpc:. --validate_out=lang=go:. $<
	# sed -i '' 's/json:"-"/json:"-" bson:"-"/g' bcs-helm-manager.pb.go
	# sed -i '' 's/json:"ID,omitempty"/json:"ID,omitempty" bson:"_id,omitempty"/g' bcs-helm-manager.pb.go

%.pb.gw.go: %.proto
	protoc -I../../third_party --proto_path=. --micro_out=. \
	--grpc-gateway_out=allow_delete_body=true,logtostderr=true,register_func_suffix=Gw:. $<

%.swagger.json: %.proto
	protoc -I../../third_party --proto_path=. --swagger_out=allow_delete_body=true,logtostderr=true:. $<

clean:
	rm -f $(OBJ) $(GWOBJ) $(SWAGGEROBJ) *.pb.micro.go *.pb.validate.go