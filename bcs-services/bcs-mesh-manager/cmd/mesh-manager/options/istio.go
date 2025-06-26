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

package options

import (
	"fmt"
	"slices"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// IstioConfig istio config
type IstioConfig struct {
	IstioVersions   map[string]*IstioVersion  `json:"istioVersions"`
	FeatureConfigs  map[string]*FeatureConfig `json:"featureConfigs"`
	ChartValuesPath string                    `json:"chartValuesPath"`
	ChartRepo       string                    `json:"chartRepo"`
}

// IstioVersion istio version
type IstioVersion struct {
	Name         string `json:"name"`
	ChartVersion string `json:"chartVersion"`
	KubeVersion  string `json:"kubeVersion"`
	Enabled      bool   `json:"enabled"`
}

// FeatureConfig feature config
type FeatureConfig struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	DefaultValue    string   `json:"defaultValue"`
	AvailableValues []string `json:"availableValues"`
	IstioVersion    string   `json:"istioVersion"` // 支持的istio版本
	Enabled         bool     `json:"enabled"`
}

// Validate validate istio config
func (c *IstioConfig) Validate() error {
	// validate istio feature config
	for _, feature := range c.FeatureConfigs {
		if !slices.Contains(common.SupportedFeatures, feature.Name) {
			return fmt.Errorf("feature %s is not supported", feature.Name)
		}
	}
	return nil
}
