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

package independent

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
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListNamespaces implement for ListNamespaces interface
func (c *IndependentNamespaceAction) ListNamespaces(ctx context.Context,
	req *proto.ListNamespacesRequest, resp *proto.ListNamespacesResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorx.NewClusterErr(err.Error())
	}
	quotaMap := map[string]corev1.ResourceQuota{}
	if quotaList, e := client.CoreV1().ResourceQuotas("").List(ctx, metav1.ListOptions{}); e == nil {
		for _, quota := range quotaList.Items {
			if quota.GetName() == quota.GetNamespace() {
				quotaMap[quota.GetName()] = quota
			}
		}
	}
	variablesMap, err := batchListNamespaceVariables(ctx, req.GetProjectCode(), req.GetClusterID(), nsList.Items)
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}
	retDatas := []*proto.NamespaceData{}
	for _, ns := range nsList.Items {
		retData := &proto.NamespaceData{
			Name:        ns.GetName(),
			Uid:         string(ns.GetUID()),
			Status:      string(ns.Status.Phase),
			CreateTime:  ns.GetCreationTimestamp().Format(constant.TimeLayout),
			Labels:      []*proto.Label{},
			Annotations: []*proto.Annotation{},
			IsSystem:    stringx.StringInSlice(ns.GetName(), config.GlobalConf.SystemConfig.SystemNameSpaces),
		}

		for k, v := range ns.Labels {
			retData.Labels = append(retData.Labels, &proto.Label{Key: k, Value: v})
		}
		for k, v := range ns.Annotations {
			retData.Annotations = append(retData.Annotations, &proto.Annotation{Key: k, Value: v})
		}
		// get quota
		if quota, ok := quotaMap[ns.GetName()]; ok {
			retData.Quota, retData.Used, retData.CpuUseRate, retData.MemoryUseRate = quotautils.TransferToProto(&quota)
		}
		// get variables
		retData.Variables = variablesMap[ns.GetName()]
		// get managers
		managers := []string{}
		if creator, exists := ns.Annotations[constant.AnnotationKeyCreator]; exists {
			managers = append(managers, creator)
		} else {
			cluster, err := clustermanager.GetCluster(ctx, req.ClusterID)
			if err != nil {
				return err
			}
			managers = append(managers, cluster.Creator)
		}
		retData.Managers = managers
		retDatas = append(retDatas, retData)
	}
	resp.Data = retDatas
	go func() {
		if err := common.SyncNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), nsList.Items); err != nil {
			logging.Error("sync namespaces %s/%s failed, err:%s",
				req.GetProjectCode(), req.GetClusterID(), err.Error())
		}
	}()
	return nil
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
