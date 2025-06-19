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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	blog.Infof("GetConfigChartValues chartValuesPath: %s, component: %s, chartVersion: %s",
		chartValuesPath, component, chartVersion)
	if chartValuesPath == "" {
		return "", nil
	}
	// 首先尝试直接匹配chartVersion
	commentValuesFilename := fmt.Sprintf("%s-%s", component, valuesFile)

	targetPath := filepath.Join(chartValuesPath, chartVersion, commentValuesFilename)
	blog.Infof("get chartVersion: %s, targetPath: %s", chartVersion, targetPath)
	if _, err := os.Stat(targetPath); err == nil {
		content, err := os.ReadFile(targetPath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	// 如果直接匹配失败，尝试匹配主版本号（例如：1.20.1 或 1.18-bcs.2 -> 1.20 或 1.18）
	baseVersion := chartVersion
	if idx := strings.Index(baseVersion, "-"); idx >= 0 {
		baseVersion = baseVersion[:idx]
	}
	parts := strings.Split(baseVersion, ".")
	if len(parts) >= 2 {
		majorMinorVersion := strings.Join(parts[:2], ".")
		targetPath = filepath.Join(chartValuesPath, majorMinorVersion, commentValuesFilename)
		blog.Infof("get majorMinorVersion: %s, targetPath: %s", majorMinorVersion, targetPath)
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
	installOption *common.IstioInstallOption,
) (string, error) {
	values, err := GetConfigChartValues(
		installOption.ChartValuesPath,
		common.ComponentIstioBase,
		installOption.ChartVersion,
	)
	if err != nil {
		return "", err
	}
	blog.Infof("getBaseValues values: %s for cluster: %s, mesh: %s, network: %s",
		values, installOption.PrimaryClusters, installOption.MeshID, installOption.NetworkID)

	return values, nil
}

// GenIstiodValues 获取istiod组件的配置的values
func GenIstiodValues(
	installModel string,
	remotePilotAddress string,
	installOption *common.IstioInstallOption,
) (string, error) {
	values, err := GetConfigChartValues(installOption.ChartValuesPath, common.ComponentIstiod, installOption.ChartVersion)
	if err != nil {
		blog.Errorf("get istiod values failed: %s", err)
		return "", err
	}
	blog.Infof("get istiod values: %s for cluster: %s, mesh: %s, network: %s",
		values, installOption.PrimaryClusters, installOption.MeshID, installOption.NetworkID)
	clusterName := strings.ToLower(installOption.PrimaryClusters[0])
	primaryClusterName := strings.ToLower(installOption.PrimaryClusters[0])
	installValues := &common.IstiodInstallValues{
		Global: &common.IstiodGlobalConfig{
			MeshID:  &installOption.MeshID,
			Network: &installOption.NetworkID,
		},
		MultiCluster: &common.IstiodMultiClusterConfig{
			ClusterName: &clusterName,
		},
	}
	// 获取安装参数
	// 主集群
	if installModel == common.IstioInstallModePrimary {
		installValues.Global.ExternalIstiod = pointer.Bool(true)
	}

	// 从集群
	if installModel == common.IstioInstallModeRemote {
		installValues.IstiodRemote = &common.IstiodRemoteConfig{
			Enabled:       pointer.Bool(true),
			InjectionPath: pointer.String(fmt.Sprintf("/inject/cluster/%s/net/%s", primaryClusterName, installOption.NetworkID)),
		}
		installValues.Pilot.ConfigMap = pointer.Bool(false)
		if installValues.Telemetry == nil {
			installValues.Telemetry = &common.IstiodTelemetryConfig{}
		}
		installValues.Telemetry.Enabled = pointer.Bool(false)
		installValues.Global.ConfigCluster = pointer.Bool(true)
		installValues.Global.RemotePilotAddress = pointer.String(remotePilotAddress)
		installValues.Global.OmitSidecarInjectorConfigMap = pointer.Bool(true)
	}
	// proxy resource
	err = GenIstiodValuesBySidecarResource(installOption, installValues)
	if err != nil {
		blog.Errorf("gen istiod values by sidecar resource failed: %s", err)
		return "", err
	}
	// 填充feature的参数配置
	err = GenIstiodValuesByFeature(installOption.FeatureConfigs, installValues)
	if err != nil {
		blog.Errorf("gen istiod values by feature failed: %s", err)
		return "", err
	}

	// 填充observability的参数配置
	err = GenIstiodValuesByObservability(installOption.Version, installOption.ObservabilityConfig, installValues)
	if err != nil {
		blog.Errorf("gen istiod values by observability failed: %s", err)
		return "", err
	}
	// 填充高可用配置
	err = GenIstiodValuesByHighAvailability(installOption.HighAvailability, installValues)
	if err != nil {
		blog.Errorf("gen istiod values by high availability failed: %s", err)
		return "", err
	}

	var valuesMap map[string]interface{}
	if yamlErr := yaml.Unmarshal([]byte(values), &valuesMap); yamlErr != nil {
		blog.Errorf("unmarshal istiod values failed: %s", yamlErr)
		return "", yamlErr
	}
	customValues, err := yaml.Marshal(installValues)
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
		mergedValues, installOption.PrimaryClusters, installOption.MeshID, installOption.NetworkID)
	return mergedValues, nil
}

// GenIstiodValuesBySidecarResource 根据sidecarResourceConfig生成istiod的values
func GenIstiodValuesBySidecarResource(
	installOption *common.IstioInstallOption,
	installValues *common.IstiodInstallValues,
) error {
	if installOption.SidecarResourceConfig != nil {
		if installOption.SidecarResourceConfig.CpuRequest != "" {
			if installValues.Global == nil {
				installValues.Global = &common.IstiodGlobalConfig{}
			}
			if installValues.Global.Proxy == nil {
				installValues.Global.Proxy = &common.IstioProxyConfig{}
			}
			if installOption.SidecarResourceConfig.CpuRequest != "" {
				installValues.Global.Proxy.Resources.Requests[v1.ResourceCPU] =
					resource.MustParse(installOption.SidecarResourceConfig.CpuRequest)
			}
			if installOption.SidecarResourceConfig.CpuLimit != "" {
				installValues.Global.Proxy.Resources.Limits[v1.ResourceCPU] =
					resource.MustParse(installOption.SidecarResourceConfig.CpuLimit)
			}
			if installOption.SidecarResourceConfig.MemoryRequest != "" {
				installValues.Global.Proxy.Resources.Requests[v1.ResourceMemory] =
					resource.MustParse(installOption.SidecarResourceConfig.MemoryRequest)
			}
			if installOption.SidecarResourceConfig.MemoryLimit != "" {
				installValues.Global.Proxy.Resources.Limits[v1.ResourceMemory] =
					resource.MustParse(installOption.SidecarResourceConfig.MemoryLimit)
			}
		}
	}
	return nil
}

// GenIstiodValuesByFeature 根据featureConfigs生成istiod的values
func GenIstiodValuesByFeature(
	featureConfigs map[string]*meshmanager.FeatureConfig,
	installValues *common.IstiodInstallValues,
) error {
	for featureName, featureConfig := range featureConfigs {
		switch featureName {
		case common.FeatureOutboundTrafficPolicy:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			installValues.MeshConfig.OutboundTrafficPolicy = &common.OutboundTrafficPolicy{
				Mode: pointer.String(featureConfig.Value),
			}
		case common.FeatureHoldApplicationUntilProxyStarts:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{
				HoldApplicationUntilProxyStarts: pointer.Bool(featureConfig.Value == "true"),
			}
		case common.FeatureExitOnZeroActiveConnections:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{
				ExitOnZeroActiveConnections: pointer.Bool(featureConfig.Value == "true"),
			}
		case common.FeatureIstioMetaDnsCapture:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installValues.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsCapture = pointer.String(featureConfig.Value)
		case common.FeatureIstioMetaDnsAutoAllocate:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installValues.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsAutoAllocate = pointer.String(featureConfig.Value)
		case common.FeatureIstioMetaHttp10:
			if installValues.Pilot == nil {
				installValues.Pilot = &common.IstiodPilotConfig{}
			}
			if installValues.Pilot.Env == nil {
				installValues.Pilot.Env = make(map[string]string)
			}
			installValues.Pilot.Env["PILOT_HTTP10"] = featureConfig.Value
		case common.FeatureExcludeIPRanges:
			if installValues.Global == nil {
				installValues.Global = &common.IstiodGlobalConfig{}
			}
			if installValues.Global.Proxy == nil {
				installValues.Global.Proxy = &common.IstioProxyConfig{}
			}
			installValues.Global.Proxy.ExcludeIPRanges = pointer.String(featureConfig.Value)
		}
	}
	return nil
}

// GenIstiodValuesByObservability 根据observabilityConfig生成istiod的values
func GenIstiodValuesByObservability(
	istioVersion string,
	observabilityConfig *meshmanager.ObservabilityConfig,
	installValues *common.IstiodInstallValues,
) error {
	if observabilityConfig == nil {
		return nil
	}
	if observabilityConfig.LogCollectorConfig != nil {
		// 日志采集配置，如果启用则配置，否则清空
		if observabilityConfig.LogCollectorConfig.Enabled {
			installValues.MeshConfig.AccessLogFile = pointer.String(common.AccessLogFileStdout)
			if observabilityConfig.LogCollectorConfig.AccessLogFormat != "" {
				installValues.MeshConfig.AccessLogFormat = &observabilityConfig.LogCollectorConfig.AccessLogFormat
			}
			if observabilityConfig.LogCollectorConfig.AccessLogEncoding != "" {
				installValues.MeshConfig.AccessLogEncoding = &observabilityConfig.LogCollectorConfig.AccessLogEncoding
			}
		} else {
			installValues.MeshConfig.AccessLogFile = nil
			installValues.MeshConfig.AccessLogFormat = nil
			installValues.MeshConfig.AccessLogEncoding = nil
		}
	}

	// 全链路追踪配置，如果启用则配置，否则清空
	// 需要区分istio 1.21以上版本通过Telemetry API
	if observabilityConfig.TracingConfig != nil {
		err := GenIstiodValuesByTracing(istioVersion, observabilityConfig.TracingConfig, installValues)
		if err != nil {
			blog.Errorf("gen istiod values by tracing failed: %s", err)
			return err
		}
	}
	return nil
}

// GenIstiodValuesByTracing 根据tracingConfig生成istiod的values
func GenIstiodValuesByTracing(
	istioVersion string,
	tracingConfig *meshmanager.TracingConfig,
	installValues *common.IstiodInstallValues,
) error {
	if tracingConfig == nil {
		return nil
	}
	// istio 1.21以上版本通过Telemetry API
	if IsVersionSupported(istioVersion, "1.21") {
		// TODO: 通过Telemetry API生成istiod的values
		blog.Warnf("istio version %s is supported by Telemetry API, please set tracing config by Telemetry API", istioVersion)
		return nil
	}

	// 关闭全链路追踪
	if !tracingConfig.Enabled {
		installValues.MeshConfig.EnableTracing = pointer.Bool(false)
		return nil
	}
	installValues.MeshConfig.EnableTracing = pointer.Bool(true)

	// istio 1.21以下版本通过Zipkin生成istiod的values
	if installValues.MeshConfig == nil {
		installValues.MeshConfig = &common.IstiodMeshConfig{}
	}
	if installValues.MeshConfig.DefaultConfig == nil {
		installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
	}
	installValues.MeshConfig.DefaultConfig.TracingConfig = &common.TracingConfig{
		Zipkin: &common.ZipkinConfig{
			Address: pointer.String(tracingConfig.Endpoint),
		},
	}
	return nil
}

// GenIstiodValuesByHighAvailability 根据highAvailabilityConfig生成istiod的values
func GenIstiodValuesByHighAvailability(
	highAvailability *meshmanager.HighAvailability,
	installValues *common.IstiodInstallValues,
) error {
	if highAvailability == nil {
		return nil
	}
	// 副本
	if installValues.Pilot == nil {
		installValues.Pilot = &common.IstiodPilotConfig{}
	}
	installValues.Pilot.ReplicaCount = pointer.Int32(highAvailability.ReplicaCount)

	// HPA
	if highAvailability.AutoscaleEnabled {
		installValues.Pilot.AutoscaleEnabled = pointer.Bool(true)
		installValues.Pilot.AutoscaleMin = pointer.Int32(highAvailability.AutoscaleMin)
		installValues.Pilot.AutoscaleMax = pointer.Int32(highAvailability.AutoscaleMax)
		installValues.Pilot.CPU = &common.HPACPUConfig{
			TargetAverageUtilization: pointer.Int32(highAvailability.TargetCPUAverageUtilizationPercent),
		}
	} else {
		installValues.Pilot.AutoscaleEnabled = pointer.Bool(false)
	}

	// pilot资源设置
	if highAvailability.ResourceConfig != nil {
		if highAvailability.ResourceConfig.CpuRequest != "" {
			if installValues.Pilot.Resources == nil {
				installValues.Pilot.Resources = &v1.ResourceRequirements{}
			}
			installValues.Pilot.Resources.Requests[v1.ResourceCPU] =
				resource.MustParse(highAvailability.ResourceConfig.CpuRequest)
		}
		if highAvailability.ResourceConfig.CpuLimit != "" {
			if installValues.Pilot.Resources == nil {
				installValues.Pilot.Resources = &v1.ResourceRequirements{}
			}
			installValues.Pilot.Resources.Limits[v1.ResourceCPU] =
				resource.MustParse(highAvailability.ResourceConfig.CpuLimit)
		}
		if highAvailability.ResourceConfig.MemoryRequest != "" {
			if installValues.Pilot.Resources == nil {
				installValues.Pilot.Resources = &v1.ResourceRequirements{}
			}
			installValues.Pilot.Resources.Requests[v1.ResourceMemory] =
				resource.MustParse(highAvailability.ResourceConfig.MemoryRequest)
		}
		if highAvailability.ResourceConfig.MemoryLimit != "" {
			if installValues.Pilot.Resources == nil {
				installValues.Pilot.Resources = &v1.ResourceRequirements{}
			}
			installValues.Pilot.Resources.Limits[v1.ResourceMemory] =
				resource.MustParse(highAvailability.ResourceConfig.MemoryLimit)
		}
	}
	// 专属节点
	if highAvailability.DedicatedNode != nil {
		if highAvailability.DedicatedNode.Enabled {
			if installValues.Pilot.NodeSelector == nil {
				installValues.Pilot.NodeSelector = make(map[string]string)
			}
			for k, v := range highAvailability.DedicatedNode.NodeLabels {
				installValues.Pilot.NodeSelector[k] = v
			}
		} else {
			installValues.Pilot.NodeSelector = nil
		}
	}
	return nil
}
