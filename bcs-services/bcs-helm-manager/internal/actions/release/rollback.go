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

package release

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewRollbackReleaseAction return a new RollbackReleaseAction instance
func NewRollbackReleaseAction(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *RollbackReleaseAction {
	return &RollbackReleaseAction{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// RollbackReleaseAction provides the actions to do rollback release
type RollbackReleaseAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.RollbackReleaseReq
	resp *helmmanager.RollbackReleaseResp
}

// Handle the rollback process
func (r *RollbackReleaseAction) Handle(ctx context.Context,
	req *helmmanager.RollbackReleaseReq, resp *helmmanager.RollbackReleaseResp) error {

	if req == nil || resp == nil {
		blog.Errorf("rollback release failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	r.ctx = ctx
	r.req = req
	r.resp = resp

	if err := r.req.Validate(); err != nil {
		blog.Errorf("rollback release failed, invalid request, %s, param: %v", err.Error(), r.req)
		r.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	return r.rollback()
}

func (r *RollbackReleaseAction) rollback() error {
	releaseName := r.req.GetName()
	releaseNamespace := r.req.GetNamespace()
	revision := r.req.GetRevision()
	clusterID := r.req.GetClusterID()
	opName := r.req.GetOperator()

	handler := r.releaseHandler.Cluster(clusterID)
	// 执行rollback操作
	_, err := handler.Rollback(
		r.ctx,
		release.HelmRollbackConfig{
			Name:      releaseName,
			Namespace: releaseNamespace,
			Revision:  int(revision),
		})
	if err != nil {
		blog.Errorf("rollback release failed, %s, "+
			"clusterID: %s, namespace: %s, name: %s, rollback to revision %d, operator: %s",
			err.Error(), clusterID, releaseNamespace, releaseName, revision, opName)
		r.setResp(common.ErrHelmManagerRollbackActionFailed, err.Error())
		return nil
	}

	// 查询当前release的revision
	_, record, err := handler.List(r.ctx,
		release.ListOption{
			Size:      1,
			Namespace: releaseNamespace,
			Name:      releaseName,
		})
	if err != nil {
		blog.Errorf("rollback release get current revision failed, %s, "+
			"clusterID: %s, namespace: %s, name: %s, rollback to revision %d, operator: %s",
			err.Error(), clusterID, releaseNamespace, releaseName, revision, opName)
		r.setResp(common.ErrHelmManagerRollbackActionFailed, err.Error())
		return nil
	}
	if len(record) == 0 {
		blog.Errorf("rollback release get current revision failed, release not found, "+
			"clusterID: %s, namespace: %s, name: %s, rollback to revision %d, operator: %s",
			clusterID, releaseNamespace, releaseName, revision, opName)
		r.setResp(common.ErrHelmManagerRollbackActionFailed, "release not found")
		return nil
	}
	currentRevision := record[0].Revision

	// 存储release信息到store中, 首先先删掉原来的同revision的数据
	if err = r.model.DeleteRelease(r.ctx, clusterID, releaseNamespace, releaseNamespace, currentRevision); err != nil {
		blog.Errorf("rollback release, delete release in store failed, %s, "+
			"clusterID: %s, namespace: %s, name: %s, revision: %d, operator: %s",
			err.Error(), clusterID, releaseNamespace, releaseName, currentRevision, opName)
		r.setResp(common.ErrHelmManagerRollbackActionFailed, err.Error())
		return nil
	}
	if err = r.model.CreateRelease(r.ctx, &entity.Release{
		Name:       releaseName,
		Namespace:  releaseNamespace,
		ClusterID:  clusterID,
		Revision:   currentRevision,
		RollbackTo: int(revision),
	}); err != nil {
		blog.Errorf("rollback release, create release in store failed, %s, "+
			"clusterID: %s, namespace: %s, name: %s, revision: %d, operator: %s",
			err.Error(), clusterID, releaseNamespace, releaseName, currentRevision, opName)
		r.setResp(common.ErrHelmManagerRollbackActionFailed, err.Error())
		return nil
	}

	blog.Infof("rollback release successfully, rollback to revision %d and current revision is %d, "+
		"clusterID: %s, namespace: %s, name: %s, operator: %s",
		revision, currentRevision, clusterID, releaseNamespace, releaseName, opName)
	r.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (r *RollbackReleaseAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	r.resp.Code = &code
	r.resp.Message = &msg
	r.resp.Result = err.OK()
}
