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

// ListClients list client
func (s *Service) ListClients(ctx context.Context, req *pbcs.ListClientsReq) (
	*pbcs.ListClientsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	items, err := s.client.DS.ListClients(kt.RpcCtx(), &pbds.ListClientsReq{
		BizId:             req.GetBizId(),
		AppId:             req.GetAppId(),
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
		Search: &pbclient.ClientQueryCondition{
			Uid:                 req.GetSearch().GetUid(),
			Ip:                  req.GetSearch().GetIp(),
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		Order: &pbds.ListClientsReq_Order{
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

	resp := &pbcs.ListClientsResp{
		Count:   items.Count,
		Details: items.Details,
	}

	return resp, nil
}

// ClientPullTrendAnalyze 客户端拉取数量趋势统计
func (s *Service) ClientPullTrendAnalyze(ctx context.Context, req *pbcs.ClientPullTrendAnalyzeReq) (
	*pbcs.ClientPullTrendAnalyzeResp, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	item, err := s.client.DS.ClientPullTrendAnalyze(kt.RpcCtx(), &pbds.ClientPullTrendAnalyzeReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
		},
	})
	if err != nil {
		return nil, err
	}

	return &pbcs.ClientPullTrendAnalyzeResp{Details: item.Details}, nil
}

// ClientStatisticsAnalyze 客户端配置版本统计、拉取成功率统计、失败原因统计、客户端组件信息统计
func (s *Service) ClientStatisticsAnalyze(ctx context.Context, req *pbcs.ClientStatisticsAnalyzeReq) (
	*pbcs.ClientStatisticsAnalyzeResp, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	item, err := s.client.DS.ClientStatisticsAnalyze(kt.RpcCtx(), &pbds.ClientStatisticsAnalyzeReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
		ChartType:         req.GetChartType(),
	})

	if err != nil {
		return nil, err
	}

	return &pbcs.ClientStatisticsAnalyzeResp{
		Details: item.Details,
	}, nil
}

// ClientTagsAndExtraInfoAnalyze 客户端标签、客户端附加信息分布
func (s *Service) ClientTagsAndExtraInfoAnalyze(ctx context.Context, req *pbcs.ClientTagsAndExtraInfoAnalyzeReq) (
	*pbcs.ClientTagsAndExtraInfoAnalyzeResp, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	items, err := s.client.DS.ClientTagsAndExtraInfoAnalyze(kt.RpcCtx(), &pbds.ClientTagsAndExtraInfoAnalyzeReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
		},
	})

	if err != nil {
		return nil, err
	}

	return &pbcs.ClientTagsAndExtraInfoAnalyzeResp{Details: items.Details}, nil
}
