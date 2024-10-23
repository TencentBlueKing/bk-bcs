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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbaudit "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/audit"
	pbstrategy "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/strategy"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ListAudits list audits.
func (s *Service) ListAudits(ctx context.Context, req *pbds.ListAuditsReq) (*pbds.ListAuditsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	if !req.All && req.AppId == 0 {
		return nil, fmt.Errorf("app_id must have a value: %d", req.AppId)
	}

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	aas, count, err := s.dao.AuditDao().ListAuditsAppStrategy(grpcKit, req)
	if err != nil {
		return nil, err
	}

	var details []*pbaudit.ListAuditsAppStrategy
	for _, value := range aas {
		details = append(details, &pbaudit.ListAuditsAppStrategy{
			Audit: &pbaudit.Audit{
				Id: value.Audit.ID,
				Spec: &pbaudit.AuditSpec{
					ResType:     value.Audit.ResourceType,
					Action:      value.Audit.Action,
					Rid:         "", // 暂时用不到
					AppCode:     "", // 暂时用不到
					Detail:      "", // 暂时用不到
					Operator:    value.Audit.Operator,
					ResInstance: value.Audit.ResInstance,
					OperateWay:  value.Audit.OperateWay,
					Status:      value.Audit.Status,
					IsCompare:   value.Audit.IsCompare,
				},
				Attachment: &pbaudit.AuditAttachment{
					BizId: value.Audit.BizID,
					AppId: value.Audit.AppID,
					ResId: value.Audit.ResourceID,
				},
				Revision: &pbaudit.Revision{
					CreatedAt: value.Audit.CreatedAt.Format(time.DateTime),
				},
			},
			Strategy: &pbstrategy.AuditStrategy{
				PublishType:      value.Strategy.PublishType,
				PublishTime:      value.Strategy.PublishTime,
				PublishStatus:    value.Strategy.PublishStatus,
				RejectReason:     value.Strategy.RejectReason,
				Approver:         value.Strategy.Approver,
				ApproverProgress: value.Strategy.ApproverProgress,
				UpdatedAt:        value.Strategy.UpdatedAt.Format(time.DateTime),
				Reviser:          value.Strategy.Reviser,
				ReleaseId:        value.Strategy.ReleaseId,
				Scope:            pbstrategy.PbScope(&value.Strategy.Scope),
				Creator:          value.Strategy.Creator,
			},
			App: &pbapp.AuditApp{
				Name:    value.App.Name,
				Creator: value.App.Creator,
			},
		})
	}

	resp := &pbds.ListAuditsResp{
		Count:   uint32(count),
		Details: details,
	}

	return resp, nil
}
