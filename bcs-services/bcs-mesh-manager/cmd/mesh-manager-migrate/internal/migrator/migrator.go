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

// Package migrator 用于执行迁移逻辑
package migrator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager-migrate/internal/config"
	util "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager-migrate/internal/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshManagerStore "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
)

// Migrator istio迁移工具
type Migrator struct {
	config           *config.MigrateConfig
	meshManagerModel meshManagerStore.MeshManagerModel
}

// New 创建迁移器
func New(cfg *config.MigrateConfig) *Migrator {
	return &Migrator{
		config: cfg,
	}
}

// Init 初始化迁移器
func (m *Migrator) Init() error {
	initializer := []func() error{
		m.initMeshManagerModel,
	}

	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}
	return nil
}

// initMongoModel 通用的MongoDB初始化函数
func (m *Migrator) initMongoModel(
	database string,
	modelName string,
	storeInitializer func(*mongo.DB) interface{},
) error {
	if len(m.config.Mongo.Address) == 0 {
		return fmt.Errorf("mongo endpoints cannot be empty")
	}
	if len(database) == 0 {
		return fmt.Errorf("mongo database cannot be empty")
	}

	// get mongo password
	password := m.config.Mongo.Password

	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(m.config.Mongo.Address, ","),
		Replicaset:            m.config.Mongo.Replicaset,
		AuthDatabase:          m.config.Mongo.AuthDatabase,
		ConnectTimeoutSeconds: int(m.config.Mongo.ConnectTimeout),
		Database:              database,
		Username:              m.config.Mongo.Username,
		Password:              password,
		MaxPoolSize:           uint64(m.config.Mongo.MaxPoolSize),
		MinPoolSize:           uint64(m.config.Mongo.MinPoolSize),
	}

	// init mongo db
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		log.Printf("init %s mongo db failed, err %s", modelName, err.Error())
		return err
	}

	// ping mongo to check connection
	if err = mongoDB.Ping(); err != nil {
		log.Printf("ping %s mongo db failed, err %s", modelName, err.Error())
		return err
	}
	log.Printf("init %s mongo db successfully", modelName)

	// init store
	storeInitializer(mongoDB)
	log.Printf("init %s store successfully", modelName)

	return nil
}

const (
	meshManagerModel = "meshmanager"
)

func (m *Migrator) initMeshManagerModel() error {
	return m.initMongoModel(
		m.config.Mongo.Database,
		meshManagerModel,
		func(db *mongo.DB) interface{} {
			m.meshManagerModel = meshManagerStore.New(db)
			return nil
		},
	)
}

// Migrate 执行迁移逻辑
func (m *Migrator) Migrate(opts *MigrateOptions) error {
	log.Printf("start to migrate mesh %s", opts.MeshName)
	ctx := context.Background()

	// 构建 mesh 对象
	log.Printf("start to build mesh istio")
	meshIstio, err := m.buildMeshIstio(opts)
	if err != nil {
		return fmt.Errorf("failed to build mesh istio: %v", err)
	}
	log.Printf("build mesh istio successfully")

	// 插入 mesh 对象
	if err := m.meshManagerModel.Create(ctx, meshIstio); err != nil {
		return fmt.Errorf("failed to create mesh istio: %v", err)
	}
	log.Printf("create mesh istio successfully")

	return nil
}

// buildMeshIstio 构建 mesh istio 对象
// nolint:funlen
func (m *Migrator) buildMeshIstio(
	opts *MigrateOptions,
) (*entity.MeshIstio, error) {
	meshIstio := &entity.MeshIstio{}
	appVersion, err := util.GetAppVersion(opts.PrimaryClusterID, opts.IstiodReleaseName, opts.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm app version: %v", err)
	}
	meshIstio.Version = m.getIstioVersion(appVersion)
	meshIstio.Name = opts.MeshName
	meshIstio.ProjectCode = opts.ProjectCode
	meshIstio.Version = m.getIstioVersion(appVersion)
	meshIstio.ChartVersion = appVersion
	meshIstio.Status = common.IstioStatusRunning
	meshIstio.CreateTime = time.Now().UnixMilli()
	meshIstio.UpdateTime = time.Now().UnixMilli()
	meshIstio.CreateBy = opts.BcsUsername
	meshIstio.IsDeleted = false
	meshIstio.PrimaryClusters = []string{opts.PrimaryClusterID}
	meshIstio.ControlPlaneMode = common.ControlPlaneModeIndependent
	meshIstio.Description = opts.Description

	// 获取 istiod 的 values
	values, err := util.GetValues(opts.PrimaryClusterID, opts.IstiodReleaseName, opts.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm release: %v", err)
	}

	istiodValues := &common.IstiodInstallValues{}
	if err = yaml.Unmarshal(values, istiodValues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal helm values to IstiodInstallValues: %v", err)
	}
	// 从 values 中获取 network ID 和 meshID
	if istiodValues.Global != nil {
		if istiodValues.Global.Network != nil {
			meshIstio.NetworkID = *istiodValues.Global.Network
		}
		if istiodValues.Global.MeshID != nil && *istiodValues.Global.MeshID != "" {
			meshIstio.MeshID = *istiodValues.Global.MeshID
		} else {
			meshIstio.MeshID = utils.GenMeshID()
		}
	}
	// 从 release values 中提取配置
	if istiodValues.Revision != nil {
		meshIstio.Revision = *istiodValues.Revision
	}

	// 多集群配置
	if opts.MultiClusterEnabled {
		log.Printf("multi cluster enabled")
		meshIstio.ClusterMode = common.MultiClusterModePrimaryRemote
		meshIstio.DifferentNetwork = true
		meshIstio.MultiClusterEnabled = true

		// 获取东西向网关的 values
		values, err := util.GetValues(opts.PrimaryClusterID, opts.GatewaysReleaseName, opts.KubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get helm release: %v", err)
		}
		eastwestGatewayValues := &common.EastwestGatewayValues{}
		if err := yaml.Unmarshal(values, eastwestGatewayValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal helm values to EastwestGatewayValues: %v", err)
		}
		// 获取东西向网关的 clbID
		if eastwestGatewayValues.Service.Annotations == nil {
			return nil, fmt.Errorf("failed to get clbID from eastwestgateway, service annotations is empty")
		}

		clbID, ok := eastwestGatewayValues.Service.Annotations[utils.KeyServiceAnnotationCLBID]
		if !ok {
			return nil, fmt.Errorf("failed to get clbID from eastwestgateway, service annotations is empty")
		}
		meshIstio.ClbID = clbID
		if opts.RemoteClusters != "" {
			remoteClusters := strings.Split(opts.RemoteClusters, ",")
			meshIstio.RemoteClusters = make([]*entity.RemoteCluster, 0)
			for _, cluster := range remoteClusters {
				clusterID := strings.TrimSpace(cluster)
				if clusterID == "" {
					continue
				}
				remoteCluster := &entity.RemoteCluster{
					ClusterID: clusterID,
					Status:    common.IstioStatusRunning,
					JoinTime:  time.Now().UnixMilli(),
				}

				meshIstio.RemoteClusters = append(meshIstio.RemoteClusters, remoteCluster)
			}
		}
	}
	meshIstio.FeatureConfigs = m.buildFeatureConfigs(istiodValues)
	meshIstio.SidecarResourceConfig = m.buildSidecarResourceConfig(istiodValues)
	meshIstio.HighAvailability = m.buildHighAvailabilityConfig(istiodValues)
	meshIstio.ObservabilityConfig = m.buildObservabilityConfig(istiodValues, opts)

	meshIstio.ReleaseNames = m.buildReleaseNames(opts)

	return meshIstio, nil
}

// buildFeatureConfigs 构建特性配置
func (m *Migrator) buildFeatureConfigs(values *common.IstiodInstallValues) map[string]*entity.FeatureConfig {
	featureConfigs := make(map[string]*entity.FeatureConfig)

	// 获取默认特性配置模板
	defaultConfigs := common.GetDefaultFeatureConfigs()
	for featureKey, template := range defaultConfigs {
		featureConfigs[featureKey] = &entity.FeatureConfig{
			Name:            template.Name,
			Description:     template.Description,
			Value:           template.DefaultValue,
			DefaultValue:    template.DefaultValue,
			AvailableValues: template.AvailableValues,
		}
	}

	// 根据 values 更新实际配置值
	if values.MeshConfig != nil {
		// 出站流量策略
		if values.MeshConfig.OutboundTrafficPolicy != nil && values.MeshConfig.OutboundTrafficPolicy.Mode != nil {
			if config, exists := featureConfigs[common.FeatureOutboundTrafficPolicy]; exists {
				config.Value = *values.MeshConfig.OutboundTrafficPolicy.Mode
			}
		}

		if values.MeshConfig.DefaultConfig != nil {
			// 应用等待代理启动
			if values.MeshConfig.DefaultConfig.HoldApplicationUntilProxyStarts != nil {
				if config, exists := featureConfigs[common.FeatureHoldApplicationUntilProxyStarts]; exists {
					config.Value = fmt.Sprintf("%t", *values.MeshConfig.DefaultConfig.HoldApplicationUntilProxyStarts)
				}
			}

			// ProxyMetadata 相关配置
			if values.MeshConfig.DefaultConfig.ProxyMetadata != nil {
				metadata := values.MeshConfig.DefaultConfig.ProxyMetadata
				// 零连接时退出
				if metadata.ExitOnZeroActiveConnections != nil {
					if config, exists := featureConfigs[common.FeatureExitOnZeroActiveConnections]; exists {
						config.Value = *metadata.ExitOnZeroActiveConnections
					}
				}

				// DNS 捕获
				if metadata.IstioMetaDnsCapture != nil {
					if config, exists := featureConfigs[common.FeatureIstioMetaDnsCapture]; exists {
						config.Value = *metadata.IstioMetaDnsCapture
					}
				}

				// DNS 自动分配
				if metadata.IstioMetaDnsAutoAllocate != nil {
					if config, exists := featureConfigs[common.FeatureIstioMetaDnsAutoAllocate]; exists {
						config.Value = *metadata.IstioMetaDnsAutoAllocate
					}
				}
			}
		}
	}

	// Pilot 环境变量相关配置
	if values.Pilot != nil && values.Pilot.Env != nil {
		// HTTP/1.0 支持
		if http10, ok := values.Pilot.Env[common.EnvPilotHTTP10]; ok {
			if config, exists := featureConfigs[common.FeatureIstioMetaHttp10]; exists {
				config.Value = http10
			}
		}
	}

	// 排除 IP 范围配置
	if values.Global != nil && values.Global.Proxy != nil && values.Global.Proxy.ExcludeIPRanges != nil {
		if config, exists := featureConfigs[common.FeatureExcludeIPRanges]; exists {
			config.Value = *values.Global.Proxy.ExcludeIPRanges
		}
	}

	return featureConfigs
}

// buildSidecarResourceConfig 构建sidecar资源配置
func (m *Migrator) buildSidecarResourceConfig(values *common.IstiodInstallValues) *entity.ResourceConfig {
	if values.Global != nil && values.Global.Proxy != nil && values.Global.Proxy.Resources != nil {
		resources := values.Global.Proxy.Resources
		config := &entity.ResourceConfig{
			CpuRequest:    "100m",
			CpuLimit:      "2000m",
			MemoryRequest: "128Mi",
			MemoryLimit:   "1024Mi",
		}

		if resources.Requests != nil {
			if resources.Requests.CPU != nil {
				config.CpuRequest = *resources.Requests.CPU
			}
			if resources.Requests.Memory != nil {
				config.MemoryRequest = *resources.Requests.Memory
			}
		}
		if resources.Limits != nil {
			if resources.Limits.CPU != nil {
				config.CpuLimit = *resources.Limits.CPU
			}
			if resources.Limits.Memory != nil {
				config.MemoryLimit = *resources.Limits.Memory
			}
		}

		return config
	}
	return nil
}

// buildHighAvailabilityConfig 构建高可用配置
func (m *Migrator) buildHighAvailabilityConfig(values *common.IstiodInstallValues) *entity.HighAvailability {
	if values.Pilot != nil {
		haConfig := &entity.HighAvailability{
			AutoscaleEnabled:                   true,
			AutoscaleMin:                       1,
			AutoscaleMax:                       5,
			ReplicaCount:                       1,
			TargetCPUAverageUtilizationPercent: 80,
			ResourceConfig: &entity.ResourceConfig{
				CpuRequest:    "500m",
				MemoryRequest: "2048Mi",
			},
			DedicatedNode: &entity.DedicatedNode{
				Enabled:    false,
				NodeLabels: make(map[string]string),
			},
		}

		// 自动扩缩容配置
		if values.Pilot.AutoscaleEnabled != nil {
			haConfig.AutoscaleEnabled = *values.Pilot.AutoscaleEnabled
		}
		if values.Pilot.AutoscaleMin != nil {
			haConfig.AutoscaleMin = *values.Pilot.AutoscaleMin
		}
		if values.Pilot.AutoscaleMax != nil {
			haConfig.AutoscaleMax = *values.Pilot.AutoscaleMax
		}

		// 副本数配置
		if values.Pilot.ReplicaCount != nil {
			haConfig.ReplicaCount = *values.Pilot.ReplicaCount
		}

		// CPU 配置
		if values.Pilot.CPU != nil && values.Pilot.CPU.TargetAverageUtilization != nil {
			haConfig.TargetCPUAverageUtilizationPercent = *values.Pilot.CPU.TargetAverageUtilization
		}

		// 资源配置
		if values.Pilot.Resources != nil {
			if values.Pilot.Resources.Requests != nil {
				if values.Pilot.Resources.Requests.CPU != nil {
					haConfig.ResourceConfig.CpuRequest = *values.Pilot.Resources.Requests.CPU
				}
				if values.Pilot.Resources.Requests.Memory != nil {
					haConfig.ResourceConfig.MemoryRequest = *values.Pilot.Resources.Requests.Memory
				}
			}
			if values.Pilot.Resources.Limits != nil {
				if values.Pilot.Resources.Limits.CPU != nil {
					haConfig.ResourceConfig.CpuLimit = *values.Pilot.Resources.Limits.CPU
				}
				if values.Pilot.Resources.Limits.Memory != nil {
					haConfig.ResourceConfig.MemoryLimit = *values.Pilot.Resources.Limits.Memory
				}
			}
		}

		// 节点选择器
		if len(values.Pilot.NodeSelector) > 0 {
			haConfig.DedicatedNode.Enabled = true
			for k, v := range values.Pilot.NodeSelector {
				haConfig.DedicatedNode.NodeLabels[k] = v
			}
		}

		return haConfig
	}
	return nil
}

// buildObservabilityConfig 构建observability配置
func (m *Migrator) buildObservabilityConfig(
	values *common.IstiodInstallValues,
	opts *MigrateOptions,
) *entity.ObservabilityConfig {
	obsConfig := &entity.ObservabilityConfig{
		MetricsConfig: &entity.MetricsConfig{
			MetricsEnabled:             false,
			ControlPlaneMetricsEnabled: false,
			DataPlaneMetricsEnabled:    false,
		},
		LogCollectorConfig: &entity.LogCollectorConfig{
			Enabled:           false,
			AccessLogEncoding: common.AccessLogEncodingTEXT,
			AccessLogFormat:   "",
		},
		TracingConfig: &entity.TracingConfig{
			Enabled:              true,
			Endpoint:             "zipkin.istio-system:9411",
			BkToken:              "",
			TraceSamplingPercent: 1,
		},
	}

	// 指标配置
	obsConfig.MetricsConfig = &entity.MetricsConfig{
		MetricsEnabled:             opts.MetricsEnabled,
		ControlPlaneMetricsEnabled: opts.ControlPlaneMetricsEnabled,
		DataPlaneMetricsEnabled:    opts.DataPlaneMetricsEnabled,
	}

	// 从 MeshConfig 中提取日志配置
	if values.MeshConfig != nil {
		// 访问日志文件配置
		if values.MeshConfig.AccessLogFile != nil {
			// 如果访问日志文件不为空，则表示启用日志收集
			obsConfig.LogCollectorConfig.Enabled = *values.MeshConfig.AccessLogFile != ""
		}
		// 访问日志编码
		if values.MeshConfig.AccessLogEncoding != nil {
			obsConfig.LogCollectorConfig.AccessLogEncoding = *values.MeshConfig.AccessLogEncoding
		}
		// 访问日志格式
		if values.MeshConfig.AccessLogFormat != nil {
			obsConfig.LogCollectorConfig.AccessLogFormat = *values.MeshConfig.AccessLogFormat
		}
	}

	// 从 MeshConfig 中提取追踪配置
	if values.MeshConfig != nil {
		// 追踪开关
		if values.MeshConfig.EnableTracing != nil {
			obsConfig.TracingConfig.Enabled = *values.MeshConfig.EnableTracing
		}

		// 迁移参数中未设置追踪端点，则从values中获取
		if opts.TracingEndpoint != "" {
			obsConfig.TracingConfig.Endpoint = opts.TracingEndpoint
		} else if values.MeshConfig.DefaultConfig != nil &&
			values.MeshConfig.DefaultConfig.TracingConfig != nil &&
			values.MeshConfig.DefaultConfig.TracingConfig.Zipkin != nil &&
			values.MeshConfig.DefaultConfig.TracingConfig.Zipkin.Address != nil {
			obsConfig.TracingConfig.Endpoint = *values.MeshConfig.DefaultConfig.TracingConfig.Zipkin.Address
		}

		// 从提供的参数中设置 token
		if opts.BkToken != "" {
			obsConfig.TracingConfig.BkToken = opts.BkToken
		}
	}

	// 采样率配置
	if values.Pilot != nil && values.Pilot.TraceSampling != nil {
		// Pilot 中的采样率是 0.0-1.0，需要转换为百分比
		obsConfig.TracingConfig.TraceSamplingPercent = int32(*values.Pilot.TraceSampling * 100)
	}

	return obsConfig
}

// buildReleaseNames 构建ReleaseNames map
func (m *Migrator) buildReleaseNames(opts *MigrateOptions) map[string]map[string]string {
	releaseNames := make(map[string]map[string]string)

	allClusters := []string{opts.PrimaryClusterID}
	if opts.MultiClusterEnabled && opts.RemoteClusters != "" {
		remoteClusters := strings.Split(opts.RemoteClusters, ",")
		for _, clusterID := range remoteClusters {
			if strings.TrimSpace(clusterID) != "" {
				allClusters = append(allClusters, strings.TrimSpace(clusterID))
			}
		}
	}

	for _, clusterID := range allClusters {
		releaseNames[clusterID] = map[string]string{
			common.ComponentIstioBase: opts.BaseReleaseName,
			common.ComponentIstiod:    opts.IstiodReleaseName,
		}
		// 多集群模式下，存储东西向网关的release name
		if opts.MultiClusterEnabled {
			releaseNames[clusterID][common.ComponentIstioGateway] = opts.GatewaysReleaseName
		}
	}

	return releaseNames
}

// getIstioVersion 获取istio版本，从chart版本中提取istio版本号
func (m *Migrator) getIstioVersion(chartVersion string) string {
	if chartVersion == "" {
		return ""
	}
	parts := strings.Split(chartVersion, "-")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
