/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package option for controller
package option

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

// ControllerOption options for controller
type ControllerOption struct {
	// PodNamespace pod namespace
	PodNamespace string

	// ImageTag image tag used by controller
	ImageTag string

	// Address address for server
	Address string

	// PodIPs contains ipv4 and ipv6 address get from status.podIPs
	PodIPs []string

	// Port port for server
	Port int

	// MetricPort port for metric server
	MetricPort int

	// Cloud cloud mod
	Cloud string

	// Region cloud region
	Region string

	// ElectionNamespace election namespace
	ElectionNamespace string

	// IsNamespaceScope if the ingress can only be associated with the service and workload in the same namespace
	IsNamespaceScope bool

	// LogConfig for blog
	conf.LogConfig

	// IsTCPUDPReuse if the loadbalancer provider support tcp udp port reuse
	// if enabled, we will find protocol info in 4 layer listener name
	IsTCPUDPPortReuse bool

	// IsBulkMode if use bulk interface for cloud lb
	IsBulkMode bool

	// PortBindingCheckInterval check interval for portbinding
	PortBindingCheckInterval time.Duration

	// ServerCertFile server cert file path
	ServerCertFile string
	// ServerKeyFile server key file path
	ServerKeyFile string

	// KubernetesQPS the qps of k8s client request
	KubernetesQPS int
	// KubernetesBurst the burst of k8s client request
	KubernetesBurst int

	// ConflictCheckOpen if false, skip all conflict checking about ingress and port pool
	ConflictCheckOpen bool

	// NodePortBindingNs namespace that node portbinding will be created in,
	// and if node's annotation have not related portpool namespace, will use NodePortBindingNs as default
	NodePortBindingNs string

	// HttpServerPort port for http api
	HttpServerPort uint

	// NodeInfoExporterOpen 如果为true，将会记录集群中的节点信息
	NodeInfoExporterOpen bool

	// NodeExternalWorkerEnable 是否启用NodeExternalWorker插件（该插件在指定节点上启动daemonset pod， 探测节点的公网IP）
	NodeExternalWorkerEnable bool

	// NodeExternalIPConfigmap 节点公网IP configmap名称
	NodeExternalIPConfigmap string

	// LBCacheExpiration lb缓存过期时间，单位分钟
	LBCacheExpiration int

	// ListenerAutoReconcileSeconds != 0时， 间隔若干秒reconcile一次listener, 默认1200s（20分钟）
	ListenerAutoReconcileSeconds int

	// Conf HttpServer conf
	Conf Conf
	// ServCert http server cert
	ServCert ServCert
}

// Conf 服务配置
type Conf struct {
	ServCert        ServCert
	InsecureAddress string
	InsecurePort    uint
	VerifyClientTLS bool
}

// ServCert 服务证书配置
type ServCert struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// SetFromEnv set options by environment
func (op *ControllerOption) SetFromEnv() {
	// get env var name for tcp and udp port reuse
	isTCPUDPPortReuseStr := os.Getenv(constant.EnvNameIsTCPUDPPortReuse)
	if len(isTCPUDPPortReuseStr) != 0 {
		blog.Infof("env option %s is %s", constant.EnvNameIsTCPUDPPortReuse, isTCPUDPPortReuseStr)
		isTCPUDPPortReuse, err := strconv.ParseBool(isTCPUDPPortReuseStr)
		if err != nil {
			blog.Errorf("parse bool string %s failed, err %s", isTCPUDPPortReuseStr, err.Error())
			os.Exit(1)
		}
		if isTCPUDPPortReuse {
			op.IsTCPUDPPortReuse = isTCPUDPPortReuse
		}
	}

	// get env var name for bulk mode
	isBulkModeStr := os.Getenv(constant.EnvNameIsBulkMode)
	if len(isBulkModeStr) != 0 {
		blog.Infof("env option %s is %s", constant.EnvNameIsBulkMode, isBulkModeStr)
		isBulkMode, err := strconv.ParseBool(isBulkModeStr)
		if err != nil {
			blog.Errorf("parse bool string %s failed, err %s", isBulkModeStr, err.Error())
			os.Exit(1)
		}
		if isBulkMode {
			op.IsBulkMode = isBulkMode
		}
	}

	podIPs := os.Getenv(constant.EnvNamePodIPs)
	if len(podIPs) == 0 {
		blog.Errorf("empty pod ip")
		podIPs = op.Address
	}
	blog.Infof("pod ips: %s", podIPs)
	op.PodIPs = strings.Split(podIPs, ",")

	imageTag := os.Getenv(constant.EnvNameImageTag)
	if len(imageTag) == 0 {
		blog.Errorf("empty image tag")
	}
	op.ImageTag = imageTag

	op.PodNamespace = os.Getenv(constant.EnvIngressPodNamespace)
}

// BindFromCommandLine 读取命令行参数并绑定
func (op *ControllerOption) BindFromCommandLine() {
	var checkIntervalStr string
	var verbosity int
	flag.StringVar(&op.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&op.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.IntVar(&op.Port, "port", 8080, "por for controller")
	flag.StringVar(&op.Cloud, "cloud", "tencentcloud", "cloud mode for controller")
	flag.StringVar(&op.Region, "region", "", "default cloud region for controller")
	flag.StringVar(&op.ElectionNamespace, "election_namespace", "bcs-system", "namespace for leader election")
	flag.BoolVar(&op.IsNamespaceScope, "is_namespace_scope", false,
		"if the ingress can only be associated with the service and workload in the same namespace")
	flag.StringVar(&checkIntervalStr, "portbinding_check_interval", "3m",
		"check interval of port binding, golang time format")

	flag.StringVar(&op.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&op.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&op.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&op.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&op.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.IntVar(&verbosity, "v", 0, "log level for V logs")
	flag.StringVar(&op.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&op.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&op.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.StringVar(&op.ServerCertFile, "server_cert_file", "", "server cert file for webhook server")
	flag.StringVar(&op.ServerKeyFile, "server_key_file", "", "server key file for webhook server")

	flag.IntVar(&op.KubernetesQPS, "kubernetes_qps", 100, "the qps of k8s client request")
	flag.IntVar(&op.KubernetesBurst, "kubernetes_burst", 200, "the burst of k8s client request")

	flag.BoolVar(&op.ConflictCheckOpen, "conflict_check_open", true, "if false, "+
		"skip all conflict checking about ingress and port pool")
	flag.BoolVar(&op.NodeInfoExporterOpen, "node_info_exporter_open", false, "if true, "+
		"bcs-ingress-controller will record node info in cluster")
	flag.BoolVar(&op.NodeExternalWorkerEnable, "node_external_worker_enable", false,
		"enable node_external_worker plugin or not")
	flag.StringVar(&op.NodeExternalIPConfigmap, "node_external_ip_configmap", "",
		"name of node public external ip configmap")
	flag.StringVar(&op.NodePortBindingNs, "node_portbinding_ns", "default",
		"namespace that node portbinding will be created in ")

	flag.UintVar(&op.HttpServerPort, "http_svr_port", 8088, "port for ingress controller http server")
	flag.IntVar(&op.LBCacheExpiration, "lb_cache_expiration", 60, "lb cache expiration, unit: minute ")

	flag.IntVar(&op.ListenerAutoReconcileSeconds, "listener_auto_reconcile_seconds", 1200,
		"seconds to auto reconcile listeners")

	flag.Parse()

	op.Verbosity = int32(verbosity)

	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		fmt.Printf("check interval %s invalid", checkIntervalStr)
		os.Exit(1)
	}
	op.PortBindingCheckInterval = checkInterval
}
