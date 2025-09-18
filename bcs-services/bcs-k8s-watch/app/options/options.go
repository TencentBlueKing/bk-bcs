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
	"errors"
	"io/ioutil"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	jsoniter "github.com/json-iterator/go"
)

// DefaultConfig default config
type DefaultConfig struct {
	Environment string `json:"environment"`
	// ClusterID is only used when CluterIDSource was set to "config"
	ClusterID string `json:"clusterID"`
	HostIP    string `json:"hostIP"`
}

// validate validates DefaultConfig and set proper default values
func (c *DefaultConfig) validate() error { // nolint
	if c.ClusterID == "" {
		return errors.New("must set ClusterID when ClusterIDSource was set to 'config'")
	}
	return nil
}

// TLS tls config
type TLS struct {
	CAFile   string `json:"ca-file"`
	CertFile string `json:"cert-file"`
	KeyFile  string `json:"key-file"`
	Password string `json:"password"`
}

// BCSConfig configuration for bcs service discovery
type BCSConfig struct {
	// bcs zookeeper host list, split by comma
	ZkHosts string `json:"zk"`
	TLS     TLS    `json:"tls"`

	// NetServiceZKHosts is zookeepers hosts for netservice discovery, split by comma
	NetServiceZKHosts string `json:"netservice-zookeepers"`
	// CustomStorageEndpoints, split by comma
	CustomStorageEndpoints string `json:"custom-storage-endpoints"`
	// CustomNetServiceEndpoints is custom target netservice endpoints, split by comma
	CustomNetServiceEndpoints string `json:"custom-netservice-endpoints"`

	// whether the k8s cluster and bcs-k8s-watch is in external network
	IsExternal bool `json:"is-external"`
	// authorization token
	CustomStorageEndpointToken string `json:"custom-storage-endpoints-token"`

	// WriterQueueLen show writer module chan queue length for data distribute, default 10240
	WriterQueueLen int64 `json:"writerQueueLen"`
	// PodQueueNum run many queue to distribute Pod event in due to increase storage qps
	PodQueueNum int `json:"podQueueNum"`
}

// K8sConfig for installation out of cluster
type K8sConfig struct {
	Kubeconfig string `json:"kubeconfig"`
	Master     string `json:"master"`
	TLS        TLS    `json:"tls"`
}

// WatchResource 指定监听的资源
type WatchResource struct {
	// 监听指定的namespace，暂时支持一个
	Namespace         string            `json:"namespace"`
	DisableCRD        bool              `json:"disable_crd"`
	DisableNetservice bool              `json:"disable_netservice"`
	LabelSelectors    map[string]string `json:"label_selectors"` // map[resourceType]LabelSelector
}

// WatchConfig k8s-watch config
type WatchConfig struct {
	Default          DefaultConfig `json:"default"`
	BCS              BCSConfig     `json:"bcs"`
	K8s              K8sConfig     `json:"k8s"`
	FilterConfigPath string        `json:"filterConfigPath"`
	WatchResource    WatchResource `json:"watch_resource"`
	conf.FileConfig
	conf.ProcessConfig
	conf.LogConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ServerOnlyCertConfig

	DebugMode bool `json:"debug_mode"`
}

// NewWatchOptions init watch config
func NewWatchOptions() *WatchConfig {
	return &WatchConfig{}
}

// IsWatchManagedFields watch fields
var IsWatchManagedFields bool

// FilterConfig the file config
type FilterConfig struct {
	APIResourceSpecification []APIResourceFilter `json:"apiResourceSpecification"`
	APIResourceException     []APIResourceFilter `json:"apiResourceException"`
	K8sGroupVersionWhiteList []string            `json:"k8sResourceWhiteList"`
	CrdGroupVersionWhiteList []string            `json:"crdResourceWhiteList"`
	CrdVersionSupport        string              `json:"crdVersionSupport"`
	NamespaceFilters         []string            `json:"resourceNamespaceFilters"`
	NameFilters              []string            `json:"resourceNameFilters"`
	APIResourceLists         []ApiResourceList   `json:"apiResourceLists"`
	IsWatchManagedFields     bool                `json:"isFilterManagedFields"`
	DataMaskConfigList       []MaskerConfig      `json:"resourceMaskers"`
}

// APIResourceFilter api resource exception
type APIResourceFilter struct {
	GroupVersion  string   `json:"groupVersion"`
	ResourceKinds []string `json:"resourceKinds"`
}

// ParseFilter parse filter config from file
func (wc *WatchConfig) ParseFilter() *FilterConfig {
	filter := &FilterConfig{}
	bytes, err := ioutil.ReadFile(wc.FilterConfigPath)
	if err != nil {
		glog.Warnf("open filter config file (%s) failed: %s, will not use resource filter", wc.FilterConfigPath, err.Error())
		return nil
	}
	if err := jsoniter.Unmarshal(bytes, filter); err != nil {
		glog.Warnf("unmarshal config file (%s) failed: %s, will not use resource filter", wc.FilterConfigPath, err.Error())
		return nil
	}
	IsWatchManagedFields = filter.IsWatchManagedFields
	return filter
}

// MaskerConfig config for data mask
type MaskerConfig struct {
	Kind      string   `json:"kind"`
	Namespace string   `json:"namespace"`
	Path      []string `json:"path"`
}
