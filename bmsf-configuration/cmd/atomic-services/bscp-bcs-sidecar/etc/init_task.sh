#!/bin/bash

# log dir
if [[ -n ${BSCP_BCSSIDECAR_LOG_DIR} ]]; then
    export BSCP_BCSSIDECAR_LOG_DIR="${BSCP_BCSSIDECAR_LOG_DIR}/${BCS_POD_ID}"
    install -dv ${BSCP_BCSSIDECAR_LOG_DIR}
fi

# handle sigs.
trap 'killall bk-bscp-bcs-sidecar; exit' SIGTERM

# monitor bk-bscp-bcs-sidecar
while true
do
    num=`ps -ef | grep bk-bscp-bcs-sidecar | grep -v grep | wc -l`
    if [ $num == 0 ]; then
        cd /bk-bscp/
        /bk-bscp/bk-bscp-bcs-sidecar run $@ --configfile etc/sidecar.yaml &
    fi
    sleep 1 & wait $!
done
