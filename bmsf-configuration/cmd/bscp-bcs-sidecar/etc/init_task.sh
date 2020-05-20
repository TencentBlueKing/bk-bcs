#!/bin/bash

# monitor bk-bscp-bcs-sidecar
while true
do
    num=`ps -ef | grep bk-bscp-bcs-sidecar | grep -v grep | wc -l`
    if [ $num == 0 ]; then
        cd /bk-bscp/
        /bk-bscp/bk-bscp-bcs-sidecar run $@ --configfile etc/sidecar.yaml &
    fi
    sleep 10
done
