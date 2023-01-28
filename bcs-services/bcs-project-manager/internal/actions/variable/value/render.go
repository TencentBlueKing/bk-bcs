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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// RenderAction action for render variables
type RenderAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.RenderVariablesRequest
	resp  *proto.RenderVariablesResponse
}

// NewRenderVariablesAction new render variables action
func NewRenderVariablesAction(model store.ProjectModel) *RenderAction {
	return &RenderAction{
		model: model,
	}
}

// Do render variables request
func (ca *RenderAction) Do(ctx context.Context,
	req *proto.RenderVariablesRequest, resp *proto.RenderVariablesResponse) error {
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp
	variables := []*proto.VariableValue{}

	listCond := make(operator.M)
	listCond[vdm.FieldKeyProjectCode] = ca.req.GetProjectCode()
	variableDefinitions, _, err := ca.model.ListVariableDefinitions(
		ca.ctx, operator.NewLeafCondition(operator.Eq, listCond), &page.Pagination{All: true})
	if err != nil {
		logging.Error("list variable definition in project %s from db failed, err: %s",
			ca.req.GetProjectCode(), err.Error())
		return err
	}
	vdList := []vdm.VariableDefinition{}
	if req.GetKeyList() == "" {
		vdList = variableDefinitions
	} else {
		for _, vd := range variableDefinitions {
			if stringx.StringInSlice(vd.Key, stringx.SplitString(req.GetKeyList())) {
				vdList = append(vdList, vd)
			}
		}
	}
	for _, vd := range vdList {
		variable := &proto.VariableValue{
			Id:        vd.ID,
			Key:       vd.Key,
			Name:      vd.Name,
			Scope:     vd.Scope,
			ClusterID: ca.req.GetClusterID(),
			Namespace: ca.req.GetNamespace(),
		}
		if vd.Scope == vdm.VariableScopeGlobal {
			variable.Value = vd.Default
			variables = append(variables, variable)
			continue
		}
		var namespace string
		if vd.Scope == vdm.VariableScopeNamespace {
			namespace = ca.req.GetNamespace()
		}
		variableValue, err := ca.model.GetVariableValue(ca.ctx, vd.ID, ca.req.GetClusterID(), namespace, vd.Scope)
		if err == nil {
			variable.Value = variableValue.Value
		} else if err == drivers.ErrTableRecordNotFound {
			variable.Value = vd.Default
		} else {
			logging.Error("get variable value %s/%s from db failed, err: %s",
				ca.req.GetProjectCode(), vd.Key, err.Error())
			return err
		}
		variables = append(variables, variable)
	}
	ca.resp.Data = variables
	return nil
}
