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

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// DeleteNamespace implement for DeleteNamespace interface
func (a *SharedNamespaceAction) DeleteNamespace(ctx context.Context,
	req *proto.DeleteNamespaceRequest, resp *proto.DeleteNamespaceResponse) error {
	// if itsm is not enable, delete namespace directly
	if !config.GlobalConf.ITSM.Enable {
		ia := independent.NewIndependentNamespaceAction(a.model)
		return ia.DeleteNamespace(ctx, req, resp)
	}
	var username string
	if authUser, err := middleware.GetUserFromContext(ctx); err == nil {
		username = authUser.GetUsername()
	}
	namespace := &nsm.Namespace{
		ProjectCode: req.GetProjectCode(),
		ClusterID:   req.GetClusterID(),
		Name:        req.GetNamespace(),
		Creator:     username,
	}
	itsmResp, err := itsm.SubmitDeleteNamespaceTicket(username,
		req.GetProjectCode(), req.GetClusterID(), req.GetNamespace())
	if err != nil {
		logging.Error("itsm create ticket failed, err: %s", err.Error())
		return err
	}
	namespace.ItsmTicketType = nsm.ItsmTicketTypeDelete
	namespace.ItsmTicketURL = itsmResp.TicketURL
	namespace.ItsmTicketStatus = nsm.ItsmTicketStatusCreated
	namespace.ItsmTicketSN = itsmResp.SN
	if err := a.model.CreateNamespace(ctx, namespace); err != nil {
		logging.Error("create namespace %s/%s in db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return err
	}
	return nil
}
