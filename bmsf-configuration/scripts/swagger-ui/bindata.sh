#!/bin/sh

# PROJECT ROOT DIR
ROOTDIR=$GOPATH/src/bk-bscp

# SWAGGER UI EMBEDDED FILES
SWAGGERUI="third_party/swagger-ui/..."

# SOURCE CODE FOR EMBEDDED FILES
SOURCECODE="pkg/swagger-ui/data/datafile.go"

# BINDATA WITH ASSETFS
# go get github.com/elazarl/go-bindata-assetfs/...
BINDATA="go-bindata-assetfs"

# BINDATA FLAGS
BINDATA_FLAGS="--nocompress -pkg swaggerui"

# GEN SWAGGER UI EMBEDDED DATAFILES
$BINDATA $BINDATA_FLAGS -o $ROOTDIR/$SOURCECODE $ROOTDIR/$SWAGGERUI
