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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// UpdateNamespacesVariablesAction ...
type UpdateNamespacesVariablesAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateNamespacesVariablesRequest
}

// NewUpdateNamespacesVariablesAction new update cluster variables action
func NewUpdateNamespacesVariablesAction(model store.ProjectModel) *UpdateNamespacesVariablesAction {
	return &UpdateNamespacesVariablesAction{
		model: model,
	}
}

// Do ...
func (la *UpdateNamespacesVariablesAction) Do(ctx context.Context,
	req *proto.UpdateNamespacesVariablesRequest) error {
	la.ctx = ctx
	la.req = req

	if err := la.updateNamespaceVariables(); err != nil {
		return errorx.NewDBErr(err.Error())
	}
	return nil
}

func (la *UpdateNamespacesVariablesAction) updateNamespaceVariables() error {
	_, err := la.model.GetProject(la.ctx, la.req.GetProjectCode())
	if err != nil {
		logging.Info("get project from db failed, err: %s", err.Error())
		return err
	}
	variableDefinition, err := la.model.GetVariableDefinition(la.ctx, la.req.GetVariableID())
	if err != nil {
		logging.Info("get variable definition from db failed, err: %s", err.Error())
		return err
	}
	if variableDefinition.Scope != vdm.VariableScopeNamespace {
		return fmt.Errorf("variable %s scope is %s rather than namespace",
			la.req.GetVariableID(), variableDefinition.Scope)
	}
	var username string
	if authUser, err := middleware.GetUserFromContext(la.ctx); err == nil {
		username = authUser.GetUsername()
	}
	entries := la.req.GetData()
	for _, entry := range entries {
		if err := la.model.UpsertVariableValue(la.ctx, &vvm.VariableValue{
			VariableID: la.req.GetVariableID(),
			ClusterID:  entry.ClusterID,
			Namespace:  entry.Namespace,
			Value:      entry.Value,
			Scope:      vdm.VariableScopeNamespace,
			UpdateTime: time.Now().Format(time.RFC3339),
			Updater:    username,
		}); err != nil {
			return err
		}
	}
	return nil
}
