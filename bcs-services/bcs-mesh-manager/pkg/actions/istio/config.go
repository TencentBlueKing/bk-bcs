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

package istio

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstioConfigAction action for list istio config
type ListIstioConfigAction struct {
	istioConfig *options.IstioConfig
	req         *meshmanager.ListIstioConfigRequest
	resp        *meshmanager.ListIstioConfigResponse
}

// NewListIstioConfigAction create list istio config action
func NewListIstioConfigAction(istioConfig *options.IstioConfig) *ListIstioConfigAction {
	return &ListIstioConfigAction{
		istioConfig: istioConfig,
	}
}

// Handle processes the istio config request
func (l *ListIstioConfigAction) Handle(
	ctx context.Context,
	req *meshmanager.ListIstioConfigRequest,
	resp *meshmanager.ListIstioConfigResponse,
) error {
	l.req = req
	l.resp = resp

	// 获取版本配置并输出
	if l.istioConfig == nil {
		blog.Errorf("list istio config failed, istio config is nil")
		l.setResp(common.ParamErrorCode, "istio config is nil", nil)
		return nil
	}

	// 设置成功响应
	l.setResp(common.SuccessCode, "", l.buildConfigData())
	blog.Infof("list istio config successfully")
	return nil
}

// setResp sets the response with code, message and data
func (l *ListIstioConfigAction) setResp(code uint32, message string, data *meshmanager.IstioConfigData) {
	l.resp.Code = code
	l.resp.Message = message
	l.resp.Data = data
}

// buildConfigData 构建配置数据
func (l *ListIstioConfigAction) buildConfigData() *meshmanager.IstioConfigData {
	// 输出版本
	istioVersions := []*meshmanager.IstioVersion{}
	for version, istioVersionConfig := range l.istioConfig.IstioVersions {
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

	return &meshmanager.IstioConfigData{
		IstioVersions:         istioVersions,
		SidecarResourceConfig: common.GetDefaultSidecarResourceConfig(),
		HighAvailability:      common.GetDefaultHighAvailabilityConfig(),
		FeatureConfigs:        l.buildFeaturesForVersion(istioVersions, l.istioConfig.FeatureConfigs),
	}
}

// buildFeaturesForVersion 根据版本和全局 featureConfig 构建 features 列表
func (l *ListIstioConfigAction) buildFeaturesForVersion(
	istioVersions []*meshmanager.IstioVersion,
	featureConfigs map[string]*options.FeatureConfig,
) map[string]*meshmanager.FeatureConfig {
	features := make(map[string]*meshmanager.FeatureConfig)

	// 首先处理配置文件中定义的特性
	for _, feature := range featureConfigs {
		if !feature.Enabled {
			continue
		}
		// 提取版本字符串列表
		versionStrings := make([]string, 0, len(istioVersions))
		for _, version := range istioVersions {
			versionStrings = append(versionStrings, version.Version)
		}
		// 使用公共函数过滤版本
		supportVersions := l.filterSupportedVersions(versionStrings, feature.IstioVersion)
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

	// 然后处理 common.SupportedFeatures 中但没有在配置文件中配置的特性
	for _, featureName := range common.SupportedFeatures {
		// 如果该特性已经在配置文件中定义了，跳过
		if _, exists := features[featureName]; exists {
			blog.Infof("feature %s is already defined, skip", featureName)
			continue
		}

		// 创建默认的特性配置
		supportVersions := l.getAllSupportVersions(istioVersions)
		defaultFeature := l.getDefaultFeatureConfigByName(featureName, supportVersions)
		if defaultFeature != nil {
			features[featureName] = defaultFeature
		}
	}

	return features
}

// getAllSupportVersions 获取所有版本列表
func (l *ListIstioConfigAction) getAllSupportVersions(istioVersions []*meshmanager.IstioVersion) []string {
	supportVersions := make([]string, 0, len(istioVersions))
	for _, version := range istioVersions {
		supportVersions = append(supportVersions, version.Version)
	}
	return supportVersions
}

// buildDefaultFeatureConfig 根据模板和版本列表构建特性配置
func (l *ListIstioConfigAction) buildDefaultFeatureConfig(
	template *common.DefaultFeatureConfigTemplate,
	supportVersions []string,
) *meshmanager.FeatureConfig {
	if template == nil || len(supportVersions) == 0 {
		return nil
	}

	// 如果模板有版本要求，需要筛选支持的版本
	filteredVersions := supportVersions
	if template.SupportIstioVersion != "" {
		filteredVersions = l.filterSupportedVersions(supportVersions, template.SupportIstioVersion)
		// 如果没有版本满足要求，返回nil
		if len(filteredVersions) == 0 {
			return nil
		}
	}

	return &meshmanager.FeatureConfig{
		Name:            template.Name,
		Description:     template.Description,
		DefaultValue:    template.DefaultValue,
		AvailableValues: template.AvailableValues,
		SupportVersions: filteredVersions,
	}
}

// getDefaultFeatureConfigByName 根据特性名称获取默认配置
func (l *ListIstioConfigAction) getDefaultFeatureConfigByName(
	featureName string,
	supportVersions []string,
) *meshmanager.FeatureConfig {
	templates := common.GetDefaultFeatureConfigs()
	template, exists := templates[featureName]
	if !exists {
		// 对于未知的特性，创建一个基本的默认配置
		blog.Errorf("unknown feature: %s", featureName)
		return nil
	}

	return l.buildDefaultFeatureConfig(template, supportVersions)
}

// filterSupportedVersions 根据semver表达式筛选支持的版本
func (l *ListIstioConfigAction) filterSupportedVersions(versions []string, semverExpr string) []string {
	if semverExpr == "" {
		return versions
	}

	var filteredVersions []string
	for _, version := range versions {
		if utils.IsVersionSupported(version, semverExpr) {
			filteredVersions = append(filteredVersions, version)
		}
	}
	return filteredVersions
}
