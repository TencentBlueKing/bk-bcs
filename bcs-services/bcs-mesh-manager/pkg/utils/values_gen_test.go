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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"gopkg.in/yaml.v2"
	pointer "k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

func TestMergeValues(t *testing.T) {
	cases := []struct {
		name          string
		defaultValues string
		customValues  string
		expect        string
	}{
		{
			name: "simple merge",
			defaultValues: `a: 1
b: 2`,
			customValues: `b: 3
c: 4`,
			expect: "a: 1\nb: 3\nc: 4\n",
		},
		{
			name: "nested merge",
			defaultValues: `a:
  b: 1
  c: 2
x: 5`,
			customValues: `a:
  c: 3
d: 4`,
			expect: "a:\n  b: 1\n  c: 3\nx: 5\nd: 4\n",
		},
		{
			name: "custom overrides default",
			defaultValues: `foo: bar
bar: baz`,
			customValues: `foo: newbar`,
			expect:       "foo: newbar\nbar: baz\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := MergeValues(c.defaultValues, c.customValues)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var expectMap, resultMap map[string]interface{}
			if err := yaml.Unmarshal([]byte(c.expect), &expectMap); err != nil {
				t.Fatalf("unmarshal expect failed: %v", err)
			}
			if err := yaml.Unmarshal([]byte(result), &resultMap); err != nil {
				t.Fatalf("unmarshal result failed: %v", err)
			}
			if !reflect.DeepEqual(expectMap, resultMap) {
				t.Errorf("merge failed.\nExpected:\n%v\nGot:\n%v", expectMap, resultMap)
			}
		})
	}
}

func TestGenIstiodValues(t *testing.T) {
	dir := t.TempDir()
	// mock values.yaml 文件
	istiodYaml := "global:\n  bar: baz\nmeshConfig:\n  outboundTrafficPolicy:\n    mode: ALLOW_ANY"
	os.MkdirAll(dir+"/1.24", 0755)
	os.WriteFile(dir+"/1.24/istiod-values.yaml", []byte(istiodYaml), 0644)

	tests := []struct {
		name           string
		installMode    string
		remotePilot    string
		installOption  *common.IstioInstallOption
		expectedFields []string
		notExpected    []string
	}{
		{
			name:        "basic primary cluster configuration",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-123",
				NetworkID:       "net-456",
				FeatureConfigs: map[string]*meshmanager.FeatureConfig{
					"outboundTrafficPolicy": {
						Name:  "outboundTrafficPolicy",
						Value: "REGISTRY_ONLY",
					},
				},
			},
			expectedFields: []string{"mesh-123", "net-456", "REGISTRY_ONLY", "externalIstiod: true"},
		},
		{
			name:        "sidecar resource configuration",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-sidecar",
				NetworkID:       "net-sidecar",
				SidecarResourceConfig: &meshmanager.ResourceConfig{
					CpuRequest:    &wrappers.StringValue{Value: "100m"},
					CpuLimit:      &wrappers.StringValue{Value: "200m"},
					MemoryRequest: &wrappers.StringValue{Value: "128Mi"},
					MemoryLimit:   &wrappers.StringValue{Value: "256Mi"},
				},
			},
			expectedFields: []string{"mesh-sidecar", "net-sidecar", "proxy:", "resources:", "limits:", "requests:"},
		},
		{
			name:        "high availability configuration",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-ha",
				NetworkID:       "net-ha",
				HighAvailability: &meshmanager.HighAvailability{
					ReplicaCount:                       &wrappers.Int32Value{Value: 3},
					AutoscaleEnabled:                   &wrappers.BoolValue{Value: true},
					AutoscaleMin:                       &wrappers.Int32Value{Value: 2},
					AutoscaleMax:                       &wrappers.Int32Value{Value: 5},
					TargetCPUAverageUtilizationPercent: &wrappers.Int32Value{Value: 80},
					ResourceConfig: &meshmanager.ResourceConfig{
						CpuRequest:    &wrappers.StringValue{Value: "500m"},
						CpuLimit:      &wrappers.StringValue{Value: "1000m"},
						MemoryRequest: &wrappers.StringValue{Value: "512Mi"},
						MemoryLimit:   &wrappers.StringValue{Value: "1Gi"},
					},
					DedicatedNode: &meshmanager.DedicatedNode{
						Enabled: &wrappers.BoolValue{Value: true},
						NodeLabels: map[string]string{
							"node-type": "istio-control",
							"zone":      "az1",
						},
					},
				},
			},
			expectedFields: []string{
				"mesh-ha", "net-ha", "replicaCount: 3", "autoscaleEnabled: true",
				"autoscaleMin: 2", "autoscaleMax: 5", "targetAverageUtilization: 80",
				"pilot:", "resources:", "node-type", "istio-control", "zone", "az1",
			},
		},
		{
			name:        "observability configuration with tracing",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-obs",
				NetworkID:       "net-obs",
				Version:         "1.24.0",
				ObservabilityConfig: &meshmanager.ObservabilityConfig{
					LogCollectorConfig: &meshmanager.LogCollectorConfig{
						Enabled:           &wrappers.BoolValue{Value: true},
						AccessLogEncoding: &wrappers.StringValue{Value: "JSON"},
						AccessLogFormat:   &wrappers.StringValue{Value: `{"timestamp":"%START_TIME%"}`},
					},
					TracingConfig: &meshmanager.TracingConfig{
						Enabled:              &wrappers.BoolValue{Value: true},
						Endpoint:             &wrappers.StringValue{Value: "http://jaeger-collector.istio-system:14268/api/traces"},
						BkToken:              &wrappers.StringValue{Value: "test-token"},
						TraceSamplingPercent: &wrappers.Int32Value{Value: 10},
					},
				},
			},
			expectedFields: []string{
				"mesh-obs", "net-obs", "accessLogFile", "/dev/stdout",
				"JSON", `"timestamp"`, "enableTracing: true", "jaeger-collector.istio-system",
				"traceSampling: 0.1",
			},
		},
		{
			name:        "observability configuration with legacy tracing (< 1.21)",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-legacy",
				NetworkID:       "net-legacy",
				Version:         "1.20.0", // 版本 < 1.21，使用 Zipkin
				ObservabilityConfig: &meshmanager.ObservabilityConfig{
					TracingConfig: &meshmanager.TracingConfig{
						Enabled:              &wrappers.BoolValue{Value: true},
						Endpoint:             &wrappers.StringValue{Value: "http://zipkin.istio-system:9411/api/v2/spans"},
						TraceSamplingPercent: &wrappers.Int32Value{Value: 5},
					},
				},
			},
			expectedFields: []string{
				"mesh-legacy", "net-legacy", "enableTracing: true",
				"zipkin.istio-system:9411", "traceSampling: 0.05",
			},
		},
		{
			name:        "comprehensive feature configuration",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-features",
				NetworkID:       "net-features",
				FeatureConfigs: map[string]*meshmanager.FeatureConfig{
					common.FeatureOutboundTrafficPolicy: {
						Name:  common.FeatureOutboundTrafficPolicy,
						Value: "REGISTRY_ONLY",
					},
					common.FeatureHoldApplicationUntilProxyStarts: {
						Name:  common.FeatureHoldApplicationUntilProxyStarts,
						Value: "true",
					},
					common.FeatureExitOnZeroActiveConnections: {
						Name:  common.FeatureExitOnZeroActiveConnections,
						Value: "true",
					},
					common.FeatureIstioMetaDnsCapture: {
						Name:  common.FeatureIstioMetaDnsCapture,
						Value: "true",
					},
					common.FeatureIstioMetaDnsAutoAllocate: {
						Name:  common.FeatureIstioMetaDnsAutoAllocate,
						Value: "true",
					},
					common.FeatureIstioMetaHttp10: {
						Name:  common.FeatureIstioMetaHttp10,
						Value: "1",
					},
					common.FeatureExcludeIPRanges: {
						Name:  common.FeatureExcludeIPRanges,
						Value: "10.0.0.0/8,172.16.0.0/12",
					},
				},
			},
			expectedFields: []string{
				"mesh-features", "net-features", "REGISTRY_ONLY",
				"holdApplicationUntilProxyStarts: true",
				"EXIT_ON_ZERO_ACTIVE_CONNECTIONS: true",
				"ISTIO_META_DNS_CAPTURE: \"true\"",
				"PILOT_HTTP10",
				"excludeIPRanges", "10.0.0.0/8,172.16.0.0/12",
			},
		},
		{
			name:        "remote cluster configuration",
			installMode: common.IstioInstallModeRemote,
			remotePilot: "pilot.istio-system.svc.cluster.local",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-remote",
				NetworkID:       "net-remote",
			},
			expectedFields: []string{
				"mesh-remote", "net-remote", "configCluster: true",
				"remotePilotAddress", "pilot.istio-system.svc.cluster.local",
				"istiodRemote:", "enabled: true", "injectionPath",
				"configMap: false", "telemetry:", "enabled: false",
			},
		},
		{
			name:        "resource configuration with actual values verification",
			installMode: common.IstioInstallModePrimary,
			remotePilot: "",
			installOption: &common.IstioInstallOption{
				ChartValuesPath: dir,
				ChartVersion:    "1.24",
				PrimaryClusters: []string{"primary-cluster"},
				MeshID:          "mesh-resource",
				NetworkID:       "net-resource",
				SidecarResourceConfig: &meshmanager.ResourceConfig{
					CpuRequest:    &wrappers.StringValue{Value: "50m"},
					CpuLimit:      &wrappers.StringValue{Value: "100m"},
					MemoryRequest: &wrappers.StringValue{Value: "64Mi"},
					MemoryLimit:   &wrappers.StringValue{Value: "128Mi"},
				},
				HighAvailability: &meshmanager.HighAvailability{
					ReplicaCount: &wrappers.Int32Value{Value: 2},
					ResourceConfig: &meshmanager.ResourceConfig{
						CpuRequest:    &wrappers.StringValue{Value: "200m"},
						CpuLimit:      &wrappers.StringValue{Value: "500m"},
						MemoryRequest: &wrappers.StringValue{Value: "256Mi"},
						MemoryLimit:   &wrappers.StringValue{Value: "512Mi"},
					},
				},
			},
			expectedFields: []string{
				"mesh-resource", "net-resource",
				// Sidecar 资源配置 - 验证实际值而不是格式信息
				"proxy:", "resources:",
				"cpu: 50m", "cpu: 100m", "memory: 64Mi", "memory: 128Mi",
				// Pilot 资源配置 - 验证实际值
				"pilot:", "replicaCount: 2",
				"cpu: 200m", "cpu: 500m", "memory: 256Mi", "memory: 512Mi",
			},
			// 确保不包含旧的格式信息
			notExpected: []string{
				"format: DecimalSI", "format: BinarySI",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenIstiodValues(tt.installMode, tt.remotePilot, tt.installOption)
			if err != nil {
				t.Fatalf("GenIstiodValues error: %v", err)
			}

			t.Logf("Test: %s\nResult:\n%s", tt.name, result)

			// 检查期望的字段
			for _, expected := range tt.expectedFields {
				if !strings.Contains(result, expected) {
					t.Errorf("expected field '%s' not found in result", expected)
				}
			}

			// 检查不应该存在的字段
			for _, notExpected := range tt.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("unexpected field '%s' found in result", notExpected)
				}
			}

			// 验证生成的 YAML 是否有效
			var yamlMap map[string]interface{}
			if err := yaml.Unmarshal([]byte(result), &yamlMap); err != nil {
				t.Errorf("generated YAML is invalid: %v", err)
			}
		})
	}
}

// TestGenIstiodValuesByComponents 测试各个组件的值生成函数
func TestGenIstiodValuesByComponents(t *testing.T) {
	t.Run("TestGenIstiodValuesBySidecarResource", func(t *testing.T) {
		installValues := &common.IstiodInstallValues{}
		resourceConfig := &meshmanager.ResourceConfig{
			CpuRequest:    &wrappers.StringValue{Value: "100m"},
			CpuLimit:      &wrappers.StringValue{Value: "200m"},
			MemoryRequest: &wrappers.StringValue{Value: "128Mi"},
			MemoryLimit:   &wrappers.StringValue{Value: "256Mi"},
		}

		err := GenIstiodValuesBySidecarResource(resourceConfig, installValues)
		if err != nil {
			t.Fatalf("GenIstiodValuesBySidecarResource error: %v", err)
		}

		if installValues.Global == nil || installValues.Global.Proxy == nil || installValues.Global.Proxy.Resources == nil {
			t.Fatal("sidecar resources not set")
		}

		resources := installValues.Global.Proxy.Resources
		if resources.Requests == nil || resources.Limits == nil {
			t.Fatal("requests or limits not set")
		}

		// 验证 CPU 和内存设置
		if resources.Requests.CPU == nil || *resources.Requests.CPU != "100m" {
			t.Errorf("expected CPU request 100m, got %v", resources.Requests.CPU)
		}

		if resources.Limits.CPU == nil || *resources.Limits.CPU != "200m" {
			t.Errorf("expected CPU limit 200m, got %v", resources.Limits.CPU)
		}
	})

	t.Run("TestGenIstiodValuesByHighAvailability", func(t *testing.T) {
		installValues := &common.IstiodInstallValues{}
		haConfig := &meshmanager.HighAvailability{
			ReplicaCount:                       &wrappers.Int32Value{Value: 3},
			AutoscaleEnabled:                   &wrappers.BoolValue{Value: true},
			AutoscaleMin:                       &wrappers.Int32Value{Value: 2},
			AutoscaleMax:                       &wrappers.Int32Value{Value: 5},
			TargetCPUAverageUtilizationPercent: &wrappers.Int32Value{Value: 80},
			ResourceConfig: &meshmanager.ResourceConfig{
				CpuRequest:    &wrappers.StringValue{Value: "500m"},
				CpuLimit:      &wrappers.StringValue{Value: "1000m"},
				MemoryRequest: &wrappers.StringValue{Value: "512Mi"},
				MemoryLimit:   &wrappers.StringValue{Value: "1Gi"},
			},
		}

		err := GenIstiodValuesByHighAvailability(haConfig, installValues)
		if err != nil {
			t.Fatalf("GenIstiodValuesByHighAvailability error: %v", err)
		}

		if installValues.Pilot == nil {
			t.Fatal("pilot config not set")
		}

		pilot := installValues.Pilot
		if *pilot.ReplicaCount != 3 {
			t.Errorf("expected replica count 3, got %d", *pilot.ReplicaCount)
		}

		if !*pilot.AutoscaleEnabled {
			t.Error("expected autoscale enabled")
		}

		if *pilot.AutoscaleMin != 2 || *pilot.AutoscaleMax != 5 {
			t.Errorf("expected autoscale min/max 2/5, got %d/%d", *pilot.AutoscaleMin, *pilot.AutoscaleMax)
		}
	})

	t.Run("TestGenIstiodValuesByObservability", func(t *testing.T) {
		installValues := &common.IstiodInstallValues{}
		obsConfig := &meshmanager.ObservabilityConfig{
			LogCollectorConfig: &meshmanager.LogCollectorConfig{
				Enabled:           &wrappers.BoolValue{Value: true},
				AccessLogEncoding: &wrappers.StringValue{Value: "JSON"},
				AccessLogFormat:   &wrappers.StringValue{Value: `{"timestamp":"%START_TIME%"}`},
			},
			TracingConfig: &meshmanager.TracingConfig{
				Enabled:              &wrappers.BoolValue{Value: true},
				Endpoint:             &wrappers.StringValue{Value: "http://jaeger-collector.istio-system:14268/api/traces"},
				BkToken:              &wrappers.StringValue{Value: "test-token"},
				TraceSamplingPercent: &wrappers.Int32Value{Value: 10},
			},
		}

		err := GenIstiodValuesByObservability("1.24.0", obsConfig, installValues)
		if err != nil {
			t.Fatalf("GenIstiodValuesByObservability error: %v", err)
		}

		if installValues.MeshConfig == nil {
			t.Fatal("mesh config not set")
		}

		meshConfig := installValues.MeshConfig
		if meshConfig.AccessLogFile == nil || *meshConfig.AccessLogFile != "/dev/stdout" {
			t.Error("access log file not set correctly")
		}

		if meshConfig.AccessLogEncoding == nil || *meshConfig.AccessLogEncoding != "JSON" {
			t.Error("access log encoding not set correctly")
		}
	})

	t.Run("TestGenIstiodValuesByTracingCleanup", func(t *testing.T) {
		// 创建测试用的 installValues，包含追踪相关字段
		installValues := &common.IstiodInstallValues{
			MeshConfig: &common.IstiodMeshConfig{
				EnableTracing: pointer.Bool(true),
				ExtensionProviders: []*common.ExtensionProvider{
					{
						Name: pointer.String("otel-tracing"),
						OpenTelemetry: &common.OpenTelemetryConfig{
							Service: pointer.String("jaeger-collector.istio-system"),
							Port:    pointer.Int32(14268),
						},
					},
				},
				DefaultConfig: &common.DefaultConfig{
					TracingConfig: &common.TracingConfig{
						Zipkin: &common.ZipkinConfig{
							Address: pointer.String("http://zipkin:9411/api/v2/spans"),
						},
					},
				},
			},
			Pilot: &common.IstiodPilotConfig{
				TraceSampling: pointer.Float64(0.1),
			},
		}

		// 创建禁用的追踪配置
		tracingConfig := &meshmanager.TracingConfig{
			Enabled: &wrappers.BoolValue{Value: false},
		}

		// 调用 GenIstiodValuesByTracing
		err := GenIstiodValuesByTracing("1.21.0", tracingConfig, installValues)
		if err != nil {
			t.Fatalf("GenIstiodValuesByTracing error: %v", err)
		}

		// 验证 EnableTracing 被设置为 false
		if installValues.MeshConfig.EnableTracing == nil || *installValues.MeshConfig.EnableTracing != false {
			t.Error("expected EnableTracing to be false")
		}

		// 验证其他字段保持不变（不期望被清理）
		if installValues.MeshConfig.ExtensionProviders == nil {
			t.Error("expected ExtensionProviders to remain unchanged")
		}

		if installValues.MeshConfig.DefaultConfig == nil || installValues.MeshConfig.DefaultConfig.TracingConfig == nil {
			t.Error("expected TracingConfig to remain unchanged")
		}

		if installValues.Pilot == nil || installValues.Pilot.TraceSampling == nil {
			t.Error("expected TraceSampling to remain unchanged")
		}

		// 测试旧版本（< 1.21）关闭追踪的逻辑
		installValues2 := &common.IstiodInstallValues{
			MeshConfig: &common.IstiodMeshConfig{
				EnableTracing: pointer.Bool(true),
				DefaultConfig: &common.DefaultConfig{
					TracingConfig: &common.TracingConfig{
						Zipkin: &common.ZipkinConfig{
							Address: pointer.String("http://zipkin:9411/api/v2/spans"),
						},
					},
				},
			},
			Pilot: &common.IstiodPilotConfig{
				TraceSampling: pointer.Float64(0.1),
			},
		}

		err = GenIstiodValuesByTracing("1.20.0", tracingConfig, installValues2)
		if err != nil {
			t.Fatalf("GenIstiodValuesByTracing error: %v", err)
		}

		// 验证 EnableTracing 被设置为 false
		if installValues2.MeshConfig.EnableTracing == nil || *installValues2.MeshConfig.EnableTracing != false {
			t.Error("expected EnableTracing to be false for legacy version")
		}

		// 验证其他字段保持不变（不期望被清理）
		if installValues2.MeshConfig.DefaultConfig == nil || installValues2.MeshConfig.DefaultConfig.TracingConfig == nil {
			t.Error("expected TracingConfig to remain unchanged for legacy version")
		}

		if installValues2.Pilot == nil || installValues2.Pilot.TraceSampling == nil {
			t.Error("expected TraceSampling to remain unchanged for legacy version")
		}
	})

}

func TestGetConfigChartValues(t *testing.T) {
	dir := t.TempDir()
	component := "base"
	chartVersion := "1.18-bcs.2"
	majorMinor := "1.18"
	filename := component + "-values.yaml"

	// 1. 精确版本匹配
	verDir := filepath.Join(dir, chartVersion)
	os.MkdirAll(verDir, 0755)
	verFile := filepath.Join(verDir, filename)
	os.WriteFile(verFile, []byte("exact: true\n"), 0644)

	// 2. 主版本号匹配
	majorDir := filepath.Join(dir, majorMinor)
	os.MkdirAll(majorDir, 0755)
	majorFile := filepath.Join(majorDir, filename)
	os.WriteFile(majorFile, []byte("major: true\n"), 0644)

	t.Run("exact match", func(t *testing.T) {
		val, err := GetConfigChartValues(dir, component, chartVersion)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "exact: true\n" {
			t.Errorf("expected exact match, got: %q", val)
		}
	})

	t.Run("major.minor match", func(t *testing.T) {
		val, err := GetConfigChartValues(dir, component, "1.18-bcs.3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "major: true\n" {
			t.Errorf("expected major.minor match, got: %q", val)
		}
	})

	t.Run("not found", func(t *testing.T) {
		val, err := GetConfigChartValues(dir, component, "2.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "" {
			t.Errorf("expected empty string for not found, got: %q", val)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		val, err := GetConfigChartValues("", component, chartVersion)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "" {
			t.Errorf("expected empty string for empty path, got: %q", val)
		}
	})
}

func TestMergeValuesWithAutoscaleDisabled(t *testing.T) {
	// 测试当 autoscaleEnabled 为 false 时，相关字段是否被正确删除
	defaultValues := `pilot:
  autoscaleEnabled: true
  autoscaleMin: 3
  autoscaleMax: 6
  cpu:
    targetAverageUtilization: 80
  replicaCount: 4
  resources:
    requests:
      cpu: "600m"
      memory: "512Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
global:
  externalIstiod: true`

	customValues := `pilot:
  autoscaleEnabled: false
  replicaCount: 4`

	result, err := MergeValues(defaultValues, customValues)
	if err != nil {
		t.Fatalf("MergeValues error: %v", err)
	}

	t.Logf("Merge result: %s", result)

	// 解析结果
	var resultMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &resultMap); err != nil {
		t.Fatalf("unmarshal result failed: %v", err)
	}

	t.Logf("Parsed result map: %+v", resultMap)

	// 检查 pilot 配置
	pilotConfig, ok := resultMap["pilot"].(map[interface{}]interface{})
	if !ok {
		t.Logf("pilot config type assertion failed, actual type: %T", resultMap["pilot"])
		t.Logf("pilot config value: %+v", resultMap["pilot"])
		t.Fatal("pilot config not found in result")
	}

	t.Logf("Pilot config: %+v", pilotConfig)

	// 验证 autoscaleEnabled 字段应该保留（因为这是用户设置的）
	if _, exists := pilotConfig["autoscaleEnabled"]; !exists {
		t.Error("autoscaleEnabled should be preserved as it's user-specified")
	}

	// 验证相关字段应该保留（MergeValues 只是合并，不处理字段删除逻辑）
	// 字段删除逻辑应该在 actions 层通过 ProcessFieldKey 来处理
	for _, k := range []string{"autoscaleMin", "autoscaleMax", "cpu"} {
		if _, exists := pilotConfig[k]; !exists {
			t.Errorf("%s should be preserved by MergeValues (field deletion is handled by ProcessFieldKey)", k)
		}
	}

	// 验证其他字段应该保留
	for _, k := range []string{"replicaCount", "resources"} {
		found := false
		for key := range pilotConfig {
			if ks, ok := key.(string); ok && ks == k {
				found = true
			}
		}
		if !found {
			t.Errorf("%s should be preserved but was deleted", k)
		}
	}

	t.Logf("Test passed! Result: %s", result)
}

func TestMergeValuesWithAutoscaleEnabled(t *testing.T) {
	// 测试当 autoscaleEnabled 为 true 时，相关字段应该保留
	defaultValues := `pilot:
  autoscaleEnabled: false
  autoscaleMin: 3
  autoscaleMax: 6
  cpu:
    targetAverageUtilization: 80
  replicaCount: 4
  resources:
    requests:
      cpu: "600m"
      memory: "512Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
global:
  externalIstiod: true`

	customValues := `pilot:
  autoscaleEnabled: true
  autoscaleMin: 2
  autoscaleMax: 5
  cpu:
    targetAverageUtilization: 70
  replicaCount: 4`

	result, err := MergeValues(defaultValues, customValues)
	if err != nil {
		t.Fatalf("MergeValues error: %v", err)
	}

	// 解析结果
	var resultMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &resultMap); err != nil {
		t.Fatalf("unmarshal result failed: %v", err)
	}

	// 检查 pilot 配置
	pilotConfig, ok := resultMap["pilot"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("pilot config not found in result")
	}

	// 验证 autoscaleEnabled 字段应该保留
	if autoscaleEnabled, exists := pilotConfig["autoscaleEnabled"]; !exists {
		t.Error("autoscaleEnabled should be preserved but was deleted")
	} else if !autoscaleEnabled.(bool) {
		t.Error("autoscaleEnabled should be true")
	}

	// 验证 autoscaleMin 字段应该保留
	if autoscaleMin, exists := pilotConfig["autoscaleMin"]; !exists {
		t.Error("autoscaleMin should be preserved but was deleted")
	} else if autoscaleMin.(int) != 2 {
		t.Errorf("autoscaleMin should be 2, got %v", autoscaleMin)
	}

	// 验证 autoscaleMax 字段应该保留
	if autoscaleMax, exists := pilotConfig["autoscaleMax"]; !exists {
		t.Error("autoscaleMax should be preserved but was deleted")
	} else if autoscaleMax.(int) != 5 {
		t.Errorf("autoscaleMax should be 5, got %v", autoscaleMax)
	}

	// 验证 cpu 字段应该保留
	if cpu, exists := pilotConfig["cpu"]; !exists {
		t.Error("cpu should be preserved but was deleted")
	} else {
		cpuMap, ok := cpu.(map[interface{}]interface{})
		if !ok {
			t.Error("cpu should be a map")
		} else if targetUtil, exists := cpuMap["targetAverageUtilization"]; !exists {
			t.Error("targetAverageUtilization should be preserved")
		} else if targetUtil.(int) != 70 {
			t.Errorf("targetAverageUtilization should be 70, got %v", targetUtil)
		}
	}

	t.Logf("Test passed! Result: %s", result)
}

func TestMapOperations(t *testing.T) {
	t.Run("TestGetMapValue", func(t *testing.T) {
		// 测试 map[string]interface{} 类型
		stringMap := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		if val, exists := getMapValue(stringMap, "key1"); !exists || val != "value1" {
			t.Error("getMapValue failed for string key")
		}

		if val, exists := getMapValue(stringMap, "key2"); !exists || val != 123 {
			t.Error("getMapValue failed for numeric value")
		}

		if _, exists := getMapValue(stringMap, "nonexistent"); exists {
			t.Error("getMapValue should return false for nonexistent key")
		}

		// 测试 map[interface{}]interface{} 类型
		interfaceMap := map[interface{}]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		if val, exists := getMapValue(interfaceMap, "key1"); !exists || val != "value1" {
			t.Error("getMapValue failed for interface{} map")
		}
	})

	t.Run("TestDeleteMapKey", func(t *testing.T) {
		// 测试 map[string]interface{} 类型
		stringMap := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}

		deleteMapKey(stringMap, "key1")
		if _, exists := stringMap["key1"]; exists {
			t.Error("deleteMapKey failed for string map")
		}
		if _, exists := stringMap["key2"]; !exists {
			t.Error("deleteMapKey should not delete other keys")
		}

		// 测试 map[interface{}]interface{} 类型
		interfaceMap := map[interface{}]interface{}{
			"key1": "value1",
			"key2": "value2",
		}

		deleteMapKey(interfaceMap, "key1")
		if _, exists := getMapValue(interfaceMap, "key1"); exists {
			t.Error("deleteMapKey failed for interface{} map")
		}
	})

	t.Run("TestEnsureMapKeyExists", func(t *testing.T) {
		// 测试 map[string]interface{} 类型
		stringMap := map[string]interface{}{
			"key1": "value1",
		}

		ensureMapKeyExists(stringMap, "key2")
		if _, exists := stringMap["key2"]; !exists {
			t.Error("ensureMapKeyExists failed for string map")
		}

		// 测试 map[interface{}]interface{} 类型
		interfaceMap := map[interface{}]interface{}{
			"key1": "value1",
		}

		ensureMapKeyExists(interfaceMap, "key2")
		if _, exists := getMapValue(interfaceMap, "key2"); !exists {
			t.Error("ensureMapKeyExists failed for interface{} map")
		}
	})
}
