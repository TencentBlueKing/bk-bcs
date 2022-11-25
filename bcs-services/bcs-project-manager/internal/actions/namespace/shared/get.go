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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetNamespace implement for GetNamespace interface
func (a *SharedNamespaceAction) GetNamespace(ctx context.Context,
	req *proto.GetNamespaceRequest, resp *proto.GetNamespaceResponse) error {
	staging, err := a.model.GetNamespace(ctx,
		req.GetProjectCode(), req.GetClusterID(), req.GetName(), nsm.ItsmTicketTypeCreate)
	if err != nil && err != drivers.ErrTableRecordNotFound {
		logging.Error("get staging namespace failed, err: %s", err.Error())
		return errorx.NewDBErr(err)
	}
	if err == nil {
		// get staging namespace from db
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
		resp.Data = retData
		return nil
	}
	// get exist namespaces from cluster
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	namespace, err := client.CoreV1().Namespaces().Get(ctx, req.GetName(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logging.Error("get namespace %s in cluster %s failed, err: %s", req.GetName(), req.GetClusterID(), err.Error())
		return err
	}
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, "命名空间不存在")
	}
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
	modifyStagging, err := a.model.GetNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetName(),
		nsm.ItsmTicketTypeUpdate)
	if modifyStagging != nil {

		retData.ItsmTicketType = modifyStagging.ItsmTicketType
		retData.ItsmTicketSN = modifyStagging.ItsmTicketSN
		retData.ItsmTicketStatus = modifyStagging.ItsmTicketStatus
		retData.ItsmTicketURL = modifyStagging.ItsmTicketURL
	}
	deleteStagging, err := a.model.GetNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetName(),
		nsm.ItsmTicketTypeDelete)
	if deleteStagging != nil {

		retData.ItsmTicketType = deleteStagging.ItsmTicketType
		retData.ItsmTicketSN = deleteStagging.ItsmTicketSN
		retData.ItsmTicketStatus = deleteStagging.ItsmTicketStatus
		retData.ItsmTicketURL = deleteStagging.ItsmTicketURL
	}
	resp.Data = retData
	return nil
}
