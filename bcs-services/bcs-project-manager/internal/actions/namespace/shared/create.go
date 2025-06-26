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
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NamespacePrefix namespace in shared cluster must be prefixed by it
var NamespacePrefix = envs.BCSNamespacePrefix + "-%s-"

// CreateNamespace implement for CreateNamespace interface
func (a *SharedNamespaceAction) CreateNamespace(ctx context.Context,
	req *proto.CreateNamespaceRequest, resp *proto.CreateNamespaceResponse) error {
	if err := a.validateCreate(ctx, req); err != nil {
		return err
	}
	// if itsm is not enable, create namespace directly
	if !config.GlobalConf.ITSM.Enable {
		ia := independent.NewIndependentNamespaceAction(a.model)
		req.Annotations = append(req.Annotations, &proto.Annotation{
			Key:   config.GlobalConf.SharedClusterConfig.AnnoKeyProjCode,
			Value: req.GetProjectCode(),
		})
		return ia.CreateNamespace(ctx, req, resp)
	}
	var username string
	if authUser, gErr := middleware.GetUserFromContext(ctx); gErr == nil {
		username = authUser.GetUsername()
	}
	namespace := &nsm.Namespace{
		ProjectCode: req.GetProjectCode(),
		ClusterID:   req.GetClusterID(),
		Name:        req.GetName(),
		CreateTime:  time.Now().Format(time.RFC3339),
		Creator:     username,
		Managers:    username,
		ResourceQuota: &nsm.Quota{
			CPURequests:    req.GetQuota().GetCpuRequests(),
			CPULimits:      req.GetQuota().GetCpuLimits(),
			MemoryRequests: req.GetQuota().GetMemoryRequests(),
			MemoryLimits:   req.GetQuota().GetMemoryLimits(),
		},
	}
	variables := []*nsm.Variable{}
	for _, variable := range req.GetVariables() {
		variables = append(variables, &nsm.Variable{
			VariableID: variable.Id,
			ClusterID:  variable.ClusterID,
			Namespace:  variable.Namespace,
			Key:        variable.Key,
			Value:      variable.Value,
		})
	}
	namespace.Variables = variables
	cpuLimits, err := resource.ParseQuantity(req.GetQuota().GetCpuLimits())
	if err != nil {
		logging.Error("parse quantity cpu limits failed, err: %s", err.Error())
		return err
	}
	memoryLimits, err := resource.ParseQuantity(req.GetQuota().GetMemoryLimits())
	if err != nil {
		logging.Error("parse quantity cpu limits failed, err: %s", err.Error())
		return err
	}
	// memoryLimits.Value() return unit is byte， needs to be converted to Gi (divide 2^30)
	// TODO ITSM 传入 tenantID
	itsmResp, err := v2.SubmitCreateNamespaceTicket(ctx, username, req.GetProjectCode(), req.GetClusterID(), req.GetName(),
		int(cpuLimits.Value()), int(memoryLimits.Value()/int64(math.Pow(2, 30))))
	if err != nil {
		logging.Error("itsm create ticket failed, err: %s", err.Error())
		return err
	}
	namespace.ItsmTicketType = nsm.ItsmTicketTypeCreate
	namespace.ItsmTicketURL = itsmResp.TicketURL
	namespace.ItsmTicketStatus = nsm.ItsmTicketStatusCreated
	namespace.ItsmTicketSN = itsmResp.SN
	if err := a.model.CreateNamespace(ctx, namespace); err != nil {
		logging.Error("create namespace %s/%s in db failed, err: %s", req.GetClusterID(), req.GetName(), err.Error())
		return err
	}
	return nil
}

func (a *SharedNamespaceAction) validateCreate(ctx context.Context, req *proto.CreateNamespaceRequest) error {
	// check is namespace name valid
	if !strings.HasPrefix(req.Name, fmt.Sprintf(NamespacePrefix, req.ProjectCode)) {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("共享集群命名空间必须以 %s-[projectCode]- 开头",
			envs.BCSNamespacePrefix))
	}
	// check is namespace name exists
	stagings, _ := a.model.ListNamespacesByItsmTicketType(ctx,
		req.ProjectCode, req.ClusterID, []string{nsm.ItsmTicketTypeCreate})
	for _, staging := range stagings {
		if staging.Name == req.Name {
			return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("命名空间 [%s] 已存在", req.Name))
		}
	}
	client, err := clientset.GetClientGroup().Client(req.ClusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.ClusterID, err.Error())
		return err
	}
	_, err = client.CoreV1().Namespaces().Get(ctx, req.Name, metav1.GetOptions{})
	if err == nil {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("命名空间 [%s] 已存在", req.Name))
	}
	if !errors.IsNotFound(err) {
		logging.Error("get namespace in cluster %s failed, err: %s", req.ClusterID, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// check resourceQuota validate
	if err := quotautils.ValidateResourceQuota(req.Quota); err != nil {
		return err
	}
	return nil
}
