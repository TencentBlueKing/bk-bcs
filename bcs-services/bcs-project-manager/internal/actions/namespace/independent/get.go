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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetNamespace implement for GetNamespace interface
func (c *IndependentNamespaceAction) GetNamespace(ctx context.Context,
	req *proto.GetNamespaceRequest, resp *proto.GetNamespaceResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	ns, err := client.CoreV1().Namespaces().Get(ctx, req.GetNamespace(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return errorx.NewClusterErr(err.Error())
	}
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, "命名空间不存在")
	}
	retData := &proto.NamespaceData{
		Name:        ns.GetName(),
		Uid:         string(ns.GetUID()),
		Status:      string(ns.Status.Phase),
		CreateTime:  ns.GetCreationTimestamp().Format(constant.TimeLayout),
		Labels:      []*proto.Label{},
		Annotations: []*proto.Annotation{},
	}
	for k, v := range ns.Labels {
		retData.Labels = append(retData.Labels, &proto.Label{Key: k, Value: v})
	}
	for k, v := range ns.Annotations {
		retData.Annotations = append(retData.Annotations, &proto.Annotation{Key: k, Value: v})
	}
	// get quota
	quota, err := getNamespaceQuota(ctx, req.GetClusterID(), ns.GetName(), client)
	if err != nil {
		return err
	}
	if quota != nil {
		retData.Quota, retData.Used, retData.CpuUseRate, retData.MemoryUseRate = quotautils.TransferToProto(quota)
	}

	// get variables
	variables, err := listNamespaceVariables(ctx, req.GetProjectCode(), req.GetClusterID(), ns.GetName())
	if err != nil {
		logging.Error("get namespace %s/%s variables failed, err: %s", req.GetClusterID(), ns.GetName(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	// get managers
	managers := []string{}
	if creator, exists := ns.Annotations[constant.AnnotationKeyCreator]; exists {
		managers = append(managers, creator)
	} else {
		cluster, err := clustermanager.GetCluster(req.ClusterID)
		if err != nil {
			return err
		}
		managers = append(managers, cluster.Creator)
	}
	retData.Managers = managers
	retData.Variables = variables
	resp.Data = retData
	return nil
}

func getNamespaceQuota(ctx context.Context, clusterID, namespace string, clientset *kubernetes.Clientset) (
	*corev1.ResourceQuota, error) {
	quota, err := clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, namespace, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logging.Error("get resourceQuota %s/%s failed, err: %s", clusterID, namespace, err.Error())
		return nil, errorx.NewClusterErr(err.Error())
	}

	// get quota
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
