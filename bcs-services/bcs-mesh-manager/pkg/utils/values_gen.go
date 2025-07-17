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
	// OtelTracingName 全链路追踪的名称
	OtelTracingName = "otel-tracing"
	// OtelTracingPath 全链路追踪的path
	OtelTracingPath = "/v1/traces"
	// OtelTracingTimeout 全链路追踪的timeout
	OtelTracingTimeout = "10s"
	// OtelTracingHeader 全链路追踪的header
	OtelTracingHeader = "X-BK-TOKEN"
)

// GetConfigChartValues 从配置文件中获取istio安装的values
// path目录中包含了以chartVersion命名的文件夹，文件夹中包含了values.yaml文件
// 例如：./config/sample/istio/1.20/base-values.yaml
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
		if installValues.Pilot == nil {
			installValues.Pilot = &common.IstiodPilotConfig{}
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
	err = GenIstiodValuesBySidecarResource(installOption.SidecarResourceConfig, installValues)
	if err != nil {
		blog.Errorf("gen istiod values by sidecar resource failed: %s", err)
		return "", err
	}
	// 填充feature的参数配置
	GenIstiodValuesByFeature(installOption.FeatureConfigs, installValues)

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

// setResourceRequirement 通用的资源设置函数
func setResourceRequirement(
	resources **common.ResourceConfig,
	resourceType v1.ResourceName,
	value string,
	isLimit bool,
) error {
	if value == "" {
		return nil
	}
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return fmt.Errorf("parse quantity %s failed: %s", value, err)
	}
	if quantity.IsZero() {
		return nil
	}

	// 初始化 Resources 结构
	if *resources == nil {
		*resources = &common.ResourceConfig{}
	}

	// 设置对应的资源值
	if isLimit {
		if (*resources).Limits == nil {
			(*resources).Limits = &common.ResourceLimits{}
		}
		switch resourceType {
		case v1.ResourceCPU:
			(*resources).Limits.CPU = pointer.String(value)
		case v1.ResourceMemory:
			(*resources).Limits.Memory = pointer.String(value)
		}
	} else {
		if (*resources).Requests == nil {
			(*resources).Requests = &common.ResourceRequests{}
		}
		switch resourceType {
		case v1.ResourceCPU:
			(*resources).Requests.CPU = pointer.String(value)
		case v1.ResourceMemory:
			(*resources).Requests.Memory = pointer.String(value)
		}
	}

	return nil
}

// applyResourceConfig 应用资源配置到指定的 Resources 对象
func applyResourceConfig(
	resources **common.ResourceConfig,
	resourceConfig *meshmanager.ResourceConfig,
) error {
	if resourceConfig == nil {
		return nil
	}

	// 设置 CPU 请求
	cpuRequest := resourceConfig.CpuRequest.GetValue()
	if err := setResourceRequirement(resources, v1.ResourceCPU, cpuRequest, false); err != nil {
		return err
	}

	// 设置 CPU 限制
	cpuLimit := resourceConfig.CpuLimit.GetValue()
	if err := setResourceRequirement(resources, v1.ResourceCPU, cpuLimit, true); err != nil {
		return err
	}

	// 设置内存请求
	memoryRequest := resourceConfig.MemoryRequest.GetValue()
	if err := setResourceRequirement(resources, v1.ResourceMemory, memoryRequest, false); err != nil {
		return err
	}

	// 设置内存限制
	memoryLimit := resourceConfig.MemoryLimit.GetValue()
	if err := setResourceRequirement(resources, v1.ResourceMemory, memoryLimit, true); err != nil {
		return err
	}

	return nil
}

// GenIstiodValuesBySidecarResource 根据sidecarResourceConfig生成istiod的values
func GenIstiodValuesBySidecarResource(
	sidecarResourceConfig *meshmanager.ResourceConfig,
	installValues *common.IstiodInstallValues,
) error {
	if sidecarResourceConfig == nil {
		return nil
	}

	// 初始化 Global.Proxy 结构
	if installValues.Global == nil {
		installValues.Global = &common.IstiodGlobalConfig{}
	}
	if installValues.Global.Proxy == nil {
		installValues.Global.Proxy = &common.IstioProxyConfig{}
	}

	// 使用通用函数应用资源配置
	return applyResourceConfig(&installValues.Global.Proxy.Resources, sidecarResourceConfig)
}

// GenIstiodValuesByFeature 根据featureConfigs生成istiod的values
func GenIstiodValuesByFeature(
	featureConfigs map[string]*meshmanager.FeatureConfig,
	installValues *common.IstiodInstallValues,
) {
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
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			// nolint:lll
			installValues.MeshConfig.DefaultConfig.HoldApplicationUntilProxyStarts = pointer.Bool(featureConfig.Value == common.StringTrue)
		case common.FeatureExitOnZeroActiveConnections:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installValues.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			// nolint:lll
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.ExitOnZeroActiveConnections = pointer.Bool(featureConfig.Value == common.StringTrue)
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
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsCapture =
				pointer.String(featureConfig.Value)
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
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsAutoAllocate =
				pointer.String(featureConfig.Value)
		case common.FeatureIstioMetaHttp10:
			if installValues.Pilot == nil {
				installValues.Pilot = &common.IstiodPilotConfig{}
			}
			if installValues.Pilot.Env == nil {
				installValues.Pilot.Env = make(map[string]string)
			}
			installValues.Pilot.Env[common.EnvPilotHTTP10] = featureConfig.Value
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
		if installValues.MeshConfig == nil {
			installValues.MeshConfig = &common.IstiodMeshConfig{}
		}

		// 日志采集配置，如果启用则配置，否则不设置相关字段
		if observabilityConfig.LogCollectorConfig.Enabled.GetValue() {
			installValues.MeshConfig.AccessLogFile = pointer.String(common.AccessLogFileStdout)
			if observabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue() != "" {
				installValues.MeshConfig.AccessLogFormat =
					pointer.String(observabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue())
			}
			if observabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue() != "" {
				installValues.MeshConfig.AccessLogEncoding =
					pointer.String(observabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue())
			}
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
// nolint:funlen
func GenIstiodValuesByTracing(
	istioVersion string,
	tracingConfig *meshmanager.TracingConfig,
	installValues *common.IstiodInstallValues,
) error {
	if tracingConfig == nil {
		return nil
	}

	if installValues.MeshConfig == nil {
		installValues.MeshConfig = &common.IstiodMeshConfig{}
	}

	if !tracingConfig.Enabled.GetValue() {
		installValues.MeshConfig.EnableTracing = pointer.Bool(false)
		return nil
	}
	// istio 1.21以上版本通过Telemetry API
	if IsVersionSupported(istioVersion, ">=1.21") {
		installValues.MeshConfig.EnableTracing = pointer.Bool(true)

		// 设置采样率
		if tracingConfig.TraceSamplingPercent.GetValue() != 0 {
			if installValues.Pilot == nil {
				installValues.Pilot = &common.IstiodPilotConfig{}
			}
			installValues.Pilot.TraceSampling = pointer.Float64(float64(tracingConfig.TraceSamplingPercent.GetValue()) / 100)
		}

		if installValues.MeshConfig.ExtensionProviders == nil {
			installValues.MeshConfig.ExtensionProviders = []*common.ExtensionProvider{}
		}
		// 解析endpoint获取service、port和path
		service, port, path, err := ParseOpenTelemetryEndpoint(tracingConfig.Endpoint.GetValue())
		if err != nil {
			blog.Errorf("parse endpoint %s failed: %s", tracingConfig.Endpoint.GetValue(), err)
			return err
		}
		if path == "" {
			blog.Warnf("path is empty, use default path: %s, endpoint: %s", OtelTracingPath, tracingConfig.Endpoint.GetValue())
			path = OtelTracingPath
		}
		installValues.MeshConfig.ExtensionProviders = append(installValues.MeshConfig.ExtensionProviders,
			&common.ExtensionProvider{
				Name: pointer.String(OtelTracingName),
				OpenTelemetry: &common.OpenTelemetryConfig{
					Service: pointer.String(service),
					Port:    pointer.Int32(port),
					Http: &common.OpenTelemetryHttpConfig{
						Path:    pointer.String(path),
						Timeout: pointer.String(OtelTracingTimeout),
						Headers: map[string]string{
							OtelTracingHeader: tracingConfig.BkToken.GetValue(),
						},
					},
				},
			})
		blog.Infof("istio version %s is supported by Telemetry API, set tracing config by Telemetry API", istioVersion)
		return nil
	}

	// istio 1.21以下版本通过Zipkin生成istiod的values
	installValues.MeshConfig.EnableTracing = pointer.Bool(true)

	if installValues.MeshConfig.DefaultConfig == nil {
		installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
	}
	installValues.MeshConfig.DefaultConfig.TracingConfig = &common.TracingConfig{
		Zipkin: &common.ZipkinConfig{
			Address: pointer.String(tracingConfig.Endpoint.GetValue()),
		},
	}
	// 采样率
	if tracingConfig.TraceSamplingPercent.GetValue() != 0 {
		if installValues.Pilot == nil {
			installValues.Pilot = &common.IstiodPilotConfig{}
		}
		installValues.Pilot.TraceSampling = pointer.Float64(float64(tracingConfig.TraceSamplingPercent.GetValue()) / 100)
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
	installValues.Pilot.ReplicaCount = pointer.Int32(highAvailability.ReplicaCount.GetValue())

	// HPA
	if highAvailability.AutoscaleEnabled.GetValue() {
		installValues.Pilot.AutoscaleEnabled = pointer.Bool(true)
		installValues.Pilot.AutoscaleMin = pointer.Int32(highAvailability.AutoscaleMin.GetValue())
		installValues.Pilot.AutoscaleMax = pointer.Int32(highAvailability.AutoscaleMax.GetValue())
		installValues.Pilot.CPU = &common.HPACPUConfig{
			TargetAverageUtilization: pointer.Int32(highAvailability.TargetCPUAverageUtilizationPercent.GetValue()),
		}
	} else {
		installValues.Pilot.AutoscaleEnabled = pointer.Bool(false)
	}

	// pilot资源设置
	if highAvailability.ResourceConfig != nil {
		// 使用通用函数应用资源配置
		if err := applyResourceConfig(&installValues.Pilot.Resources, highAvailability.ResourceConfig); err != nil {
			return err
		}
	}
	// 专属节点
	if highAvailability.DedicatedNode != nil {
		if highAvailability.DedicatedNode.Enabled.GetValue() {
			if installValues.Pilot.NodeSelector == nil {
				installValues.Pilot.NodeSelector = make(map[string]string)
			}
			for k, v := range highAvailability.DedicatedNode.NodeLabels {
				installValues.Pilot.NodeSelector[k] = v
			}
			// 增加容忍, 所有节点
			if installValues.Pilot.Tolerations == nil {
				installValues.Pilot.Tolerations = make([]v1.Toleration, 0)
			}
			installValues.Pilot.Tolerations = append(installValues.Pilot.Tolerations, v1.Toleration{
				Operator: v1.TolerationOpExists,
			})
		}
	}
	return nil
}
