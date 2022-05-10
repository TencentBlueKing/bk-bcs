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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewCreateRepositoryAction return a new CreateRepositoryAction instance
func NewCreateRepositoryAction(model store.HelmManagerModel, platform repo.Platform) *CreateRepositoryAction {
	return &CreateRepositoryAction{
		model:    model,
		platform: platform,
	}
}

// CreateRepositoryAction provides the action to do create repository
type CreateRepositoryAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.CreateRepositoryReq
	resp *helmmanager.CreateRepositoryResp
}

// Handle the creating process
func (c *CreateRepositoryAction) Handle(ctx context.Context,
	req *helmmanager.CreateRepositoryReq, resp *helmmanager.CreateRepositoryResp) error {

	if req == nil || resp == nil {
		blog.Errorf("create repository failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	c.ctx = ctx
	c.req = req
	c.resp = resp

	if err := c.req.Validate(); err != nil {
		blog.Errorf("create repository failed, invalid request, %s, param: %v", err.Error(), c.req)
		c.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return c.create(c.req.GetTakeover(), &helmmanager.Repository{
		ProjectID:      c.req.ProjectID,
		Name:           c.req.Name,
		Type:           c.req.Type,
		Remote:         c.req.Remote,
		RemoteURL:      c.req.RemoteURL,
		RemoteUsername: c.req.RemoteUsername,
		RemotePassword: c.req.RemotePassword,
		Username:       c.req.Username,
		Password:       c.req.Password,
		CreateBy:       c.req.Operator,
	})
}

func (c *CreateRepositoryAction) create(takeover bool, data *helmmanager.Repository) error {
	blog.Infof("try to create repository, takeover: %t, projectID: %s, type: %s, name: %s",
		takeover, data.GetProjectID(), data.GetType(), data.GetName())

	r := &entity.Repository{}
	r.LoadFromProto(data)

	projectHandler := c.platform.User(repo.User{
		Name:     data.GetCreateBy(),
		Password: data.GetPassword(),
	}).Project(data.GetProjectID())
	if err := projectHandler.Ensure(c.ctx); err != nil {
		blog.Errorf("create repository failed, ensure project failed, %s, param: %v", err.Error(), r)
		c.setResp(common.ErrHelmManagerCreateActionFailed, err.Error(), nil)
		return nil
	}

	if err := c.createRepository2Repo(takeover, projectHandler, r); err != nil {
		c.setResp(common.ErrHelmManagerCreateActionFailed, err.Error(), nil)
		blog.Errorf("create repository failed, create to repo failed, %s, project: %s, type: %s, name: %s",
			err.Error(), data.GetProjectID(), data.GetType(), data.GetName())
		return nil
	}

	if err := c.model.CreateRepository(c.ctx, r); err != nil {
		blog.Errorf("create repository failed, create repository in model failed, %s, param: %v", err.Error(), r)
		c.setResp(common.ErrHelmManagerCreateActionFailed, err.Error(), nil)
		return nil
	}

	c.setResp(common.ErrHelmManagerSuccess, "ok", r.Transfer2Proto())
	blog.Infof("create repository successfully, takeover: %t, projectID: %s, type: %s, name: %s",
		takeover, r.ProjectID, r.Type, r.Name)
	return nil
}

func (c *CreateRepositoryAction) createRepository2Repo(
	takeover bool, projectHandler repo.ProjectHandler, data *entity.Repository) error {

	handler := projectHandler.Repository(repo.GetRepositoryType(data.Type), data.Name)
	if takeover {
		if _, err := handler.Get(c.ctx); err != nil {
			return err
		}
		return nil
	}

	repoURL, err := handler.Create(c.ctx, &repo.Repository{
		Remote:         data.Remote,
		RemoteURL:      data.RemoteURL,
		RemoteUsername: data.RemoteUsername,
		RemotePassword: data.RemotePassword,
	})
	if err != nil {
		return err
	}

	u, p, err := handler.CreateUser(c.ctx)
	if err != nil {
		return err
	}
	data.Username = u
	data.Password = p
	data.RepoURL = repoURL

	return nil
}

func (c *CreateRepositoryAction) setResp(err common.HelmManagerError, message string, r *helmmanager.Repository) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	c.resp.Code = &code
	c.resp.Message = &msg
	c.resp.Result = err.OK()
	c.resp.Data = r
}
