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
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/json-iterator/go"

	glog "bk-bcs/bcs-common/common/blog"
)

const (
	ClusterIDSourceClusterKeeper = "clusterKeeper"
	ClusterIDSourceConfig        = "config"
)

type DefaultConfig struct {
	Environment     string `json:"environment"`
	ClusterIDSource string `json:"clusterIDSource"`
	// ClusterID is only used when CluterIDSource was set to "config"
	ClusterID string `json:"clusterID"`

	HostIP string `json:"hostIP"`
}

// validate validates DefaultConfig and set proper default values
func (c *DefaultConfig) validate() error {
	if c.ClusterIDSource == "" {
		c.ClusterIDSource = ClusterIDSourceClusterKeeper
	} else if c.ClusterIDSource == ClusterIDSourceConfig {
		if c.ClusterID == "" {
			return errors.New("must set ClusterID when ClusterIDSource was set to 'config'")
		}
	} else {
		return fmt.Errorf("unsupported ClusterIDSource %s", c.ClusterIDSource)
	}
	return nil
}

type TLS struct {
	CAFile   string `json:"ca-file"`
	CertFile string `json:"cert-file"`
	KeyFile  string `json:"key-file"`
	Password string `json:"password"`
}

type BCSConfig struct {
	ZkHosts []string `json:"zk"`
	TLS     TLS      `json:"tls"`

	// NetServiceZKHosts is zookeepers hosts for netservice discovery.
	NetServiceZKHosts []string `json:"netservice-zookeepers"`

	CustomStorageEndpoints []string `json:"custom-storage-endpoints"`

	// CustomNetServiceEndpoints is custom target netservice endpoints.
	CustomNetServiceEndpoints []string `json:"custom-netservice-endpoints"`

	// whether the k8s cluster and bcs-k8s-watch is in external network
	IsExternal bool `json:"is-external"`
}

type K8sConfig struct {
	Master string `json:"master"`
	TLS    TLS    `json:"tls"`
}

type WatchConfig struct {
	Default DefaultConfig `json:"default"`
	BCS     BCSConfig     `json:"bcs"`
	K8s     K8sConfig     `json:"k8s"`
}

func ParseConfigFile(configFilePath string) (*WatchConfig, error) {
	bytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s fail! %+v", configFilePath, err)
	}

	watchConfig := &WatchConfig{}
	if err := jsoniter.Unmarshal(bytes, watchConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config file %s fail! %+v", configFilePath, err)
	}

	if err := watchConfig.Default.validate(); err != nil {
		return nil, fmt.Errorf("config file invalid: %s", err)
	}

	glog.Infof("Parse config file %s, got: %+v", configFilePath, watchConfig)

	return watchConfig, nil

}
