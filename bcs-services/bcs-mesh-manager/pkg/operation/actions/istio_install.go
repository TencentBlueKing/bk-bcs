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

// Package actions 操作包
package actions

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
)

// IstioInstallAction istio安装操作
type IstioInstallAction struct {
	model store.MeshManagerModel

	*common.IstioInstallOption
}

var _ operation.Operation = &IstioInstallAction{}

// NewIstioInstallAction 创建istio安装操作
func NewIstioInstallAction(opt *common.IstioInstallOption, model store.MeshManagerModel) *IstioInstallAction {
	return &IstioInstallAction{
		IstioInstallOption: opt,
		model:              model,
	}
}

// Action 操作名称
func (i *IstioInstallAction) Action() string {
	return "istio-install"
}

// Name 操作实例名称
func (i *IstioInstallAction) Name() string {
	return fmt.Sprintf("istio-install-%s", i.MeshID)
}

// Validate 验证参数
func (i *IstioInstallAction) Validate() error {
	// 必填字段
	if i.ProjectCode == "" {
		return fmt.Errorf("project is required")
	}
	if len(i.PrimaryClusters) == 0 {
		return fmt.Errorf("clusters is required")
	}
	if i.Version == "" {
		return fmt.Errorf("chart version is required")
	}
	if i.ChartVersion == "" {
		return fmt.Errorf("chart version is required")
	}
	if i.FeatureConfigs == nil {
		return fmt.Errorf("feature configs is required")
	}
	if i.Revision == "" {
		return fmt.Errorf("revision is required")
	}
	return nil
}

// Prepare 准备阶段
func (i *IstioInstallAction) Prepare(ctx context.Context) error {
	blog.Infof("[%s]prepare istio install", i.MeshID)
	// 这里可以做一些准备工作
	return nil
}

// Execute 执行安装
func (i *IstioInstallAction) Execute(ctx context.Context) error {
	blog.Infof("[%s]execute istio install", i.MeshID)

	// 安装主集群中的istio
	for _, cluster := range i.PrimaryClusters {
		if err := i.installIstioForPrimary(ctx, i.ChartVersion, cluster); err != nil {
			blog.Errorf("[%s]install istio for primary cluster %s failed, err: %s", i.MeshID, cluster, err)
			// 安装失败，将所有从集群状态设置为失败
			i.setAllRemoteClustersStatus(common.RemoteClusterStatusInstallFailed)
			return fmt.Errorf("install istio for primary cluster %s failed: %s", cluster, err)
		}
	}

	// 适配主从多集群模式
	if i.MultiClusterEnabled {
		// 为主集群安装eastwestgateway，当前默认一主多从模式
		primaryCluster := i.PrimaryClusters[0]
		if err := i.installEgressGateway(ctx, primaryCluster); err != nil {
			blog.Errorf("[%s]install egress gateway for primary cluster %s failed, err: %s", i.MeshID, primaryCluster, err)
			// 网关部署失败，将所有从集群状态设置为失败
			i.setAllRemoteClustersStatus(common.RemoteClusterStatusInstallFailed)
			return fmt.Errorf("install egress gateway for primary cluster %s failed: %s", primaryCluster, err)
		}

		// 从eastwestgateway的service中获取内网clb地址
		clbIP, err := k8s.GetCLBIP(ctx, primaryCluster, common.EastWestGatewayServiceName)
		if err != nil {
			blog.Errorf("[%s]get clb id failed for primary cluster %s, err: %s", i.MeshID, primaryCluster, err)
			// 获取CLB失败，将所有从集群状态设置为失败
			i.setAllRemoteClustersStatus(common.RemoteClusterStatusInstallFailed)
			return fmt.Errorf("get clb id failed for primary cluster %s: %s", primaryCluster, err)
		}
		i.CLBIP = clbIP

		// 部署从集群
		if err := i.deployRemoteClusters(ctx); err != nil {
			blog.Errorf("[%s]deploy remote clusters failed, err: %s", i.MeshID, err)
			return err
		}
	}

	// 部署可观测性资源
	if err := i.deployObservabilityResources(ctx); err != nil {
		return err
	}

	blog.Infof("[%s]istio install completed", i.MeshID)
	return nil
}

// deployObservabilityResources 部署可观测性资源
func (i *IstioInstallAction) deployObservabilityResources(ctx context.Context) error {
	// 部署链路追踪资源
	if err := i.deployTelemetry(ctx); err != nil {
		return err
	}

	// 部署监控资源
	if err := i.deployMonitoringResources(ctx); err != nil {
		return err
	}

	return nil
}

// deployTelemetry 部署链路追踪资源
func (i *IstioInstallAction) deployTelemetry(ctx context.Context) error {
	if i.ObservabilityConfig == nil || i.ObservabilityConfig.TracingConfig == nil ||
		!i.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
		return nil
	}

	// 下发Telemetry 资源
	traceSamplingPercent := 1
	if i.ObservabilityConfig.TracingConfig.TraceSamplingPercent != nil {
		traceSamplingPercent = int(i.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue())
	}

	if err := k8s.DeployTelemetry(ctx, i.PrimaryClusters, traceSamplingPercent); err != nil {
		blog.Errorf("[%s]deploy Telemetry failed for primary clusters, err: %s", i.MeshID, err)
		return fmt.Errorf("deploy Telemetry failed: %s", err)
	}

	blog.Infof("[%s]deploy Telemetry success for primary clusters", i.MeshID)
	return nil
}

// deployMonitoringResources 部署监控资源
func (i *IstioInstallAction) deployMonitoringResources(ctx context.Context) error {
	if i.ObservabilityConfig == nil || i.ObservabilityConfig.MetricsConfig == nil ||
		!i.ObservabilityConfig.MetricsConfig.MetricsEnabled.GetValue() {
		return nil
	}

	// 开启指标采集时触发流水线执行，流水线执行失败则记录日志，并继续向下执行
	err := utils.ExecutePipeline(ctx)
	if err != nil {
		blog.Errorf("[%s]execute pipeline failed, err: %s", i.MeshID, err)
		// PASS
	}

	// 获取所有需要部署监控资源的集群
	clusters := i.getAllClusters()

	// 部署控制面监控资源
	if err := i.deployControlPlaneMonitoring(ctx, clusters); err != nil {
		return err
	}

	// 部署数据面监控资源
	if err := i.deployDataPlaneMonitoring(ctx, clusters); err != nil {
		return err
	}

	return nil
}

// deployControlPlaneMonitoring 部署控制面监控资源
func (i *IstioInstallAction) deployControlPlaneMonitoring(ctx context.Context, clusters []string) error {
	if !i.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled.GetValue() {
		return nil
	}

	// 开启控制面监控，下发ServiceMonitor
	if err := k8s.DeployServiceMonitor(ctx, clusters); err != nil {
		blog.Errorf("[%s]deploy ServiceMonitor failed for clusters, err: %s", i.MeshID, err)
		return fmt.Errorf("deploy ServiceMonitor failed: %s", err)
	}

	blog.Infof("[%s]deploy ServiceMonitor success for clusters", i.MeshID)
	return nil
}

// deployDataPlaneMonitoring 部署数据面监控资源
func (i *IstioInstallAction) deployDataPlaneMonitoring(ctx context.Context, clusters []string) error {
	if !i.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled.GetValue() {
		return nil
	}

	// 开启数据面监控，下发PodMonitor
	if err := k8s.DeployPodMonitor(ctx, clusters); err != nil {
		blog.Errorf("[%s]deploy PodMonitor failed for clusters, err: %s", i.MeshID, err)
		return fmt.Errorf("deploy PodMonitor failed: %s", err)
	}

	blog.Infof("[%s]deploy PodMonitor success for clusters", i.MeshID)
	return nil
}

// getAllClusters 获取所有需要部署监控资源的集群
func (i *IstioInstallAction) getAllClusters() []string {
	clusters := make([]string, 0, len(i.PrimaryClusters)+len(i.RemoteClusters))
	clusters = append(clusters, i.PrimaryClusters...)
	for _, cluster := range i.RemoteClusters {
		clusters = append(clusters, cluster.ClusterID)
	}
	return clusters
}

// Done 完成回调
func (i *IstioInstallAction) Done(err error) {
	m := make(entity.M)
	if err != nil {
		blog.Errorf("[%s]istio install failed, err: %s", i.MeshID, err)
		m[entity.FieldKeyStatus] = common.IstioStatusInstallFailed
		m[entity.FieldKeyStatusMessage] = fmt.Sprintf("安装失败，%s", err.Error())
	} else {
		blog.Infof("[%s]istio install success", i.MeshID)
		m[entity.FieldKeyStatus] = common.IstioStatusRunning
	}
	// 转换proto RemoteCluster 为 entity RemoteCluster
	remoteClusters := make([]*entity.RemoteCluster, 0, len(i.RemoteClusters))
	for _, cluster := range i.RemoteClusters {
		entityCluster := &entity.RemoteCluster{
			ClusterID: cluster.ClusterID,
			JoinTime:  cluster.JoinTime,
			Status:    cluster.Status,
		}
		remoteClusters = append(remoteClusters, entityCluster)
	}
	// 更新从集群状态
	m["remoteClusters"] = remoteClusters

	updateErr := i.model.Update(context.TODO(), i.MeshID, m)
	if updateErr != nil {
		blog.Errorf("[%s]update mesh status failed, err: %s", i.MeshID, updateErr)
	}
}

// installIstioForPrimary 为主集群安装istio
func (i *IstioInstallAction) installIstioForPrimary(ctx context.Context, chartVersion, clusterID string) error {
	// 创建istio命名空间
	if err := k8s.CreateIstioNamespace(ctx, clusterID); err != nil {
		blog.Errorf("[%s]create istio namespace failed for cluster %s, err: %s", i.MeshID, clusterID, err)
		return fmt.Errorf("create istio namespace failed for cluster %s: %s", clusterID, err)
	}
	opt := &helm.InstallComponentOption{
		ChartVersion:  chartVersion,
		ClusterID:     clusterID,
		ComponentName: common.IstioInstallBaseName,
		ChartName:     common.ComponentIstioBase,
		ProjectCode:   i.ProjectCode,
		MeshID:        i.MeshID,
		NetworkID:     i.NetworkID,
		ChartRepo:     i.ChartRepo,
	}
	// 安装istio base
	if err := helm.InstallComponent(
		ctx,
		opt,
		func() (string, error) {
			return utils.GenBaseValues(&utils.GenBaseValuesOption{
				InstallModel:    common.IstioInstallModePrimary,
				ChartValuesPath: i.ChartValuesPath,
				ChartVersion:    i.ChartVersion,
				Revision:        i.Revision,
			})
		},
	); err != nil {
		return fmt.Errorf("install istio base failed for cluster %s: %s", clusterID, err)
	}

	// 安装istiod
	opt.ComponentName = common.IstioInstallIstiodName
	opt.ChartName = common.ComponentIstiod
	if err := helm.InstallComponent(
		ctx,
		opt,
		func() (string, error) {
			return utils.GenIstiodValues(&utils.GenIstiodValuesOption{
				InstallModel:          common.IstioInstallModePrimary,
				ClusterID:             clusterID,
				NetworkID:             i.NetworkID,
				CLBIP:                 i.CLBIP,
				PrimaryClusters:       i.PrimaryClusters,
				MeshID:                i.MeshID,
				ChartVersion:          i.ChartVersion,
				ChartValuesPath:       i.ChartValuesPath,
				Version:               i.Version,
				ObservabilityConfig:   i.ObservabilityConfig,
				HighAvailability:      i.HighAvailability,
				FeatureConfigs:        i.FeatureConfigs,
				SidecarResourceConfig: i.SidecarResourceConfig,
				Revision:              i.Revision,
			})
		},
	); err != nil {
		return fmt.Errorf("install istiod failed: %s", err)
	}

	return nil
}

// installIstioForRemote 为从集群安装istio
func (i *IstioInstallAction) installIstioForRemote(
	ctx context.Context,
	chartVersion,
	primaryClusterID,
	remoteClusterID string,
) error {
	// 创建istio命名空间
	if err := k8s.CreateIstioNamespace(ctx, remoteClusterID); err != nil {
		blog.Errorf("[%s]create istio namespace failed for cluster %s, err: %s", i.MeshID, remoteClusterID, err)
		return fmt.Errorf("create istio namespace failed for cluster %s: %s", remoteClusterID, err)
	}

	// 为从集群的istio-system命名空间添加主控制面集群注解
	primaryClusterName := strings.ToLower(primaryClusterID)
	if err := k8s.AnnotateNamespace(ctx, remoteClusterID, common.IstioNamespace, map[string]string{
		"topology.istio.io/controlPlaneClusters": primaryClusterName,
	}); err != nil {
		blog.Errorf("[%s]annotate namespace %s failed for cluster %s, err: %s",
			i.MeshID, common.IstioNamespace, remoteClusterID, err)
		return fmt.Errorf("annotate namespace failed for cluster %s: %s", remoteClusterID, err)
	}

	// 安装istio base
	opt := &helm.InstallComponentOption{
		ChartVersion:  chartVersion,
		ClusterID:     remoteClusterID,
		ComponentName: common.IstioInstallBaseName,
		ChartName:     common.ComponentIstioBase,
		ProjectCode:   i.ProjectCode,
		MeshID:        i.MeshID,
		NetworkID:     i.NetworkID,
		ChartRepo:     i.ChartRepo,
	}
	if err := helm.InstallComponent(
		ctx,
		opt,
		func() (string, error) {
			return utils.GenBaseValues(&utils.GenBaseValuesOption{
				InstallModel:    common.IstioInstallModeRemote,
				ChartValuesPath: i.ChartValuesPath,
				ChartVersion:    i.ChartVersion,
				Revision:        i.Revision,
			})
		},
	); err != nil {
		return fmt.Errorf("install istio base failed for cluster %s: %s", remoteClusterID, err)
	}

	// 安装istiod
	opt.ComponentName = common.IstioInstallIstiodName
	opt.ChartName = common.ComponentIstiod
	if err := helm.InstallComponent(
		ctx,
		opt,
		func() (string, error) {
			return utils.GenIstiodValues(&utils.GenIstiodValuesOption{
				InstallModel:    common.IstioInstallModeRemote,
				ClusterID:       remoteClusterID,
				NetworkID:       i.NetworkID,
				CLBIP:           i.CLBIP,
				PrimaryClusters: i.PrimaryClusters,
				MeshID:          i.MeshID,
				Revision:        i.Revision,
			})
		},
	); err != nil {
		return fmt.Errorf("install istiod failed for cluster %s: %s", remoteClusterID, err)
	}

	return nil
}

// installEgressGateway 安装eastwestgateway
func (i *IstioInstallAction) installEgressGateway(ctx context.Context, clusterID string) error {
	// 创建 istio-system 命名空间,如果已经存在则忽略
	if err := k8s.CreateIstioNamespace(ctx, clusterID); err != nil {
		blog.Errorf("[%s]create istio namespace failed for cluster %s, err: %s", i.MeshID, clusterID, err)
		return fmt.Errorf("create istio namespace failed for cluster %s: %s", clusterID, err)
	}
	// 安装eastwestgateway
	if err := helm.InstallComponent(
		ctx,
		&helm.InstallComponentOption{
			ChartVersion:  i.ChartVersion,
			ClusterID:     clusterID,
			ComponentName: common.IstioInstallIstioGatewayName,
			ChartName:     common.ComponentIstioGateway,
			ProjectCode:   i.ProjectCode,
			MeshID:        i.MeshID,
			NetworkID:     i.NetworkID,
			ChartRepo:     i.ChartRepo,
		},
		func() (string, error) {
			return utils.GenEgressGatewayValues(&utils.GenEgressGatewayValuesOption{
				ChartValuesPath: i.ChartValuesPath,
				ChartVersion:    i.ChartVersion,
				CLBID:           i.CLBID,
				NetworkID:       i.NetworkID,
				Revision:        i.Revision,
			})
		},
	); err != nil {
		return fmt.Errorf("install egress gateway failed: %s", err)
	}

	if err := k8s.DeployResourceByYAML(ctx, clusterID, common.GetGatewayYAML(),
		common.GatewayKind, common.GatewayName); err != nil {
		blog.Errorf("[%s]deploy gateway failed for cluster %s, err: %s", i.MeshID, clusterID, err)
		return fmt.Errorf("deploy gateway failed for cluster %s: %s", clusterID, err)
	}
	if err := k8s.DeployResourceByYAML(ctx, clusterID,
		common.GetVirtualServiceYAML(i.Revision),
		common.VirtualServiceKind, common.VirtualServiceName); err != nil {
		blog.Errorf("[%s]deploy virtual service failed for cluster %s, err: %s", i.MeshID, clusterID, err)
		return fmt.Errorf("deploy virtual service failed for cluster %s: %s", clusterID, err)
	}
	return nil
}

// deployRemoteClusters 部署从集群
func (i *IstioInstallAction) deployRemoteClusters(ctx context.Context) error {
	// 并发安装从集群istio
	var (
		wg sync.WaitGroup
		mu sync.Mutex
		// 记录从集群网格安装结果
		successResults = make(map[string]bool)
	)

	for _, cluster := range i.RemoteClusters {
		clusterID := cluster.ClusterID
		wg.Add(1)
		go func(clusterID string) {
			defer wg.Done()

			if err := i.installIstioForRemote(ctx, i.ChartVersion, i.PrimaryClusters[0], clusterID); err != nil {
				blog.Errorf("[%s]install istio for remote cluster %s failed, err: %s", i.MeshID, clusterID, err)
				mu.Lock()
				successResults[clusterID] = false
				mu.Unlock()
				return
			}

			mu.Lock()
			successResults[clusterID] = true
			mu.Unlock()
		}(clusterID)
	}

	// 等待所有从集群安装完成
	wg.Wait()

	// 根据安装结果更新从集群状态
	for _, cluster := range i.RemoteClusters {
		if success, exists := successResults[cluster.ClusterID]; exists {
			if success {
				cluster.Status = common.RemoteClusterStatusRunning
			} else {
				cluster.Status = common.RemoteClusterStatusInstallFailed
			}
		}
	}

	return nil
}

// setAllRemoteClustersStatus 设置所有从集群的状态
func (i *IstioInstallAction) setAllRemoteClustersStatus(status string) {
	for _, cluster := range i.RemoteClusters {
		cluster.Status = status
	}
}
