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

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
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
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
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
			ClientVersion:       req.GetSearch().GetClientVersion(),
			StartPullTime:       req.GetSearch().GetStartPullTime(),
			EndPullTime:         req.GetSearch().GetEndPullTime(),
			ClientType:          req.GetSearch().GetClientType(),
			FailedReason:        req.GetSearch().GetFailedReason(),
			ClientIds:           req.GetSearch().GetClientIds(),
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

	var details []*pbcs.ListClientsResp_Item

	for _, v := range items.GetDetails() {
		details = append(details, &pbcs.ListClientsResp_Item{
			Client:            v.GetClient(),
			CpuUsageStr:       v.GetCpuUsageStr(),
			CpuMaxUsageStr:    v.GetCpuMaxUsageStr(),
			MemoryUsageStr:    v.GetMemoryUsageStr(),
			MemoryMaxUsageStr: v.GetMemoryMaxUsageStr(),
		})
	}

	resp := &pbcs.ListClientsResp{
		Count:   items.Count,
		Details: details,
	}

	return resp, nil
}

// ClientConfigVersionStatistics 客户端配置版本统计
func (s *Service) ClientConfigVersionStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientConfigVersionStatistics(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
	})
}

// ClientPullTrendStatistics 客户端拉取趋势统计
func (s *Service) ClientPullTrendStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientPullTrendStatistics(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
		PullTime:          req.GetPullTime(),
		IsDuplicates:      req.GetIsDuplicates(),
	})
}

// ClientPullStatistics 客户端拉取信息统计
func (s *Service) ClientPullStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientPullStatistics(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
	})
}

// ClientLabelStatistics 客户端标签统计
func (s *Service) ClientLabelStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientLabelStatistics(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
		PrimaryKey:        req.GetPrimaryKey(),
		ForeignKeys:       req.GetForeignKeys(),
	})
}

// ClientAnnotationStatistics 客户端附加信息统计
func (s *Service) ClientAnnotationStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientAnnotationStatistics(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
	})
}

// ClientVersionStatistics 客户端版本统计
func (s *Service) ClientVersionStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientVersionStatistics(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
	})
}

// ListClientLabelAndAnnotation 列出客户端标签和注释
func (s *Service) ListClientLabelAndAnnotation(ctx context.Context, req *pbcs.ListClientLabelAndAnnotationReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ListClientLabelAndAnnotation(kt.RpcCtx(), &pbds.ListClientLabelAndAnnotationReq{
		BizId:             req.GetBizId(),
		AppId:             req.GetAppId(),
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
	})
}

// ClientSpecificFailedReason 统计客户端失败详细原因
func (s *Service) ClientSpecificFailedReason(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	return s.client.DS.ClientSpecificFailedReason(kt.RpcCtx(), &pbclient.ClientCommonReq{
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
		Search: &pbclient.ClientQueryCondition{
			Label:               req.GetSearch().GetLabel(),
			CurrentReleaseName:  req.GetSearch().GetCurrentReleaseName(),
			TargetReleaseName:   req.GetSearch().GetTargetReleaseName(),
			ReleaseChangeStatus: req.GetSearch().GetReleaseChangeStatus(),
			Annotations:         req.GetSearch().GetAnnotations(),
			ClientVersion:       req.GetSearch().GetClientVersion(),
			ClientType:          req.GetSearch().GetClientType(),
			FailedReason:        req.GetSearch().GetFailedReason(),
			OnlineStatus:        req.GetSearch().GetOnlineStatus(),
		},
		LastHeartbeatTime: req.GetLastHeartbeatTime(),
	})
}

// RetryClients 重试客户端执行版本变更回调
func (s *Service) RetryClients(ctx context.Context, req *pbcs.RetryClientsReq) (*pbcs.RetryClientsResp, error) {

	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Publish, ResourceID: req.AppId}, BizID: req.BizId},
	}

	if err := s.authorizer.Authorize(kt, res...); err != nil {
		return nil, err
	}

	if !req.All && len(req.ClientIds) == 0 {
		return nil, errf.Errorf(errf.InvalidArgument, i18n.T(kt, "client ids is empty"))
	}

	_, err := s.client.DS.RetryClients(kt.RpcCtx(), &pbds.RetryClientsReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		All:       req.All,
		ClientIds: req.ClientIds,
	})

	if err != nil {
		return nil, err
	}

	return &pbcs.RetryClientsResp{}, nil

}
