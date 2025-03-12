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
	"strings"

	"gopkg.in/yaml.v3"
)

// NewBcsEstimatorAgentValues create a new bcs-estimator-agent values
func NewBcsEstimatorAgentValues(clusterId string) *BcsEstimatorAgentValues {
	clusterId = strings.ToLower(clusterId)

	return &BcsEstimatorAgentValues{
		ExtraArgs: map[string]interface{}{
			"leader-elect": "true",
			"clusterid":    clusterId,
			"port":         9090,
			"v":            4,
		},
		Provider: &EstimatorAgentProvider{
			Name: "bcs",
			ConfigContent: &ConfigContent{
				SchedulerPlugins: []SchedulerPlugin{
					{Name: "PrioritySort"},
					{Name: "NodeUnschedulable"},
					{Name: "NodeName"},
					{Name: "NodePorts"},
					{Name: "DefaultBinder"},
					{Name: "TaintToleration", Weight: 3},
					{Name: "NodeAffinity", Weight: 2},
					{Name: "NodeResourcesFit", Weight: 1},
					{Name: "PodTopologySpread", Weight: 2},
					{Name: "InterPodAffinity", Weight: 2},
					{Name: "NodeResourcesBalancedAllocation", Weight: 1},
					{Name: "ImageLocality", Weight: 1},
				},
				KubeApiQps:   100,
				KubeApiBurst: 200,
				KubeConfig:   "",
			},
		},
	}
}

// BcsEstimatorAgentValues values for bcs-estimator-agent
type BcsEstimatorAgentValues struct {
	ExtraArgs map[string]interface{}  `yaml:"extraArgs"`
	Provider  *EstimatorAgentProvider `yaml:"provider"`
}

// EstimatorAgentProvider estimator agent provider
type EstimatorAgentProvider struct {
	Name             string         `yaml:"name"`
	ConfigContentStr string         `yaml:"configContent"`
	ConfigContent    *ConfigContent `yaml:"-"`
}

// Yaml return the yaml format string
func (b *BcsEstimatorAgentValues) Yaml() string {
	b.Provider.ConfigContentStr = b.Provider.ConfigContent.Yaml()

	result, _ := yaml.Marshal(b)
	return string(result)
}

// SetKubeConfig set kubeconfig
func (b *BcsEstimatorAgentValues) SetKubeConfig(kubeconfig string) {
	if b.Provider == nil {
		b.Provider = &EstimatorAgentProvider{
			ConfigContent: &ConfigContent{},
		}
	}
	if b.Provider.ConfigContent == nil {
		b.Provider.ConfigContent = &ConfigContent{}
	}
	b.Provider.ConfigContent.KubeConfig = kubeconfig
}

// ConfigContent config content
type ConfigContent struct {
	SchedulerPlugins []SchedulerPlugin `yaml:"schedulerPlugins"`
	KubeApiQps       int               `yaml:"kube_api_qps"`
	KubeApiBurst     int               `yaml:"kube_api_burst"`
	KubeConfig       string            `yaml:"kubeconfig"`
}

// SchedulerPlugin scheduler plugin
type SchedulerPlugin struct {
	Name   string `yaml:"name"`
	Weight int    `yaml:"configContent,omitempty"`
}

// Yaml return the yaml format string
func (c *ConfigContent) Yaml() string {
	result, _ := yaml.Marshal(c)
	return string(result)
}
