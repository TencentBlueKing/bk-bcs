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

// Package istio implements the istio management actions
package istio

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// DeleteIstioAction action for deleting istio
type DeleteIstioAction struct {
	model store.MeshManagerModel
	req   *meshmanager.DeleteIstioRequest
	resp  *meshmanager.DeleteIstioResponse
}

// NewDeleteIstioAction create delete istio action
func NewDeleteIstioAction(model store.MeshManagerModel) *DeleteIstioAction {
	return &DeleteIstioAction{
		model: model,
	}
}

// Handle processes the mesh deletion request
func (d *DeleteIstioAction) Handle(
	ctx context.Context,
	req *meshmanager.DeleteIstioRequest,
	resp *meshmanager.DeleteIstioResponse,
) error {
	d.req = req
	d.resp = resp

	// 执行统一校验
	meshIstio, err := d.Validate(ctx)
	if err != nil {
		blog.Errorf("validate delete istio request failed, meshID: %s, err: %s", d.req.MeshID, err)
		if customErr, ok := err.(*common.CodeMessageError); ok {
			d.setResp(customErr.GetCode(), customErr.GetMessageWithErr())
		} else {
			d.setResp(common.InnerErrorCode, err.Error())
		}
		return nil
	}

	if err := d.delete(ctx, meshIstio); err != nil {
		blog.Errorf("delete istio failed, meshID: %s, projectCode: %s, err: %s",
			d.req.MeshID, meshIstio.ProjectCode, err)
		if customErr, ok := err.(*common.CodeMessageError); ok {
			d.setResp(customErr.GetCode(), customErr.GetMessageWithErr())
		} else {
			d.setResp(common.InnerErrorCode, err.Error())
		}
		return nil
	}

	d.setResp(common.SuccessCode, "istio delete success")
	return nil
}

// setResp sets the response with code and message
func (d *DeleteIstioAction) setResp(code uint32, message string) {
	d.resp.Code = code
	d.resp.Message = message
}

// Validate validates the delete request and mesh status
func (d *DeleteIstioAction) Validate(ctx context.Context) (*entity.MeshIstio, error) {
	// 校验请求参数
	if err := d.req.Validate(); err != nil {
		blog.Errorf("request parameter validation failed, meshID: %s, err: %s", d.req.MeshID, err)
		return nil, common.NewCodeMessageError(common.InvalidRequestErrorCode, "invalid request parameters", err)
	}

	meshIstio, err := d.model.Get(ctx, d.req.MeshID)
	if err != nil {
		blog.Errorf("get mesh failed, meshID: %s, err: %s", d.req.MeshID, err)
		return nil, common.NewCodeMessageError(common.DBErrorCode, "get mesh failed", err)
	}
	if meshIstio == nil {
		blog.Errorf("mesh not found, meshID: %s", d.req.MeshID)
		return nil, common.NewCodeMessageError(common.NotFoundErrorCode, "mesh not found", nil)
	}

	// 检查mesh状态是否允许删除
	if meshIstio.Status != common.IstioStatusRunning {
		blog.Errorf("mesh status is not running, meshID: %s, current status: %s", d.req.MeshID, meshIstio.Status)
		return nil, common.NewCodeMessageError(common.InnerErrorCode, "mesh status is not running, cannot delete", nil)
	}

	// 检查集群中是否存在Istio资源，如果存在则不允许删除
	allClusters := make([]string, 0, len(meshIstio.PrimaryClusters)+len(meshIstio.RemoteClusters))
	allClusters = append(allClusters, meshIstio.PrimaryClusters...)
	allClusters = append(allClusters, meshIstio.RemoteClusters...)

	for _, clusterID := range allClusters {
		exists, err := k8s.CheckIstioResourceExists(ctx, clusterID)
		if err != nil {
			blog.Errorf("check istio resources failed, meshID: %s, clusterID: %s, err: %s",
				d.req.MeshID, clusterID, err)
			return nil, common.NewCodeMessageError(common.InnerErrorCode, "check istio resources failed", err)
		}

		if exists {
			blog.Errorf("cluster still has istio resources, meshID: %s, clusterID: %s",
				d.req.MeshID, clusterID)
			return nil, common.NewCodeMessageError(
				common.InnerErrorCode,
				fmt.Sprintf("cluster %s still has istio resources", clusterID),
				nil,
			)
		}
	}

	return meshIstio, nil
}

// delete implements the business logic for deleting istio
func (d *DeleteIstioAction) delete(ctx context.Context, meshIstio *entity.MeshIstio) error {
	// 更新mesh状态为删除中
	updateFields := entity.M{
		entity.FieldKeyStatus:     common.IstioStatusUninstalling,
		entity.FieldKeyUpdateTime: time.Now().Unix(),
	}
	if err := d.model.Update(ctx, d.req.MeshID, updateFields); err != nil {
		errMsg := fmt.Sprintf("update mesh status failed, meshID: %s", d.req.MeshID)
		blog.Errorf("%s, err: %s", errMsg, err)
		return common.NewCodeMessageError(common.DBErrorCode, errMsg, err)
	}

	// 异步删除istio
	action := actions.NewIstioUninstallAction(
		&actions.IstioUninstallOption{
			Model:           d.model,
			ProjectCode:     meshIstio.ProjectCode,
			MeshID:          d.req.MeshID,
			PrimaryClusters: meshIstio.PrimaryClusters,
			RemoteClusters:  meshIstio.RemoteClusters,
		},
	)
	// 异步执行，5分钟超时
	_, err := operation.GlobalOperator.Dispatch(action, 5*time.Minute)
	if err != nil {
		errMsg := fmt.Sprintf("dispatch istio uninstall action failed, meshID: %s, projectCode: %s",
			d.req.MeshID, meshIstio.ProjectCode)
		blog.Errorf("%s, err: %s", errMsg, err)
		return common.NewCodeMessageError(common.InnerErrorCode, errMsg, err)
	}

	// 返回删除结果
	d.setResp(common.SuccessCode, "删除中")
	return nil
}
