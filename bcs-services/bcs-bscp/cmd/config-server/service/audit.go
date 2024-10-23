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
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// ListAudits list audits
func (s *Service) ListAudits(ctx context.Context, req *pbcs.ListAuditsReq) (
	*pbcs.ListAuditsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Find, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListAuditsReq{
		BizId:       req.BizId,
		AppId:       req.AppId,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Operate:     req.Operate,
		OperateWay:  req.OperateWay,
		Start:       req.Start,
		Limit:       req.Limit,
		All:         req.All,
		Name:        req.Name,
		ResInstance: req.ResInstance,
		Operator:    req.Operator,
		Id:          req.Id,
	}
	// 前端组件以逗号分开
	if req.Action != "" {
		r.Action = strings.Split(req.Action, ",")
	}
	if req.Status != "" {
		r.Status = strings.Split(req.Status, ",")
	}
	if req.ResourceType != "" {
		r.ResourceType = strings.Split(req.ResourceType, ",")
	}
	rp, err := s.client.DS.ListAudits(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListAuditsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
