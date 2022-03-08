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

// NewGetRepositoryAction return a new GetRepositoryAction instance
func NewGetRepositoryAction(model store.HelmManagerModel) *GetRepositoryAction {
	return &GetRepositoryAction{
		model: model,
	}
}

// GetRepositoryAction provides the actions to get repository
type GetRepositoryAction struct {
	ctx context.Context

	model store.HelmManagerModel

	req  *helmmanager.GetRepositoryReq
	resp *helmmanager.GetRepositoryResp
}

// Handle the getting process
func (g *GetRepositoryAction) Handle(ctx context.Context,
	req *helmmanager.GetRepositoryReq, resp *helmmanager.GetRepositoryResp) error {

	if req == nil || resp == nil {
		blog.Errorf("get repository failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get repository failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return g.get(g.req.GetProjectID(), g.req.GetName())
}

func (g *GetRepositoryAction) get(projectID, name string) error {
	r, err := g.model.GetRepository(g.ctx, projectID, name)
	if err != nil {
		blog.Errorf("get repository failed, %s, projectID: %s, name: %s", err.Error(), projectID, name)
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	g.setResp(common.ErrHelmManagerSuccess, "ok", r.Transfer2Proto())
	blog.Infof("get repository successfully, projectID: %s, name: %s", r.ProjectID, r.Name)
	return nil
}

func (g *GetRepositoryAction) setResp(err common.HelmManagerError, message string, r *helmmanager.Repository) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
