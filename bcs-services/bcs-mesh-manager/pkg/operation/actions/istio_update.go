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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
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
	ChartName           string
	ChartVersion        *string
	ChartRepo           *string
	PrimaryClusters     []string
	RemoteClusters      []string
	UpdateValues        *common.IstiodInstallValues
	ObservabilityConfig *meshmanager.ObservabilityConfig
	UpdateValuesOptions *utils.UpdateValuesOptions
	Version             string
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
	return nil
}

// Prepare 准备阶段
func (i *IstioUpdateAction) Prepare(ctx context.Context) error {
	// 暂时无需预处理
	return nil
}

// Execute 执行更新
func (i *IstioUpdateAction) Execute(ctx context.Context) error {
	// 更新主集群的istio
	for _, cluster := range i.PrimaryClusters {
		if err := i.updatePrimaryCluster(ctx, cluster); err != nil {
			blog.Errorf("[%s]update primary cluster istio failed, clusterID: %s, err: %s", i.MeshID, cluster, err)
			return err
		}
	}
	// TODO: 更新远程集群的istio

	// 合并主从集群列表
	clusters := utils.MergeSlices(i.PrimaryClusters, i.RemoteClusters)

	// 更新集群依赖资源（PodMonitor, ServiceMonitor, Telemetry）
	for _, cluster := range clusters {
		if err := i.updateClusterResource(ctx, cluster); err != nil {
			blog.Errorf("[%s]update cluster resource failed for cluster %s, err: %s",
				*i.MeshID, cluster, err)
			// 注意：这里不返回错误，继续更新其他资源
			blog.Warnf("[%s]continue updating Istio components despite cluster resource update failure", *i.MeshID)
		}
	}
	return nil
}

func (i *IstioUpdateAction) updatePrimaryCluster(ctx context.Context, clusterID string) error {
	// 获取Release名称
	istiodReleaseName, err := i.Model.GetReleaseName(ctx, *i.MeshID, clusterID, common.ComponentIstiod)
	if err != nil {
		blog.Errorf("[%s]get release name failed, clusterID: %s, err: %s", i.MeshID, clusterID, err)
		return fmt.Errorf("get release name failed: %s", err)
	}

	// 获取istiod的values.yaml配置信息
	releaseDetail, err := helm.GetReleaseDetail(
		ctx,
		&helmmanager.GetReleaseDetailV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &clusterID,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        istiodReleaseName,
		},
	)
	if err != nil || releaseDetail == nil {
		blog.Errorf("[%s]get release detail failed, clusterID: %s", i.MeshID, clusterID)
		return fmt.Errorf("get release detail failed, clusterID: %s", clusterID)
	}

	if len(releaseDetail.Data.Values) == 0 {
		blog.Errorf("[%s]release values is empty, clusterID: %s", i.MeshID, clusterID)
		return fmt.Errorf("release values is empty, clusterID: %s", clusterID)
	}
	values := releaseDetail.Data.Values[0]
	var customValues string
	customValuesBytes, err := yaml.Marshal(i.UpdateValues)
	if err != nil {
		blog.Errorf("[%s]marshal install values failed, err: %s", i.MeshID, err)
		return err
	}
	customValues = string(customValuesBytes)

	// utils.MergeValues 的合并以values为基准
	// values中存在的字段无法被customValues覆盖
	// 合并前需要先处理values中本次需要移除的字段
	newValues, err := utils.ProcessValues(values, i.UpdateValuesOptions)
	if err != nil {
		blog.Errorf("[%s]process field key failed, clusterID: %s, err: %s", i.MeshID, clusterID, err)
		return err
	}

	mergedValues, err := utils.MergeValues(newValues, customValues)
	if err != nil {
		blog.Errorf("[%s]merge values failed, clusterID: %s, err: %s", i.MeshID, clusterID, err)
		return err
	}
	blog.Infof("[%s]merged values: %s", i.MeshID, mergedValues)

	_, err = helm.Upgrade(
		ctx,
		&helmmanager.UpgradeReleaseV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &clusterID,
			Chart:       &i.ChartName,
			Repository:  i.ChartRepo,
			Version:     i.ChartVersion,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        istiodReleaseName,
			Values:      []string{mergedValues},
		},
	)
	if err != nil {
		blog.Errorf("[%s]upgrade istiod failed, clusterID: %s, err: %s", i.MeshID, clusterID, err)
		return err
	}

	return nil
}

// updateClusterResource 更新集群依赖的资源
// nolint:funlen
func (i *IstioUpdateAction) updateClusterResource(ctx context.Context, clusterID string) error {
	blog.Infof("[%s]updating cluster resources for cluster %s", *i.MeshID, clusterID)

	if i.ObservabilityConfig == nil {
		blog.Infof("[%s]no observability config provided, skipping cluster resource update", *i.MeshID)
		return nil
	}

	// 更新 ServiceMonitor 资源（控制面监控）
	if i.ObservabilityConfig.MetricsConfig != nil {
		if i.ObservabilityConfig.MetricsConfig.MetricsEnabled.GetValue() &&
			i.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled.GetValue() {
			// 启用控制面监控，部署 ServiceMonitor
			if err := k8s.DeployResourceByYAML(
				ctx,
				clusterID,
				common.GetServiceMonitorYAML(),
				"ServiceMonitor",
				common.ServiceMonitorName,
			); err != nil {
				blog.Errorf("[%s]deploy ServiceMonitor failed for cluster %s, err: %s", *i.MeshID, clusterID, err)
				return err
			}
		} else {
			// 禁用控制面监控，删除 ServiceMonitor
			if err := k8s.DeleteResource(
				ctx,
				clusterID,
				"ServiceMonitor",
				common.ServiceMonitorName,
			); err != nil {
				blog.Errorf("[%s]delete ServiceMonitor failed for cluster %s, err: %s", *i.MeshID, clusterID, err)
				return err
			}
		}

		// 更新 PodMonitor 资源（数据面监控）
		if i.ObservabilityConfig.MetricsConfig.MetricsEnabled.GetValue() &&
			i.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled.GetValue() {
			// 启用数据面监控，部署 PodMonitor
			if err := k8s.DeployResourceByYAML(
				ctx,
				clusterID,
				common.GetPodMonitorYAML(),
				"PodMonitor",
				common.PodMonitorName,
			); err != nil {
				blog.Errorf("[%s]deploy PodMonitor failed for cluster %s, err: %s", *i.MeshID, clusterID, err)
				return err
			}
		} else {
			// 禁用数据面监控，删除 PodMonitor
			if err := k8s.DeleteResource(
				ctx,
				clusterID,
				"PodMonitor",
				common.PodMonitorName,
			); err != nil {
				blog.Errorf("[%s]delete PodMonitor failed for cluster %s, err: %s", *i.MeshID, clusterID, err)
				return err
			}
		}
	}

	// 更新 Telemetry 资源（链路追踪）， 只有大于1.21的版本才支持链路追踪
	if i.ObservabilityConfig.TracingConfig != nil && utils.IsVersionSupported(i.Version, "1.21") {
		if i.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
			// 启用链路追踪，部署 Telemetry
			traceSamplingPercent := 1
			if i.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() != 0 {
				traceSamplingPercent = int(i.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue())
			}
			if err := k8s.DeployResourceByYAML(
				ctx,
				clusterID,
				common.GetTelemetryYAML(traceSamplingPercent),
				"Telemetry",
				common.TelemetryName,
			); err != nil {
				blog.Errorf("[%s]deploy Telemetry failed for cluster %s, err: %s", *i.MeshID, clusterID, err)
				return err
			}
		} else {
			// 禁用链路追踪或版本不支持，删除 Telemetry
			if err := k8s.DeleteResource(
				ctx,
				clusterID,
				"Telemetry",
				common.TelemetryName,
			); err != nil {
				blog.Errorf("[%s]delete Telemetry failed for cluster %s, err: %s", *i.MeshID, clusterID, err)
				return err
			}
		}
	}

	blog.Infof("[%s]cluster resources update completed for cluster %s", *i.MeshID, clusterID)
	return nil
}

// Done 完成回调
func (i *IstioUpdateAction) Done(err error) {
	if err != nil {
		blog.Errorf("[%s]istio update operation failed, err: %s", i.MeshID, err)
		updateErr := i.Model.Update(context.TODO(), *i.MeshID, entity.M{
			entity.FieldKeyStatus:        common.IstioStatusUpdateFailed,
			entity.FieldKeyStatusMessage: fmt.Sprintf("更新失败，%s", err.Error()),
		})
		if updateErr != nil {
			blog.Errorf("[%s]update mesh status failed, err: %s", *i.MeshID, updateErr)
		}
	} else {
		updateErr := i.Model.Update(context.TODO(), *i.MeshID, entity.M{
			entity.FieldKeyStatus:        common.IstioStatusRunning,
			entity.FieldKeyStatusMessage: "",
		})
		if updateErr != nil {
			blog.Errorf("[%s]update mesh status failed, err: %s", *i.MeshID, updateErr)
		}
	}
}
