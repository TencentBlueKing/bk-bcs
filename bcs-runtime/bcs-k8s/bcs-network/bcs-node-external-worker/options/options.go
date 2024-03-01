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

// Package options xxx
package options

import (
	"flag"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

const (
	// EnvNameNodeName xxx
	EnvNameNodeName = "NODE_NAME"
	// EnvNamePodNamespace xxx
	EnvNamePodNamespace = "POD_NAMESPACE"
)

// Options xxx
type Options struct {
	// ExternalIPWebURL URL to get external IP
	ExternalIPWebURL string
	// ListenPort Listen to check if external IP is valid
	ListenPort uint
	// ListenAddress Listen to check if external IP is valid
	ListenAddress string

	// Address address for controller
	Address string

	// Namespace where controller locate
	Namespace string

	// NodeName which node controller locate
	NodeName string

	// ExternalIPConfigMapName which configmap to store external IP
	ExternalIPConfigMapName string

	// LogConfig for blog
	conf.LogConfig
}

// BindFromCommandLine xxx
func (op *Options) BindFromCommandLine() {
	var verbosity int

	flag.StringVar(&op.Address, "address", "127.0.0.1", "address for controller")
	flag.UintVar(&op.ListenPort, "listen_port", 80, "Listen port to check if external IP is valid")
	flag.StringVar(&op.ListenAddress, "listen_address", "0.0.0.0", "Listen address to check if external IP is valid")
	flag.StringVar(&op.ExternalIPWebURL, "external_ip_web_url", "http://myexternalip.com/raw", "external ip")
	flag.StringVar(&op.ExternalIPConfigMapName, "node_external_ip_configmap",
		"bcs-ingress-controller-node-external-ip-configmap", "configmap to store external ip info")

	flag.StringVar(&op.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&op.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&op.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&op.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&op.AlsoToStdErr, "alsologtostderr", true, "log to standard error as well as files")
	flag.IntVar(&verbosity, "v", 3, "log level for V logs")
	flag.StringVar(&op.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&op.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&op.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.Parse()
	op.Verbosity = int32(verbosity)

}

// SetFromEnv set form env
func (op *Options) SetFromEnv() {
	nodeName := os.Getenv(EnvNameNodeName)
	if len(nodeName) == 0 {
		blog.Errorf("empty node name")
		os.Exit(1)
	}

	podNamespace := os.Getenv(EnvNamePodNamespace)
	if len(podNamespace) == 0 {
		blog.Errorf("empty pod namespace")
		os.Exit(1)
	}

	op.NodeName = nodeName
	op.Namespace = podNamespace
}
