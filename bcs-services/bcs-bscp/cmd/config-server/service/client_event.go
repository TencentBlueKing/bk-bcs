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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// ListClientEvents List client details query
func (s *Service) ListClientEvents(ctx context.Context, req *pbcs.ListClientEventsReq) (
	*pbcs.ListClientEventsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	items, err := s.client.DS.ListClientEvents(kt.RpcCtx(), &pbds.ListClientEventsReq{
		BizId:       req.GetBizId(),
		AppId:       req.GetAppId(),
		ClientId:    req.GetClientId(),
		All:         req.GetAll(),
		Limit:       req.GetLimit(),
		Start:       req.GetStart(),
		SearchValue: req.GetSearchValue(),
		StartTime:   req.GetStartTime(),
		EndTime:     req.GetEndTime(),
		Order: &pbds.ListClientEventsReq_Order{
			Asc:  req.GetOrder().GetAsc(),
			Desc: req.GetOrder().GetDesc(),
		},
	})
	if err != nil {
		return nil, err
	}

	resp := &pbcs.ListClientEventsResp{
		Count:   items.Count,
		Details: items.Details,
	}

	return resp, nil
}
