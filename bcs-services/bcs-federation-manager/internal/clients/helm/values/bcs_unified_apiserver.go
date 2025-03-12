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
	// BcsUnifiedAPIServerServicePort is the port of bcs-unified-apiserver service
	BcsUnifiedAPIServerServicePort = 443
	// BcsUnifiedAPIServerServiceType is the type of bcs-unified-apiserver service
	BcsUnifiedAPIServerServiceType = "LoadBalancer"
	// BcsUnifiedAPIServerServiceLbIdKey is the annotations of bcs-unified-apiserver service
	BcsUnifiedAPIServerServiceLbIdKey = "service.kubernetes.io/tke-existed-lbid"
	// BcsUnifiedAPIServerServiceSubnetIdKey is the annotations of bcs-unified-apiserver service
	BcsUnifiedAPIServerServiceSubnetIdKey = "service.kubernetes.io/qcloud-loadbalancer-internal-subnetid"
)

// NewBcsUnifiedAPIServerValues create a new BcsUnifiedAPIServerValues
func NewBcsUnifiedAPIServerValues() *BcsUnifiedAPIServerValues {
	v := &BcsUnifiedAPIServerValues{}
	v.Service.Annotations = make(map[string]string)
	v.Service.Port = BcsUnifiedAPIServerServicePort
	v.Service.Type = BcsUnifiedAPIServerServiceType
	return v
}

// BcsUnifiedAPIServerValues is the values.yaml for bcs-unified-apiserver
// BcsUnifiedAPIServerValues  > default values config(charts.DefaultValues) > values in repository
type BcsUnifiedAPIServerValues struct {
	Config struct {
		BcsConf struct {
			Host         string `yaml:"host,omitempty"`
			Token        string `yaml:"token,omitempty"`
			JwtPublicKey string `yaml:"jwt_public_key,omitempty"`
		} `yaml:"bcs_conf,omitempty"`
		Apiserver struct {
			FederationHostClusterId string `yaml:"federation_host_cluster_id"`
			StoreMode               string `yaml:"store_mode,omitempty"`
			WebhookAddress          string `yaml:"webhook_address,omitempty"`
		} `yaml:"apiserver"`
	} `yaml:"config,omitempty"`
	Service struct {
		Type        string            `yaml:"type,omitempty"`
		Port        int               `yaml:"port,omitempty"`
		Annotations map[string]string `yaml:"annotations,omitempty"`
	} `yaml:"service,omitempty"`
}

// Yaml return the yaml format string
func (b *BcsUnifiedAPIServerValues) Yaml() string {
	result, _ := yaml.Marshal(b)
	return string(result)
}

// SetFederationClusterId set the cluster id for bcs-unified-apiserver service
func (b *BcsUnifiedAPIServerValues) SetFederationClusterId(clusterId string) error {
	if clusterId == "" {
		return fmt.Errorf("clusterId is empty")
	}
	b.Config.Apiserver.FederationHostClusterId = clusterId
	return nil
}

// SetUserToken set the user token for bcs-unified-apiserver service
func (b *BcsUnifiedAPIServerValues) SetUserToken(token string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	b.Config.BcsConf.Token = token
	return nil
}

// SetLoadbalancerId set the lb id for bcs-unified-apiserver service
func (b *BcsUnifiedAPIServerValues) SetLoadbalancerId(id string) error {
	// if id not begin with "lb-", return error
	if strings.HasPrefix(id, "lb-") {
		b.Service.Annotations[BcsUnifiedAPIServerServiceLbIdKey] = id
	} else if strings.HasPrefix(id, "subnet-") {
		b.Service.Annotations[BcsUnifiedAPIServerServiceSubnetIdKey] = id
	} else {
		return fmt.Errorf("LoadBalancer id must begin with lb- or subnet-")
	}

	return nil
}
