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

package definition

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vd "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
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
	req *proto.ListClusterVariablesRequest) ([]*proto.ClusterVariable, error) {
	la.ctx = ctx
	la.req = req

	variables, err := la.listClusterVariables()
	if err != nil {
		return nil, errorx.NewDBErr(err)
	}

	return variables, nil
}

func (la *ListClusterVariablesAction) listClusterVariables() ([]*proto.ClusterVariable, error) {
	project, err := la.model.GetProject(la.ctx, la.req.GetProjectCode())
	if err != nil {
		logging.Info("get project from db failed, err: %s", err.Error())
		return nil, err
	}
	variableDefinition, err := la.model.GetVariableDefinition(la.ctx, la.req.GetVariableID())
	if err != nil {
		logging.Info("get variable definition from db failed, err: %s", err.Error())
		return nil, err
	}
	if variableDefinition.Scope != vd.VariableDefinitionScopeCluster {
		return nil, fmt.Errorf("variable %s scope is %s rather than cluster",
			la.req.GetVariableID(), variableDefinition.Scope)
	}
	cli, closeCon, err := clustermanager.GetClusterManagerClient()
	if err != nil {
		logging.Info("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer func() {
		if closeCon != nil {
			closeCon()
		}
	}()
	req := &clustermanager.ListClusterReq{
		ProjectID: project.ProjectID,
	}
	resp, err := cli.ListCluster(context.Background(), req)
	if err != nil {
		logging.Info("list cluster from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	clusters := resp.GetData()
	var variables []*proto.ClusterVariable
	var value string
	for _, cluster := range clusters {
		variableValue, err := la.model.GetVariableValue(la.ctx,
			la.req.GetProjectCode(), la.req.GetVariableID(), cluster.ClusterID, "", vd.VariableDefinitionScopeCluster)
		if err == drivers.ErrTableRecordNotFound {
			logging.Info("cannot get variable by id %s, clusterID %s", la.req.GetVariableID(), cluster.ClusterID)
			value = variableDefinition.Default
		} else if err != nil {
			logging.Info("get variable value from db failed, err: %s", err.Error())
			return nil, err
		} else {
			value = variableValue.Value
		}
		variable := &proto.ClusterVariable{
			ClusterID:   cluster.ClusterID,
			ClusterName: cluster.ClusterName,
			Value:       value,
		}
		variables = append(variables, variable)
	}
	return variables, nil
}
