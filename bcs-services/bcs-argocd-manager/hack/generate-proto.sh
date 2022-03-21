#! /usr/bin/env bash

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE})/..; pwd)

export GO111MODULE=off

GOPATH=${GOPATH:-$(go env GOPATH)}

PACKAGES=(
    github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1
)

APIMACHINERY_PKGS=(
    k8s.io/apimachinery/pkg/util/intstr
    k8s.io/apimachinery/pkg/api/resource
    -k8s.io/apimachinery/pkg/apis/meta/v1
)

export GO111MODULE=on

go-to-protobuf \
    --go-header-file hack/boilerplate.go.txt \
    --packages $(IFS=, ; echo "${PACKAGES[*]}") \
    --apimachinery-packages=$(IFS=, ; echo "${APIMACHINERY_PKGS[*]}") \
    --proto-import=./vendor \
    --proto-import=./third-party/protoc-include \

PROTO_FILES=(
    ./pkg/sdk/instance/instance.proto
    ./pkg/sdk/project/project.proto
    ./pkg/sdk/plugin/plugin.proto
)
for i in ${PROTO_FILES[@]}; do
    protoc \
        -I ./third-party/protoc-include \
        -I ./pkg/apis/tkex/v1alpha1 \
        -I ./vendor \
        -I $GOPATH/src \
        -I . \
        --go_out=:$GOPATH/src \
        $i
done
for i in ${PROTO_FILES[@]}; do
    protoc \
        -I ./third-party/protoc-include \
        -I ./pkg/apis/tkex/v1alpha1 \
        -I ./vendor \
        -I $GOPATH/src \
        -I . \
        --micro_out=:$GOPATH/src \
        $i
done

