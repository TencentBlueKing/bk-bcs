/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repository

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewDeleteRepositoryAction return a new DeleteRepositoryAction instance
func NewDeleteRepositoryAction(model store.HelmManagerModel) *DeleteRepositoryAction {
	return &DeleteRepositoryAction{
		model: model,
	}
}

// DeleteRepositoryAction provides the action to do delete repository
type DeleteRepositoryAction struct {
	ctx context.Context

	model store.HelmManagerModel

	req  *helmmanager.DeleteRepositoryReq
	resp *helmmanager.DeleteRepositoryResp
}

// Handle the deleting process
func (d *DeleteRepositoryAction) Handle(ctx context.Context,
	req *helmmanager.DeleteRepositoryReq, resp *helmmanager.DeleteRepositoryResp) error {

	if req == nil || resp == nil {
		blog.Errorf("delete repository failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	d.ctx = ctx
	d.req = req
	d.resp = resp

	if err := d.req.Validate(); err != nil {
		blog.Errorf("delete repository failed, invalid request, %s, param: %v", err.Error(), d.req)
		d.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	return d.delete(d.req.GetProjectID(), d.req.GetName())
}

func (d *DeleteRepositoryAction) delete(projectID, name string) error {
	if projectID == "" || name == "" {
		blog.Errorf("delete repository failed, get empty projectID or name")
		d.setResp(common.ErrHelmManagerRequestParamInvalid, "projectID or name can not be empty")
		return nil
	}

	if err := d.model.DeleteRepository(d.ctx, projectID, name); err != nil {
		blog.Errorf("delete repository failed, %s, projectID: %s, name: %s", err.Error(), projectID, name)
		d.setResp(common.ErrHelmManagerDeleteActionFailed, err.Error())
		return nil
	}

	d.setResp(common.ErrHelmManagerSuccess, "ok")
	blog.Infof("delete repository successfully, projectID: %s, name: %s", projectID, name)
	return nil
}

func (d *DeleteRepositoryAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	d.resp.Code = &code
	d.resp.Message = &msg
	d.resp.Result = err.OK()
}
