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
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

const (
	// EnvNamePodIPs pod ip env name
	EnvNamePodIPs = "POD_IPS"
)

// ControllerOption options for controller
type ControllerOption struct {
	// Address address for server
	Address string

	// PodIPs contains ipv4 and ipv6 address get from status.podIPs
	PodIPs []string

	// Port port for server
	Port int

	// MetricPort port for metric server
	MetricPort int

	// LogConfig for blog
	conf.LogConfig

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
	podIPs := os.Getenv(EnvNamePodIPs)
	if len(podIPs) == 0 {
		blog.Errorf("empty pod ip")
		podIPs = op.Address
	}
	blog.Infof("pod ips: %s", podIPs)
	op.PodIPs = strings.Split(podIPs, ",")
}

// BindFromCommandLine 读取命令行参数并绑定
func (op *ControllerOption) BindFromCommandLine() {
	var verbosity int
	flag.StringVar(&op.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&op.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.IntVar(&op.Port, "port", 8080, "por for controller")

	flag.StringVar(&op.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&op.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&op.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&op.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&op.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.IntVar(&verbosity, "v", 0, "log level for V logs")
	flag.StringVar(&op.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&op.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&op.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.IntVar(&op.KubernetesQPS, "kubernetes_qps", 100, "the qps of k8s client request")
	flag.IntVar(&op.KubernetesBurst, "kubernetes_burst", 200, "the burst of k8s client request")

	flag.BoolVar(&op.ConflictCheckOpen, "conflict_check_open", true, "if false, "+
		"skip all conflict checking about ingress and port pool")
	flag.BoolVar(&op.NodeInfoExporterOpen, "node_info_exporter_open", false, "if true, "+
		"bcs-ingress-controller will record node info in cluster")
	flag.StringVar(&op.NodePortBindingNs, "node_portbinding_ns", "default",
		"namespace that node portbinding will be created in ")

	flag.UintVar(&op.HttpServerPort, "http_svr_port", 8088, "port for ingress controller http server")
	flag.Parse()

	op.Verbosity = int32(verbosity)
}
