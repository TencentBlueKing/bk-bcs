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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// CreateOtherQuota directly manages an additional ResourceQuota in a shared namespace.
func (a *SharedNamespaceAction) CreateOtherQuota(ctx context.Context,
	req *proto.CreateOtherQuotaRequest, resp *proto.OtherQuotaResponse) error {
	ia := independent.NewIndependentNamespaceAction(a.model)
	return ia.CreateOtherQuota(ctx, req, resp)
}

// UpdateOtherQuota directly manages an additional ResourceQuota in a shared namespace.
func (a *SharedNamespaceAction) UpdateOtherQuota(ctx context.Context,
	req *proto.UpdateOtherQuotaRequest, resp *proto.OtherQuotaResponse) error {
	ia := independent.NewIndependentNamespaceAction(a.model)
	return ia.UpdateOtherQuota(ctx, req, resp)
}

// DeleteOtherQuota directly manages an additional ResourceQuota in a shared namespace.
func (a *SharedNamespaceAction) DeleteOtherQuota(ctx context.Context,
	req *proto.DeleteOtherQuotaRequest, resp *proto.OtherQuotaResponse) error {
	ia := independent.NewIndependentNamespaceAction(a.model)
	return ia.DeleteOtherQuota(ctx, req, resp)
}
