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

// Package values xxx
package values

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// params for build bcs-clusternet-agent values
	BcsClusternetAgentParentURL             = "https://kubernetes.default.svc.cluster.local"
	BcsClusternetAgentSyncMode              = "Pull"
	BcsClusternetAgentFeatureGates          = "SocketConnection=false,AppPusher=false"
	BcsClusternetAgentClusterLabelsFormat   = "clusters.clusternet.io/cluster-name=%s,subscription.bkbcs.tencent.com/clusterid=%s"
	BcsClusternetAgentProviderName          = "bcs"
	BcsClusternetAgentParentKubeApiFlowRate = 3
	BcsClusternetAgentProviderKubeApiBurst  = 100
	BcsClusternetAgentProviderKubeApiQps    = 50
	BcsClusternetAgentVerb                  = 4
	BcsClusternetAgentProviderWorkerCount   = 10
)

// NewBcsClusternetAgentValues xxx
func NewBcsClusternetAgentValues(clusterId string) *BcsClusternetAgentValues {
	clusterId = strings.ToLower(clusterId)

	return &BcsClusternetAgentValues{
		ExtraArgs: map[string]interface{}{
			"feature-gates":             BcsClusternetAgentFeatureGates,
			"cluster-sync-mode":         BcsClusternetAgentSyncMode,
			"cluster-labels":            fmt.Sprintf(BcsClusternetAgentClusterLabelsFormat, clusterId, clusterId),
			"cluster-reg-name":          clusterId,
			"parent-kube-api-flow-rate": BcsClusternetAgentParentKubeApiFlowRate,
			"provider-kube-api-burst":   BcsClusternetAgentProviderKubeApiBurst,
			"provider-kube-api-qps":     BcsClusternetAgentProviderKubeApiQps,
			"v":                         BcsClusternetAgentVerb,
			"worker-count":              BcsClusternetAgentProviderWorkerCount,
		},
		ParentURL:         BcsClusternetAgentParentURL,
		RegistrationToken: "",
	}
}

// BcsClusternetAgentValues values for bcs-clusternet-agent
type BcsClusternetAgentValues struct {
	ExtraArgs         map[string]interface{}   `yaml:"extraArgs"`
	ParentURL         string                   `yaml:"parentURL"`
	RegistrationToken string                   `yaml:"registrationToken"`
	Provider          *ClusternetAgentProvider `yaml:"provider"`
}

// ClusternetAgentProvider xxx
type ClusternetAgentProvider struct {
	KubeConfig string `yaml:"kubeconfig"`
	Name       string `yaml:"name"`
}

// Yaml return the yaml format string
func (b *BcsClusternetAgentValues) Yaml() string {
	result, _ := yaml.Marshal(b)
	return string(result)
}

// SetKubeConfig set provider kubeconfig
func (b *BcsClusternetAgentValues) SetKubeConfig(kubeconfig string) {
	b.Provider = &ClusternetAgentProvider{
		KubeConfig: kubeconfig,
		Name:       BcsClusternetAgentProviderName,
	}
}

// SetRegistrationToken set registration token
func (b *BcsClusternetAgentValues) SetRegistrationToken(token string) {
	b.RegistrationToken = token
}
