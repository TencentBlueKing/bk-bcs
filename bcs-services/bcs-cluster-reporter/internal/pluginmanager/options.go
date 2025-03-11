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

// Package pluginmanager xxx
package pluginmanager

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"sync"
)

const (
	RunModeOnce   = "once"
	RunModeDaemon = "daemon"

	TKECluster = "tke"
)

// Config Options bcs log options
type Config struct {
	ClusterConfigs  map[string]*ClusterConfig
	NodeConfig      NodeConfig
	InClusterConfig ClusterConfig
}

// ClusterConfig xxx
type ClusterConfig struct {
	Config      *rest.Config
	ClusterID   string
	BusinessID  string
	Master      []string
	BCSCluster  clustermanager.Cluster
	ClusterType string
	ClientSet   *kubernetes.Clientset
	MetricSet   *metricsclientset.Clientset
	Version     string

	// net
	ServiceCidr   string
	Cidr          []string
	MaskSize      int
	ServiceMaxNum int
	ServiceNum    int

	// node 集群总pod数，包含master
	NodeNum    int
	EkletNum   int
	VnodeNum   int
	WindowsNum int
	NodeInfo   map[string]plugin.NodeInfo
	ALLEKLET   bool

	// mutex
	sync.Mutex
}

// NodeConfig xxx
type NodeConfig struct {
	Config        *rest.Config
	ClientSet     *kubernetes.Clientset
	NodeName      string
	Node          *v1.Node
	HostPath      string
	KubernetesSvc string
	KubeletParams map[string]string
}

// Validate validate options
func (o *Config) Validate() error {
	// if len(o.KubeMaster) == 0 {
	//	return fmt.Errorf("kube_master cannot be empty")
	// }
	// if len(o.Kubeconfig) == 0 {
	//	return fmt.Errorf("kubeconfig cannot be empty")
	// }
	return nil
}
