#!/usr/bin/env bash
# 用途：构建并推送docker镜像

# 安全模式
set -euo pipefail

# 通用脚本框架变量
PROGRAM=$(basename "$0")
EXITCODE=0

ALL=1
HELM_MANAGER=0
VERSION=latest
PUSH=0
REGISTRY=docker.io
USERNAME=
PASSWORD=

cd $(dirname $0)
WORKING_DIR=$(pwd)
ROOT_DIR=${WORKING_DIR%/*}

usage () {
    cat <<EOF
用法:
    $PROGRAM [OPTIONS]...

            [ --helmmanager         [可选] 打包helmmanager镜像 ]
            [ -v, --version         [可选] 镜像版本tag, 默认latest ]
            [ -p, --push            [可选] 推送镜像到docker远程仓库，默认不推送 ]
            [ -r, --registry        [可选] docker仓库地址, 默认docker.io ]
            [ --username            [可选] docker仓库用户名 ]
            [ --password            [可选] docker仓库密码 ]
            [ -h, --help            [可选] 查看脚本帮助 ]
EOF
}

usage_and_exit () {
    usage
    exit "$1"
}

log () {
    echo "$@"
}

error () {
    echo "$@" 1>&2
    usage_and_exit 1
}

warning () {
    echo "$@" 1>&2
    EXITCODE=$((EXITCODE + 1))
}

# 解析命令行参数，长短混合模式
(( $# == 0 )) && usage_and_exit 1
while (( $# > 0 )); do
    case "$1" in
        --helmmanager )
            ALL=0
            HELM_MANAGER=1
            ;;
        -v | --version )
            shift
            VERSION=$1
            ;;
        -p | --push )
            PUSH=1
            ;;
        -r | --registry )
            shift
            REGISTRY=$1
            ;;
        --username )
            shift
            USERNAME=$1
            ;;
        --password )
            shift
            PASSWORD=$1
            ;;
        --help | -h | '-?' )
            usage_and_exit 0
            ;;
        -*)
            error "不可识别的参数: $1"
            ;;
        *)
            break
            ;;
    esac
    shift
done

if [[ $PUSH -eq 1 && -n "$USERNAME" ]] ; then
    docker login --username $USERNAME --password $PASSWORD $REGISTRY
    log "docker login成功"
fi

# 创建临时目录
mkdir -p $WORKING_DIR/tmp
tmp_dir=$WORKING_DIR/tmp
# 执行退出时自动清理tmp目录
trap 'rm -rf $tmp_dir' EXIT TERM

# 编译
log "编译service..."
cd $ROOT_DIR
export GO111MODULE=on
export PATH=$GOPATH/bin:$PATH
VERSION=$VERSION GOOS=linux GOARCH=amd64 make -j build
cd $WORKING_DIR

# 构建helm manager镜像
if [[ $ALL -eq 1 || $HELM_MANAGER -eq 1 ]] ; then
    log "构建helm manager镜像..."
    rm -rf tmp/*
    cp -rf bcs-helm-manager/* tmp/
    cp -rf $ROOT_DIR/../build/bcs.$VERSION/bcs-services/bcs-helm-manager/* tmp/
    docker build -f tmp/Dockerfile -t $REGISTRY/bcs/bcs-helm-manager:$VERSION tmp --no-cache --network=host
fi

echo "BUILD SUCCESSFUL!"

if [[ $PUSH -eq 1 ]]; then
    log "推送镜像到docker远程仓库"
    if [[ $ALL -eq 1 || $HELM_MANAGER -eq 1 ]] ; then
        docker push $REGISTRY/bcs/bcs-helm-manager:$VERSION
    fi
fi