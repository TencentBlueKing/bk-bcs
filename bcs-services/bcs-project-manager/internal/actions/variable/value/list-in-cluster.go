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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListClusterVariablesAction ...
type ListClusterVariablesAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListClusterVariablesRequest
}

// NewListClusterVariablesAction new list cluster variables action
func NewListClusterVariablesAction(model store.ProjectModel) *ListClusterVariablesAction {
	return &ListClusterVariablesAction{
		model: model,
	}
}

// Do ...
func (la *ListClusterVariablesAction) Do(ctx context.Context,
	req *proto.ListClusterVariablesRequest) ([]*proto.VariableValue, error) {
	la.ctx = ctx
	la.req = req

	variables, err := la.listClusterVariables()
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}

	return variables, nil
}

func (la *ListClusterVariablesAction) listClusterVariables() ([]*proto.VariableValue, error) {
	listCond := make(operator.M)
	listCond[vdm.FieldKeyProjectCode] = la.req.GetProjectCode()
	listCond[vdm.FieldKeyScope] = vdm.VariableScopeCluster

	variableDefinitions, _, err := la.model.ListVariableDefinitions(
		la.ctx, operator.NewLeafCondition(operator.Eq, listCond),
		&page.Pagination{Sort: map[string]int{vdm.FieldKeyCreateTime: -1}, All: true})
	if err != nil {
		logging.Error("get variable definitions from db failed, err: %s", err.Error())
		return nil, err
	}
	var variables []*proto.VariableValue
	variableValues, err := la.model.ListVariableValuesInCluster(la.ctx,
		la.req.GetClusterID())
	if err != nil {
		return variables, err
	}
	exists := make(map[string]vvm.VariableValue, len(variableValues))
	for _, value := range variableValues {
		exists[value.VariableID] = value
	}
	for _, variableDefinition := range variableDefinitions {
		variable := &proto.VariableValue{
			Id:   variableDefinition.ID,
			Name: variableDefinition.Name,
			Key:  variableDefinition.Key,
		}
		if value, ok := exists[variable.Id]; ok {
			variable.Value = value.Value
		} else {
			variable.Value = variableDefinition.Default
		}
		variables = append(variables, variable)
	}
	return variables, nil
}
