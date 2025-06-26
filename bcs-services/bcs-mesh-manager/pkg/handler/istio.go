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
	istioaction "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/actions/istio"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) ListIstio(
	ctx context.Context,
	req *meshmanager.ListIstioRequest,
	resp *meshmanager.ListIstioResponse,
) error {
	action := istioaction.NewListIstioAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// InstallIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) InstallIstio(
	ctx context.Context,
	req *meshmanager.IstioRequest,
	resp *meshmanager.InstallIstioResponse,
) error {
	action := istioaction.NewInstallIstioAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// UpdateIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) UpdateIstio(
	ctx context.Context,
	req *meshmanager.IstioRequest,
	resp *meshmanager.UpdateIstioResponse,
) error {
	action := istioaction.NewUpdateIstioAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// DeleteIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) DeleteIstio(
	ctx context.Context,
	req *meshmanager.DeleteIstioRequest,
	resp *meshmanager.DeleteIstioResponse,
) error {
	action := istioaction.NewDeleteIstioAction(m.model)
	return action.Handle(ctx, req, resp)
}

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
	for version, istioVersionConfig := range istioConfig.IstioVersions {
		if !istioVersionConfig.Enabled {
			continue
		}
		istioVersions = append(istioVersions, &meshmanager.IstioVersion{
			Name:         istioVersionConfig.Name,
			Version:      version,
			ChartVersion: istioVersionConfig.ChartVersion,
			KubeVersion:  istioVersionConfig.KubeVersion,
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
	featureConfigs map[string]*options.FeatureConfig,
) map[string]*meshmanager.FeatureConfig {
	features := make(map[string]*meshmanager.FeatureConfig)
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
		features[feature.Name] = &meshmanager.FeatureConfig{
			Name:            feature.Name,
			Description:     feature.Description,
			DefaultValue:    feature.DefaultValue,
			AvailableValues: feature.AvailableValues,
			SupportVersions: supportVersions,
		}
	}
	return features
}

// GetIstioDetail implements meshmanager.MeshManagerHandler
func (m *MeshManager) GetIstioDetail(
	ctx context.Context,
	req *meshmanager.GetIstioDetailRequest,
	resp *meshmanager.GetIstioDetailResponse,
) error {
	action := istioaction.NewGetIstioDetailAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}
