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

package types

import (
	"fmt"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

const (
	// ResourceApplicationType application definition
	ResourceApplicationType = "application"
	// ResourceCollectorType collector definition
	ResourceCollectorType = "collector"
	// ResourceMetricType metric definition
	ResourceMetricType = "metric"
	// ResourceTaskType task definition, describe a outer-defined metric collect settings
	ResourceTaskType = "task"
)

// Metric metric数据结构定义
type Metric struct {
	Version   string `json:"version"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	TLSConfig TLSCollectorCfg `json:"tlsConfig"`

	// mesos network
	NetworkMode interface{} `json:"networkMode"`
	NetworkType string      `json:"networkType"`

	// k8s network
	HostNetwork bool   `json:"hostNetwork"`
	DnsPolicy   string `json:"dnsPolicy"`
	// k8s secret
	Secrets []interface{} `json:"imagePullSecrets"`

	ClusterID   string `json:"clusterID"`
	ClusterType string `json:"clusterType,omitempty"`

	DataID     int               `json:"dataID"`
	Port       uint              `json:"port"`
	URI        string            `json:"uri"`
	Method     string            `json:"method"`
	Head       map[string]string `json:"head"`
	Parameters map[string]string `json:"parameters"`
	Selector   map[string]string `json:"selector"`
	Frequency  int               `json:"frequency"`
	Timeout    int               `json:"timeout"`

	ImageBase string `json:"imageBase"`

	// prometheus
	MetricType            MetricType        `json:"metricType"`
	PrometheusConstLabels map[string]string `json:"constLabels"`
}

// ApplicationCollectorCfg application collector configuration
type ApplicationCollectorCfg struct {
	Version   string         `json:"version"`   // collector configuration version
	Name      string         `json:"name"`      // collector configuration name for application
	Namespace string         `json:"namespace"` // collector configuration namespace for application
	Cfg       []CollectorCfg `json:"cfg"`       // collector configuration
}

// collectorCfg collector configuration
type CollectorCfg struct {
	CfgKey     string            `json:"cfgKey"`
	Version    string            `json:"version"`
	Meta       btypes.ObjectMeta `json:"meta"`   // pod meta define
	DataID     int               `json:"dataid"` // report data to bluking data plat by dataid
	IP         string            `json:"ip"`
	Port       uint              `json:"port"`
	Scheme     string            `json:"scheme"`
	Address    string            `json:"address"`    // http request URL
	Head       map[string]string `json:"head"`       // http head
	Parameters map[string]string `json:"parameters"` // http request parameters
	Method     string            `json:"method"`     // http method ,the value only can  be POST or GET
	Frequency  int               `json:"frequency"`  // the frequency of data collection
	Timeout    int               `json:"timeout"`

	TLSConfig TLSCollectorCfg `json:"tlsConfig"`

	MetricType            MetricType        `json:"metricType"`
	PrometheusConstLabels map[string]string `json:"constLabels"`
}

// data struct of MetricTask in storage
type StorageTaskIf struct {
	Data *MetricTask `json:"data"`
}

// MetricTask define a extra settings that import from API and will be added into collect settings
type MetricTask struct {
	ClusterID string            `json:"clusterID"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Selector  map[string]string `json:"selector"`
	Pods      []*MetricTaskPod  `json:"pods"`
}

type MetricTaskPod struct {
	IP   string            `json:"ip"`
	Port uint              `json:"port"`
	Meta btypes.ObjectMeta `json:"meta"`
}

type MetricType string

const (
	MetricPrometheus MetricType = "prometheus"
)

type ClusterType int

const (
	ClusterUnknown ClusterType = iota
	ClusterMesos
	ClusterK8S
	BcsComponents
)

var (
	clusterNames = map[ClusterType]string{
		ClusterMesos:  "mesos",
		ClusterK8S:    "k8s",
		BcsComponents: "bcs-components",
	}
)

func (ct ClusterType) String() string {
	return clusterNames[ct]
}

func (ct ClusterType) GetContainerTypeName() string {
	switch ct {
	case ClusterMesos:
		return "taskgroup"
	case ClusterK8S:
		return "pod"
	}
	return ""
}

func GetClusterType(s string) ClusterType {
	for k, v := range clusterNames {
		if v == s {
			return k
		}
	}
	return ClusterUnknown
}

const (
	BcsComponentsClusterId = "bcs_unique_const_clusterid"
	BcsComponentsSchemeKey = "io.tencent.bcs.metric.component.scheme"
)

var (
	DeleteCollectorNotExist = fmt.Errorf("delete collector not exist")
)
