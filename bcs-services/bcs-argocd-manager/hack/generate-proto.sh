#! /usr/bin/env bash

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE})/..; pwd)

export GO111MODULE=off

GOPATH=${GOPATH:-$(go env GOPATH)}

PACKAGE=github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1

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
    --packages ${PACKAGE} \
    --apimachinery-packages=$(IFS=, ; echo "${APIMACHINERY_PKGS[*]}") \
    --proto-import=./vendor \
    --proto-import=${GOPATH}/src \
    --proto-import=./third_party/protoc-include \
    --output-base "${PROJECT_ROOT}/generated"

cp ${PROJECT_ROOT}/generated/${PACKAGE}/* ${PROJECT_ROOT}/pkg/apis/tkex/v1alpha1

# generate go files from proto files
# .pb.go | .pb.gw.go | .pb.micro.go
PROTO_FILES=(
    instance
    project
    plugin
)
for i in ${PROTO_FILES[@]}; do
    protoc \
        -I ./third_party/protoc-include/ \
        -I ./vendor/ \
        -I ${GOPATH}/src/ \
        -I ${PROJECT_ROOT}/generated/ \
        -I . \
        --go_out=plugins=grpc:pkg/sdk/${i} \
        --micro_out=:pkg/sdk/${i} \
        --grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:pkg/sdk/${i} \
        --swagger_out=logtostderr=true:. \
        pkg/sdk/${i}/${i}.proto
done

protoc \
  -I ./plugins/proto/ \
  --go_out=./plugins/proto/ \
  ./plugins/proto/*.proto

rm -rf ${PROJECT_ROOT}/generated

echo "generate proto files success"