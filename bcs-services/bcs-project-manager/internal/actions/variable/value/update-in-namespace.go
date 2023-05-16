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

// UpdateNamespaceVariablesAction ...
type UpdateNamespaceVariablesAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateNamespaceVariablesRequest
}

// NewUpdateNamespaceVariablesAction new update cluster variables action
func NewUpdateNamespaceVariablesAction(model store.ProjectModel) *UpdateNamespaceVariablesAction {
	return &UpdateNamespaceVariablesAction{
		model: model,
	}
}

// Do ...
func (la *UpdateNamespaceVariablesAction) Do(ctx context.Context,
	req *proto.UpdateNamespaceVariablesRequest) error {
	la.ctx = ctx
	la.req = req

	if err := la.updateNamespaceVariables(); err != nil {
		return errorx.NewDBErr(err.Error())
	}
	return nil
}

func (la *UpdateNamespaceVariablesAction) updateNamespaceVariables() error {
	if la.req.GetClusterID() == "" {
		return errorx.NewParamErr("clusterID cannot be empty")
	}
	if la.req.GetNamespace() == "" {
		return errorx.NewParamErr("namespace cannot be empty")
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
			Namespace:  la.req.GetNamespace(),
			Value:      variable.Value,
			Scope:      vdm.VariableScopeNamespace,
			UpdateTime: time.Now().Format(time.RFC3339),
			Updater:    username,
		}); err != nil {
			return err
		}
	}
	return nil
}
