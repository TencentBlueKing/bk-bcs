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

// Mesh represents a service mesh entity in database
type Mesh struct {
	// Basic information
	MeshID       string `bson:"meshID" json:"meshID" validate:"required"`
	MeshName     string `bson:"meshName" json:"meshName" validate:"required"`
	ProjectID    string `bson:"projectID" json:"projectID" validate:"required"`
	ProjectCode  string `bson:"projectCode" json:"projectCode" validate:"required"`
	Description  string `bson:"description" json:"description"`
	ChartVersion string `bson:"chartVersion" json:"chartVersion" validate:"required"`
	Status       string `bson:"status" json:"status" validate:"required"`
	CreateTime   int64  `bson:"createTime" json:"createTime"`
	UpdateTime   int64  `bson:"updateTime" json:"updateTime"`
	CreateBy     string `bson:"createBy" json:"createBy"`
	UpdateBy     string `bson:"updateBy" json:"updateBy"`

	// Mesh configuration
	ControlPlaneMode string   `bson:"controlPlaneMode" json:"controlPlaneMode"`
	ClusterMode      string   `bson:"clusterMode" json:"clusterMode"`
	PrimaryClusters  []string `bson:"primaryClusters" json:"primaryClusters"`
	RemoteClusters   []string `bson:"remoteClusters" json:"remoteClusters"`
	DifferentNetwork bool     `bson:"differentNetwork" json:"differentNetwork"`

	// Service discovery information
	ServiceDiscovery *ServiceDiscovery `bson:"serviceDiscovery" json:"serviceDiscovery"`

	// Feature configurations
	FeatureConfigs map[string]*FeatureConfig `bson:"featureConfigs" json:"featureConfigs"`

	// Resource and observability configurations
	SidecarResourceConfig *ResourceConfig     `bson:"sidecarResourceConfig" json:"sidecarResourceConfig"`
	HighAvailability      *HighAvailability   `bson:"highAvailability" json:"highAvailability"`
	LogCollectorConfig    *LogCollectorConfig `bson:"logCollectorConfig" json:"logCollectorConfig"`
	TracingConfig         *TracingConfig      `bson:"tracingConfig" json:"tracingConfig"`
}

// ResourceConfig represents resource configuration for sidecar
type ResourceConfig struct {
	CpuRequest    string `bson:"cpuRequest" json:"cpuRequest"`
	CpuLimit      string `bson:"cpuLimit" json:"cpuLimit"`
	MemoryRequest string `bson:"memoryRequest" json:"memoryRequest"`
	MemoryLimit   string `bson:"memoryLimit" json:"memoryLimit"`
}

// DedicatedNodeLabel represents dedicated node label configuration
type DedicatedNodeLabel struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

// HighAvailability represents high availability configuration
type HighAvailability struct {
	AutoscaleEnabled   bool                `bson:"autoscaleEnabled" json:"autoscaleEnabled"`
	AutoscaleMin       int32               `bson:"autoscaleMin" json:"autoscaleMin"`
	AutoscaleMax       int32               `bson:"autoscaleMax" json:"autoscaleMax"`
	ReplicaCount       int32               `bson:"replicaCount" json:"replicaCount"`
	ResourceConfig     *ResourceConfig     `bson:"resourceConfig" json:"resourceConfig"`
	DedicatedNodeLabel *DedicatedNodeLabel `bson:"dedicatedNodeLabel" json:"dedicatedNodeLabel"`
}

// LogCollectorConfig represents log collector configuration
type LogCollectorConfig struct {
	Enabled           bool   `bson:"enabled" json:"enabled"`
	AccessLogEncoding string `bson:"accessLogEncoding" json:"accessLogEncoding"`
	AccessLogFormat   string `bson:"accessLogFormat" json:"accessLogFormat"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled  bool   `bson:"enabled" json:"enabled"`
	Endpoint string `bson:"endpoint" json:"endpoint"`
	BkToken  string `bson:"bkToken" json:"bkToken"`
}

// ServiceDiscovery represents the service discovery configuration of the mesh
type ServiceDiscovery struct {
	// List of cluster IDs that this mesh is bound to
	Clusters []string `bson:"clusters" json:"clusters"`
	// Auto-injection namespace mapping
	// key: clusterID, value: list of namespaces that need auto-injection in this cluster
	AutoInjectNS map[string][]string `bson:"autoInjectNS" json:"autoInjectNS"`
	// Disabled injection pods mapping
	// key: clusterID, value: map[namespace][]podName
	DisabledInjectPods map[string]map[string][]string `bson:"disabledInjectPods" json:"disabledInjectPods"`
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

// Transfer2Proto converts Mesh entity to proto message
func (m *Mesh) Transfer2Proto() *meshmanager.MeshListItem {
	// 转换基本字段
	proto := m.transferBasicFields()

	// 转换服务发现配置
	proto.ServiceDiscovery = m.transferServiceDiscovery()

	// 转换配置相关字段
	proto.SidecarResourceConfig = m.transferSidecarResourceConfig()
	proto.HighAvailability = m.transferHighAvailability()
	proto.LogCollectorConfig = m.transferLogCollectorConfig()
	proto.TracingConfig = m.transferTracingConfig()
	proto.FeatureConfigs = m.transferFeatureConfigs()

	return proto
}

// transferBasicFields 转换基本字段
func (m *Mesh) transferBasicFields() *meshmanager.MeshListItem {
	return &meshmanager.MeshListItem{
		MeshID:           m.MeshID,
		MeshName:         m.MeshName,
		ProjectID:        m.ProjectID,
		ProjectCode:      m.ProjectCode,
		Description:      m.Description,
		ChartVersion:     m.ChartVersion,
		Status:           meshmanager.MeshStatus(meshmanager.MeshStatus_value[m.Status]),
		CreateTime:       m.CreateTime,
		UpdateTime:       m.UpdateTime,
		CreateBy:         m.CreateBy,
		UpdateBy:         m.UpdateBy,
		ControlPlaneMode: meshmanager.ControlPlaneMode(meshmanager.ControlPlaneMode_value[m.ControlPlaneMode]),
		ClusterMode:      meshmanager.ClusterMode(meshmanager.ClusterMode_value[m.ClusterMode]),
		PrimaryClusters:  m.PrimaryClusters,
		RemoteClusters:   m.RemoteClusters,
		DifferentNetwork: m.DifferentNetwork,
	}
}

// transferServiceDiscovery 转换服务发现配置
func (m *Mesh) transferServiceDiscovery() *meshmanager.ServiceDiscovery {
	if m.ServiceDiscovery == nil {
		return nil
	}

	protoServiceDiscovery := &meshmanager.ServiceDiscovery{
		Clusters:           m.ServiceDiscovery.Clusters,
		AutoInjectNS:       make(map[string]*meshmanager.NamespaceList),
		DisabledInjectPods: make(map[string]*meshmanager.NamespacePods),
	}

	// 转换 AutoInjectNS
	for clusterID, namespaces := range m.ServiceDiscovery.AutoInjectNS {
		protoServiceDiscovery.AutoInjectNS[clusterID] = &meshmanager.NamespaceList{
			Namespaces: namespaces,
		}
	}

	// 转换 DisabledInjectPods
	for clusterID, namespacePods := range m.ServiceDiscovery.DisabledInjectPods {
		protoNamespacePods := &meshmanager.NamespacePods{
			NamespacePods: make(map[string]*meshmanager.PodList),
		}
		for namespace, pods := range namespacePods {
			protoNamespacePods.NamespacePods[namespace] = &meshmanager.PodList{
				Pods: pods,
			}
		}
		protoServiceDiscovery.DisabledInjectPods[clusterID] = protoNamespacePods
	}

	return protoServiceDiscovery
}

// transferFeatureConfigs 转换特性配置
func (m *Mesh) transferFeatureConfigs() map[string]*meshmanager.FeatureConfig {
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
func (m *Mesh) transferSidecarResourceConfig() *meshmanager.ResourceConfig {
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
func (m *Mesh) transferHighAvailability() *meshmanager.HighAvailability {
	if m.HighAvailability == nil {
		return nil
	}

	protoHighAvailability := &meshmanager.HighAvailability{
		AutoscaleEnabled: m.HighAvailability.AutoscaleEnabled,
		AutoscaleMin:     m.HighAvailability.AutoscaleMin,
		AutoscaleMax:     m.HighAvailability.AutoscaleMax,
		ReplicaCount:     m.HighAvailability.ReplicaCount,
	}

	if m.HighAvailability.ResourceConfig != nil {
		protoHighAvailability.ResourceConfig = &meshmanager.ResourceConfig{
			CpuRequest:    m.HighAvailability.ResourceConfig.CpuRequest,
			CpuLimit:      m.HighAvailability.ResourceConfig.CpuLimit,
			MemoryRequest: m.HighAvailability.ResourceConfig.MemoryRequest,
			MemoryLimit:   m.HighAvailability.ResourceConfig.MemoryLimit,
		}
	}

	if m.HighAvailability.DedicatedNodeLabel != nil {
		protoHighAvailability.DedicatedNodeLabel = &meshmanager.DedicatedNodeLabel{
			Key:   m.HighAvailability.DedicatedNodeLabel.Key,
			Value: m.HighAvailability.DedicatedNodeLabel.Value,
		}
	}

	return protoHighAvailability
}

// transferLogCollectorConfig 转换日志收集配置
func (m *Mesh) transferLogCollectorConfig() *meshmanager.LogCollectorConfig {
	if m.LogCollectorConfig == nil {
		return nil
	}
	encoding := meshmanager.AccessLogEncoding(
		meshmanager.AccessLogEncoding_value[m.LogCollectorConfig.AccessLogEncoding])
	return &meshmanager.LogCollectorConfig{
		Enabled:           m.LogCollectorConfig.Enabled,
		AccessLogEncoding: encoding,
		AccessLogFormat:   m.LogCollectorConfig.AccessLogFormat,
	}
}

// transferTracingConfig 转换链路追踪配置
func (m *Mesh) transferTracingConfig() *meshmanager.TracingConfig {
	if m.TracingConfig == nil {
		return nil
	}
	return &meshmanager.TracingConfig{
		Enabled:  m.TracingConfig.Enabled,
		Endpoint: m.TracingConfig.Endpoint,
		BkToken:  m.TracingConfig.BkToken,
	}
}

// TransferFromProto converts InstallIstioRequest to Mesh entity
func (m *Mesh) TransferFromProto(req *meshmanager.InstallIstioRequest) {
	// 转换基本字段
	m.MeshName = req.MeshName
	m.ProjectID = req.ProjectID
	m.ProjectCode = req.ProjectCode
	m.Description = req.Description
	m.ChartVersion = req.ChartVersion
	m.Status = meshmanager.MeshStatus_MESH_STATUS_INSTALLING.String()
	m.ControlPlaneMode = req.ControlPlaneMode.String()
	m.ClusterMode = req.ClusterMode.String()
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
			AutoscaleEnabled: req.HighAvailability.AutoscaleEnabled,
			AutoscaleMin:     req.HighAvailability.AutoscaleMin,
			AutoscaleMax:     req.HighAvailability.AutoscaleMax,
			ReplicaCount:     req.HighAvailability.ReplicaCount,
		}

		if req.HighAvailability.ResourceConfig != nil {
			m.HighAvailability.ResourceConfig = &ResourceConfig{
				CpuRequest:    req.HighAvailability.ResourceConfig.CpuRequest,
				CpuLimit:      req.HighAvailability.ResourceConfig.CpuLimit,
				MemoryRequest: req.HighAvailability.ResourceConfig.MemoryRequest,
				MemoryLimit:   req.HighAvailability.ResourceConfig.MemoryLimit,
			}
		}

		if req.HighAvailability.DedicatedNodeLabel != nil {
			m.HighAvailability.DedicatedNodeLabel = &DedicatedNodeLabel{
				Key:   req.HighAvailability.DedicatedNodeLabel.Key,
				Value: req.HighAvailability.DedicatedNodeLabel.Value,
			}
		}
	}

	// 转换日志收集配置
	if req.LogCollectorConfig != nil {
		m.LogCollectorConfig = &LogCollectorConfig{
			Enabled:           req.LogCollectorConfig.Enabled,
			AccessLogEncoding: req.LogCollectorConfig.AccessLogEncoding.String(),
			AccessLogFormat:   req.LogCollectorConfig.AccessLogFormat,
		}
	}

	// 转换链路追踪配置
	if req.TracingConfig != nil {
		m.TracingConfig = &TracingConfig{
			Enabled:  req.TracingConfig.Enabled,
			Endpoint: req.TracingConfig.Endpoint,
			BkToken:  req.TracingConfig.BkToken,
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

// UpdateFromProto converts UpdateMeshRequest to update fields
func (m *Mesh) UpdateFromProto(req *meshmanager.UpdateMeshRequest) M {
	updateFields := m.updateBasicFields(req)
	m.updateServiceDiscovery(req, updateFields)
	m.updateResourceConfigs(req, updateFields)
	m.updateFeatureConfigs(req, updateFields)
	return updateFields
}

// updateBasicFields updates basic fields from request
func (m *Mesh) updateBasicFields(req *meshmanager.UpdateMeshRequest) M {
	return M{
		"description":      req.Description,
		"primaryClusters":  req.PrimaryClusters,
		"remoteClusters":   req.RemoteClusters,
		"differentNetwork": req.DifferentNetwork,
		"updateTime":       time.Now().Unix(),
		"updateBy":         "system", // TODO: get from context
	}
}

// updateServiceDiscovery updates service discovery configuration
func (m *Mesh) updateServiceDiscovery(req *meshmanager.UpdateMeshRequest, updateFields M) {
	if req.ServiceDiscovery == nil {
		return
	}

	serviceDiscovery := &ServiceDiscovery{
		Clusters:           req.ServiceDiscovery.Clusters,
		AutoInjectNS:       make(map[string][]string),
		DisabledInjectPods: make(map[string]map[string][]string),
	}

	// Convert AutoInjectNS
	for clusterID, namespaceList := range req.ServiceDiscovery.AutoInjectNS {
		serviceDiscovery.AutoInjectNS[clusterID] = namespaceList.Namespaces
	}

	// Convert DisabledInjectPods
	for clusterID, namespacePods := range req.ServiceDiscovery.DisabledInjectPods {
		clusterPods := make(map[string][]string)
		for namespace, podList := range namespacePods.NamespacePods {
			clusterPods[namespace] = podList.Pods
		}
		serviceDiscovery.DisabledInjectPods[clusterID] = clusterPods
	}

	updateFields["serviceDiscovery"] = serviceDiscovery
}

// updateResourceConfigs updates resource related configurations
func (m *Mesh) updateResourceConfigs(req *meshmanager.UpdateMeshRequest, updateFields M) {
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
			AccessLogEncoding: req.LogCollectorConfig.AccessLogEncoding.String(),
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
func (m *Mesh) convertHighAvailability(ha *meshmanager.HighAvailability) *HighAvailability {
	highAvailability := &HighAvailability{
		AutoscaleEnabled: ha.AutoscaleEnabled,
		AutoscaleMin:     ha.AutoscaleMin,
		AutoscaleMax:     ha.AutoscaleMax,
		ReplicaCount:     ha.ReplicaCount,
	}

	if ha.ResourceConfig != nil {
		highAvailability.ResourceConfig = &ResourceConfig{
			CpuRequest:    ha.ResourceConfig.CpuRequest,
			CpuLimit:      ha.ResourceConfig.CpuLimit,
			MemoryRequest: ha.ResourceConfig.MemoryRequest,
			MemoryLimit:   ha.ResourceConfig.MemoryLimit,
		}
	}

	if ha.DedicatedNodeLabel != nil {
		highAvailability.DedicatedNodeLabel = &DedicatedNodeLabel{
			Key:   ha.DedicatedNodeLabel.Key,
			Value: ha.DedicatedNodeLabel.Value,
		}
	}

	return highAvailability
}

// updateFeatureConfigs updates feature configurations
func (m *Mesh) updateFeatureConfigs(req *meshmanager.UpdateMeshRequest, updateFields M) {
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
