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

package option

import (
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

	// ScenarioGitRefreshFreq refresh frequency of git repo
	ScenarioGitRefreshFreq time.Duration

	// ScenarioGitUserName username for git
	ScenarioGitUserName string
	// ScenarioGitSecret secret for git
	ScenarioGitSecret string
}
