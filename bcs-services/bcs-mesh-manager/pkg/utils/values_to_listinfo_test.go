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
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

func TestConvertValuesToListItem(t *testing.T) {
	tests := []struct {
		name         string
		meshIstio    *entity.MeshIstio
		istiodValues *common.IstiodInstallValues
		wantErr      bool
		validate     func(t *testing.T, result *meshmanager.IstioDetailInfo)
	}{
		{
			name:         "nil meshIstio",
			meshIstio:    nil,
			istiodValues: &common.IstiodInstallValues{},
			wantErr:      true,
		},
		{
			name:         "nil istiodValues",
			meshIstio:    &entity.MeshIstio{},
			istiodValues: nil,
			wantErr:      true,
		},
		{
			name: "basic conversion",
			meshIstio: &entity.MeshIstio{
				MeshID:           "test-mesh",
				Name:             "test-istio",
				ProjectCode:      "test-code",
				NetworkID:        "test-network",
				Description:      "test description",
				Version:          "1.24.0",
				ChartVersion:     "1.24.0",
				Status:           "RUNNING",
				ControlPlaneMode: "PRIMARY",
				ClusterMode:      "SINGLE",
				PrimaryClusters:  []string{"cluster1"},
			},
			istiodValues: &common.IstiodInstallValues{},
			wantErr:      false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				if result.MeshID != "test-mesh" {
					t.Errorf("expected MeshID 'test-mesh', got '%s'", result.MeshID)
				}
				if result.Name != "test-istio" {
					t.Errorf("expected Name 'test-istio', got '%s'", result.Name)
				}
				if result.Version != "1.24.0" {
					t.Errorf("expected Version '1.24.0', got '%s'", result.Version)
				}
			},
		},
		{
			name: "sidecar resource configuration",
			meshIstio: &entity.MeshIstio{
				MeshID:       "test-mesh",
				Name:         "test-istio",
				NetworkID:    "test-network",
				Version:      "1.24.0",
				ChartVersion: "1.24.0",
			},
			istiodValues: &common.IstiodInstallValues{
				Global: &common.IstiodGlobalConfig{
					Proxy: &common.IstioProxyConfig{
						Resources: &common.ResourceConfig{
							Requests: &common.ResourceRequests{
								CPU:    strPtr("50m"),
								Memory: strPtr("64Mi"),
							},
							Limits: &common.ResourceLimits{
								CPU:    strPtr("100m"),
								Memory: strPtr("128Mi"),
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				if result.SidecarResourceConfig == nil {
					t.Fatal("expected SidecarResourceConfig to be set")
				}
				if result.SidecarResourceConfig.CpuRequest.GetValue() != "50m" {
					t.Errorf("expected CPU request '50m', got '%s'", result.SidecarResourceConfig.CpuRequest.GetValue())
				}
				if result.SidecarResourceConfig.MemoryLimit.GetValue() != "128Mi" {
					t.Errorf("expected Memory limit '128Mi', got '%s'", result.SidecarResourceConfig.MemoryLimit.GetValue())
				}
			},
		},
		{
			name: "high availability configuration",
			meshIstio: &entity.MeshIstio{
				MeshID:       "test-mesh",
				Name:         "test-istio",
				NetworkID:    "test-network",
				Version:      "1.24.0",
				ChartVersion: "1.24.0",
			},
			istiodValues: &common.IstiodInstallValues{
				Pilot: &common.IstiodPilotConfig{
					ReplicaCount:     int32Ptr(3),
					AutoscaleEnabled: boolPtr(true),
					AutoscaleMin:     int32Ptr(2),
					AutoscaleMax:     int32Ptr(5),
					CPU: &common.HPACPUConfig{
						TargetAverageUtilization: int32Ptr(80),
					},
					Resources: &common.ResourceConfig{
						Requests: &common.ResourceRequests{
							CPU:    strPtr("200m"),
							Memory: strPtr("256Mi"),
						},
						Limits: &common.ResourceLimits{
							CPU:    strPtr("500m"),
							Memory: strPtr("512Mi"),
						},
					},
					NodeSelector: map[string]string{
						"node-type": "istio",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				if result.HighAvailability == nil {
					t.Fatal("expected HighAvailability to be set")
				}
				if result.HighAvailability.ReplicaCount.GetValue() != 3 {
					t.Errorf("expected ReplicaCount 3, got %d", result.HighAvailability.ReplicaCount.GetValue())
				}
				if !result.HighAvailability.AutoscaleEnabled.GetValue() {
					t.Error("expected AutoscaleEnabled to be true")
				}
				if result.HighAvailability.AutoscaleMin.GetValue() != 2 {
					t.Errorf("expected AutoscaleMin 2, got %d", result.HighAvailability.AutoscaleMin.GetValue())
				}
				if result.HighAvailability.TargetCPUAverageUtilizationPercent.GetValue() != 80 {
					t.Errorf("expected TargetCPUAverageUtilizationPercent 80, got %d",
						result.HighAvailability.TargetCPUAverageUtilizationPercent.GetValue())
				}
				if result.HighAvailability.ResourceConfig == nil {
					t.Fatal("expected ResourceConfig to be set")
				}
				if result.HighAvailability.ResourceConfig.CpuRequest.GetValue() != "200m" {
					t.Errorf("expected CPU request '200m', got '%s'",
						result.HighAvailability.ResourceConfig.CpuRequest.GetValue())
				}
				if result.HighAvailability.DedicatedNode == nil {
					t.Fatal("expected DedicatedNode to be set")
				}
				if !result.HighAvailability.DedicatedNode.Enabled.GetValue() {
					t.Error("expected DedicatedNode.Enabled to be true")
				}
				if result.HighAvailability.DedicatedNode.NodeLabels["node-type"] != "istio" {
					t.Errorf("expected node label 'node-type=istio', got '%s'",
						result.HighAvailability.DedicatedNode.NodeLabels["node-type"])
				}
			},
		},
		{
			name: "observability configuration - tracing enabled (>= 1.21)",
			meshIstio: &entity.MeshIstio{
				MeshID:       "test-mesh",
				Name:         "test-istio",
				NetworkID:    "test-network",
				Version:      "1.24.0",
				ChartVersion: "1.24.0",
			},
			istiodValues: &common.IstiodInstallValues{
				MeshConfig: &common.IstiodMeshConfig{
					EnableTracing: boolPtr(true),
					ExtensionProviders: []*common.ExtensionProvider{
						{
							Name: strPtr(OtelTracingName),
							OpenTelemetry: &common.OpenTelemetryConfig{
								Service: strPtr("jaeger-collector"),
								Port:    int32Ptr(14268),
								Http: &common.OpenTelemetryHttpConfig{
									Path: strPtr("/api/traces"),
									Headers: map[string]string{
										OtelTracingHeader: "test-token",
									},
								},
							},
						},
					},
					AccessLogFile:     strPtr("/dev/stdout"),
					AccessLogFormat:   strPtr("json"),
					AccessLogEncoding: strPtr("TEXT"),
				},
				Pilot: &common.IstiodPilotConfig{
					TraceSampling: float64Ptr(0.1),
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				if result.ObservabilityConfig == nil {
					t.Fatal("expected ObservabilityConfig to be set")
				}
				if result.ObservabilityConfig.TracingConfig == nil {
					t.Fatal("expected TracingConfig to be set")
				}
				if !result.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
					t.Error("expected TracingConfig.Enabled to be true")
				}
				expectedEndpoint := "jaeger-collector:14268/api/traces"
				if result.ObservabilityConfig.TracingConfig.Endpoint.GetValue() != expectedEndpoint {
					t.Errorf("expected endpoint '%s', got '%s'", expectedEndpoint,
						result.ObservabilityConfig.TracingConfig.Endpoint.GetValue())
				}
				if result.ObservabilityConfig.TracingConfig.BkToken.GetValue() != "test-token" {
					t.Errorf("expected token 'test-token', got '%s'",
						result.ObservabilityConfig.TracingConfig.BkToken.GetValue())
				}
				if result.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() != 10 {
					t.Errorf("expected TraceSamplingPercent 10, got %d",
						result.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue())
				}
				if result.ObservabilityConfig.LogCollectorConfig == nil {
					t.Fatal("expected LogCollectorConfig to be set")
				}
				if !result.ObservabilityConfig.LogCollectorConfig.Enabled.GetValue() {
					t.Error("expected LogCollectorConfig.Enabled to be true")
				}
				if result.ObservabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue() != "json" {
					t.Errorf("expected AccessLogFormat 'json', got '%s'",
						result.ObservabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue())
				}
			},
		},
		{
			name: "observability configuration - tracing with zipkin (< 1.21)",
			meshIstio: &entity.MeshIstio{
				MeshID:       "test-mesh",
				Name:         "test-istio",
				NetworkID:    "test-network",
				Version:      "1.20.0",
				ChartVersion: "1.20.0",
			},
			istiodValues: &common.IstiodInstallValues{
				MeshConfig: &common.IstiodMeshConfig{
					EnableTracing: boolPtr(true),
					DefaultConfig: &common.DefaultConfig{
						TracingConfig: &common.TracingConfig{
							Zipkin: &common.ZipkinConfig{
								Address: strPtr("zipkin.istio-system:9411/api/v2/spans"),
							},
						},
					},
				},
				Pilot: &common.IstiodPilotConfig{
					TraceSampling: float64Ptr(0.05),
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				if result.ObservabilityConfig == nil {
					t.Fatal("expected ObservabilityConfig to be set")
				}
				if result.ObservabilityConfig.TracingConfig == nil {
					t.Fatal("expected TracingConfig to be set")
				}
				if !result.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
					t.Error("expected TracingConfig.Enabled to be true")
				}
				expectedEndpoint := "zipkin.istio-system:9411/api/v2/spans"
				if result.ObservabilityConfig.TracingConfig.Endpoint.GetValue() != expectedEndpoint {
					t.Errorf("expected endpoint '%s', got '%s'", expectedEndpoint,
						result.ObservabilityConfig.TracingConfig.Endpoint.GetValue())
				}
				if result.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() != 5 {
					t.Errorf("expected TraceSamplingPercent 5, got %d",
						result.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue())
				}
			},
		},
		{
			name: "feature configurations",
			meshIstio: &entity.MeshIstio{
				MeshID:       "test-mesh",
				Name:         "test-istio",
				NetworkID:    "test-network",
				Version:      "1.24.0",
				ChartVersion: "1.24.0",
			},
			istiodValues: &common.IstiodInstallValues{
				MeshConfig: &common.IstiodMeshConfig{
					OutboundTrafficPolicy: &common.OutboundTrafficPolicy{
						Mode: strPtr("REGISTRY_ONLY"),
					},
					DefaultConfig: &common.DefaultConfig{
						HoldApplicationUntilProxyStarts: boolPtr(true),
						ProxyMetadata: &common.ProxyMetadata{
							ExitOnZeroActiveConnections: strPtr("true"),
							IstioMetaDnsCapture:         strPtr("true"),
							IstioMetaDnsAutoAllocate:    strPtr("true"),
						},
					},
				},
				Pilot: &common.IstiodPilotConfig{
					Env: map[string]string{
						common.EnvPilotHTTP10: "true",
					},
				},
				Global: &common.IstiodGlobalConfig{
					Proxy: &common.IstioProxyConfig{
						ExcludeIPRanges: strPtr("10.0.0.0/8,172.16.0.0/12"),
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				if result.FeatureConfigs == nil {
					t.Fatal("expected FeatureConfigs to be set")
				}

				// Check outbound traffic policy
				if config, ok := result.FeatureConfigs[common.FeatureOutboundTrafficPolicy]; ok {
					if config.Value != "REGISTRY_ONLY" {
						t.Errorf("expected outbound traffic policy 'REGISTRY_ONLY', got '%s'", config.Value)
					}
				} else {
					t.Error("expected outbound traffic policy feature config")
				}

				// Check hold application until proxy starts
				if config, ok := result.FeatureConfigs[common.FeatureHoldApplicationUntilProxyStarts]; ok {
					if config.Value != "true" {
						t.Errorf("expected hold application 'true', got '%s'", config.Value)
					}
				} else {
					t.Error("expected hold application feature config")
				}

				// Check exit on zero active connections
				if config, ok := result.FeatureConfigs[common.FeatureExitOnZeroActiveConnections]; ok {
					if config.Value != "true" {
						t.Errorf("expected exit on zero connections 'true', got '%s'", config.Value)
					}
				} else {
					t.Error("expected exit on zero connections feature config")
				}

				// Check DNS capture
				if config, ok := result.FeatureConfigs[common.FeatureIstioMetaDnsCapture]; ok {
					if config.Value != "true" {
						t.Errorf("expected DNS capture 'true', got '%s'", config.Value)
					}
				} else {
					t.Error("expected DNS capture feature config")
				}

				// Check HTTP 1.0 support
				if config, ok := result.FeatureConfigs[common.FeatureIstioMetaHttp10]; ok {
					if config.Value != "true" {
						t.Errorf("expected HTTP 1.0 'true', got '%s'", config.Value)
					}
				} else {
					t.Error("expected HTTP 1.0 feature config")
				}

				// Check exclude IP ranges
				if config, ok := result.FeatureConfigs[common.FeatureExcludeIPRanges]; ok {
					if config.Value != "10.0.0.0/8,172.16.0.0/12" {
						t.Errorf("expected exclude IP ranges '10.0.0.0/8,172.16.0.0/12', got '%s'", config.Value)
					}
				} else {
					t.Error("expected exclude IP ranges feature config")
				}
			},
		},
		{
			name: "complete configuration",
			meshIstio: &entity.MeshIstio{
				MeshID:           "complete-mesh",
				Name:             "complete-istio",
				ProjectCode:      "complete-code",
				NetworkID:        "complete-network",
				Description:      "complete test",
				Version:          "1.24.0",
				ChartVersion:     "1.24.0",
				Status:           "RUNNING",
				StatusMessage:    "All components are healthy",
				ControlPlaneMode: "PRIMARY",
				ClusterMode:      "MULTI",
				PrimaryClusters:  []string{"cluster1", "cluster2"},
				RemoteClusters:   []*entity.RemoteCluster{{ClusterID: "cluster3"}},
				DifferentNetwork: true,
				CreateTime:       1640995200,
				UpdateTime:       1640995300,
				CreateBy:         "admin",
				UpdateBy:         "admin",
			},
			istiodValues: &common.IstiodInstallValues{
				Global: &common.IstiodGlobalConfig{
					MeshID:             strPtr("complete-mesh"),
					Network:            strPtr("complete-network"),
					ConfigCluster:      boolPtr(true),
					ExternalIstiod:     boolPtr(true),
					RemotePilotAddress: strPtr("pilot.istio-system.svc.cluster.local"),
					Proxy: &common.IstioProxyConfig{
						Resources: &common.ResourceConfig{
							Requests: &common.ResourceRequests{
								CPU:    strPtr("100m"),
								Memory: strPtr("128Mi"),
							},
							Limits: &common.ResourceLimits{
								CPU:    strPtr("200m"),
								Memory: strPtr("256Mi"),
							},
						},
						ExcludeIPRanges: strPtr("192.168.0.0/16"),
					},
					MultiCluster: &common.IstiodMultiClusterConfig{
						ClusterName: strPtr("complete-cluster"),
					},
				},
				Pilot: &common.IstiodPilotConfig{
					ReplicaCount:     int32Ptr(3),
					AutoscaleEnabled: boolPtr(true),
					AutoscaleMin:     int32Ptr(2),
					AutoscaleMax:     int32Ptr(10),
					TraceSampling:    float64Ptr(0.01),
					ConfigMap:        boolPtr(false),
					CPU: &common.HPACPUConfig{
						TargetAverageUtilization: int32Ptr(70),
					},
					Resources: &common.ResourceConfig{
						Requests: &common.ResourceRequests{
							CPU:    strPtr("500m"),
							Memory: strPtr("1Gi"),
						},
						Limits: &common.ResourceLimits{
							CPU:    strPtr("1000m"),
							Memory: strPtr("2Gi"),
						},
					},
					NodeSelector: map[string]string{
						"node-type":   "control-plane",
						"istio-ready": "true",
					},
					Env: map[string]string{
						common.EnvPilotHTTP10: "false",
						"CUSTOM_ENV":          "custom-value",
					},
				},
				MeshConfig: &common.IstiodMeshConfig{
					EnableTracing: boolPtr(true),
					ExtensionProviders: []*common.ExtensionProvider{
						{
							Name: strPtr(OtelTracingName),
							OpenTelemetry: &common.OpenTelemetryConfig{
								Service: strPtr("otel-collector"),
								Port:    int32Ptr(4318),
								Http: &common.OpenTelemetryHttpConfig{
									Path: strPtr("/v1/traces"),
									Headers: map[string]string{
										OtelTracingHeader: "complete-token",
									},
								},
							},
						},
					},
					OutboundTrafficPolicy: &common.OutboundTrafficPolicy{
						Mode: strPtr("ALLOW_ANY"),
					},
					DefaultConfig: &common.DefaultConfig{
						HoldApplicationUntilProxyStarts: boolPtr(false),
						ProxyMetadata: &common.ProxyMetadata{
							ExitOnZeroActiveConnections: strPtr("false"),
							IstioMetaDnsCapture:         strPtr("false"),
							IstioMetaDnsAutoAllocate:    strPtr("false"),
						},
					},
					AccessLogFile:     strPtr("/dev/stdout"),
					AccessLogFormat:   strPtr("text"),
					AccessLogEncoding: strPtr("JSON"),
				},
				Telemetry: &common.IstiodTelemetryConfig{
					Enabled: boolPtr(false),
				},
				IstiodRemote: &common.IstiodRemoteConfig{
					Enabled:       boolPtr(true),
					InjectionPath: strPtr("/inject"),
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *meshmanager.IstioDetailInfo) {
				// Validate basic fields
				if result.MeshID != "complete-mesh" {
					t.Errorf("expected MeshID 'complete-mesh', got '%s'", result.MeshID)
				}
				if result.DifferentNetwork != true {
					t.Error("expected DifferentNetwork to be true")
				}
				if len(result.PrimaryClusters) != 2 {
					t.Errorf("expected 2 primary clusters, got %d", len(result.PrimaryClusters))
				}

				// Validate sidecar resource config
				if result.SidecarResourceConfig == nil {
					t.Fatal("expected SidecarResourceConfig to be set")
				}
				if result.SidecarResourceConfig.CpuRequest.GetValue() != "100m" {
					t.Errorf("expected sidecar CPU request '100m', got '%s'",
						result.SidecarResourceConfig.CpuRequest.GetValue())
				}

				// Validate high availability
				if result.HighAvailability == nil {
					t.Fatal("expected HighAvailability to be set")
				}
				if result.HighAvailability.ReplicaCount.GetValue() != 3 {
					t.Errorf("expected ReplicaCount 3, got %d", result.HighAvailability.ReplicaCount.GetValue())
				}
				if result.HighAvailability.AutoscaleMax.GetValue() != 10 {
					t.Errorf("expected AutoscaleMax 10, got %d", result.HighAvailability.AutoscaleMax.GetValue())
				}

				// Validate observability
				if result.ObservabilityConfig == nil {
					t.Fatal("expected ObservabilityConfig to be set")
				}
				if !result.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
					t.Error("expected tracing to be enabled")
				}
				expectedEndpoint := "otel-collector:4318/v1/traces"
				if result.ObservabilityConfig.TracingConfig.Endpoint.GetValue() != expectedEndpoint {
					t.Errorf("expected endpoint '%s', got '%s'", expectedEndpoint,
						result.ObservabilityConfig.TracingConfig.Endpoint.GetValue())
				}
				if result.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() != 1 {
					t.Errorf("expected TraceSamplingPercent 1, got %d",
						result.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue())
				}

				// Validate feature configs
				if result.FeatureConfigs == nil {
					t.Fatal("expected FeatureConfigs to be set")
				}
				if config, ok := result.FeatureConfigs[common.FeatureOutboundTrafficPolicy]; ok {
					if config.Value != "ALLOW_ANY" {
						t.Errorf("expected outbound traffic policy 'ALLOW_ANY', got '%s'", config.Value)
					}
				} else {
					t.Error("expected outbound traffic policy feature config")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValuesToIstioDetailInfo(tt.meshIstio, tt.istiodValues)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
