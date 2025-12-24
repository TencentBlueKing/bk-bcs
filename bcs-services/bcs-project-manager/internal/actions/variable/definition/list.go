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

package definition

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	timeutil "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/time"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListAction ...
type ListAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListVariableDefinitionsRequest
}

// NewListAction new list variable definitions action
func NewListAction(model store.ProjectModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

// Do ...
func (la *ListAction) Do(ctx context.Context,
	req *proto.ListVariableDefinitionsRequest) (*map[string]interface{}, error) {
	la.ctx = ctx
	la.req = req
	variables := []*vdm.VariableDefinition{}
	// inject system build in variables
	var systems []*vdm.VariableDefinition
	if la.req.GetScope() == "" {
		systems = vdm.FilterSystemVariables(
			[]string{vdm.VariableScopeGlobal, vdm.VariableScopeCluster, vdm.VariableScopeNamespace}, req.GetSearchKey())
	} else {
		systems = vdm.FilterSystemVariables([]string{la.req.GetScope()}, req.GetSearchKey())
	}
	if la.req.Offset == 0 {
		variables = append(variables, systems...)
		la.req.Limit -= int64(len(systems))
	} else {
		la.req.Offset -= int64(len(systems))
	}

	definitions, total, err := la.listVariableDefinitions()
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}
	variables = append(variables, definitions...)

	data := map[string]interface{}{
		"total":   uint32(total) + uint32(len(systems)),
		"results": variables,
	}
	return &data, nil
}

func (la *ListAction) listVariableDefinitions() ([]*vdm.VariableDefinition, int64, error) {
	var cond *operator.Condition
	condM := make(operator.M)
	condM["projectCode"] = la.req.GetProjectCode()
	if la.req.GetScope() != "" {
		condM["scope"] = la.req.GetScope()
	}
	condEq := operator.NewLeafCondition(operator.Eq, condM)
	// 通过变量key进行模糊查询
	condSearch := make(operator.M)
	var condC *operator.Condition
	if la.req.GetSearchKey() != "" {
		condSearch["key"] = primitive.Regex{Pattern: la.req.GetSearchKey(), Options: "i"}
		condC = operator.NewLeafCondition(operator.Con, condSearch)
		cond = operator.NewBranchCondition(operator.And, condC, condEq)
	} else {
		cond = condEq
	}

	// 查询变量信息
	definitions, total, err := la.model.ListVariableDefinitions(la.ctx, cond, &page.Pagination{
		Sort: map[string]int{vdm.FieldKeyCreateTime: -1}, Limit: la.req.Limit, Offset: la.req.Offset, All: la.req.All,
	})
	if err != nil {
		return nil, total, err
	}
	definitionList := []*vdm.VariableDefinition{}
	for i := range definitions {
		defi := definitions[i]
		defi.CreateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, defi.CreateTime)
		defi.UpdateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, defi.UpdateTime)
		definitionList = append(definitionList, &defi)
	}
	return definitionList, total, nil
}
