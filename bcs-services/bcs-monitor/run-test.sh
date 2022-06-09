#!/bin/bash
# 本地测试工具

function prometheus() {
    prometheus \
        --config.file=./etc/prometheus_dev.yml \
        --log.level=debug \
        --storage.tsdb.path=./data/prometheus \
        --storage.tsdb.min-block-duration=2h \
        --storage.tsdb.max-block-duration=2h \
        --web.listen-address=127.0.0.1:9090 \
        --web.enable-lifecycle
}

function api() {
    set -ex
    ./bin/bcs-monitor api \
        --http-address 0.0.0.0:11902 \
        --advertise-address clb:11902 \
        --config ./etc/config_dev.yaml
}

function query {
    ./bin/bcs-monitor query \
    --http-address 0.0.0.0:10902 \
    --advertise-address clb:11902 \
    --config ./etc/config_dev.yaml \
    --credential-config ./etc/credentials_dev.yaml \
    --credential-config ./etc/credentials_mgr_dev.yaml \
    --store 127.0.0.1:19901 \
    --store 127.0.0.1:11901
}

function storegw() {
    ./bin/bcs-monitor storegw \
        --config ./etc/config_dev.yaml
}

############
# Main Loop #
#############
echo "start run $1"
$1
