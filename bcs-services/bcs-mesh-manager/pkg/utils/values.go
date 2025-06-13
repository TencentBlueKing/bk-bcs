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

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	pointer "k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

const (
	valuesFile = "values.yaml"
)

// GetConfigChartValues 从配置文件中获取istio安装的values
// path目录中包含了以chartVersion命名的文件夹，文件夹中包含了values.yaml文件
// 例如：./config/sample/istio/1.20/values.yaml
// 如果chartVersion中包含了小版本，但是没有对应的文件夹（例如：1.20.1的文件夹），则可以从1.20的文件夹中获取values
func GetConfigChartValues(chartValuesPath, component, chartVersion string) (string, error) {
	if chartValuesPath == "" {
		return "", nil
	}
	// 首先尝试直接匹配chartVersion
	commentValuesFilename := fmt.Sprintf("%s-%s", component, valuesFile)

	targetPath := filepath.Join(chartValuesPath, chartVersion, commentValuesFilename)
	if _, err := os.Stat(targetPath); err == nil {
		content, err := os.ReadFile(targetPath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	// 如果直接匹配失败，尝试匹配主版本号（例如：1.20.1 -> 1.20）
	parts := strings.Split(chartVersion, ".")
	if len(parts) >= 2 {
		majorMinorVersion := strings.Join(parts[:2], ".")
		targetPath = filepath.Join(chartValuesPath, majorMinorVersion, commentValuesFilename)
		if _, err := os.Stat(targetPath); err == nil {
			content, err := os.ReadFile(targetPath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
	}

	// 如果都没有找到，返回空字符串
	return "", nil
}

// MergeValues 合并defaultValues和customValues
func MergeValues(defaultValues, customValues string) (string, error) {
	var defaultValuesMap map[string]interface{}
	var customValuesMap map[string]interface{}

	if err := yaml.Unmarshal([]byte(defaultValues), &defaultValuesMap); err != nil {
		return "", err
	}

	if err := yaml.Unmarshal([]byte(customValues), &customValuesMap); err != nil {
		return "", err
	}

	// 递归合并 customValuesMap 到 defaultValuesMap，customValuesMap 字段覆盖 defaultValuesMap
	if err := mergo.Merge(&defaultValuesMap, customValuesMap, mergo.WithOverride); err != nil {
		return "", err
	}

	merged, err := yaml.Marshal(defaultValuesMap)
	if err != nil {
		return "", err
	}
	return string(merged), nil
}

// GenBaseValues 获取base组件的配置的values
func GenBaseValues(
	chartValuesPath string,
	chartVersion,
	clusterID,
	meshID,
	networkID string,
) (string, error) {
	values, err := GetConfigChartValues(chartValuesPath, common.ComponentIstioBase, chartVersion)
	if err != nil {
		return "", err
	}
	blog.Infof("getBaseValues values: %s for cluster: %s, mesh: %s, network: %s",
		values, clusterID, meshID, networkID)

	return values, nil
}

// GenIstiodValues 获取istiod组件的配置的values
func GenIstiodValues(
	chartValuesPath string,
	installModel,
	chartVersion,
	primaryClusterID,
	remotePilotAddress,
	clusterID,
	meshID,
	networkID string,
	featureConfigs map[string]*meshmanager.FeatureConfig,
) (string, error) {
	values, err := GetConfigChartValues(chartValuesPath, common.ComponentIstiod, chartVersion)
	fmt.Println("values", values)
	if err != nil {
		blog.Errorf("get istiod values failed: %s", err)
		return "", err
	}
	clusterName := strings.ToLower(clusterID)
	primaryClusterName := strings.ToLower(primaryClusterID)
	installArgs := &common.IstiodInstallArgs{
		Global: &common.IstiodGlobalConfig{
			MeshID:  &meshID,
			Network: &networkID,
		},
		MultiCluster: &common.IstiodMultiClusterConfig{
			ClusterName: &clusterName,
		},
	}
	// 获取安装参数
	// 主集群
	if installModel == common.IstioInstallModePrimary {
		installArgs.Global.ExternalIstiod = pointer.Bool(true)
	}

	// 从集群
	if installModel == common.IstioInstallModeRemote {
		installArgs.IstiodRemote = &common.IstiodRemoteConfig{
			Enabled:       pointer.Bool(true),
			InjectionPath: pointer.String(fmt.Sprintf("/inject/cluster/%s/net/%s", primaryClusterName, networkID)),
		}
		installArgs.Pilot.ConfigMap = pointer.Bool(false)
		installArgs.Telemetry.Enabled = pointer.Bool(false)
		installArgs.Global.ConfigCluster = pointer.Bool(true)
		installArgs.Global.RemotePilotAddress = pointer.String(remotePilotAddress)
		installArgs.Global.OmitSidecarInjectorConfigMap = pointer.Bool(true)
	}

	// 填充自定义的参数配置
	err = GenIstiodValuesByFeature(featureConfigs, installArgs)
	if err != nil {
		blog.Errorf("gen istiod values by feature failed: %s", err)
		return "", err
	}

	var valuesMap map[string]interface{}
	if yamlErr := yaml.Unmarshal([]byte(values), &valuesMap); yamlErr != nil {
		blog.Errorf("unmarshal istiod values failed: %s", yamlErr)
		return "", yamlErr
	}
	customValues, err := yaml.Marshal(installArgs)
	if err != nil {
		blog.Errorf("marshal istiod values failed: %s", err)
		return "", err
	}
	mergedValues, err := MergeValues(values, string(customValues))
	if err != nil {
		blog.Errorf("merge istiod values failed: %s", err)
		return "", err
	}

	blog.Infof("gen istiod values: %s for cluster: %s, mesh: %s, network: %s",
		mergedValues, clusterID, meshID, networkID)
	return mergedValues, nil
}

// GenIstiodValuesByFeature 根据featureConfigs生成istiod的values
func GenIstiodValuesByFeature(
	featureConfigs map[string]*meshmanager.FeatureConfig,
	installArgs *common.IstiodInstallArgs,
) error {
	for featureName, featureConfig := range featureConfigs {
		switch featureName {
		case common.FeatureOutboundTrafficPolicy:
			if installArgs.MeshConfig == nil {
				installArgs.MeshConfig = &common.IstiodMeshConfig{}
			}
			installArgs.MeshConfig.OutboundTrafficPolicy = &common.OutboundTrafficPolicy{
				Mode: pointer.String(featureConfig.Value),
			}
		case common.FeatureHoldApplicationUntilProxyStarts:
			if installArgs.MeshConfig == nil {
				installArgs.MeshConfig = &common.IstiodMeshConfig{}
			}
			installArgs.MeshConfig.DefaultConfig = &common.DefaultConfig{
				HoldApplicationUntilProxyStarts: pointer.Bool(featureConfig.Value == "true"),
			}
		case common.FeatureExitOnZeroActiveConnections:
			if installArgs.MeshConfig == nil {
				installArgs.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installArgs.MeshConfig.DefaultConfig == nil {
				installArgs.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			installArgs.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{
				ExitOnZeroActiveConnections: pointer.Bool(featureConfig.Value == "true"),
			}
		case common.FeatureIstioMetaDnsCapture:
			if installArgs.MeshConfig == nil {
				installArgs.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installArgs.MeshConfig.DefaultConfig == nil {
				installArgs.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installArgs.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installArgs.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installArgs.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsCapture = pointer.String(featureConfig.Value)
		case common.FeatureIstioMetaDnsAutoAllocate:
			if installArgs.MeshConfig == nil {
				installArgs.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installArgs.MeshConfig.DefaultConfig == nil {
				installArgs.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installArgs.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installArgs.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installArgs.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsAutoAllocate = pointer.String(featureConfig.Value)
		case common.FeatureIstioMetaHttp10:
			if installArgs.Pilot == nil {
				installArgs.Pilot = &common.IstiodPilotConfig{}
			}
			if installArgs.Pilot.Env == nil {
				installArgs.Pilot.Env = make(map[string]string)
			}
			installArgs.Pilot.Env["PILOT_HTTP10"] = featureConfig.Value
		case common.FeatureExcludeIPRanges:
			if installArgs.Global == nil {
				installArgs.Global = &common.IstiodGlobalConfig{}
			}
			if installArgs.Global.Proxy == nil {
				installArgs.Global.Proxy = &common.IstioProxyConfig{}
			}
			installArgs.Global.Proxy.ExcludeIPRanges = pointer.String(featureConfig.Value)
		}
	}
	return nil
}
