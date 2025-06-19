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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
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

	// 删除集群中的istio
	for _, cluster := range clusters {
		if err := i.uninstallIstio(ctx, cluster); err != nil {
			blog.Errorf("uninstall istio for mesh %s failed, clusterID: %s, err: %s",
				i.MeshID, cluster, err)
			return fmt.Errorf("uninstall istio for mesh %s failed: %s", i.MeshID, err)
		}
	}
	return nil
}

// Done 完成回调
func (i *IstioUninstallAction) Done(err error) {
	m := make(entity.M)
	if err != nil {
		blog.Errorf("istio uninstall operation failed for mesh %s, err: %s", i.MeshID, err)
		m[entity.FieldKeyStatus] = common.IstioStatusUninstallingFailed
		m[entity.FieldKeyStatusMessage] = err.Error()
	} else {
		m[entity.FieldKeyStatus] = common.IstioStatusUninstalled
	}
	// 更新mesh状态为已删除
	updateErr := i.Model.Update(context.TODO(), i.MeshID, m)
	if updateErr != nil {
		blog.Errorf("update mesh status failed for mesh %s, err: %s", i.MeshID, updateErr)
	}
}

// uninstallIstio 卸载istio
func (i *IstioUninstallAction) uninstallIstio(ctx context.Context, clusterID string) error {
	// 删除istio base
	if err := i.uninstallIstioComponent(ctx, clusterID, common.IstioInstallBaseName); err != nil {
		return fmt.Errorf("uninstall istio base failed: %s", err)
	}

	// 删除istiod
	if err := i.uninstallIstioComponent(ctx, clusterID, common.IstioInstallIstiodName); err != nil {
		return fmt.Errorf("uninstall istiod failed: %s", err)
	}

	return nil
}

// uninstallIstioComponent 通用的istio组件卸载函数
func (i *IstioUninstallAction) uninstallIstioComponent(ctx context.Context, clusterID, componentName string) error {
	resp, err := helm.Uninstall(ctx, &helmmanager.UninstallReleaseV1Req{
		ProjectCode: pointer.String(i.ProjectCode),
		ClusterID:   pointer.String(clusterID),
		Name:        pointer.String(componentName),
		Namespace:   pointer.String(common.IstioNamespace),
	})
	if err != nil {
		blog.Errorf("helm uninstall %s failed, meshID: %s, clusterID: %s, err: %s",
			componentName, i.MeshID, clusterID, err)
		return fmt.Errorf("uninstall %s failed: %s", componentName, err)
	}
	if resp.Result != nil && !*resp.Result {
		blog.Errorf("helm uninstall %s failed, meshID: %s, clusterID: %s, resp message: %s",
			componentName, i.MeshID, clusterID, *resp.Message)
		return fmt.Errorf("uninstall %s failed: %s", componentName, *resp.Message)
	}

	// 查询是否删除成功 查询详情 每隔5s查询一次 直到删除成功，超时2min
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			blog.Errorf("uninstall %s timeout, meshID: %s, clusterID: %s",
				componentName, i.MeshID, clusterID)
			return fmt.Errorf("uninstall %s timeout for cluster %s", componentName, clusterID)
		case <-ticker.C:
			// 查询 release 是否存在
			detail, err := helm.GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
				ProjectCode: pointer.String(i.ProjectCode),
				ClusterID:   pointer.String(clusterID),
				Name:        pointer.String(componentName),
				Namespace:   pointer.String(common.IstioNamespace),
			})
			if err != nil {
				blog.Errorf("get %s release status failed, meshID: %s, clusterID: %s, err: %v",
					componentName, i.MeshID, clusterID, err)
				return fmt.Errorf("get %s release status failed: %v", componentName, err)
			}
			if detail != nil && detail.Message != nil && *detail.Message == driver.ErrReleaseNotFound.Error() {
				return nil
			}
		}
	}
}
