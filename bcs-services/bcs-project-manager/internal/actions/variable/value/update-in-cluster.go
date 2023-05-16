/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package value

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// UpdateClusterVariablesAction ...
type UpdateClusterVariablesAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateClusterVariablesRequest
}

// NewUpdateClusterVariablesAction new update cluster variables action
func NewUpdateClusterVariablesAction(model store.ProjectModel) *UpdateClusterVariablesAction {
	return &UpdateClusterVariablesAction{
		model: model,
	}
}

// Do ...
func (la *UpdateClusterVariablesAction) Do(ctx context.Context,
	req *proto.UpdateClusterVariablesRequest) error {
	la.ctx = ctx
	la.req = req

	err := la.updateClusterVariables()
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}
	return nil
}

func (la *UpdateClusterVariablesAction) updateClusterVariables() error {
	if la.req.GetClusterID() == "" {
		return errorx.NewParamErr("clusterID cannot be empty")
	}
	var username string
	if authUser, err := middleware.GetUserFromContext(la.ctx); err == nil {
		username = authUser.GetUsername()
	}
	variables := la.req.GetData()
	for _, variable := range variables {
		if err := la.model.UpsertVariableValue(la.ctx, &vvm.VariableValue{
			VariableID: variable.Id,
			ClusterID:  la.req.GetClusterID(),
			Value:      variable.Value,
			Scope:      vdm.VariableScopeCluster,
			UpdateTime: time.Now().Format(time.RFC3339),
			Updater:    username,
		}); err != nil {
			return err
		}
	}
	return nil
}
