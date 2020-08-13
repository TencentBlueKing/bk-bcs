#!/bin/bash

module="bcs-cloud-netagent"

cd /data/bcs/${module}
chmod +x ${module}

# ready to start
/data/bcs/${module}/${module} $@

# ./bcs-cloud-netagent --help
#   -a, --address string                      IP address to listen on for this service (default "127.0.0.1")
#       --alsologtostderr                     log to standard error as well as files
#       --cloud-netservice-endpoints string   cloud netservice endpoints, split by comma
#       --cluster string                      cluster for bcs
#       --eni-mtu int                         the mtu of eni (default 1500)
#       --external-ip string                  external IP address to listen on for this service
#       --external-ipv6 string                external IPv6 address to listen on for this service
#       --external-port uint                  external port to listen on for this service
#   -f, --file string                         json file with configuration
#       --fixed-ip-workloads string           names of workloads that support fixed ip, split by comma, default[StatefulSet,GameStatefulSet] (default "StatefulSet,GameStatefulSet")
#   -h, --help                                show this help info
#       --ifaces string                       use ip of these network interfaces as node identity, split with comma or semicolon (default "eth1")
#       --insecure-address string             insecure IP address to listen on for this service
#       --insecure-port uint                  insecure port to listen on for this service
#       --ipv6-address string                 IPv6 address to listen on for this service
#       --kube-cachesync-timeout int          wait for kube cache sync timeout in seconds; (default 10) (default 10)
#       --kube-resync-peried int              resync interval for informer factory in seconds; (default 300) (default 1200)
#       --kubeconfig string                   kubeconfig for kube-apiserver, Only required if out-of-cluster.
#       --log-backtrace-at string             when logging hits line file:N, emit a stack trace
#       --log-dir string                      If non-empty, write log files in this directory (default "./logs")
#       --log-max-num int                     Max num of log file. The oldest will be removed if there is a extra file created. (default 10)
#       --log-max-size uint                   Max size (MB) per log file. (default 500)
#       --logtostderr                         log to standard error instead of files
#       --metric-port uint                    Port to listen on for metric (default 8081)
#   -p, --port uint                           Port to listen on for this service (default 8080)
#       --stderrthreshold string              logs at or above this threshold go to stderr (default "2")
#       --v int32                             log level for V logs
#       --version                             show version infomation
#       --vmodule string                      comma-separated list of pattern=N settings for file-filtered logging