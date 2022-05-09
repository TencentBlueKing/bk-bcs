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

package handler

import (
	"context"

	actionRepository "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/actions/repository"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// CreateRepository provide the actions to create a repository
func (hm *HelmManager) CreateRepository(ctx context.Context,
	req *helmmanager.CreateRepositoryReq, resp *helmmanager.CreateRepositoryResp) error {

	defer recorder(ctx, "CreateRepository", req, resp)()
	action := actionRepository.NewCreateRepositoryAction(hm.model, hm.platform)
	return action.Handle(ctx, req, resp)
}

// UpdateRepository provide the actions to update a repository
func (hm *HelmManager) UpdateRepository(ctx context.Context,
	req *helmmanager.UpdateRepositoryReq, resp *helmmanager.UpdateRepositoryResp) error {

	defer recorder(ctx, "UpdateRepository", req, resp)()
	action := actionRepository.NewUpdateRepositoryAction(hm.model)
	return action.Handle(ctx, req, resp)
}

// GetRepository provide the actions to get a repository
func (hm *HelmManager) GetRepository(ctx context.Context,
	req *helmmanager.GetRepositoryReq, resp *helmmanager.GetRepositoryResp) error {

	defer recorder(ctx, "GetRepository", req, resp)()
	action := actionRepository.NewGetRepositoryAction(hm.model)
	return action.Handle(ctx, req, resp)
}

// ListRepository provide the actions to list repositories
func (hm *HelmManager) ListRepository(ctx context.Context,
	req *helmmanager.ListRepositoryReq, resp *helmmanager.ListRepositoryResp) error {

	defer recorder(ctx, "ListRepository", req, resp)()
	action := actionRepository.NewListRepositoryAction(hm.model)
	return action.Handle(ctx, req, resp)
}

// DeleteRepository provide the actions to delete a repository
func (hm *HelmManager) DeleteRepository(ctx context.Context,
	req *helmmanager.DeleteRepositoryReq, resp *helmmanager.DeleteRepositoryResp) error {

	defer recorder(ctx, "DeleteRepository", req, resp)()
	action := actionRepository.NewDeleteRepositoryAction(hm.model)
	return action.Handle(ctx, req, resp)
}

// DeleteRepositories provide the actions to delete multi repositories
func (hm *HelmManager) DeleteRepositories(ctx context.Context,
	req *helmmanager.DeleteRepositoriesReq, resp *helmmanager.DeleteRepositoriesResp) error {

	defer recorder(ctx, "DeleteRepositories", req, resp)()
	action := actionRepository.NewDeleteRepositoriesAction(hm.model)
	return action.Handle(ctx, req, resp)
}
