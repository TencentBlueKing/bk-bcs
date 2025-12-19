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

package shared

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	nsutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/namespace"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListNamespaces implement for ListNamespaces interface
func (a *SharedNamespaceAction) ListNamespaces(ctx context.Context,
	req *proto.ListNamespacesRequest, resp *proto.ListNamespacesResponse) error {
	var retDatas []*proto.NamespaceData
	existns := map[string]nsm.Namespace{}
	// if itsm is not enable, list namespaces directly
	if config.GlobalConf.ITSM.Enable {
		// list staging creating namespaces from db
		stagings, err := a.model.ListNamespacesByItsmTicketType(ctx, req.GetProjectCode(), req.GetClusterID(),
			[]string{nsm.ItsmTicketTypeCreate, nsm.ItsmTicketTypeUpdate, nsm.ItsmTicketTypeDelete})
		if err != nil {
			logging.Error("list staging namespaces failed, err: %s", err.Error())
			return errorx.NewDBErr(err.Error())
		}
		// filter staging namespaces by its type, create as creating, update and delete as existing in cluster
		creatings := []nsm.Namespace{}
		for _, staging := range stagings {
			switch staging.ItsmTicketType {
			case nsm.ItsmTicketTypeCreate:
				creatings = append(creatings, staging)
			case nsm.ItsmTicketTypeUpdate:
				existns[staging.Name] = staging
			case nsm.ItsmTicketTypeDelete:
				existns[staging.Name] = staging
			}
		}
		// list creating namespaces from db and insert into retDatas first
		for _, namespace := range creatings {
			retDatas = append(retDatas, loadListRetDataFromDB(namespace))
		}
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
		return errorx.NewClusterErr(err.Error())
	}
	// list all quota in cluster
	quotaMap := map[string]corev1.ResourceQuota{}
	if quotaList, e := client.CoreV1().ResourceQuotas("").List(ctx, metav1.ListOptions{}); e == nil {
		for _, quota := range quotaList.Items {
			if quota.GetName() == quota.GetNamespace() {
				quotaMap[quota.GetName()] = quota
			}
		}
	}
	// filter namespaces by project code
	namespaces := nsutils.FilterNamespaces(nsList, true, req.GetProjectCode())
	namespaces = nsutils.FilterOutVcluster(namespaces)
	variablesMap, err := batchListNamespaceVariables(ctx, req.GetProjectCode(), req.GetClusterID(), namespaces)
	if err != nil {
		logging.Error("batch list variables failed, err: %s", err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	list, err := loadRetDatasFromCluster(ctx, req.ClusterID, namespaces, variablesMap, quotaMap, existns)
	if err != nil {
		return err
	}
	retDatas = append(retDatas, list...)
	resp.Data = retDatas

	go func() {
		// sync namespaces to bcs-cc
		if err := common.SyncNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), namespaces); err != nil {
			logging.Error("sync shared namespaces %s/%s failed, err:%s",
				req.GetProjectCode(), req.GetClusterID(), err.Error())
		}
	}()
	return nil
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

func batchListNamespaceVariables(ctx context.Context,
	projectCode, clusterID string, namespaces []corev1.Namespace) (map[string][]*proto.VariableValue, error) {
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
	variablesMap := make(map[string][]*proto.VariableValue)
	variableValues, err := model.ListVariableValuesInAllNamespace(ctx, clusterID)
	if err != nil {
		logging.Error("list variable values from db failed, err: %s", err.Error())
		return variablesMap, errorx.NewDBErr(err.Error())
	}
	exists := make(map[string]vvm.VariableValue)
	for _, value := range variableValues {
		exists[value.VariableID+"&"+value.Namespace] = value
	}
	for _, namespace := range namespaces {
		for _, definition := range definitions {
			variable := &proto.VariableValue{
				Id:   definition.ID,
				Name: definition.Name,
				Key:  definition.Key,
			}
			if value, ok := exists[definition.ID+"&"+namespace.GetName()]; ok {
				variable.Value = value.Value
			} else {
				variable.Value = definition.Default
			}
			if _, ok := variablesMap[namespace.GetName()]; ok {
				variablesMap[namespace.GetName()] = append(variablesMap[namespace.GetName()], variable)
			} else {
				variablesMap[namespace.GetName()] = []*proto.VariableValue{variable}
			}
		}
	}
	return variablesMap, nil
}

func loadListRetDataFromDB(namespace nsm.Namespace) *proto.NamespaceData {
	retData := &proto.NamespaceData{
		Name:     namespace.Name,
		IsSystem: stringx.StringInSlice(namespace.Name, config.GlobalConf.SystemConfig.SystemNameSpaces),
	}
	if namespace.ResourceQuota != nil {
		retData.Quota = &proto.ResourceQuota{
			CpuRequests:    namespace.ResourceQuota.CPURequests,
			CpuLimits:      namespace.ResourceQuota.CPULimits,
			MemoryRequests: namespace.ResourceQuota.MemoryRequests,
			MemoryLimits:   namespace.ResourceQuota.MemoryLimits,
		}
	}
	variables := []*proto.VariableValue{}
	for _, variable := range namespace.Variables {
		variables = append(variables, &proto.VariableValue{
			Id:    variable.VariableID,
			Key:   variable.Key,
			Value: variable.Value,
		})
	}
	retData.Variables = variables
	retData.Managers = []string{namespace.Managers}
	retData.ItsmTicketSN = namespace.ItsmTicketSN
	retData.ItsmTicketStatus = namespace.ItsmTicketStatus
	retData.ItsmTicketURL = namespace.ItsmTicketURL
	retData.ItsmTicketType = namespace.ItsmTicketType
	return retData
}

func loadRetDatasFromCluster(ctx context.Context, clusterID string, namespaces []corev1.Namespace,
	variablesMap map[string][]*proto.VariableValue, quotaMap map[string]corev1.ResourceQuota,
	existns map[string]nsm.Namespace) ([]*proto.NamespaceData, error) {
	retDatas := []*proto.NamespaceData{}
	for _, namespace := range namespaces {
		retData := &proto.NamespaceData{
			Name:       namespace.GetName(),
			Uid:        string(namespace.GetUID()),
			CreateTime: namespace.GetCreationTimestamp().Format(constant.TimeLayout),
			Status:     string(namespace.Status.Phase),
			IsSystem:   stringx.StringInSlice(namespace.GetName(), config.GlobalConf.SystemConfig.SystemNameSpaces),
		}
		// get quota
		if quota, ok := quotaMap[namespace.GetName()]; ok {
			retData.Quota, retData.Used, retData.CpuUseRate, retData.MemoryUseRate =
				quotautils.TransferToProto(&quota)
		}
		// get variables
		retData.Variables = variablesMap[namespace.GetName()]
		if ns, ok := existns[retData.GetName()]; ok {
			retData.ItsmTicketType = ns.ItsmTicketType
			retData.ItsmTicketSN = ns.ItsmTicketSN
			retData.ItsmTicketStatus = ns.ItsmTicketStatus
			retData.ItsmTicketURL = ns.ItsmTicketURL
		}
		// get managers
		managers := []string{}
		if creator, exists := namespace.Annotations[constant.AnnotationKeyCreator]; exists {
			managers = append(managers, creator)
		} else {
			cluster, err := clustermanager.GetCluster(ctx, clusterID)
			if err != nil {
				return nil, err
			}
			managers = append(managers, cluster.Creator)
		}
		retData.Managers = managers
		retDatas = append(retDatas, retData)
	}
	return retDatas, nil
}
