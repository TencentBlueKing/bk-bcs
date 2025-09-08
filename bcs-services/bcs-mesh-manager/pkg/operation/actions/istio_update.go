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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	opcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation/actions/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// IstioUpdateOption istio更新操作选项
type IstioUpdateOption struct {
	Model               store.MeshManagerModel
	ProjectCode         *string
	MeshID              *string
	NetworkID           *string
	ChartName           string
	ChartVersion        *string
	ChartValuesPath     *string
	ChartRepo           *string
	PrimaryClusters     []string
	OldRemoteClusters   []*entity.RemoteCluster
	NewRemoteClusters   []*entity.RemoteCluster
	MultiClusterEnabled *bool
	UpdateValues        *common.IstiodInstallValues
	ObservabilityConfig *meshmanager.ObservabilityConfig
	UpdateValuesOptions *utils.UpdateValuesOptions
	CLBID               *string
	Revision            *string
	OldReleaseNames     map[string]map[string]string
}

// IstioUpdateAction istio更新操作
type IstioUpdateAction struct {
	*IstioUpdateOption
}

var _ operation.Operation = &IstioUpdateAction{}

// NewIstioUpdateAction 创建istio更新操作
func NewIstioUpdateAction(opt *IstioUpdateOption) *IstioUpdateAction {
	return &IstioUpdateAction{
		IstioUpdateOption: opt,
	}
}

// Action 操作名称
func (i *IstioUpdateAction) Action() string {
	return "istio-update"
}

// Name 操作实例名称
func (i *IstioUpdateAction) Name() string {
	return fmt.Sprintf("istio-update-%s", *i.MeshID)
}

// Validate 验证参数
func (i *IstioUpdateAction) Validate() error {
	// 必填字段
	if i.ProjectCode == nil {
		return fmt.Errorf("projectCode is required")
	}
	if i.MeshID == nil {
		return fmt.Errorf("meshID is required")
	}
	if i.NetworkID == nil {
		return fmt.Errorf("networkID is required")
	}
	if i.ChartVersion == nil {
		return fmt.Errorf("chartVersion is required")
	}
	if i.ChartValuesPath == nil {
		return fmt.Errorf("chartValuesPath is required")
	}
	if i.ChartRepo == nil {
		return fmt.Errorf("chartRepo is required")
	}
	if i.MultiClusterEnabled != nil && *i.MultiClusterEnabled {
		if i.CLBID == nil {
			return fmt.Errorf("clbID is required")
		}
	}
	if i.Revision == nil {
		return fmt.Errorf("revision is required")
	}
	return nil
}

// Prepare 准备阶段
func (i *IstioUpdateAction) Prepare(ctx context.Context) error {
	// 暂时无需预处理
	return nil
}

// Execute 执行更新
func (i *IstioUpdateAction) Execute(ctx context.Context) error {
	// 更新主集群
	for _, cluster := range i.PrimaryClusters {
		if err := i.updatePrimaryCluster(ctx, cluster); err != nil {
			blog.Errorf("[%s]update primary cluster istio failed, clusterID: %s, err: %s", i.MeshID, cluster, err)
			return err
		}
	}

	// 更新东西向网关
	if err := i.updateEastWestGateway(ctx); err != nil {
		blog.Errorf("[%s]update istio gateway failed, err: %s", i.MeshID, err)
		return err
	}

	// 更新从集群
	if err := i.updateRemoteCluster(ctx); err != nil {
		blog.Errorf("[%s]update remote cluster istio failed, err: %s", i.MeshID, err)
		return err
	}

	// 更新链路追踪资源
	if err := i.updateTelemetry(ctx); err != nil {
		blog.Errorf("[%s]update telemetry failed, err: %s", i.MeshID, err)
		return err
	}

	// 更新监控资源
	remoteClusterIds := make([]string, 0, len(i.NewRemoteClusters))
	for _, cluster := range i.NewRemoteClusters {
		remoteClusterIds = append(remoteClusterIds, cluster.ClusterID)
	}
	clusters := utils.MergeSlices(i.PrimaryClusters, remoteClusterIds)
	if err := i.updateObservability(ctx, clusters); err != nil {
		blog.Errorf("[%s]update observability failed, err: %s", i.MeshID, err)
		return err
	}

	return nil
}

// updateTelemetry 更新链路追踪资源
func (i *IstioUpdateAction) updateTelemetry(ctx context.Context) error {
	if i.ObservabilityConfig == nil || i.ObservabilityConfig.TracingConfig == nil {
		return nil
	}

	// 更新链路追踪资源
	if i.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
		traceSamplingPercent := 1
		if i.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() != 0 {
			traceSamplingPercent = int(i.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue())
		}
		if err := k8s.DeployTelemetry(ctx, i.PrimaryClusters, traceSamplingPercent); err != nil {
			blog.Errorf("[%s]deploy Telemetry failed for clusters, err: %s", i.MeshID, err)
			return err
		}
	} else {
		if err := k8s.DeleteTelemetry(ctx, i.PrimaryClusters); err != nil {
			blog.Errorf("[%s]delete Telemetry failed for clusters, err: %s", i.MeshID, err)
			return err
		}
	}
	return nil
}

// updateObservability 更新监控资源
func (i *IstioUpdateAction) updateObservability(ctx context.Context, clusters []string) error {
	if i.ObservabilityConfig == nil || i.ObservabilityConfig.MetricsConfig == nil {
		return nil
	}

	// 更新监控资源
	if i.ObservabilityConfig.MetricsConfig.MetricsEnabled.GetValue() {
		if i.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled.GetValue() {
			// 启用控制面监控，部署 ServiceMonitor
			if err := k8s.DeployServiceMonitor(ctx, clusters); err != nil {
				blog.Errorf("[%s]deploy ServiceMonitor failed for clusters, err: %s", i.MeshID, err)
				return err
			}
		} else {
			// 禁用控制面监控，删除 ServiceMonitor
			if err := k8s.DeleteServiceMonitor(ctx, clusters); err != nil {
				blog.Errorf("[%s]delete ServiceMonitor failed for clusters, err: %s", i.MeshID, err)
				return err
			}
		}

		if i.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled.GetValue() {
			// 启用数据面监控，部署 PodMonitor
			if err := k8s.DeployPodMonitor(ctx, clusters); err != nil {
				blog.Errorf("[%s]deploy PodMonitor failed for clusters, err: %s", i.MeshID, err)
				return err
			}
		} else {
			// 禁用数据面监控，删除 PodMonitor
			if err := k8s.DeletePodMonitor(ctx, clusters); err != nil {
				blog.Errorf("[%s]delete PodMonitor failed for clusters, err: %s", i.MeshID, err)
				return err
			}
		}
	}
	return nil
}

// uninstallEastWestGateway 卸载东西向网关
func (i *IstioUpdateAction) uninstallEastWestGateway(ctx context.Context, releaseName string) error {

	if err := helm.UninstallIstioComponent(
		ctx,
		i.PrimaryClusters[0],
		releaseName,
		*i.ProjectCode,
		*i.MeshID,
	); err != nil {
		blog.Errorf("[%s]uninstall istio gateway failed, err: %s", *i.MeshID, err)
		return fmt.Errorf("uninstall istio gateway failed: %s", err)
	}

	return nil
}

// updateEastWestGateway 安装东西向网关
func (i *IstioUpdateAction) updateEastWestGateway(ctx context.Context) error {
	// 获取东西向网关的release name。若获取不到表示未安装东西向网关
	releaseName, err := opcommon.GetReleaseName(
		i.OldReleaseNames, i.PrimaryClusters[0], common.ComponentIstioGateway, *i.MeshID,
	)
	if err != nil {
		return err
	}
	// 关闭多集群模式且东西向网关已部署，则卸载东西向网关
	if i.MultiClusterEnabled != nil && !*i.MultiClusterEnabled {
		if releaseName != "" {
			return i.uninstallEastWestGateway(ctx, releaseName)
		}
		return nil
	}

	// 开启多集群模式且东西向网关的release name不存在，则表示未部署
	if releaseName == "" {
		return i.installEastWestGateway(ctx)
	}

	// 若已安装则更新东西向网关
	return i.upgradeEastWestGateway(ctx, releaseName)
}

// installNewGateway 安装新的网关
func (i *IstioUpdateAction) installEastWestGateway(ctx context.Context) error {
	// 创建 istio-system 命名空间,如果已经存在则忽略
	if err := k8s.CreateIstioNamespace(ctx, i.PrimaryClusters[0]); err != nil {
		blog.Errorf("[%s]create istio namespace failed for cluster %s, err: %s", *i.MeshID, i.PrimaryClusters[0], err)
		return fmt.Errorf("create istio namespace failed for cluster %s: %s", i.PrimaryClusters[0], err)
	}

	// 安装东西向网关组件
	if err := i.installGatewayComponent(ctx); err != nil {
		return err
	}

	// 部署网关相关资源
	if err := i.deployGatewayResources(ctx); err != nil {
		return err
	}

	blog.Infof("[%s]istio gateway installed successfully", *i.MeshID)
	return nil
}

// installGatewayComponent 安装网关组件
func (i *IstioUpdateAction) installGatewayComponent(ctx context.Context) error {
	if err := helm.InstallComponent(
		ctx,
		&helm.InstallComponentOption{
			ChartVersion:  *i.ChartVersion,
			ClusterID:     i.PrimaryClusters[0],
			ComponentName: common.IstioInstallIstioGatewayName,
			ChartName:     common.ComponentIstioGateway,
			ProjectCode:   *i.ProjectCode,
			MeshID:        *i.MeshID,
			NetworkID:     *i.NetworkID,
			ChartRepo:     *i.ChartRepo,
		},
		func() (string, error) {
			return utils.GenEgressGatewayValues(&utils.GenEgressGatewayValuesOption{
				ChartValuesPath: *i.ChartValuesPath,
				ChartVersion:    *i.ChartVersion,
				CLBID:           *i.CLBID,
				NetworkID:       *i.NetworkID,
				Revision:        *i.Revision,
			})
		},
	); err != nil {
		blog.Errorf("[%s]install istio gateway failed, err: %s", *i.MeshID, err)
		return fmt.Errorf("install istio gateway failed: %s", err)
	}
	blog.Infof("[%s]istio gateway installed successfully", *i.MeshID)
	return nil
}

// deployGatewayResources 部署网关相关资源
func (i *IstioUpdateAction) deployGatewayResources(ctx context.Context) error {
	// 部署Gateway资源
	if err := k8s.DeployResourceByYAML(ctx, i.PrimaryClusters[0], common.GetGatewayYAML(),
		common.GatewayKind, common.GatewayName); err != nil {
		blog.Errorf("[%s]deploy gateway failed for primary cluster %s, err: %s", i.MeshID, i.PrimaryClusters[0], err)
		return fmt.Errorf("deploy gateway failed for primary cluster %s: %s", i.PrimaryClusters[0], err)
	}

	// 部署VirtualService资源
	if err := k8s.DeployResourceByYAML(ctx, i.PrimaryClusters[0],
		common.GetVirtualServiceYAML(*i.Revision),
		common.VirtualServiceKind, common.VirtualServiceName); err != nil {
		blog.Errorf("[%s]deploy virtual service failed for primary cluster %s, err: %s",
			*i.MeshID, i.PrimaryClusters[0], err)
		return fmt.Errorf("deploy virtual service failed for primary cluster %s: %s",
			i.PrimaryClusters[0], err)
	}
	return nil
}

// upgradeExistingGateway 升级已存在的网关
func (i *IstioUpdateAction) upgradeEastWestGateway(ctx context.Context, releaseName string) error {
	// 生成网关配置值
	values, err := utils.GenEgressGatewayValues(&utils.GenEgressGatewayValuesOption{
		ChartValuesPath: *i.ChartValuesPath,
		ChartVersion:    *i.ChartVersion,
		CLBID:           *i.CLBID,
		NetworkID:       *i.NetworkID,
		Revision:        *i.Revision,
	})
	if err != nil {
		blog.Errorf("[%s]gen egress gateway values failed, err: %s", *i.MeshID, err)
		return fmt.Errorf("gen egress gateway values failed: %s", err)
	}

	// 执行升级
	_, err = helm.Upgrade(
		ctx,
		&helmmanager.UpgradeReleaseV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &i.PrimaryClusters[0],
			Chart:       pointer.String(common.ComponentIstioGateway),
			Repository:  i.ChartRepo,
			Version:     i.ChartVersion,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        &releaseName,
			Values:      []string{values},
		},
	)
	if err != nil {
		blog.Errorf("[%s]upgrade istio gateway failed, err: %s", *i.MeshID, err)
		return fmt.Errorf("upgrade istio gateway failed: %s", err)
	}

	blog.Infof("[%s]istio gateway upgraded successfully", *i.MeshID)
	return nil
}

// updatePrimaryCluster 更新主集群istio
func (i *IstioUpdateAction) updatePrimaryCluster(ctx context.Context, clusterID string) error {
	// 获取Release名称
	istiodReleaseName, err := opcommon.GetReleaseName(i.OldReleaseNames, clusterID, common.ComponentIstiod, *i.MeshID)
	if err != nil {
		return err
	}

	// 获取istiod的values.yaml配置信息
	releaseDetail, err := helm.GetReleaseDetail(
		ctx,
		&helmmanager.GetReleaseDetailV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &clusterID,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        &istiodReleaseName,
		},
	)
	if err != nil || releaseDetail == nil {
		blog.Errorf("[%s]get release detail failed, clusterID: %s", *i.MeshID, clusterID)
		return fmt.Errorf("get release detail failed, clusterID: %s", clusterID)
	}

	if len(releaseDetail.Data.Values) == 0 {
		blog.Errorf("[%s]release values is empty, clusterID: %s", *i.MeshID, clusterID)
		return fmt.Errorf("release values is empty, clusterID: %s", clusterID)
	}
	values := releaseDetail.Data.Values[0]
	var customValues string
	customValuesBytes, err := yaml.Marshal(i.UpdateValues)
	if err != nil {
		blog.Errorf("[%s]marshal install values failed, err: %s", *i.MeshID, err)
		return err
	}
	customValues = string(customValuesBytes)

	// utils.MergeValues 的合并以values为基准
	// values中存在的字段无法被customValues覆盖
	// 合并前需要先处理values中本次需要移除的字段
	newValues, err := utils.ProcessValues(values, i.UpdateValuesOptions)
	if err != nil {
		blog.Errorf("[%s]process field key failed, clusterID: %s, err: %s", *i.MeshID, clusterID, err)
		return err
	}

	mergedValues, err := utils.MergeValues(newValues, customValues)
	if err != nil {
		blog.Errorf("[%s]merge values failed, clusterID: %s, err: %s", *i.MeshID, clusterID, err)
		return err
	}

	_, err = helm.Upgrade(
		ctx,
		&helmmanager.UpgradeReleaseV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &clusterID,
			Chart:       &i.ChartName,
			Repository:  i.ChartRepo,
			Version:     i.ChartVersion,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        &istiodReleaseName,
			Values:      []string{mergedValues},
		},
	)
	if err != nil {
		blog.Errorf("[%s]upgrade istiod failed, clusterID: %s, err: %s", *i.MeshID, clusterID, err)
		return err
	}

	return nil
}

// updateRemoteCluster 更新从集群istio
func (i *IstioUpdateAction) updateRemoteCluster(ctx context.Context) error {
	// 处理需要移除的从集群
	if err := i.handleRemovedRemoteClusters(ctx); err != nil {
		return err
	}

	// 处理新增的从集群
	if err := i.handleNewRemoteClusters(ctx); err != nil {
		return err
	}

	// 更新新的从集群的监控资源
	newRemoteClusterIds := i.getNewRemoteClusterIds()
	if err := i.updateObservability(ctx, newRemoteClusterIds); err != nil {
		blog.Errorf("[%s]update observability failed for clusters, err: %s", i.MeshID, err)
		return fmt.Errorf("update observability failed for clusters: %s", err)
	}

	return nil
}

// handleRemovedRemoteClusters 处理需要移除的从集群
func (i *IstioUpdateAction) handleRemovedRemoteClusters(ctx context.Context) error {
	// 获取需要移除的集群列表
	diffClusters := i.getRemovedClusterIds()
	if len(diffClusters) == 0 {
		return nil
	}

	// 移除从集群istio
	if err := i.uninstallOldIstio(ctx, diffClusters); err != nil {
		blog.Errorf("[%s]uninstall istio failed for clusters, err: %s", i.MeshID, err)
		return fmt.Errorf("uninstall istio failed for clusters: %s", err)
	}

	// 删除需要移除的集群的监控资源
	if err := i.cleanupMonitoringResources(ctx, diffClusters); err != nil {
		return err
	}

	return nil
}

// handleNewRemoteClusters 处理新增的从集群
func (i *IstioUpdateAction) handleNewRemoteClusters(ctx context.Context) error {
	// 获取状态值为installing的从集群列表
	installingClusters := i.getInstallingRemoteClusters()
	if len(installingClusters) == 0 {
		return nil
	}

	// 获取CLB IP地址，如果失败则设置所有新集群为安装失败
	clbIP, err := i.getCLBIPForNewClusters(ctx)
	if err != nil {
		blog.Errorf("[%s]get CLB IP failed for new remote clusters, err: %s", i.MeshID, err)
		// 获取CLB IP失败，将所有新集群状态设置为安装失败
		i.setNewRemoteClustersStatus(installingClusters, common.RemoteClusterStatusInstallFailed)
		return fmt.Errorf("get CLB IP failed for new remote clusters: %s", err)
	}

	// 并发安装新的从集群
	var (
		wg             sync.WaitGroup
		mu             sync.Mutex
		successResults = make(map[string]bool)
	)

	for _, cluster := range installingClusters {
		clusterID := cluster.ClusterID
		wg.Add(1)
		go func(clusterID string) {
			defer wg.Done()

			if err := i.installNewIstio(ctx, clusterID, clbIP); err != nil {
				blog.Errorf("[%s]install istio failed for cluster %s, err: %s", i.MeshID, clusterID, err)
				mu.Lock()
				successResults[clusterID] = false
				mu.Unlock()
			} else {
				mu.Lock()
				successResults[clusterID] = true
				mu.Unlock()
			}
		}(clusterID)
	}

	// 等待所有安装完成
	wg.Wait()

	// 根据安装结果更新从集群状态
	i.updateRemoteClusterStatuses(installingClusters, successResults)

	return nil
}

// getRemovedClusterIds 获取需要移除的集群ID列表
func (i *IstioUpdateAction) getRemovedClusterIds() []string {
	oldRemoteClusterIds := make([]string, 0, len(i.OldRemoteClusters))
	for _, cluster := range i.OldRemoteClusters {
		oldRemoteClusterIds = append(oldRemoteClusterIds, cluster.ClusterID)
	}
	newRemoteClusterIds := make([]string, 0, len(i.NewRemoteClusters))
	for _, cluster := range i.NewRemoteClusters {
		newRemoteClusterIds = append(newRemoteClusterIds, cluster.ClusterID)
	}

	// 获取在旧的从集群列表中但不在新的从集群列表中的集群，即需要移除的集群
	return utils.Difference(oldRemoteClusterIds, newRemoteClusterIds)
}

// getNewRemoteClusterIds 获取新的从集群ID列表
func (i *IstioUpdateAction) getNewRemoteClusterIds() []string {
	newRemoteClusterIds := make([]string, 0, len(i.NewRemoteClusters))
	for _, cluster := range i.NewRemoteClusters {
		newRemoteClusterIds = append(newRemoteClusterIds, cluster.ClusterID)
	}
	return newRemoteClusterIds
}

// getInstallingRemoteClusters 获取状态为Installing的从集群列表
func (i *IstioUpdateAction) getInstallingRemoteClusters() []*entity.RemoteCluster {
	installingClusters := make([]*entity.RemoteCluster, 0)
	for _, cluster := range i.NewRemoteClusters {
		if cluster.Status == common.RemoteClusterStatusInstalling {
			installingClusters = append(installingClusters, cluster)
		}
	}
	return installingClusters
}

// cleanupMonitoringResources 清理监控资源
func (i *IstioUpdateAction) cleanupMonitoringResources(ctx context.Context, clusterIds []string) error {
	// 删除需要移除的集群的PodMonitor
	if err := k8s.DeletePodMonitor(ctx, clusterIds); err != nil {
		blog.Errorf("[%s]delete PodMonitor failed for clusters %v, err: %s", i.MeshID, clusterIds, err)
		return fmt.Errorf("delete PodMonitor failed for clusters %v: %s", clusterIds, err)
	}

	// 删除需要移除的集群的ServiceMonitor
	if err := k8s.DeleteServiceMonitor(ctx, clusterIds); err != nil {
		blog.Errorf("[%s]delete ServiceMonitor failed for clusters %v, err: %s", i.MeshID, clusterIds, err)
		return fmt.Errorf("delete ServiceMonitor failed for clusters %v: %s", clusterIds, err)
	}

	return nil
}

// updateRemoteClusterStatuses 更新从集群状态
func (i *IstioUpdateAction) updateRemoteClusterStatuses(
	clusters []*entity.RemoteCluster,
	successResults map[string]bool,
) {
	// 创建clusterID的集合，用于快速查找
	clusterIDSet := make(map[string]struct{})
	for _, cluster := range clusters {
		clusterIDSet[cluster.ClusterID] = struct{}{}
	}

	// 更新i.NewRemoteClusters中对应集群的状态
	for _, cluster := range i.NewRemoteClusters {
		if _, exists := clusterIDSet[cluster.ClusterID]; exists {
			if success, exists := successResults[cluster.ClusterID]; exists {
				if success {
					cluster.Status = common.RemoteClusterStatusRunning
				} else {
					cluster.Status = common.RemoteClusterStatusInstallFailed
				}
			}
		}
	}
}

// uninstallOldIstio 卸载旧集群istio
func (i *IstioUpdateAction) uninstallOldIstio(ctx context.Context, clusters []string) error {
	for _, cluster := range clusters {
		// 获取istiod release name
		istiodReleaseName, err := opcommon.GetReleaseName(
			i.OldReleaseNames, cluster, common.ComponentIstiod, *i.MeshID,
		)
		if err != nil {
			return err
		}
		if err = helm.UninstallIstioComponent(
			ctx, cluster, istiodReleaseName, *i.ProjectCode, *i.MeshID,
		); err != nil {
			blog.Errorf("[%s]uninstall istiod failed, clusterID: %s, err: %s", *i.MeshID, cluster, err)
			return err
		}
		// 获取istio base release name
		baseReleaseName, err := opcommon.GetReleaseName(
			i.OldReleaseNames, cluster, common.ComponentIstioBase, *i.MeshID,
		)
		if err != nil {
			return err
		}
		// 移除从集群istio base
		if err := helm.UninstallIstioComponent(
			ctx, cluster, baseReleaseName, *i.ProjectCode, *i.MeshID,
		); err != nil {
			blog.Errorf("[%s]uninstall istio base failed, clusterID: %s, err: %s", *i.MeshID, cluster, err)
			return err
		}
	}
	return nil
}

// installNewIstioWithCLB 安装新集群istio，并传递CLB IP
func (i *IstioUpdateAction) installNewIstio(ctx context.Context, cluster string, clbIP string) error {
	opt := &helm.InstallComponentOption{
		ChartVersion: *i.ChartVersion,
		ProjectCode:  *i.ProjectCode,
		MeshID:       *i.MeshID,
		NetworkID:    *i.NetworkID,
		ChartRepo:    *i.ChartRepo,
	}
	// 创建istio命名空间
	if err := k8s.CreateIstioNamespace(ctx, cluster); err != nil {
		blog.Errorf("[%s]create istio namespace failed for cluster %s, err: %s", *i.MeshID, cluster, err)
		return fmt.Errorf("create istio namespace failed for cluster %s: %s", cluster, err)
	}

	// 为从集群的istio-system命名空间添加主控制面集群注解
	primaryClusterName := strings.ToLower(i.PrimaryClusters[0])
	if err := k8s.AnnotateNamespace(ctx, cluster, common.IstioNamespace, map[string]string{
		"topology.istio.io/controlPlaneClusters": primaryClusterName,
	}); err != nil {
		blog.Errorf("[%s]annotate namespace %s failed for cluster %s, err: %s",
			*i.MeshID, common.IstioNamespace, cluster, err)
		return fmt.Errorf("annotate namespace failed for cluster %s: %s", cluster, err)
	}

	// 新增从集群istio base
	opt.ClusterID = cluster
	opt.ComponentName = common.IstioInstallBaseName
	opt.ChartName = common.ComponentIstioBase
	if err := helm.InstallComponent(ctx, opt, func() (string, error) {
		return utils.GenBaseValues(&utils.GenBaseValuesOption{
			InstallModel:    common.IstioInstallModeRemote,
			ChartValuesPath: *i.ChartValuesPath,
			ChartVersion:    *i.ChartVersion,
			Revision:        *i.Revision,
		})
	}); err != nil {
		blog.Errorf("[%s]install istio base failed, clusterID: %s, err: %s", *i.MeshID, cluster, err)
		return err
	}
	// 新增从集群istiod
	opt.ComponentName = common.IstioInstallIstiodName
	opt.ChartName = common.ComponentIstiod
	if err := helm.InstallComponent(ctx, opt, func() (string, error) {
		return utils.GenIstiodValues(&utils.GenIstiodValuesOption{
			InstallModel:    common.IstioInstallModeRemote,
			ClusterID:       cluster,
			NetworkID:       *i.NetworkID,
			PrimaryClusters: i.PrimaryClusters,
			MeshID:          *i.MeshID,
			CLBIP:           clbIP,
			Revision:        *i.Revision,
		})
	}); err != nil {
		blog.Errorf("[%s]install istiod failed, clusterID: %s, err: %s", *i.MeshID, cluster, err)
		return err
	}
	return nil
}

// Done 完成回调
func (i *IstioUpdateAction) Done(err error) {
	updateFields := entity.M{}
	if err != nil {
		blog.Errorf("[%s]istio update operation failed, err: %s", *i.MeshID, err)
		updateFields[entity.FieldKeyStatus] = common.IstioStatusUpdateFailed
		updateFields[entity.FieldKeyStatusMessage] = fmt.Sprintf("更新失败，%s", err.Error())
	} else {
		updateFields[entity.FieldKeyStatus] = common.IstioStatusRunning
		updateFields[entity.FieldKeyStatusMessage] = ""
	}

	// 更新从集群状态
	updateFields[entity.FieldKeyRemoteClusters] = i.NewRemoteClusters
	// 构建ReleaseNames
	remoteClusterIDs := make([]string, 0, len(i.NewRemoteClusters))
	for _, cluster := range i.NewRemoteClusters {
		remoteClusterIDs = append(remoteClusterIDs, cluster.ClusterID)
	}
	clusterIDs := utils.MergeSlices(i.PrimaryClusters, remoteClusterIDs)
	updateFields[entity.FieldKeyReleaseNames] = i.buildReleaseNames(clusterIDs)

	// 执行更新操作并检查错误
	if updateErr := i.Model.Update(context.TODO(), *i.MeshID, updateFields); updateErr != nil {
		blog.Errorf("[%s]update mesh status failed, err: %s", *i.MeshID, updateErr)
	}
}

// buildReleaseNames 构建ReleaseNames map
func (i *IstioUpdateAction) buildReleaseNames(clusters []string) map[string]map[string]string {
	releaseNames := make(map[string]map[string]string)

	// 构建集群的集合，用于快速查找
	clusterSet := make(map[string]struct{})
	for _, clusterID := range clusters {
		clusterSet[clusterID] = struct{}{}
	}

	// 存在部分服务网格通过手动部署，使用自定义的releaseName，需要保留releaseName
	if i.OldReleaseNames != nil {
		for clusterID, clusterReleases := range i.OldReleaseNames {
			if _, exists := clusterSet[clusterID]; exists {
				releaseNames[clusterID] = make(map[string]string)
				for component, releaseName := range clusterReleases {
					releaseNames[clusterID][component] = releaseName
				}
			}
		}
	}

	// 处理多集群模式关闭时的东西向网关清理
	if i.MultiClusterEnabled != nil && !*i.MultiClusterEnabled {
		for clusterID := range releaseNames {
			delete(releaseNames[clusterID], common.ComponentIstioGateway)
		}
	}

	// 为新集群设置默认Release名称
	for _, clusterID := range clusters {
		if _, exists := releaseNames[clusterID]; !exists {
			releaseNames[clusterID] = map[string]string{
				common.ComponentIstioBase: common.IstioInstallBaseName,
				common.ComponentIstiod:    common.IstioInstallIstiodName,
			}
		}
	}

	// 为多集群模式开启的集群添加东西向网关Release名称
	if i.MultiClusterEnabled != nil && *i.MultiClusterEnabled {
		for _, clusterID := range clusters {
			if _, exists := releaseNames[clusterID]; exists {
				releaseNames[clusterID][common.ComponentIstioGateway] = common.IstioInstallIstioGatewayName
			}
		}
	}

	return releaseNames
}

// getCLBIPForNewClusters 为新集群获取CLB IP地址
func (i *IstioUpdateAction) getCLBIPForNewClusters(ctx context.Context) (string, error) {
	// 从主集群的eastwestgateway service中获取内网CLB地址
	releaseName, err := opcommon.GetReleaseName(
		i.OldReleaseNames, i.PrimaryClusters[0], common.ComponentIstioGateway, *i.MeshID,
	)
	if err != nil {
		return "", fmt.Errorf("get eastwestgateway release name failed: %s", err)
	}
	clbIP, err := k8s.GetCLBIP(ctx, i.PrimaryClusters[0], releaseName)
	if err != nil {
		blog.Errorf("[%s]get CLB IP failed for primary cluster %s, err: %s", i.MeshID, i.PrimaryClusters[0], err)
		return "", fmt.Errorf("get CLB IP failed for primary cluster %s: %s", i.PrimaryClusters[0], err)
	}
	return clbIP, nil
}

// setNewRemoteClustersStatus 设置新从集群的状态
func (i *IstioUpdateAction) setNewRemoteClustersStatus(clusters []*entity.RemoteCluster, status string) {
	// 创建clusterID的集合，用于快速查找
	clusterIDSet := make(map[string]struct{})
	for _, cluster := range clusters {
		clusterIDSet[cluster.ClusterID] = struct{}{}
	}

	// 更新i.NewRemoteClusters中对应集群的状态
	for _, cluster := range i.NewRemoteClusters {
		if _, exists := clusterIDSet[cluster.ClusterID]; exists {
			cluster.Status = status
		}
	}
}
