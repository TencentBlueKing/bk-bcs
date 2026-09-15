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

// CreatePodLimitRange directly manages a Pod LimitRange in a shared namespace.
func (a *SharedNamespaceAction) CreatePodLimitRange(ctx context.Context,
	req *proto.CreatePodLimitRangeRequest, resp *proto.PodLimitRangeResponse) error {
	ia := independent.NewIndependentNamespaceAction(a.model)
	return ia.CreatePodLimitRange(ctx, req, resp)
}

// UpdatePodLimitRange directly manages a Pod LimitRange in a shared namespace.
func (a *SharedNamespaceAction) UpdatePodLimitRange(ctx context.Context,
	req *proto.UpdatePodLimitRangeRequest, resp *proto.PodLimitRangeResponse) error {
	ia := independent.NewIndependentNamespaceAction(a.model)
	return ia.UpdatePodLimitRange(ctx, req, resp)
}

// DeletePodLimitRange directly manages a Pod LimitRange in a shared namespace.
func (a *SharedNamespaceAction) DeletePodLimitRange(ctx context.Context,
	req *proto.DeletePodLimitRangeRequest, resp *proto.PodLimitRangeResponse) error {
	ia := independent.NewIndependentNamespaceAction(a.model)
	return ia.DeletePodLimitRange(ctx, req, resp)
}
