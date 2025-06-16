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

package handler

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstioVersion implements meshmanager.MeshManagerHandler
func (m *MeshManager) ListIstioVersion(
	ctx context.Context,
	req *meshmanager.ListIstioVersionRequest,
	resp *meshmanager.ListIstioVersionResponse,
) error {
	// 获取版本配置并输出
	istioConfig := m.opt.IstioConfig
	if istioConfig == nil {
		return errors.New("istio config is nil")
	}
	// 输出版本
	istioVersions := []*meshmanager.IstioVersion{}
	for _, version := range istioConfig.IstioVersions {
		if !version.Enabled {
			continue
		}
		istioVersions = append(istioVersions, &meshmanager.IstioVersion{
			Name:         version.Name,
			Version:      version.Version,
			ChartVersion: version.ChartVersion,
			KubeVersion:  version.KubeVersion,
		})
	}
	resp.Data = &meshmanager.IstioVersionAndFeatures{
		IstioVersions:  istioVersions,
		FeatureConfigs: buildFeaturesForVersion(istioVersions, istioConfig.FeatureConfigs),
	}
	return nil
}

// buildFeaturesForVersion 根据版本和全局 featureConfig 构建 features 列表
func buildFeaturesForVersion(
	istioVersions []*meshmanager.IstioVersion,
	featureConfigs []*options.FeatureConfig,
) []*meshmanager.FeatureConfig {
	features := []*meshmanager.FeatureConfig{}
	for _, feature := range featureConfigs {
		if !feature.Enabled {
			continue
		}
		supportVersions := []string{}
		for _, version := range istioVersions {
			if utils.IsVersionSupported(version.Version, feature.IstioVersion) {
				supportVersions = append(supportVersions, version.Version)
			}
		}
		if len(supportVersions) == 0 {
			continue
		}
		features = append(features, &meshmanager.FeatureConfig{
			Name:            feature.Name,
			Description:     feature.Description,
			DefaultValue:    feature.DefaultValue,
			AvailableValues: feature.AvailableValues,
			SupportVersions: supportVersions,
		})
	}
	return features
}
