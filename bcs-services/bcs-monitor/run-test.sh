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
    ./bin/bcs-monitor api \
        --config ./etc/config_dev.yaml
}

function query {
    ./bin/bcs-monitor query \
    --config ./etc/config_dev.yaml \
    --credential-config ./etc/credentials_dev.yaml \
    --credential-config ./etc/credentials_mgr_dev.yaml \
    --store 127.0.0.1:19901 \
    --store 127.0.0.1:1998
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
