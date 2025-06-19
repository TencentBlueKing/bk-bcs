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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// MeshIstio represents a service mesh istio entity in database
type MeshIstio struct {
	// Basic information
	MeshID        string `bson:"meshID" json:"meshID" validate:"required"`
	MeshName      string `bson:"meshName" json:"meshName" validate:"required"`
	NetworkID     string `bson:"networkID" json:"networkID" validate:"required"`
	ProjectID     string `bson:"projectID" json:"projectID" validate:"required"`
	ProjectCode   string `bson:"projectCode" json:"projectCode" validate:"required"`
	Description   string `bson:"description" json:"description"`
	ChartVersion  string `bson:"chartVersion" json:"chartVersion" validate:"required"`
	Status        string `bson:"status" json:"status" validate:"required"`
	StatusMessage string `bson:"statusMessage" json:"statusMessage"`
	CreateTime    int64  `bson:"createTime" json:"createTime"`
	UpdateTime    int64  `bson:"updateTime" json:"updateTime"`
	CreateBy      string `bson:"createBy" json:"createBy"`
	UpdateBy      string `bson:"updateBy" json:"updateBy"`

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
	MetricsConfig        *MetricsConfig      `bson:"metricsConfig" json:"metricsConfig"`
	LogCollectorConfig   *LogCollectorConfig `bson:"logCollectorConfig" json:"logCollectorConfig"`
	TracingConfig        *TracingConfig      `bson:"tracingConfig" json:"tracingConfig"`
	TraceSamplingPercent int32               `bson:"traceSamplingPercent" json:"traceSamplingPercent"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled                    bool `bson:"enabled" json:"enabled"`
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
	// Default value of the feature
	DefaultValue string `bson:"defaultValue" json:"defaultValue"`
	// Available values for the feature
	AvailableValues []string `bson:"availableValues" json:"availableValues"`
	// Supported versions for the feature
	SupportVersions []string `bson:"supportVersions" json:"supportVersions"`
}

// Transfer2Proto converts MeshIstio entity to proto message
func (m *MeshIstio) Transfer2Proto() *meshmanager.IstioListItem {
	// 转换基本字段
	proto := m.transferBasicFields()

	// 转换配置相关字段
	proto.SidecarResourceConfig = m.transferSidecarResourceConfig()
	proto.HighAvailability = m.transferHighAvailability()
	proto.ObservabilityConfig = m.transferObservabilityConfig()
	proto.FeatureConfigs = m.transferFeatureConfigs()

	return proto
}

// transferBasicFields 转换基本字段
func (m *MeshIstio) transferBasicFields() *meshmanager.IstioListItem {
	return &meshmanager.IstioListItem{
		MeshID:           m.MeshID,
		MeshName:         m.MeshName,
		ProjectID:        m.ProjectID,
		ProjectCode:      m.ProjectCode,
		Description:      m.Description,
		ChartVersion:     m.ChartVersion,
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
}

// transferFeatureConfigs 转换特性配置
func (m *MeshIstio) transferFeatureConfigs() map[string]*meshmanager.FeatureConfig {
	protoFeatureConfigs := make(map[string]*meshmanager.FeatureConfig)
	for name, config := range m.FeatureConfigs {
		// 只转换支持的特性
		if !slices.Contains(common.SupportedFeatures, name) {
			continue
		}

		protoFeatureConfigs[name] = &meshmanager.FeatureConfig{
			Name:            config.Name,
			Description:     config.Description,
			DefaultValue:    config.DefaultValue,
			AvailableValues: config.AvailableValues,
			SupportVersions: config.SupportVersions,
		}
	}
	return protoFeatureConfigs
}

// transferSidecarResourceConfig 转换 Sidecar 资源配置
func (m *MeshIstio) transferSidecarResourceConfig() *meshmanager.ResourceConfig {
	if m.SidecarResourceConfig == nil {
		return nil
	}
	return &meshmanager.ResourceConfig{
		CpuRequest:    m.SidecarResourceConfig.CpuRequest,
		CpuLimit:      m.SidecarResourceConfig.CpuLimit,
		MemoryRequest: m.SidecarResourceConfig.MemoryRequest,
		MemoryLimit:   m.SidecarResourceConfig.MemoryLimit,
	}
}

// transferHighAvailability 转换高可用配置
func (m *MeshIstio) transferHighAvailability() *meshmanager.HighAvailability {
	if m.HighAvailability == nil {
		return nil
	}

	protoHighAvailability := &meshmanager.HighAvailability{
		AutoscaleEnabled:                   m.HighAvailability.AutoscaleEnabled,
		AutoscaleMin:                       m.HighAvailability.AutoscaleMin,
		AutoscaleMax:                       m.HighAvailability.AutoscaleMax,
		ReplicaCount:                       m.HighAvailability.ReplicaCount,
		TargetCPUAverageUtilizationPercent: m.HighAvailability.TargetCPUAverageUtilizationPercent,
	}

	if m.HighAvailability.ResourceConfig != nil {
		protoHighAvailability.ResourceConfig = &meshmanager.ResourceConfig{
			CpuRequest:    m.HighAvailability.ResourceConfig.CpuRequest,
			CpuLimit:      m.HighAvailability.ResourceConfig.CpuLimit,
			MemoryRequest: m.HighAvailability.ResourceConfig.MemoryRequest,
			MemoryLimit:   m.HighAvailability.ResourceConfig.MemoryLimit,
		}
	}

	if m.HighAvailability.DedicatedNode != nil {
		protoHighAvailability.DedicatedNode = &meshmanager.DedicatedNode{
			Enabled:    m.HighAvailability.DedicatedNode.Enabled,
			NodeLabels: m.HighAvailability.DedicatedNode.NodeLabels,
		}
	}

	return protoHighAvailability
}

func (m *MeshIstio) transferObservabilityConfig() *meshmanager.ObservabilityConfig {
	if m.ObservabilityConfig == nil {
		return nil
	}

	protoObservabilityConfig := &meshmanager.ObservabilityConfig{
		MetricsConfig:      m.transferMetricsConfig(),
		LogCollectorConfig: m.transferLogCollectorConfig(),
		TracingConfig:      m.transferTracingConfig(),
	}

	return protoObservabilityConfig
}

func (m *MeshIstio) transferMetricsConfig() *meshmanager.MetricsConfig {
	if m.ObservabilityConfig == nil || m.ObservabilityConfig.MetricsConfig == nil {
		return nil
	}

	return &meshmanager.MetricsConfig{
		ControlPlaneMetricsEnabled: m.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled,
		DataPlaneMetricsEnabled:    m.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled,
	}
}

// transferLogCollectorConfig 转换日志收集配置
func (m *MeshIstio) transferLogCollectorConfig() *meshmanager.LogCollectorConfig {
	if m.ObservabilityConfig == nil || m.ObservabilityConfig.LogCollectorConfig == nil {
		return nil
	}
	return &meshmanager.LogCollectorConfig{
		Enabled:           m.ObservabilityConfig.LogCollectorConfig.Enabled,
		AccessLogEncoding: m.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding,
		AccessLogFormat:   m.ObservabilityConfig.LogCollectorConfig.AccessLogFormat,
	}
}

// transferTracingConfig 转换链路追踪配置
func (m *MeshIstio) transferTracingConfig() *meshmanager.TracingConfig {
	if m.ObservabilityConfig == nil || m.ObservabilityConfig.TracingConfig == nil {
		return nil
	}
	return &meshmanager.TracingConfig{
		Enabled:              m.ObservabilityConfig.TracingConfig.Enabled,
		Endpoint:             m.ObservabilityConfig.TracingConfig.Endpoint,
		BkToken:              m.ObservabilityConfig.TracingConfig.BkToken,
		TraceSamplingPercent: m.ObservabilityConfig.TraceSamplingPercent,
	}
}

// TransferFromProto converts InstallIstioRequest to MeshIstio entity
// nolint:funlen
func (m *MeshIstio) TransferFromProto(req *meshmanager.InstallIstioRequest) {
	// 转换基本字段
	m.MeshName = req.Name
	m.ProjectID = req.ProjectID
	m.ProjectCode = req.ProjectCode
	m.Description = req.Description
	m.ChartVersion = req.Version
	m.ControlPlaneMode = req.ControlPlaneMode
	m.ClusterMode = req.ClusterMode
	m.PrimaryClusters = req.PrimaryClusters
	m.RemoteClusters = req.RemoteClusters
	m.DifferentNetwork = req.DifferentNetwork
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	m.CreateBy = "system" // TODO: get from context
	m.UpdateBy = "system" // TODO: get from context

	// 转换 Sidecar 资源配置
	if req.SidecarResourceConfig != nil {
		m.SidecarResourceConfig = &ResourceConfig{
			CpuRequest:    req.SidecarResourceConfig.CpuRequest,
			CpuLimit:      req.SidecarResourceConfig.CpuLimit,
			MemoryRequest: req.SidecarResourceConfig.MemoryRequest,
			MemoryLimit:   req.SidecarResourceConfig.MemoryLimit,
		}
	}

	// 转换高可用配置
	if req.HighAvailability != nil {
		m.HighAvailability = &HighAvailability{
			AutoscaleEnabled:                   req.HighAvailability.AutoscaleEnabled,
			AutoscaleMin:                       req.HighAvailability.AutoscaleMin,
			AutoscaleMax:                       req.HighAvailability.AutoscaleMax,
			ReplicaCount:                       req.HighAvailability.ReplicaCount,
			TargetCPUAverageUtilizationPercent: req.HighAvailability.TargetCPUAverageUtilizationPercent,
		}

		if req.HighAvailability.ResourceConfig != nil {
			m.HighAvailability.ResourceConfig = &ResourceConfig{
				CpuRequest:    req.HighAvailability.ResourceConfig.CpuRequest,
				CpuLimit:      req.HighAvailability.ResourceConfig.CpuLimit,
				MemoryRequest: req.HighAvailability.ResourceConfig.MemoryRequest,
				MemoryLimit:   req.HighAvailability.ResourceConfig.MemoryLimit,
			}
		}

		if req.HighAvailability.DedicatedNode != nil {
			m.HighAvailability.DedicatedNode = &DedicatedNode{
				Enabled:    req.HighAvailability.DedicatedNode.Enabled,
				NodeLabels: req.HighAvailability.DedicatedNode.NodeLabels,
			}

		}
	}

	// 转换日志收集配置
	if m.ObservabilityConfig == nil {
		m.ObservabilityConfig = &ObservabilityConfig{}
	}
	if req.ObservabilityConfig != nil && req.ObservabilityConfig.LogCollectorConfig != nil {
		m.ObservabilityConfig.LogCollectorConfig = &LogCollectorConfig{
			Enabled:           req.ObservabilityConfig.LogCollectorConfig.Enabled,
			AccessLogEncoding: req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding,
			AccessLogFormat:   req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat,
		}
	}

	// 转换链路追踪配置
	if req.ObservabilityConfig != nil && req.ObservabilityConfig.TracingConfig != nil {
		m.ObservabilityConfig.TracingConfig = &TracingConfig{
			Enabled:              req.ObservabilityConfig.TracingConfig.Enabled,
			Endpoint:             req.ObservabilityConfig.TracingConfig.Endpoint,
			BkToken:              req.ObservabilityConfig.TracingConfig.BkToken,
			TraceSamplingPercent: req.ObservabilityConfig.TracingConfig.TraceSamplingPercent,
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
				DefaultValue:    config.DefaultValue,
				AvailableValues: config.AvailableValues,
				SupportVersions: config.SupportVersions,
			}
		}
	}
}

// UpdateFromProto converts UpdateIstioRequest to update fields
func (m *MeshIstio) UpdateFromProto(req *meshmanager.UpdateIstioRequest) M {
	updateFields := m.updateBasicFields(req)
	m.updateResourceConfigs(req, updateFields)
	m.updateFeatureConfigs(req, updateFields)
	return updateFields
}

// updateBasicFields updates basic fields from request
func (m *MeshIstio) updateBasicFields(req *meshmanager.UpdateIstioRequest) M {
	return M{
		"description":      req.Description,
		"primaryClusters":  req.PrimaryClusters,
		"remoteClusters":   req.RemoteClusters,
		"differentNetwork": req.DifferentNetwork,
		"updateTime":       time.Now().Unix(),
		"updateBy":         "system", // TODO: get from context
	}
}

// updateResourceConfigs updates resource related configurations
func (m *MeshIstio) updateResourceConfigs(req *meshmanager.UpdateIstioRequest, updateFields M) {
	// Update Sidecar resource config
	if req.SidecarResourceConfig != nil {
		updateFields["sidecarResourceConfig"] = &ResourceConfig{
			CpuRequest:    req.SidecarResourceConfig.CpuRequest,
			CpuLimit:      req.SidecarResourceConfig.CpuLimit,
			MemoryRequest: req.SidecarResourceConfig.MemoryRequest,
			MemoryLimit:   req.SidecarResourceConfig.MemoryLimit,
		}
	}

	// Update high availability config
	if req.HighAvailability != nil {
		updateFields["highAvailability"] = m.convertHighAvailability(req.HighAvailability)
	}

	// Update log collector config
	if req.LogCollectorConfig != nil {
		updateFields["logCollectorConfig"] = &LogCollectorConfig{
			Enabled:           req.LogCollectorConfig.Enabled,
			AccessLogEncoding: req.LogCollectorConfig.AccessLogEncoding,
			AccessLogFormat:   req.LogCollectorConfig.AccessLogFormat,
		}
	}

	// Update tracing config
	if req.TracingConfig != nil {
		updateFields["tracingConfig"] = &TracingConfig{
			Enabled:  req.TracingConfig.Enabled,
			Endpoint: req.TracingConfig.Endpoint,
			BkToken:  req.TracingConfig.BkToken,
		}
	}
}

// convertHighAvailability converts proto HighAvailability to entity
func (m *MeshIstio) convertHighAvailability(ha *meshmanager.HighAvailability) *HighAvailability {
	highAvailability := &HighAvailability{
		AutoscaleEnabled:                   ha.AutoscaleEnabled,
		AutoscaleMin:                       ha.AutoscaleMin,
		AutoscaleMax:                       ha.AutoscaleMax,
		ReplicaCount:                       ha.ReplicaCount,
		TargetCPUAverageUtilizationPercent: ha.TargetCPUAverageUtilizationPercent,
	}

	if ha.ResourceConfig != nil {
		highAvailability.ResourceConfig = &ResourceConfig{
			CpuRequest:    ha.ResourceConfig.CpuRequest,
			CpuLimit:      ha.ResourceConfig.CpuLimit,
			MemoryRequest: ha.ResourceConfig.MemoryRequest,
			MemoryLimit:   ha.ResourceConfig.MemoryLimit,
		}
	}

	if ha.DedicatedNode != nil {
		highAvailability.DedicatedNode = &DedicatedNode{
			Enabled:    ha.DedicatedNode.Enabled,
			NodeLabels: ha.DedicatedNode.NodeLabels,
		}
	}

	return highAvailability
}

// updateFeatureConfigs updates feature configurations
func (m *MeshIstio) updateFeatureConfigs(req *meshmanager.UpdateIstioRequest, updateFields M) {
	if len(req.FeatureConfigs) == 0 {
		return
	}

	featureConfigs := make(map[string]*FeatureConfig)
	for name, config := range req.FeatureConfigs {
		// Only save supported features
		if !slices.Contains(common.SupportedFeatures, name) {
			continue
		}
		featureConfigs[name] = &FeatureConfig{
			Name:            config.Name,
			Description:     config.Description,
			DefaultValue:    config.DefaultValue,
			AvailableValues: config.AvailableValues,
			SupportVersions: config.SupportVersions,
		}
	}
	updateFields["featureConfigs"] = featureConfigs
}
