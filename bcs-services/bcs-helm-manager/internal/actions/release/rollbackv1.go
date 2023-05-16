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
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewRollbackReleaseV1Action return a new RollbackReleaseAction instance
func NewRollbackReleaseV1Action(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *RollbackReleaseV1Action {
	return &RollbackReleaseV1Action{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// RollbackReleaseV1Action provides the actions to do rollback release
type RollbackReleaseV1Action struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.RollbackReleaseV1Req
	resp *helmmanager.RollbackReleaseV1Resp
}

// Handle the rollback process
func (r *RollbackReleaseV1Action) Handle(ctx context.Context,
	req *helmmanager.RollbackReleaseV1Req, resp *helmmanager.RollbackReleaseV1Resp) error {
	r.ctx = ctx
	r.req = req
	r.resp = resp

	if err := r.req.Validate(); err != nil {
		blog.Errorf("rollback release failed, invalid request, %s, param: %v", err.Error(), r.req)
		r.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	if err := r.rollback(); err != nil {
		blog.Errorf("rollback release %s failed, clusterID: %s, namespace: %s, targetRevision: %d, error: %s",
			r.req.GetName(), r.req.GetClusterID(), r.req.GetNamespace(), r.req.GetRevision(), err.Error())
		r.setResp(common.ErrHelmManagerRollbackActionFailed, err.Error())
		return nil
	}

	blog.Infof("dispatch release successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s, operator: %s",
		r.req.GetProjectCode(), r.req.GetClusterID(), r.req.GetNamespace(), r.req.GetName(), auth.GetUserFromCtx(r.ctx))
	r.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (r *RollbackReleaseV1Action) rollback() error {
	// check release revision exist
	_, err := r.releaseHandler.Cluster(r.req.GetClusterID()).Get(r.ctx, release.GetOption{
		Namespace: r.req.GetNamespace(), Name: r.req.GetName(), Revision: int(r.req.GetRevision()),
	})
	if err != nil {
		return fmt.Errorf("check release revision failed, %s", err.Error())
	}

	if err = r.saveDB(); err != nil {
		return fmt.Errorf("db error, %s", err.Error())
	}

	// dispatch release
	options := &actions.ReleaseRollbackActionOption{
		Model:          r.model,
		ReleaseHandler: r.releaseHandler,
		ClusterID:      r.req.GetClusterID(),
		Name:           r.req.GetName(),
		Namespace:      r.req.GetNamespace(),
		Revision:       int(r.req.GetRevision()),
		Username:       auth.GetUserFromCtx(r.ctx),
	}
	action := actions.NewReleaseRollbackAction(options)
	_, err = operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout)
	if err != nil {
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	return nil
}

func (r *RollbackReleaseV1Action) saveDB() error {
	create := false
	_, err := r.model.GetRelease(r.ctx, r.req.GetClusterID(), r.req.GetNamespace(), r.req.GetName())
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			create = true
		} else {
			return err
		}
	}

	username := auth.GetUserFromCtx(r.ctx)
	if create {
		if err := r.model.CreateRelease(r.ctx, &entity.Release{
			Name:        r.req.GetName(),
			ProjectCode: contextx.GetProjectCodeFromCtx(r.ctx),
			Namespace:   r.req.GetNamespace(),
			ClusterID:   r.req.GetClusterID(),
			CreateBy:    username,
			Status:      helmrelease.StatusPendingRollback.String(),
		}); err != nil {
			return err
		}
	} else {
		rl := entity.M{
			entity.FieldKeyUpdateBy: username,
			entity.FieldKeyStatus:   helmrelease.StatusPendingRollback.String(),
			entity.FieldKeyMessage:  "",
		}
		if err := r.model.UpdateRelease(r.ctx, r.req.GetClusterID(), r.req.GetNamespace(),
			r.req.GetName(), rl); err != nil {

		}
	}
	return nil
}

func (r *RollbackReleaseV1Action) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	r.resp.Code = &code
	r.resp.Message = &msg
	r.resp.Result = err.OK()
}
