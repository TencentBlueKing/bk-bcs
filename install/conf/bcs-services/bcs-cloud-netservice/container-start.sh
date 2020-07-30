#!/bin/bash

module="bcs-cloud-netservice"

cd /data/bcs/${module}
chmod +x ${module}

# ready to start
/data/bcs/${module}/${module} $@

# ./bcs-cloud-netservice --help
#   -a, --address string                 IP address to listen on for this service (default "127.0.0.1")
#       --alsologtostderr                log to standard error as well as files
#       --cloud-mode string              cloud mode, option [tencentcloud, aws]
#       --debug                          debug flag, open pprof
#       --external-ip string             external IP address to listen on for this service
#       --external-ipv6 string           external IPv6 address to listen on for this service
#       --external-port uint             external port to listen on for this service
#   -f, --file string                    json file with configuration
#   -h, --help                           show this help info
#       --insecure-address string        insecure IP address to listen on for this service
#       --insecure-port uint             insecure port to listen on for this service
#       --ip-clean-interval-minute int   interval minute for ip cleaner check interval, unit[minute] (default 10)
#       --ip-max-idle-minute int         max time for available ip before return to cloud; unit[minute] (default 1600)
#       --ipv6-address string            IPv6 address to listen on for this service
#       --kubeconfig string              kubeconfig for kubernetes apiserver
#       --log-backtrace-at string        when logging hits line file:N, emit a stack trace
#       --log-dir string                 If non-empty, write log files in this directory (default "./logs")
#       --log-max-num int                Max num of log file. The oldest will be removed if there is a extra file created. (default 10)
#       --log-max-size uint              Max size (MB) per log file. (default 500)
#       --logtostderr                    log to standard error instead of files
#       --metric-port uint               Port to listen on for metric (default 8081)
#   -p, --port uint                      Port to listen on for this service (default 8080)
#       --stderrthreshold string         logs at or above this threshold go to stderr (default "2")
#       --swagger-dir string             swagger dir
#       --v int32                        log level for V logs
#       --version                        show version infomation
#       --vmodule string                 comma-separated list of pattern=N settings for file-filtered logging