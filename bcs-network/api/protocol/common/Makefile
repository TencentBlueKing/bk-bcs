PROTO=$(wildcard ./*.proto)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))

GOENVPATH = $(shell go env GOPATH)

.PHONY: all clean

all:$(OBJ)

$(OBJ):$(PROTO)

%.pb.go: %.proto
	@protoc --proto_path=. --go_out=plugins=grpc:. $<

clean:
	@rm -f $(OBJ)