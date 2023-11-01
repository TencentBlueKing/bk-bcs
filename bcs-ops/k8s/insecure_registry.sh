#!/bin/bash
set -euo pipefail

# 通用脚本框架变量
PROGRAM=$(basename "$0")

# 定义需要设置为免证书信任的registry地址
REGISTRIES=()
TIMESTMP=$(date +%s)
CRI_TYPE=""
ACTION=""

usage_and_exit() {
    cat <<EOF
	免证书信任的registry地址，docker 需要
Usage:
	$PROGRAM -c containerd -a docker.example.com docker.example2.com:8080
	$PROGRAM -c docker -d docker.example.com docker.example2.com:8080
    $PROGRAM [ -h --help -?  show usage ]
			 [ -a, --add add insecure registry]
			 [ -d, --del remove insecure registry]
             [ -c, --cri-type  support docker\containerd]
EOF
    exit "$1"
}

version() {
    echo "$PROGRAM version $VERSION"
}

while (($# > 0)); do
    case "$1" in
        -a | --add)
            shift
            if [[ -z $ACTION ]]; then
                ACTION="add"
            else
                echo "ACTION already define: ${ACTION}"
                usage_and_exit 1
            fi
            while (($# > 0)) && [[ "$1" != -* ]]; do
                REGISTRIES+=("$1")
                shift
            done
            continue
            ;;
        -d | --del)
            shift
            if [[ -z $ACTION ]]; then
                ACTION="del"
            else
                echo "ACTION already define: ${ACTION}"
                usage_and_exit 1
            fi
            while (($# > 0)) && [[ "$1" != -* ]]; do
                REGISTRIES+=("$1")
                shift
            done
            continue
            ;;
        -c | --cri-type)
            shift
            CRI_TYPE=$1
            ;;
        --help | -h | '-?')
            usage_and_exit 0
            ;;
        -*)
            error "不可识别的参数: $1"
            ;;
        *)
            break
            ;;
    esac
    (($# > 0)) && shift
done

add_docker() {
    # 获取docker配置文件路径
    DOCKER_CONFIG_PATH="/etc/docker/daemon.json"

    # 文件不存在，则需要创建
    if [[ ! -f "$DOCKER_CONFIG_PATH" ]]; then
        echo "{}" >"$DOCKER_CONFIG_PATH"
    fi

    cp $DOCKER_CONFIG_PATH $DOCKER_CONFIG_PATH.registry.tmp

    registries=$(printf '"%s",' "${REGISTRIES[@]}")
    registries="[${registries%,}]"

    jq --arg k 'insecure-registries' --argjson v "$registries" '.[$k] as $insecure_registries | if $insecure_registries then reduce $v[] as $r (.; if $insecure_registries | index($r) == null then .[$k] += [$r] else . end) else .[$k] = $v end' $DOCKER_CONFIG_PATH >/tmp/docker_daemon-"${TIMESTMP}".tmp

    cp "$DOCKER_CONFIG_PATH" "$DOCKER_CONFIG_PATH.${TIMESTMP}.bak"
    mv /tmp/docker_daemon-"${TIMESTMP}".tmp "$DOCKER_CONFIG_PATH"
    cat "$DOCKER_CONFIG_PATH"

    # 重启docker服务
    systemctl reload docker
}

del_docker() {
    DOCKER_CONFIG_PATH="/etc/docker/daemon.json"

    if [[ ! -f "$DOCKER_CONFIG_PATH" ]]; then
        echo "{}" >"$DOCKER_CONFIG_PATH"
        cat $DOCKER_CONFIG_PATH
        return 0
    fi

    registries=$(printf '"%s",' "${REGISTRIES[@]}")
    registries="[${registries%,}]"

    jq --arg k 'insecure-registries' --argjson v "$registries" '.[$k] as $insecure_registries | if $insecure_registries then reduce $v[] as $r (.; if $insecure_registries | index($r) != null then .[$k] -= [$r] else . end) else .[$k] = $v end' $DOCKER_CONFIG_PATH >/tmp/docker_daemon-"${TIMESTMP}".tmp

    cp "$DOCKER_CONFIG_PATH" "$DOCKER_CONFIG_PATH.$TIMESTMP.bak"
    mv /tmp/docker_daemon-"${TIMESTMP}".tmp "$DOCKER_CONFIG_PATH"
    cat "$DOCKER_CONFIG_PATH"

    systemctl reload docker
}

add_containerd() {
    local registry
    for registry in "${REGISTRIES[@]}"; do
        CONTAINERD_HOST_DIR="/etc/containerd/certs.d/${registry}"
        mkdir -p "$CONTAINERD_HOST_DIR"
        if [[ -f $CONTAINERD_HOST_DIR/hosts.toml ]]; then
            cp "$CONTAINERD_HOST_DIR/hosts.toml" "$CONTAINERD_HOST_DIR/hosts.toml.${TIMESTMP}.bak"
        fi
        cat <<EOF >"$CONTAINERD_HOST_DIR/hosts.toml"
[host."https://$registry"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
EOF
    done
}

del_containerd() {
    local registry
    for registry in "${REGISTRIES[@]}"; do
        CONTAINERD_HOST_DIR="/etc/containerd/certs.d/${registry}"
        if [[ -f $CONTAINERD_HOST_DIR/host.toml ]]; then
            if grep -q "skip_verify = true" "$CONTAINERD_HOST_DIR"/host.toml; then
                cp "$CONTAINERD_HOST_DIR/hosts.toml" "$CONTAINERD_HOST_DIR/hosts.toml.${TIMESTMP}.bak"
                sed -i '/skip_verify = true/d' "$CONTAINERD_HOST_DIR"/host.toml
            fi
        fi
    done
}

"${ACTION}_${CRI_TYPE}"
