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
	defer recorder(ctx, "create_variable", req, resp)
	// 判断是否有创建权限
	// authUser := auth.GetAuthUserFromCtx(ctx)
	// if err := perm.CanCreateVariable(authUser, req.ProjectCode); err != nil {
	// 	return err
	// }
	// 创建变量
	ca := vda.NewCreateAction(p.model)
	vd, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	resp.Id = vd.ID
	resp.ProjectCode = vd.ProjectCode
	resp.Name = vd.Name
	resp.Key = vd.Key
	resp.Scope = vd.Scope
	resp.Default = vd.Default
	resp.Desc = vd.Description
	resp.Category = vd.Category
	return nil
}

// UpdateVariable implement for UpdateVariable interface
func (p *VariableHandler) UpdateVariable(ctx context.Context,
	req *proto.UpdateVariableRequest, resp *proto.UpdateVariableResponse) error {
	defer recorder(ctx, "create_variable", req, resp)
	// 判断是否有创建权限
	// authUser := auth.GetAuthUserFromCtx(ctx)
	// if err := perm.CanUpdateVariable(authUser, req.ProjectCode); err != nil {
	// 	return err
	// }
	// 创建变量
	ca := vda.NewUpdateAction(p.model)
	vd, err := ca.Do(ctx, req)
	if err != nil {
		return err
	}
	resp.Id = vd.ID
	resp.ProjectCode = vd.ProjectCode
	resp.Name = vd.Name
	resp.Key = vd.Key
	resp.Scope = vd.Scope
	resp.Default = vd.Default
	resp.Desc = vd.Description
	resp.Category = vd.Category
	return nil
}

// ListVariableDefinitions implement for ListVariableDefinitions interface
func (p *VariableHandler) ListVariableDefinitions(ctx context.Context,
	req *proto.ListVariableDefinitionsRequest, resp *proto.ListVariableDefinitionsResponse) error {
	defer recorder(ctx, "list_variable", req, resp)
	// 判断是否有项目查看权限
	// authUser := auth.GetAuthUserFromCtx(ctx)
	// if err := perm.CanViewProject(authUser, req.ProjectCode); err != nil {
	// 	return err
	// }
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
	defer recorder(ctx, "list_cluster_variables", req, resp)
	// 判断是否有创建权限
	// authUser := auth.GetAuthUserFromCtx(ctx)
	// if err := perm.CanViewProject(authUser, req.ProjectCode); err != nil {
	// 	return err
	// }
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
	defer recorder(ctx, "list_namespace_variables", req, resp)
	// 判断是否有创建权限
	// authUser := auth.GetAuthUserFromCtx(ctx)
	// if err := perm.CanViewProject(authUser, req.ProjectCode); err != nil {
	// 	return err
	// }
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
