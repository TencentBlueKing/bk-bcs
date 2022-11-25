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

package shared

import (
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	nsutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/namespace"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListNamespaces implement for ListNamespaces interface
func (a *SharedNamespaceAction) ListNamespaces(ctx context.Context,
	req *proto.ListNamespacesRequest, resp *proto.ListNamespacesResponse) error {
	retDatas := []*proto.NamespaceData{}
	// list staging creating namespaces from db
	stagings, err := a.model.ListNamespacesByItsmTicketType(ctx,
		req.GetProjectCode(), req.GetClusterID(), []string{nsm.ItsmTicketTypeCreate})
	if err != nil {
		logging.Error("list staging namespaces failed, err: %s", err.Error())
		return errorx.NewDBErr(err)
	}
	for _, staging := range stagings {
		retData := &proto.NamespaceData{
			Name: staging.Name,
		}
		if staging.ResourceQuota != nil {
			retData.Quota = &proto.ResourceQuota{
				CpuRequests:    staging.ResourceQuota.CPURequests,
				CpuLimits:      staging.ResourceQuota.CPULimits,
				MemoryRequests: staging.ResourceQuota.MemoryRequests,
				MemoryLimits:   staging.ResourceQuota.MemoryLimits,
			}
		}
		variables := []*proto.VariableValue{}
		for _, variable := range staging.Variables {
			variables = append(variables, &proto.VariableValue{
				Id:    variable.VariableID,
				Key:   variable.Key,
				Value: variable.Value,
			})
		}
		retData.Variables = variables
		retData.ItsmTicketSN = staging.ItsmTicketSN
		retData.ItsmTicketStatus = staging.ItsmTicketStatus
		retData.ItsmTicketURL = staging.ItsmTicketURL
		retData.ItsmTicketType = staging.ItsmTicketType
		retDatas = append(retDatas, retData)
	}
	// list exists namespaces from cluster
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		logging.Error("list namespaces in cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return errorx.NewClusterErr(err)
	}
	namespaces := nsutils.FilterNamespaces(nsList, true, req.GetProjectCode())
	// inject staging updating info to exist namespace
	modifyStaggings, err := a.model.ListNamespacesByItsmTicketType(ctx, req.GetProjectCode(), req.GetClusterID(),
		[]string{nsm.ItsmTicketTypeUpdate, nsm.ItsmTicketTypeDelete})
	existns := map[string]nsm.Namespace{}
	for _, modifyStagging := range modifyStaggings {
		existns[modifyStagging.Name] = modifyStagging
	}
	lock := &sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	for _, item := range namespaces {
		namespace := item
		g.Go(func() error {
			retData := &proto.NamespaceData{
				Name:       namespace.GetName(),
				Uid:        string(namespace.GetUID()),
				CreateTime: namespace.GetCreationTimestamp().Format(config.TimeLayout),
				Status:     string(namespace.Status.Phase),
			}
			// get quota
			quota, err := getNamespaceQuota(ctx, req.GetProjectCode(), req.GetClusterID(), namespace.GetName(), client)
			if err != nil {
				return err
			}
			if quota != nil {
				retData.Quota, retData.Used, retData.CpuUseRate, retData.MemoryUseRate = quotautils.TransferToProto(quota)
			}
			// get variables
			variables, err := listNamespaceVariables(ctx, req.GetProjectCode(), req.GetClusterID(), namespace.GetName())
			if err != nil {
				logging.Error("get namespace %s/%s variables failed, err: %s",
					req.GetClusterID(), namespace.GetName(), err.Error())
				return errorx.NewDBErr(err.Error())
			}
			retData.Variables = variables
			if ns, ok := existns[retData.GetName()]; ok {
				retData.ItsmTicketType = ns.ItsmTicketType
				retData.ItsmTicketSN = ns.ItsmTicketSN
				retData.ItsmTicketStatus = ns.ItsmTicketStatus
				retData.ItsmTicketURL = ns.ItsmTicketURL
			}
			lock.Lock()
			defer lock.Unlock()
			retDatas = append(retDatas, retData)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logging.Error("list namespaces in %s failed, err:%s", req.GetClusterID(), err.Error())
		return err
	}
	resp.Data = retDatas
	return nil
}

func getNamespaceQuota(ctx context.Context, projectCode, clusterID, namespace string, clientset *kubernetes.Clientset) (
	*corev1.ResourceQuota, error) {
	quota, err := clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, namespace, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logging.Error("get resourceQuota %s/%s failed, err: %s", clusterID, namespace, err.Error())
		return nil, errorx.NewClusterErr(err.Error())
	}

	if errors.IsNotFound(err) {
		return nil, nil
	}
	return quota, nil
}

func listNamespaceVariables(ctx context.Context,
	projectCode, clusterID, namespace string) ([]*proto.VariableValue, error) {
	model := store.GetModel()
	listCond := make(operator.M)
	listCond[vdm.FieldKeyProjectCode] = projectCode
	listCond[vdm.FieldKeyScope] = vdm.VariableScopeNamespace
	definitions, _, err := model.ListVariableDefinitions(ctx, operator.NewLeafCondition(operator.Eq, listCond),
		&page.Pagination{Sort: map[string]int{vdm.FieldKeyCreateTime: -1}, All: true})
	if err != nil {
		logging.Error("get variable definitions from db failed, err: %s", err.Error())
		return nil, errorx.NewDBErr(err.Error())
	}
	var variables []*proto.VariableValue
	variableValues, err := model.ListVariableValuesInNamespace(ctx, clusterID, namespace)
	if err != nil {
		logging.Error("list variable values from db failed, err: %s", err.Error())
		return variables, errorx.NewDBErr(err.Error())
	}
	exists := make(map[string]vvm.VariableValue, len(variableValues))
	for _, value := range variableValues {
		exists[value.VariableID] = value
	}
	for _, definition := range definitions {
		variable := &proto.VariableValue{
			Id:   definition.ID,
			Name: definition.Name,
			Key:  definition.Key,
		}
		if value, ok := exists[variable.Id]; ok {
			variable.Value = value.Value
		} else {
			variable.Value = definition.Default
		}
		variables = append(variables, variable)
	}
	return variables, nil
}
