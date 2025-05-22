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

package value

import (
	"context"
	"fmt"
	"sync"

	bcsapiClusterManager "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	nsutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/namespace"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListNamespacesVariablesAction ...
type ListNamespacesVariablesAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListNamespacesVariablesRequest
}

// NewListNamespacesVariablesAction new list cluster variables action
func NewListNamespacesVariablesAction(model store.ProjectModel) *ListNamespacesVariablesAction {
	return &ListNamespacesVariablesAction{
		model: model,
	}
}

// Do ...
func (la *ListNamespacesVariablesAction) Do(ctx context.Context,
	req *proto.ListNamespacesVariablesRequest) ([]*proto.VariableValue, error) {
	la.ctx = ctx
	la.req = req

	variables, err := la.listNamespaceVariables(ctx)
	if err != nil {
		return nil, err
	}
	return variables, nil
}

func (la *ListNamespacesVariablesAction) listNamespaceVariables(ctx context.Context) ([]*proto.VariableValue, error) {
	project, err := la.model.GetProject(la.ctx, la.req.GetProjectCode())
	if err != nil {
		logging.Info("get project from db failed, err: %s", err.Error())
		return nil, errorx.NewDBErr(err.Error())
	}
	variableDefinition, err := la.model.GetVariableDefinition(la.ctx, la.req.GetVariableID())
	if err != nil {
		logging.Info("get variable definition from db failed, err: %s", err.Error())
		return nil, errorx.NewDBErr(err.Error())
	}
	if variableDefinition.Scope != vdm.VariableScopeNamespace {
		return nil, fmt.Errorf("variable %s scope is %s rather than namespace",
			la.req.GetVariableID(), variableDefinition.Scope)
	}
	clusters, err := clustermanager.ListClusters(ctx, project.ProjectID)
	if err != nil {
		return nil, err
	}
	// concurrently list namespace variables from cluster
	variables := []*proto.VariableValue{}
	lock := &sync.Mutex{}
	g, ctx := errgroup.WithContext(la.ctx)
	la.ctx = ctx
	for _, cluster := range clusters {
		g.Go(func(cluster *bcsapiClusterManager.Cluster) func() error {
			return func() error {
				vs, err := la.concurrencyList(cluster, variableDefinition)
				if err != nil {
					return err
				}
				lock.Lock()
				defer lock.Unlock()
				variables = append(variables, vs...)
				return nil
			}
		}(cluster))
	}
	if err := g.Wait(); err != nil {
		logging.Error("list variables failed, err:%s", err.Error())
		return nil, err
	}
	return variables, nil
}

func (la *ListNamespacesVariablesAction) concurrencyList(cluster *bcsapiClusterManager.Cluster,
	variableDefinition *vdm.VariableDefinition) ([]*proto.VariableValue, error) {
	client, err := clientset.GetClientGroup().Client(cluster.GetClusterID())
	if err != nil {
		logging.Error("get client for cluster %s failed, err: ", cluster.GetClusterID())
		return nil, err
	}
	nsList, err := client.CoreV1().Namespaces().List(la.ctx, metav1.ListOptions{})
	if err != nil {
		logging.Error("list namespaces in cluster %s failed, err: ", cluster.GetClusterID())
		return nil, err
	}
	// if cluster is shared, filter namespace list in project
	namespaces := nsutils.FilterNamespaces(nsList, cluster.GetIsShared(), la.req.GetProjectCode())
	variableValues, err := la.model.ListNamespaceVariableValues(la.ctx,
		la.req.GetVariableID(), cluster.GetClusterID())
	if err != nil {
		logging.Info("get variable values from db failed, err: %s", err.Error())
		return nil, err
	}
	exists := make(map[string]vvm.VariableValue, len(variableValues))
	for _, value := range variableValues {
		exists[value.Namespace] = value
	}
	variables := []*proto.VariableValue{}
	for _, ns := range namespaces {
		variable := &proto.VariableValue{
			ClusterID:   cluster.GetClusterID(),
			ClusterName: cluster.GetClusterName(),
			Namespace:   ns.GetName(),
		}
		if value, ok := exists[variable.Namespace]; ok {
			variable.Value = value.Value
		} else {
			variable.Value = variableDefinition.Default
		}
		variables = append(variables, variable)
	}
	return variables, nil
}
