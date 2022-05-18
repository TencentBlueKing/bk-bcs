#!/bin/bash
# 本地测试工具

function run_prometheus() {
    prometheus \
        --config.file=./etc/prometheus_dev.yml \
        --log.level=debug \
        --storage.tsdb.path=./data/prometheus \
        --storage.tsdb.min-block-duration=2h \
        --storage.tsdb.max-block-duration=2h \
        --web.listen-address=127.0.0.1:9090 \
        --web.enable-lifecycle
}

function run_api() {
    ./bin/bcs-monitor api \
        --config ./etc/dev_bcs-monitor.yml
}

function run_storegw() {
    ./bin/bcs-monitor storegw \
        --config ./etc/dev_bcs-monitor.yml
}

############
# Main Loop #
#############
echo "start run $1"
$1
