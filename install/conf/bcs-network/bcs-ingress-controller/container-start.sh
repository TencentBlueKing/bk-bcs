#!/bin/bash

module="bcs-ingress-controller"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [[ $BCS_CONFIG_TYPE == "render" ]]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#ready to start
/data/bcs/${module}/${module} $@

# Usage of ./bcs-ingress-controller:
#   -address string
#         address for controller (default "127.0.0.1")
#   -alsologtostderr
#         log to standard error as well as files
#   -cloud string
#         cloud mode for bcs network controller (default "tencentcloud")
#   -kubeconfig string
#         Paths to a kubeconfig. Only required if out-of-cluster.
#   -log_backtrace_at string
#         when logging hits line file:N, emit a stack trace
#   -log_dir string
#         If non-empty, write log files in this directory (default "./logs")
#   -log_max_num int
#         Max num of log file. (default 10)
#   -log_max_size uint
#         Max size (MB) per log file. (default 500)
#   -logtostderr
#         log to standard error instead of files
#   -master --kubeconfig
#         (Deprecated: switch to --kubeconfig) The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
#   -metric_port int
#         metric port for controller (default 8081)
#   -port int
#         por for controller (default 8080)
#   -stderrthreshold string
#         logs at or above this threshold go to stderr (default "2")
#   -v int
#         log level for V logs
#   -vmodule string
#         comma-separated list of pattern=N settings for file-filtered logging