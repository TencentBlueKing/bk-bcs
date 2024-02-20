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
 *
 */

// Package option xxx
package option

import (
	"flag"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
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
	// ProbePort port for probe
	ProbePort int

	// HttpServerPort http server port
	HttpServerPort uint

	// ElectionNamespace election namespace
	ElectionNamespace string

	// LogConfig for blog
	conf.LogConfig

	// KubernetesQPS the qps of k8s client request
	KubernetesQPS int
	// KubernetesBurst the burst of k8s client request
	KubernetesBurst int

	// ScenarioPath path to store scenario config
	ScenarioPath string

	// ScenarioGitRefreshFreq 定时pull所有仓库， 判断场景是否更新
	ScenarioGitRefreshFreq time.Duration

	// ScenarioGitUserName username for git
	ScenarioGitUserName string
	// ScenarioGitSecret secret for git
	ScenarioGitSecret string

	// RepoRefreshFreq 根据集群内AppMonitor刷新Repo缓存 (如某个Repo已经不被任何AppMonitor引用，则从缓存中删除)
	RepoRefreshFreq time.Duration

	// ArgoAdminNamespace admin namespace for argo
	ArgoAdminNamespace string
}

// BindFromCommandLine bind from
func (c *ControllerOption) BindFromCommandLine() {
	var verbosity int
	var scenarioRefreshFreqSec int64
	var repoRefreshFreqSec int64
	flag.IntVar(&c.MetricPort, "metrics_port", 8080, "The address the metric endpoint binds to.")
	flag.IntVar(&c.ProbePort, "health_probe_port", 8081, "The address the probe endpoint binds to.")

	// log config
	flag.IntVar(&verbosity, "v", 3, "log level for V logs")
	flag.StringVar(&c.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&c.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&c.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&c.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&c.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.StringVar(&c.Address, "address", "0.0.0.0", "address for controller")
	flag.StringVar(&c.ScenarioPath, "scenario_path", "/data/bcs", "Store scenario templates")
	flag.Int64Var(&scenarioRefreshFreqSec, "scenario_refresh_req", 60, "refresh frequency ")
	flag.UintVar(&c.HttpServerPort, "http_server_port", 8088, "http server port")

	flag.Int64Var(&repoRefreshFreqSec, "repo_refresh_req", 600, "refresh frequency ")
	flag.StringVar(&c.ArgoAdminNamespace, "argo_admin_namespace", "default", "argo admin namespace")
	c.ScenarioGitRefreshFreq = time.Second * time.Duration(scenarioRefreshFreqSec)
	c.RepoRefreshFreq = time.Second * time.Duration(repoRefreshFreqSec)
	c.Verbosity = int32(verbosity)
	flag.Parse()
}
