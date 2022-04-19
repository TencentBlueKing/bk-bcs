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

package options

import (
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
)

const (
	defaultBindAddress     = "0.0.0.0"
	defaultHealthCheckPort = 9090
	defaultMetricsPort     = 9091
)

// Options has all the params needed to run a MCS agent
type Options struct {
	LeaderElection       componentbaseconfig.LeaderElectionConfiguration
	AgentID              string
	KubeconfigPath       string
	ParentKubeconfigPath string
	BindAddress          string
	HealthCheckPort      int
	MetricsPort          int
}

// NewOptions builds an default scheduler options.
func NewOptions() *Options {
	return &Options{
		LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
			LeaderElect:       true,
			ResourceLock:      resourcelock.LeasesResourceLock,
			ResourceNamespace: "kube-system",
		},
	}
}

// AddFlags adds flags of scheduler to the specified FlagSet
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.BindAddress, "bind-address", defaultBindAddress,
		"The IP address on which to listen for the --secure-port port.")
	fs.IntVar(&o.HealthCheckPort, "health-check-port", defaultHealthCheckPort,
		"The port on which to serve health checks.")
	fs.IntVar(&o.MetricsPort, "metrics-port", defaultMetricsPort, "The port on which to serve metrics.")
	fs.BoolVar(&o.LeaderElection.LeaderElect, "leader-elect", true, "Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability.")
	fs.StringVar(&o.LeaderElection.ResourceNamespace, "leader-elect-resource-namespace", "kube-system", "The namespace of resource object that is used for locking during leader election.")
	fs.StringVar(&o.AgentID, "agent-id", "", "The agent id of MCS agent.")
	fs.StringVar(&o.KubeconfigPath, "kubeconfig", "", "The path to the kubeconfig file to use for talking to the apiserver.")
	fs.StringVar(&o.ParentKubeconfigPath, "parent-kubeconfig", "", "The path to the kubeconfig file to use for talking to the apiserver. for parent cluster.")
}
