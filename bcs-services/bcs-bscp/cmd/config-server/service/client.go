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
	pbclient "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// ListClientMetrics list client metrics query
func (s *Service) ListClientMetrics(ctx context.Context, req *pbcs.ListClientMetricsReq) (
	*pbcs.ListClientMetricsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	items, err := s.client.DS.ListClientMetrics(kt.RpcCtx(), &pbds.ListClientMetricsReq{
		BizId:             req.GetBizId(),
		AppId:             req.GetAppId(),
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
		Search: &pbclient.ClientMetricQueryCondition{
			Uid:                 req.GetSearch().GetUid(),
			Ip:                  req.GetSearch().GetIp(),
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		Order: &pbds.ListClientMetricsReq_Order{
			Desc: req.GetOrder().GetDesc(),
			Asc:  req.GetOrder().GetAsc(),
		},
		Start: req.GetStart(),
		Limit: req.GetLimit(),
		All:   req.GetAll(),
	})
	if err != nil {
		return nil, err
	}

	resp := &pbcs.ListClientMetricsResp{
		Count:   items.Count,
		Details: items.Details,
	}

	return resp, nil
}
