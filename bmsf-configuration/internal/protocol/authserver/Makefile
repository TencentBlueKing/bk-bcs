PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

GWVER ?= $(shell grep -rn "grpc-gateway" ../../../go.mod | awk -F ' ' '{print $$3}')

.PHONY: all clean

all:$(OBJ)

$(OBJ):$(PROTO)

%.pb.go: %.proto
	@protoc --proto_path=. --proto_path=../../ --proto_path=$(GOENVPATH)/src --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/third_party/googleapis --proto_path=$(GOENVPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(GWVER)/ --go_out=plugins=grpc:. $<

clean:
	@rm -f $(OBJ)
