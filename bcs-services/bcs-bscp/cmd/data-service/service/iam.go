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

package service

import (
	"context"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ListInstances list instances used to iam pull resource callback.
func (s *Service) ListInstances(ctx context.Context, req *pbds.ListInstancesReq) (*pbds.ListInstancesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opts := &types.ListInstancesOption{
		ResourceType: req.ResourceType,
		ParentType:   req.ParentType,
		ParentID:     req.ParentId,
		Page:         req.Page.BasePage(),
	}

	details, err := s.dao.IAM().ListInstances(kt, opts)
	if err != nil {
		logs.Errorf("list instances failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListInstancesResp{
		Count:   details.Count,
		Details: pbds.PbInstanceResources(details.Details),
	}

	return resp, nil
}

// FetchInstanceInfo used to iam pull resource info callback.
func (s *Service) FetchInstanceInfo(ctx context.Context, req *pbds.FetchInstanceInfoReq) (
	*pbds.FetchInstanceInfoResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opts := &types.FetchInstanceInfoOption{
		ResourceType: req.ResourceType,
		IDs:          req.Ids,
	}

	details, err := s.dao.IAM().FetchInstanceInfo(kt, opts)
	if err != nil {
		logs.Errorf("list instances failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.FetchInstanceInfoResp{
		Details: pbds.PbInstanceInfo(details.Details),
	}

	return resp, nil
}
