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

// Package value xxx
package value

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListClustersVariablesAction ...
type ListClustersVariablesAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListClustersVariablesRequest
}

// NewListClustersVariablesAction new list cluster variables action
func NewListClustersVariablesAction(model store.ProjectModel) *ListClustersVariablesAction {
	return &ListClustersVariablesAction{
		model: model,
	}
}

// Do ...
func (la *ListClustersVariablesAction) Do(ctx context.Context,
	req *proto.ListClustersVariablesRequest) ([]*proto.VariableValue, error) {
	la.ctx = ctx
	la.req = req

	variables, err := la.listClusterVariables(ctx)
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}

	return variables, nil
}

func (la *ListClustersVariablesAction) listClusterVariables(ctx context.Context) ([]*proto.VariableValue, error) {
	project, err := la.model.GetProject(la.ctx, la.req.GetProjectCode())
	if err != nil {
		logging.Error("get project from db failed, err: %s", err.Error())
		return nil, err
	}
	variableDefinition, err := la.model.GetVariableDefinition(la.ctx, la.req.GetVariableID())
	if err != nil {
		logging.Error("get variable definition from db failed, err: %s", err.Error())
		return nil, err
	}
	if variableDefinition.Scope != vdm.VariableScopeCluster {
		return nil, fmt.Errorf("variable %s scope is %s rather than cluster",
			la.req.GetVariableID(), variableDefinition.Scope)
	}
	clusters, err := clustermanager.ListClusters(ctx, project.ProjectID)
	if err != nil {
		return nil, err
	}
	var variables []*proto.VariableValue
	variableValues, err := la.model.ListClusterVariableValues(la.ctx,
		la.req.GetVariableID())
	if err != nil {
		return variables, err
	}
	exists := make(map[string]vvm.VariableValue, len(variableValues))
	for _, value := range variableValues {
		exists[value.ClusterID] = value
	}
	for _, cluster := range clusters {
		variable := &proto.VariableValue{
			ClusterID:   cluster.ClusterID,
			ClusterName: cluster.ClusterName,
		}
		if value, ok := exists[variable.ClusterID]; ok {
			variable.Value = value.Value
		} else {
			variable.Value = variableDefinition.Default
		}
		variables = append(variables, variable)
	}
	return variables, nil
}
