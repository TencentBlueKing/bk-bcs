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

package permitcheck

import (
	"context"
	"net/http"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

type permitContext struct {
	context.Context
	projName    string
	projID      string
	projPermits map[RSAction]bool
	argoProj    *v1alpha1.AppProject
	userName    string
}

func contextGetProjID(ctx context.Context) string {
	v, ok := ctx.(*permitContext)
	if !ok {
		return ""
	}
	return v.projID
}

func contextGetProjName(ctx context.Context) string {
	v, ok := ctx.(*permitContext)
	if !ok {
		return ""
	}
	return v.projName
}

func contextGetProjPermits(ctx context.Context) map[RSAction]bool {
	v, ok := ctx.(*permitContext)
	if !ok {
		return make(map[RSAction]bool)
	}
	return v.projPermits
}

func (c *checker) createPermitContext(ctx context.Context, project string) (*permitContext, int, error) {
	// directly return if ever created
	v, ok := ctx.(*permitContext)
	if ok && v.projName == project {
		return v, http.StatusOK, nil
	}
	// insert operate user if check project permission
	user := ctxutils.User(ctx)
	go c.db.UpdateActivityUserWithName(&dao.ActivityUserItem{
		Project: project, User: user.GetUser(),
	})

	argoProj, projID, statusCode, err := c.getProjectWithID(ctx, project)
	if err != nil {
		return nil, statusCode, err
	}
	result, err := c.getBCSMultiProjectPermission(ctx, []string{projID}, []RSAction{ProjectViewRSAction,
		ProjectEditRSAction, ProjectDeleteRSAction})
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get project '%s' permission failed",
			project)
	}
	projPermits, ok := result[projID]
	if !ok || len(projPermits) == 0 {
		return nil, http.StatusBadRequest, errors.Errorf("get project '%s' permission for user '%s "+
			"not have result'", project, user.GetUser())
	}
	pctx := &permitContext{
		Context:     ctx,
		projName:    project,
		projID:      projID,
		argoProj:    argoProj,
		projPermits: projPermits,
		userName:    user.GetUser(),
	}
	return pctx, http.StatusOK, nil
}
