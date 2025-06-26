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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
)

// IstioUpdateOption istio更新操作选项
type IstioUpdateOption struct {
	Model           store.MeshManagerModel
	ProjectCode     *string
	MeshID          *string
	ChartName       string
	ChartVersion    *string
	ChartRepo       *string
	PrimaryClusters []string
	RemoteClusters  []string
	UpdateFields    entity.M
	ChartValuesPath string
	UpdateValues    *common.IstiodInstallValues
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
	return nil
}

// 更新主集群
func (i *IstioUpdateAction) updatePrimaryCluster(ctx context.Context, clusterID string) error {
	// 获取当前集群实际istiod的values（用户可能在集群手动更新过参数，导致和从数据库中查询的数据不一致）
	releaseDetail, err := helm.GetReleaseDetail(
		ctx,
		&helmmanager.GetReleaseDetailV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &clusterID,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        pointer.String(common.IstioInstallIstiodName),
		},
	)
	if err != nil || releaseDetail == nil {
		blog.Errorf("[%s]get release detail failed, clusterID: %s", i.MeshID, clusterID)
		return fmt.Errorf("get release detail failed, clusterID: %s", clusterID)
	}

	// 从 releaseDetail 中获取当前集群的istiod的values
	if len(releaseDetail.Data.Values) == 0 {
		blog.Errorf("[%s]release values is empty, clusterID: %s", i.MeshID, clusterID)
		return fmt.Errorf("release values is empty, clusterID: %s", clusterID)
	}
	values := releaseDetail.Data.Values[0]
	var customValues string
	// 将UpdateValues转换为YAML
	customValuesBytes, err := yaml.Marshal(i.UpdateValues)
	if err != nil {
		blog.Errorf("[%s]marshal install values failed, err: %s", i.MeshID, err)
		return err
	}
	customValues = string(customValuesBytes)

	// 通过 utils.MergeValues 合并 values
	mergedValues, err := utils.MergeValues(values, customValues)
	if err != nil {
		blog.Errorf("[%s]merge values failed, clusterID: %s, err: %s", i.MeshID, clusterID, err)
		return err
	}

	// 用新的values更新istiod（通过helm upgrade）
	_, err = helm.Upgrade(
		ctx,
		&helmmanager.UpgradeReleaseV1Req{
			ProjectCode: i.ProjectCode,
			ClusterID:   &clusterID,
			Chart:       &i.ChartName,
			Repository:  i.ChartRepo,
			Version:     i.ChartVersion,
			Namespace:   pointer.String(common.IstioNamespace),
			Name:        pointer.String(common.IstioInstallIstiodName),
			Values:      []string{mergedValues},
		},
	)
	if err != nil {
		blog.Errorf("[%s]upgrade istiod failed, clusterID: %s, err: %s", i.MeshID, clusterID, err)
		return err
	}

	return nil
}

// Done 完成回调
func (i *IstioUpdateAction) Done(err error) {
	if err != nil {
		blog.Errorf("[%s]istio update operation failed, err: %s", i.MeshID, err)
		i.UpdateFields[entity.FieldKeyStatus] = common.IstioStatusUpdateFailed
		i.UpdateFields[entity.FieldKeyStatusMessage] = fmt.Sprintf("更新失败，%s", err.Error())
	} else {
		i.UpdateFields[entity.FieldKeyStatus] = common.IstioStatusRunning
		i.UpdateFields[entity.FieldKeyStatusMessage] = "更新成功"
	}
	updateErr := i.Model.Update(context.TODO(), *i.MeshID, i.UpdateFields)
	if updateErr != nil {
		blog.Errorf("[%s]update mesh status failed, err: %s", *i.MeshID, updateErr)
	}
}
