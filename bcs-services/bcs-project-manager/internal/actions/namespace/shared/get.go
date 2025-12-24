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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetNamespace implement for GetNamespace interface
func (a *SharedNamespaceAction) GetNamespace(ctx context.Context,
	req *proto.GetNamespaceRequest, resp *proto.GetNamespaceResponse) error {
	// if itsm is not enable, get namespace directly
	if !config.GlobalConf.ITSM.Enable {
		ia := independent.NewIndependentNamespaceAction(a.model)
		return ia.GetNamespace(ctx, req, resp)
	}
	projectCode := req.GetProjectCode()
	clusterID := req.GetClusterID()
	name := req.GetNamespace()
	// get approving namespace from db
	staging, err := a.model.GetNamespace(ctx, projectCode, clusterID, name)
	if err != nil && err != drivers.ErrTableRecordNotFound {
		logging.Error("get namespace %s/%s/%s failed, err: %s", projectCode, clusterID, name, err.Error())
		return err
	}
	if staging != nil && staging.ItsmTicketType == nsm.ItsmTicketTypeCreate {
		resp.Data = constructCreatingNamespace(staging)
		return nil
	}
	// get exist namespaces from cluster
	client, err := clientset.GetClientGroup().Client(clusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", clusterID, err.Error())
		return err
	}
	namespace, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logging.Error("get namespace %s in cluster %s failed, err: %s",
			name, clusterID, err.Error())
		return err
	}
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, "命名空间不存在")
	}
	retData := &proto.NamespaceData{
		Name:       namespace.GetName(),
		Uid:        string(namespace.GetUID()),
		CreateTime: namespace.GetCreationTimestamp().UTC().Format(time.RFC3339),
		Status:     string(namespace.Status.Phase),
		IsSystem:   stringx.StringInSlice(namespace.GetName(), config.GlobalConf.SystemConfig.SystemNameSpaces),
	}
	// get quota
	// nolint
	if quota, err := getNamespaceQuota(ctx, clusterID, namespace.GetName(), client); err != nil {
		return err
	} else if quota != nil {
		retData.Quota, retData.Used, retData.CpuUseRate, retData.MemoryUseRate = quotautils.TransferToProto(quota)
	}
	// get variables
	variables, err := listNamespaceVariables(ctx, projectCode, clusterID, namespace.GetName())
	if err != nil {
		logging.Error("get namespace %s/%s variables failed, err: %s",
			clusterID, namespace.GetName(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	retData.Variables = variables
	// get managers
	managers := []string{}
	if creator, exists := namespace.Annotations[constant.AnnotationKeyCreator]; exists {
		managers = append(managers, creator)
	} else {
		cluster, err := clustermanager.GetCluster(ctx, req.ClusterID)
		if err != nil {
			return err
		}
		managers = append(managers, cluster.Creator)
	}
	retData.Managers = managers
	if staging != nil {
		retData.ItsmTicketType = staging.ItsmTicketType
		retData.ItsmTicketSN = staging.ItsmTicketSN
		retData.ItsmTicketStatus = staging.ItsmTicketStatus
		retData.ItsmTicketURL = staging.ItsmTicketURL
	}
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

	if errors.IsNotFound(err) {
		return nil, nil
	}
	return quota, nil
}

func constructCreatingNamespace(staging *nsm.Namespace) *proto.NamespaceData {
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
	// get managers
	retData.Managers = []string{staging.Creator}
	retData.Variables = variables
	retData.ItsmTicketSN = staging.ItsmTicketSN
	retData.ItsmTicketStatus = staging.ItsmTicketStatus
	retData.ItsmTicketURL = staging.ItsmTicketURL
	retData.ItsmTicketType = staging.ItsmTicketType
	return retData
}
