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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	opcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation/actions/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
)

// IstioUninstallOption istio卸载操作选项
type IstioUninstallOption struct {
	Model           store.MeshManagerModel
	ProjectCode     string
	MeshID          string
	PrimaryClusters []string
	RemoteClusters  []string
	OldReleaseNames map[string]map[string]string
}

// IstioUninstallAction istio卸载操作
type IstioUninstallAction struct {
	*IstioUninstallOption
}

var _ operation.Operation = &IstioUninstallAction{}

// NewIstioUninstallAction 创建istio卸载操作
func NewIstioUninstallAction(opt *IstioUninstallOption) *IstioUninstallAction {
	return &IstioUninstallAction{
		IstioUninstallOption: opt,
	}
}

// Action 操作名称
func (i *IstioUninstallAction) Action() string {
	return "istio-uninstall"
}

// Name 操作实例名称
func (i *IstioUninstallAction) Name() string {
	return fmt.Sprintf("istio-uninstall-%s", i.MeshID)
}

// Validate 验证参数
func (i *IstioUninstallAction) Validate() error {
	// 必填字段
	if i.ProjectCode == "" {
		return fmt.Errorf("projectCode is required")
	}
	if i.MeshID == "" {
		return fmt.Errorf("meshID is required")
	}
	// 主集群不能为空
	if len(i.PrimaryClusters) == 0 {
		return fmt.Errorf("primaryClusters is required")
	}
	// 校验istio相关资源是否存在
	// 检查集群中是否存在Istio资源，如果存在则不允许删除
	allClusters := make([]string, 0, len(i.PrimaryClusters)+len(i.RemoteClusters))
	allClusters = append(allClusters, i.PrimaryClusters...)
	allClusters = append(allClusters, i.RemoteClusters...)

	for _, cluster := range allClusters {
		exists, details, err := k8s.CheckIstioResourceExists(context.TODO(), cluster)
		if err != nil {
			blog.Errorf("check istio resources failed, meshID: %s, clusterID: %s, err: %s",
				i.MeshID, cluster, err)
			return fmt.Errorf("check istio resources failed, meshID: %s, clusterID: %s, err: %s",
				i.MeshID, cluster, err)
		}
		if exists {
			return fmt.Errorf("cluster %s still has istio resources: %s",
				cluster, strings.Join(details, ", "))
		}
	}
	return nil
}

// Prepare 准备阶段
func (i *IstioUninstallAction) Prepare(ctx context.Context) error {
	// 暂时无需预处理
	return nil
}

// Execute 执行删除
func (i *IstioUninstallAction) Execute(ctx context.Context) error {
	// 合并主从集群列表
	clusters := make([]string, 0, len(i.PrimaryClusters)+len(i.RemoteClusters))
	clusters = append(clusters, i.PrimaryClusters...)
	clusters = append(clusters, i.RemoteClusters...)

	for _, cluster := range clusters {
		// 删除istio
		if err := i.uninstallIstio(ctx, cluster); err != nil {
			blog.Errorf("[%s]uninstall istio for cluster %s failed, err: %s",
				i.MeshID, cluster, err)
			return fmt.Errorf("uninstall istio for cluster %s failed: %s", cluster, err)
		}

		// 删除集群的PodMonitor
		if err := k8s.DeletePodMonitor(ctx, []string{cluster}); err != nil {
			blog.Errorf("[%s]delete PodMonitor failed for cluster %s, err: %s", i.MeshID, cluster, err)
			return fmt.Errorf("delete PodMonitor failed for cluster %s: %s", cluster, err)
		}

		// 删除集群的ServiceMonitor
		if err := k8s.DeleteServiceMonitor(ctx, []string{cluster}); err != nil {
			blog.Errorf("[%s]delete ServiceMonitor failed for cluster %s, err: %s", i.MeshID, cluster, err)
			return fmt.Errorf("delete ServiceMonitor failed for cluster %s: %s", cluster, err)
		}
	}

	// 删除主集群的Telemetry
	if err := k8s.DeleteTelemetry(ctx, i.PrimaryClusters); err != nil {
		blog.Errorf("[%s]delete Telemetry failed for primary clusters, err: %s", i.MeshID, err)
		return fmt.Errorf("delete Telemetry failed for primary clusters: %s", err)
	}

	// 删除主集群的东西向网关
	if err := i.uninstallEgressGateway(ctx, i.PrimaryClusters[0]); err != nil {
		blog.Errorf("[%s]uninstall egress gateway failed, err: %s", i.MeshID, err)
		return fmt.Errorf("uninstall egress gateway failed: %s", err)
	}

	// 尝试删除istio crd
	for _, cluster := range clusters {
		if err := k8s.DeleteIstioCrd(ctx, cluster); err != nil {
			blog.Errorf("[%s]delete istio crd failed, err: %s", i.MeshID, err)
			return fmt.Errorf("delete istio crd failed: %s", err)
		}
	}

	return nil
}

// uninstallEgressGateway 卸载东西向网关
func (i *IstioUninstallAction) uninstallEgressGateway(ctx context.Context, clusterID string) error {
	// 获取东西向网关的release name
	releaseName, err := opcommon.GetReleaseName(i.OldReleaseNames, clusterID, common.ComponentIstioGateway, i.MeshID)
	if err != nil {
		return err
	}
	// 如果东西向网关的release name不存在，则表示未部署
	if releaseName == "" {
		return nil
	}

	// 删除东西向网关
	if err := helm.UninstallIstioComponent(
		ctx, clusterID,
		releaseName,
		i.ProjectCode,
		i.MeshID,
	); err != nil {
		return fmt.Errorf("uninstall egress gateway failed: %s", err)
	}
	return nil
}

// Done 完成回调
func (i *IstioUninstallAction) Done(err error) {
	m := make(entity.M)
	if err != nil {
		blog.Errorf("[%s]istio uninstall operation failed, err: %s", i.MeshID, err)
		m[entity.FieldKeyStatus] = common.IstioStatusUninstallingFailed
		m[entity.FieldKeyStatusMessage] = fmt.Sprintf("删除失败，%s", err.Error())
	} else {
		blog.Infof("[%s]istio uninstall success", i.MeshID)
		m[entity.FieldKeyStatus] = common.IstioStatusUninstalled
		m[entity.FieldKeyIsDeleted] = true
		m[entity.FieldKeyStatusMessage] = "删除成功"
	}
	// 更新mesh状态为已删除
	updateErr := i.Model.Update(context.TODO(), i.MeshID, m)
	if updateErr != nil {
		blog.Errorf("[%s]update mesh status failed, err: %s", i.MeshID, updateErr)
	}
}

// uninstallIstio 卸载istio
func (i *IstioUninstallAction) uninstallIstio(ctx context.Context, clusterID string) error {
	// 获取Release名称
	baseReleaseName, err := opcommon.GetReleaseName(i.OldReleaseNames, clusterID, common.ComponentIstioBase, i.MeshID)
	if err != nil {
		return err
	}

	istiodReleaseName, err := opcommon.GetReleaseName(i.OldReleaseNames, clusterID, common.ComponentIstiod, i.MeshID)
	if err != nil {
		return err
	}

	// 删除istio base
	if err := helm.UninstallIstioComponent(ctx, clusterID, baseReleaseName, i.ProjectCode, i.MeshID); err != nil {
		return fmt.Errorf("uninstall istio base failed: %s", err)
	}

	// 删除istiod
	if err := helm.UninstallIstioComponent(ctx, clusterID, istiodReleaseName, i.ProjectCode, i.MeshID); err != nil {
		return fmt.Errorf("uninstall istiod failed: %s", err)
	}

	return nil
}
