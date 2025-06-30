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

package entity

import (
	"slices"
	"time"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// MeshIstio represents a service mesh istio entity in database
type MeshIstio struct {
	// Basic information
	MeshID        string `bson:"meshID" json:"meshID" validate:"required"`
	Name          string `bson:"name" json:"name" validate:"required"`
	NetworkID     string `bson:"networkID" json:"networkID" validate:"required"`
	ProjectID     string `bson:"projectID" json:"projectID" validate:"required"`
	ProjectCode   string `bson:"projectCode" json:"projectCode" validate:"required"`
	Description   string `bson:"description" json:"description"`
	Version       string `bson:"version" json:"version" validate:"required"`
	ChartVersion  string `bson:"chartVersion" json:"chartVersion" validate:"required"`
	Status        string `bson:"status" json:"status" validate:"required"`
	StatusMessage string `bson:"statusMessage" json:"statusMessage"`
	CreateTime    int64  `bson:"createTime" json:"createTime"`
	UpdateTime    int64  `bson:"updateTime" json:"updateTime"`
	CreateBy      string `bson:"createBy" json:"createBy"`
	UpdateBy      string `bson:"updateBy" json:"updateBy"`
	IsDeleted     bool   `bson:"isDeleted" json:"isDeleted"`

	// Mesh configuration
	ControlPlaneMode string   `bson:"controlPlaneMode" json:"controlPlaneMode"`
	ClusterMode      string   `bson:"clusterMode" json:"clusterMode"`
	PrimaryClusters  []string `bson:"primaryClusters" json:"primaryClusters"`
	RemoteClusters   []string `bson:"remoteClusters" json:"remoteClusters"`
	DifferentNetwork bool     `bson:"differentNetwork" json:"differentNetwork"`

	// Feature configurations
	FeatureConfigs map[string]*FeatureConfig `bson:"featureConfigs" json:"featureConfigs"`

	// Resource and observability configurations
	SidecarResourceConfig *ResourceConfig      `bson:"sidecarResourceConfig" json:"sidecarResourceConfig"`
	HighAvailability      *HighAvailability    `bson:"highAvailability" json:"highAvailability"`
	ObservabilityConfig   *ObservabilityConfig `bson:"observabilityConfig" json:"observabilityConfig"`
}

// ResourceConfig represents resource configuration for sidecar
type ResourceConfig struct {
	CpuRequest    string `bson:"cpuRequest" json:"cpuRequest"`
	CpuLimit      string `bson:"cpuLimit" json:"cpuLimit"`
	MemoryRequest string `bson:"memoryRequest" json:"memoryRequest"`
	MemoryLimit   string `bson:"memoryLimit" json:"memoryLimit"`
}

// DedicatedNode represents dedicated node configuration
type DedicatedNode struct {
	Enabled    bool              `bson:"enabled" json:"enabled"`
	NodeLabels map[string]string `bson:"nodeLabels" json:"nodeLabels"`
}

// HighAvailability represents high availability configuration
type HighAvailability struct {
	AutoscaleEnabled                   bool            `bson:"autoscaleEnabled" json:"autoscaleEnabled"`
	AutoscaleMin                       int32           `bson:"autoscaleMin" json:"autoscaleMin"`
	AutoscaleMax                       int32           `bson:"autoscaleMax" json:"autoscaleMax"`
	ReplicaCount                       int32           `bson:"replicaCount" json:"replicaCount"`
	TargetCPUAverageUtilizationPercent int32           `bson:"targetCPUAverageUtilizationPercent" json:"targetCPUAverageUtilizationPercent"` // nolint:lll
	ResourceConfig                     *ResourceConfig `bson:"resourceConfig" json:"resourceConfig"`
	DedicatedNode                      *DedicatedNode  `bson:"dedicatedNode" json:"dedicatedNode"`
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	MetricsConfig      *MetricsConfig      `bson:"metricsConfig" json:"metricsConfig"`
	LogCollectorConfig *LogCollectorConfig `bson:"logCollectorConfig" json:"logCollectorConfig"`
	TracingConfig      *TracingConfig      `bson:"tracingConfig" json:"tracingConfig"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	ControlPlaneMetricsEnabled bool `bson:"controlPlaneMetricsEnabled" json:"controlPlaneMetricsEnabled"`
	DataPlaneMetricsEnabled    bool `bson:"dataPlaneMetricsEnabled" json:"dataPlaneMetricsEnabled"`
}

// LogCollectorConfig represents log collector configuration
type LogCollectorConfig struct {
	Enabled           bool   `bson:"enabled" json:"enabled"`
	AccessLogEncoding string `bson:"accessLogEncoding" json:"accessLogEncoding"`
	AccessLogFormat   string `bson:"accessLogFormat" json:"accessLogFormat"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled              bool   `bson:"enabled" json:"enabled"`
	Endpoint             string `bson:"endpoint" json:"endpoint"`
	BkToken              string `bson:"bkToken" json:"bkToken"`
	TraceSamplingPercent int32  `bson:"traceSamplingPercent" json:"traceSamplingPercent"`
}

// FeatureConfig represents a feature configuration
type FeatureConfig struct {
	// Feature name
	Name string `bson:"name" json:"name"`
	// Feature description
	Description string `bson:"description" json:"description"`
	// Feature value
	Value string `bson:"value" json:"value"`
	// Default value of the feature
	DefaultValue string `bson:"defaultValue" json:"defaultValue"`
	// Available values for the feature
	AvailableValues []string `bson:"availableValues" json:"availableValues"`
	// Supported versions for the feature
	SupportVersions []string `bson:"supportVersions" json:"supportVersions"`
}

// Transfer2ProtoForDetail converts MeshIstio entity to proto message
// nolint:funlen
func (m *MeshIstio) Transfer2ProtoForDetail() *meshmanager.IstioDetailInfo {
	// TODO: 考虑直接序列化转换数据，避免逐个赋值
	istioDetailInfo := &meshmanager.IstioDetailInfo{
		MeshID:           m.MeshID,
		Name:             m.Name,
		ProjectID:        m.ProjectID,
		ProjectCode:      m.ProjectCode,
		NetworkID:        m.NetworkID,
		Description:      m.Description,
		ChartVersion:     m.ChartVersion,
		Version:          m.Version,
		Status:           m.Status,
		StatusMessage:    m.StatusMessage,
		CreateTime:       m.CreateTime,
		UpdateTime:       m.UpdateTime,
		CreateBy:         m.CreateBy,
		UpdateBy:         m.UpdateBy,
		ControlPlaneMode: m.ControlPlaneMode,
		ClusterMode:      m.ClusterMode,
		PrimaryClusters:  m.PrimaryClusters,
		RemoteClusters:   m.RemoteClusters,
		DifferentNetwork: m.DifferentNetwork,
	}

	// 转换 Sidecar 资源配置
	if m.SidecarResourceConfig != nil {
		istioDetailInfo.SidecarResourceConfig = &meshmanager.ResourceConfig{
			CpuRequest:    wrapperspb.String(m.SidecarResourceConfig.CpuRequest),
			CpuLimit:      wrapperspb.String(m.SidecarResourceConfig.CpuLimit),
			MemoryRequest: wrapperspb.String(m.SidecarResourceConfig.MemoryRequest),
			MemoryLimit:   wrapperspb.String(m.SidecarResourceConfig.MemoryLimit),
		}
	}

	// 转换高可用配置
	if m.HighAvailability != nil {
		istioDetailInfo.HighAvailability = &meshmanager.HighAvailability{
			AutoscaleEnabled:                   wrapperspb.Bool(m.HighAvailability.AutoscaleEnabled),
			AutoscaleMin:                       wrapperspb.Int32(m.HighAvailability.AutoscaleMin),
			AutoscaleMax:                       wrapperspb.Int32(m.HighAvailability.AutoscaleMax),
			ReplicaCount:                       wrapperspb.Int32(m.HighAvailability.ReplicaCount),
			TargetCPUAverageUtilizationPercent: wrapperspb.Int32(m.HighAvailability.TargetCPUAverageUtilizationPercent),
		}

		if m.HighAvailability.ResourceConfig != nil {
			istioDetailInfo.HighAvailability.ResourceConfig = &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String(m.HighAvailability.ResourceConfig.CpuRequest),
				CpuLimit:      wrapperspb.String(m.HighAvailability.ResourceConfig.CpuLimit),
				MemoryRequest: wrapperspb.String(m.HighAvailability.ResourceConfig.MemoryRequest),
				MemoryLimit:   wrapperspb.String(m.HighAvailability.ResourceConfig.MemoryLimit),
			}
		}

		if m.HighAvailability.DedicatedNode != nil {
			istioDetailInfo.HighAvailability.DedicatedNode = &meshmanager.DedicatedNode{
				Enabled:    wrapperspb.Bool(m.HighAvailability.DedicatedNode.Enabled),
				NodeLabels: m.HighAvailability.DedicatedNode.NodeLabels,
			}
		}
	}

	// 转换可观测性配置
	if m.ObservabilityConfig != nil {
		istioDetailInfo.ObservabilityConfig = &meshmanager.ObservabilityConfig{}

		// 转换指标配置
		if m.ObservabilityConfig.MetricsConfig != nil {
			istioDetailInfo.ObservabilityConfig.MetricsConfig = &meshmanager.MetricsConfig{
				ControlPlaneMetricsEnabled: wrapperspb.Bool(m.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled),
				DataPlaneMetricsEnabled:    wrapperspb.Bool(m.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled),
			}
		}

		// 转换日志收集配置
		if m.ObservabilityConfig.LogCollectorConfig != nil {
			istioDetailInfo.ObservabilityConfig.LogCollectorConfig = &meshmanager.LogCollectorConfig{
				Enabled:           wrapperspb.Bool(m.ObservabilityConfig.LogCollectorConfig.Enabled),
				AccessLogEncoding: wrapperspb.String(m.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding),
				AccessLogFormat:   wrapperspb.String(m.ObservabilityConfig.LogCollectorConfig.AccessLogFormat),
			}
		}

		// 转换链路追踪配置
		if m.ObservabilityConfig.TracingConfig != nil {
			istioDetailInfo.ObservabilityConfig.TracingConfig = &meshmanager.TracingConfig{
				Enabled:              wrapperspb.Bool(m.ObservabilityConfig.TracingConfig.Enabled),
				Endpoint:             wrapperspb.String(m.ObservabilityConfig.TracingConfig.Endpoint),
				BkToken:              wrapperspb.String(m.ObservabilityConfig.TracingConfig.BkToken),
				TraceSamplingPercent: wrapperspb.Int32(m.ObservabilityConfig.TracingConfig.TraceSamplingPercent),
			}
		}
	}

	// 转换特性配置
	if len(m.FeatureConfigs) > 0 {
		istioDetailInfo.FeatureConfigs = make(map[string]*meshmanager.FeatureConfig)
		for name, config := range m.FeatureConfigs {
			// 只转换支持的特性
			if !slices.Contains(common.SupportedFeatures, name) {
				continue
			}
			istioDetailInfo.FeatureConfigs[name] = &meshmanager.FeatureConfig{
				Name:            config.Name,
				Description:     config.Description,
				Value:           config.Value,
				DefaultValue:    config.DefaultValue,
				AvailableValues: config.AvailableValues,
				SupportVersions: config.SupportVersions,
			}
		}
	}

	return istioDetailInfo
}

// Transfer2ProtoForListItems converts MeshIstio entity to proto message
func (m *MeshIstio) Transfer2ProtoForListItems() *meshmanager.IstioListItem {
	istioListItem := &meshmanager.IstioListItem{
		MeshID:          m.MeshID,
		Name:            m.Name,
		ProjectID:       m.ProjectID,
		ProjectCode:     m.ProjectCode,
		Version:         m.Version,
		Status:          m.Status,
		StatusMessage:   m.StatusMessage,
		CreateTime:      m.CreateTime,
		ChartVersion:    m.ChartVersion,
		PrimaryClusters: m.PrimaryClusters,
		RemoteClusters:  m.RemoteClusters,
	}
	return istioListItem
}

// TransferFromProto converts InstallIstioRequest to MeshIstio entity
// nolint:funlen
func (m *MeshIstio) TransferFromProto(req *meshmanager.IstioRequest) {
	// 转换基本字段
	m.Name = req.Name.GetValue()
	m.ProjectID = req.ProjectID.GetValue()
	m.ProjectCode = req.ProjectCode.GetValue()
	m.Description = req.Description.GetValue()
	m.Version = req.Version.GetValue()
	m.ControlPlaneMode = req.ControlPlaneMode.GetValue()
	m.ClusterMode = req.ClusterMode.GetValue()
	m.PrimaryClusters = req.PrimaryClusters
	m.RemoteClusters = req.RemoteClusters
	m.DifferentNetwork = req.DifferentNetwork.GetValue()
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	m.CreateBy = "system" // TODO: get from context
	m.UpdateBy = "system" // TODO: get from context

	// 转换 Sidecar 资源配置
	if req.SidecarResourceConfig != nil {
		m.SidecarResourceConfig = &ResourceConfig{
			CpuRequest:    req.SidecarResourceConfig.CpuRequest.GetValue(),
			CpuLimit:      req.SidecarResourceConfig.CpuLimit.GetValue(),
			MemoryRequest: req.SidecarResourceConfig.MemoryRequest.GetValue(),
			MemoryLimit:   req.SidecarResourceConfig.MemoryLimit.GetValue(),
		}
	}

	// 转换高可用配置
	if req.HighAvailability != nil {
		m.HighAvailability = &HighAvailability{
			AutoscaleEnabled:                   req.HighAvailability.AutoscaleEnabled.GetValue(),
			AutoscaleMin:                       req.HighAvailability.AutoscaleMin.GetValue(),
			AutoscaleMax:                       req.HighAvailability.AutoscaleMax.GetValue(),
			ReplicaCount:                       req.HighAvailability.ReplicaCount.GetValue(),
			TargetCPUAverageUtilizationPercent: req.HighAvailability.TargetCPUAverageUtilizationPercent.GetValue(),
		}

		if req.HighAvailability.ResourceConfig != nil {
			m.HighAvailability.ResourceConfig = &ResourceConfig{
				CpuRequest:    req.HighAvailability.ResourceConfig.CpuRequest.GetValue(),
				CpuLimit:      req.HighAvailability.ResourceConfig.CpuLimit.GetValue(),
				MemoryRequest: req.HighAvailability.ResourceConfig.MemoryRequest.GetValue(),
				MemoryLimit:   req.HighAvailability.ResourceConfig.MemoryLimit.GetValue(),
			}
		}

		if req.HighAvailability.DedicatedNode != nil {
			m.HighAvailability.DedicatedNode = &DedicatedNode{
				Enabled:    req.HighAvailability.DedicatedNode.Enabled.GetValue(),
				NodeLabels: req.HighAvailability.DedicatedNode.NodeLabels,
			}

		}
	}

	// 可观测性配置
	if req.ObservabilityConfig != nil {
		m.ObservabilityConfig = &ObservabilityConfig{}
		if req.ObservabilityConfig.LogCollectorConfig != nil {
			m.ObservabilityConfig.LogCollectorConfig = &LogCollectorConfig{
				Enabled:           req.ObservabilityConfig.LogCollectorConfig.Enabled.GetValue(),
				AccessLogEncoding: req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue(),
				AccessLogFormat:   req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue(),
			}
		}
		if req.ObservabilityConfig.TracingConfig != nil {
			m.ObservabilityConfig.TracingConfig = &TracingConfig{
				Enabled:              req.ObservabilityConfig.TracingConfig.Enabled.GetValue(),
				Endpoint:             req.ObservabilityConfig.TracingConfig.Endpoint.GetValue(),
				BkToken:              req.ObservabilityConfig.TracingConfig.BkToken.GetValue(),
				TraceSamplingPercent: req.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue(),
			}
		}
		if req.ObservabilityConfig.MetricsConfig != nil {
			m.ObservabilityConfig.MetricsConfig = &MetricsConfig{
				ControlPlaneMetricsEnabled: req.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled.GetValue(),
				DataPlaneMetricsEnabled:    req.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled.GetValue(),
			}
		}
	}

	// 转换特性配置
	if len(req.FeatureConfigs) > 0 {
		m.FeatureConfigs = make(map[string]*FeatureConfig)
		for name, config := range req.FeatureConfigs {
			// 只保存支持的特性
			if !slices.Contains(common.SupportedFeatures, name) {
				continue
			}
			m.FeatureConfigs[name] = &FeatureConfig{
				Name:            config.Name,
				Description:     config.Description,
				Value:           config.Value,
				DefaultValue:    config.DefaultValue,
				AvailableValues: config.AvailableValues,
				SupportVersions: config.SupportVersions,
			}
		}
	}
}
