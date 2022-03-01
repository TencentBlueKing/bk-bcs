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

// NewDeleteRepositoriesAction return a new DeleteRepositoriesAction instance
func NewDeleteRepositoriesAction(model store.HelmManagerModel) *DeleteRepositoriesAction {
	return &DeleteRepositoriesAction{
		model: model,
	}
}

// DeleteRepositoriesAction provides the action to do delete multi repositories
type DeleteRepositoriesAction struct {
	ctx context.Context

	model store.HelmManagerModel

	req  *helmmanager.DeleteRepositoriesReq
	resp *helmmanager.DeleteRepositoriesResp
}

// Handle the multiple deleting process
func (d *DeleteRepositoriesAction) Handle(ctx context.Context,
	req *helmmanager.DeleteRepositoriesReq, resp *helmmanager.DeleteRepositoriesResp) error {

	if req == nil || resp == nil {
		blog.Errorf("get repository failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	d.ctx = ctx
	d.req = req
	d.resp = resp

	if err := d.req.Validate(); err != nil {
		blog.Errorf("delete multi repositories failed, invalid request, %s, param: %v", err.Error(), d.req)
		d.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	return d.delete(d.req.GetProjectID(), d.req.GetNames())
}

func (d *DeleteRepositoriesAction) delete(projectID string, names []string) error {
	if projectID == "" || len(names) == 0 {
		blog.Errorf("delete multi repositories failed, get empty projectID or names")
		d.setResp(common.ErrHelmManagerRequestParamInvalid, "projectID or names can not be empty")
		return nil
	}

	if err := d.model.DeleteRepositories(d.ctx, projectID, names); err != nil {
		blog.Errorf("delete multi repositories failed, %s, projectID: %s, names: %s",
			err.Error(), projectID, names)
		d.setResp(common.ErrHelmManagerDeleteActionFailed, err.Error())
		return nil
	}

	d.setResp(common.ErrHelmManagerSuccess, "ok")
	blog.Infof("delete multi repositories successfully, projectID: %s, names: %s", projectID, names)
	return nil
}

func (d *DeleteRepositoriesAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	d.resp.Code = &code
	d.resp.Message = &msg
	d.resp.Result = err.OK()
}
