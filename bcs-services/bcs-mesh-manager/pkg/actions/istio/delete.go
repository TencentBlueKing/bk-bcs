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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
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

	if err := d.req.Validate(); err != nil {
		blog.Errorf("delete mesh failed, invalid request, %s, param: %v", err.Error(), d.req)
		d.setResp(common.ParamErrorCode, err.Error())
		return nil
	}

	if err := d.delete(ctx); err != nil {
		blog.Errorf("delete mesh failed, %s, meshID: %s", err.Error(), d.req.MeshID)
		d.setResp(common.DBErrorCode, err.Error())
		return nil
	}

	d.setResp(common.SuccessCode, "")
	blog.Infof("delete mesh successfully, meshID: %s", d.req.MeshID)
	return nil
}

// setResp sets the response with code and message
func (d *DeleteIstioAction) setResp(code uint32, message string) {
	d.resp.Code = code
	d.resp.Message = message
}

// delete implements the business logic for deleting istio
func (d *DeleteIstioAction) delete(ctx context.Context) error {
	return d.model.Delete(ctx, d.req.MeshID)
}
