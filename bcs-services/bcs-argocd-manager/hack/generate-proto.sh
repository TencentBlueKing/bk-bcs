#! /usr/bin/env bash

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE})/..; pwd)

export GO111MODULE=off

GOPATH=${GOPATH:-$(go env GOPATH)}

PACKAGES=(
    github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1
)

APIMACHINERY_PKGS=(
    +k8s.io/apimachinery/pkg/util/intstr
    +k8s.io/apimachinery/pkg/api/resource
    +k8s.io/apimachinery/pkg/runtime/schema
    +k8s.io/apimachinery/pkg/runtime
    k8s.io/apimachinery/pkg/apis/meta/v1
    k8s.io/api/core/v1
)

export GO111MODULE=on

# generate proto files from go apis files
# pkg/apis/tkex/v1alpha1/xxx_types.go => pkg/apis/tkex/v1alpha1/generated.proto
go-to-protobuf \
    --go-header-file hack/boilerplate.go.txt \
    --packages $(IFS=, ; echo "${PACKAGES[*]}") \
    --apimachinery-packages=$(IFS=, ; echo "${APIMACHINERY_PKGS[*]}") \
    --proto-import=./vendor \
    --proto-import=./third_party/protoc-include \

# generate go files from proto files
# .pb.go | .pb.gw.go | .pb.micro.go
PROTO_FILES=(
    ./pkg/sdk/instance/instance.proto
    ./pkg/sdk/project/project.proto
    ./pkg/sdk/plugin/plugin.proto
)
for i in ${PROTO_FILES[@]}; do
    protoc \
        -I ./third_party/protoc-include/ \
        -I ./vendor/ \
        -I $GOPATH/src/ \
        -I . \
        --go_out=plugins=grpc:$GOPATH/src \
        --micro_out=:$GOPATH/src \
        --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:$GOPATH/src \
        --swagger_out=logtostderr=true:. \
        $i
done