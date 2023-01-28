#!/bin/sh

# app informations.
APPBIN="python"
APPARGS="-m SimpleHTTPServer 8001"
BINPATH="/usr/bin"

# start app.
start() {
    echo -n $"Start ${APPBIN}..."

    # make log-dir if not exists.
    if [[ ! -d "./log" ]];then
        mkdir ./log
    fi

    ${BINPATH}/${APPBIN} ${APPARGS} > ./log/std.log 2>&1 &
    echo
    return
}

# stop app.
stop() {
    pid=`ps -ef | grep ${APPBIN} | grep -v grep | grep -v tail | awk '{print $2}'`
    if [ -n "$pid" ];then
        kill $pid
    fi

    return
}

# show daemon status.
status() {
    ps -ef | grep -w "${APPBIN}" | grep -v grep | grep -v -w sh
}

# switch cmd.
case "$1" in
    start)
        status && exit 0
        $1
    ;;
    stop)
        status || exit 0
        $1
    ;;
    status)
        $1
    ;;
    *)
        echo $"Usage: $0 {start|stop|status}"
        exit 2
esac