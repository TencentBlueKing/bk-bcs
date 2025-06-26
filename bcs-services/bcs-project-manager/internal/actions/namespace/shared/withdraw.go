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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// WithdrawNamespace implement for WithdrawNamespace interface
func (a *SharedNamespaceAction) WithdrawNamespace(ctx context.Context,
	req *proto.WithdrawNamespaceRequest, resp *proto.WithdrawNamespaceResponse) error {
	namespace, err := a.model.GetNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace())
	if err != nil {
		logging.Error("get staging namespace %s/%s failed, err: %s",
			req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	authUser, err := middleware.GetUserFromContext(ctx)
	if err != nil || authUser.GetUsername() != namespace.Creator {
		return errorx.NewReadableErr(errorx.PermDeniedErr, "仅提单人能撤回")
	}
	if err := v2.WithdrawTicket(ctx, authUser.Username, namespace.ItsmTicketSN); err != nil {
		return err
	}
	return a.model.DeleteNamespace(ctx, namespace.ProjectCode, namespace.ClusterID, namespace.Name)
}
