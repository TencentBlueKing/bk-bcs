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

package handler

import (
	"context"

	vda "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/variable/definition"
	vva "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/variable/value"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// VariableHandler ...
type VariableHandler struct {
	model store.ProjectModel
}

// NewVariable return a variable service hander
func NewVariable(model store.ProjectModel) *VariableHandler {
	return &VariableHandler{
		model: model,
	}
}

// CreateVariable implement for CreateVariable interface
func (p *VariableHandler) CreateVariable(ctx context.Context,
	req *proto.CreateVariableRequest, resp *proto.CreateVariableResponse) error {
	ca := vda.NewCreateAction(p.model)
	vd, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	retData := &proto.CreateVariableData{
		Id:          vd.ID,
		ProjectCode: vd.ProjectCode,
		Name:        vd.Name,
		Key:         vd.Key,
		Scope:       vd.Scope,
		Default:     vd.Default,
		Desc:        vd.Description,
		Category:    vd.Category,
	}
	resp.Code = 0
	resp.Data = retData
	resp.Message = "ok"
	return nil
}

// UpdateVariable implement for UpdateVariable interface
func (p *VariableHandler) UpdateVariable(ctx context.Context,
	req *proto.UpdateVariableRequest, resp *proto.UpdateVariableResponse) error {
	ca := vda.NewUpdateAction(p.model)
	vd, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	retData := &proto.UpdateVariableData{
		Id:          vd.ID,
		ProjectCode: vd.ProjectCode,
		Name:        vd.Name,
		Key:         vd.Key,
		Scope:       vd.Scope,
		Default:     vd.Default,
		Desc:        vd.Description,
		Category:    vd.Category,
	}
	resp.Code = 0
	resp.Data = retData
	resp.Message = "ok"
	return nil
}

// ListVariableDefinitions implement for ListVariableDefinitions interface
func (p *VariableHandler) ListVariableDefinitions(ctx context.Context,
	req *proto.ListVariableDefinitionsRequest, resp *proto.ListVariableDefinitionsResponse) error {
	ca := vda.NewListAction(p.model)
	vd, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	// 返回
	respData := &proto.ListVariableDefinitionData{Total: (*vd)["total"].(uint32)}
	var vds []*proto.VariableDefinition
	if result, ok := (*vd)["results"].([]*vdm.VariableDefinition); ok {
		for _, v := range result {
			vds = append(vds, v.Transfer2Proto())
		}
	}
	respData.Results = vds
	resp.Data = respData
	return nil
}

// ListClusterVariables implement for ListClusterVariables interface
func (p *VariableHandler) ListClusterVariables(ctx context.Context,
	req *proto.ListClusterVariablesRequest, resp *proto.ListClusterVariablesResponse) error {
	ca := vva.NewListClusterVariablesAction(p.model)
	variables, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	// 返回
	respData := &proto.ListClusterVariablesData{
		Total:   uint32(len(variables)),
		Results: variables,
	}
	resp.Data = respData
	return nil
}

// ListNamespaceVariables implement for ListNamespaceVariables interface
func (p *VariableHandler) ListNamespaceVariables(ctx context.Context,
	req *proto.ListNamespaceVariablesRequest, resp *proto.ListNamespaceVariablesResponse) error {
	ca := vva.NewListNamespaceVariablesAction(p.model)
	variables, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	// 返回
	respData := &proto.ListNamespaceVariablesData{
		Total:   uint32(len(variables)),
		Results: variables,
	}
	resp.Data = respData
	return nil
}

// DeleteVariableDefinitions implement for DeleteVariableDefinitions interface
func (p *VariableHandler) DeleteVariableDefinitions(ctx context.Context,
	req *proto.DeleteVariableDefinitionsRequest, resp *proto.DeleteVariableDefinitionsResponse) error {
	ca := vda.NewDeleteAction(p.model)
	return ca.Do(ctx, req, resp)
}

// UpdateClusterVariables implement for UpdateClusterVariables interface
func (p *VariableHandler) UpdateClusterVariables(ctx context.Context,
	req *proto.UpdateClusterVariablesRequest, resp *proto.UpdateClusterVariablesResponse) error {
	ua := vva.NewUpdateClusterVariablesAction(p.model)
	return ua.Do(ctx, req)
}

// UpdateNamespaceVariables implement for UpdateNamespaceVariables interface
func (p *VariableHandler) UpdateNamespaceVariables(ctx context.Context,
	req *proto.UpdateNamespaceVariablesRequest, resp *proto.UpdateNamespaceVariablesResponse) error {
	ua := vva.NewUpdateNamespaceVariablesAction(p.model)
	return ua.Do(ctx, req)
}

// ImportVariables implement for ImportVariables interface
func (p *VariableHandler) ImportVariables(ctx context.Context,
	req *proto.ImportVariablesRequest, resp *proto.ImportVariablesResponse) error {
	ia := vda.NewImportVariablesAction(p.model)
	return ia.Do(ctx, req)
}
