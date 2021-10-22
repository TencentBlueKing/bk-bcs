#!/bin/bash 

cd /data/bcs/bcs-gamedeployment-operator
chmod +x bcs-gamedeployment-operator
#start operator
./bcs-gamedeployment-operator --v=5

#Usage of ./bcs-gamestatefulset-operator:
#  -add_dir_header
#        If true, adds the file directory to the header
#  -alsologtostderr
#        log to standard error as well as files
#  -kubeConfig string
#        Path to a kubeConfig. Only required if out-of-cluster.
#  -leader-elect
#        Enable leader election (default true)
#  -leader-elect-componentname string
#        The component name for event resource (default "gamestatefulset")
#  -leader-elect-lease-duration duration
#        The leader-elect LeaseDuration (default 15s)
#  -leader-elect-name string
#        The resourcelock name (default "gamestatefulset")
#  -leader-elect-namespace string
#        The resourcelock namespace (default "bcs-system")
#  -leader-elect-renew-deadline duration
#        The leader-elect RenewDeadline (default 10s)
#  -leader-elect-retry-period duration
#        The leader-elect RetryPeriod (default 2s)
#  -log_backtrace_at value
#        when logging hits line file:N, emit a stack trace
#  -log_dir string
#        If non-empty, write log files in this directory
#  -log_file string
#        If non-empty, use this log file
#  -log_file_max_size uint
#        Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
#  -logtostderr
#        log to standard error instead of files (default true)
#  -master string
#        The address of the Kubernetes API server. Overrides any value in kubeConfig. Only required if out-of-cluster.
#  -min-resync-period duration
#        The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod.
#  -skip_headers
#        If true, avoid header prefixes in the log messages
#  -skip_log_headers
#        If true, avoid headers when opening log files
#  -stderrthreshold value
#        logs at or above this threshold go to stderr (default 2)
#  -v value
#        number for the log level verbosity
#  -vmodule value
#        comma-separated list of pattern=N settings for file-filtered logging
