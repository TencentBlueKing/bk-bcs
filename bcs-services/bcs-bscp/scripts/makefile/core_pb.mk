# Currently, Makefile is only compiled for proto files under pkg/protocol/core/xx/.
# If you adjust the file directory, Makefile will have a problem.

PROTO=$(wildcard ./*.proto)
VERSION=$(shell protoc --version)

OBJ:=$(patsubst %.proto, %.pb.go, $(PROTO))

all:
    ifeq ("$(VERSION)","libprotoc 22.0")
		@protoc --proto_path=. --proto_path=../../../../ --proto_path=../../../../pkg/thirdparty/protobuf/ --go_opt=paths=source_relative --go_out=. --go-grpc_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. *.proto
    else
		@echo "make pb failed, protoc version not 22.0"
		exit 1
    endif

clean:
	@rm -f $(OBJ)
