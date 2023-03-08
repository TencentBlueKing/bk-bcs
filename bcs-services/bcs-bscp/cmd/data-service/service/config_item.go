/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"time"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateConfigItem create config item.
func (s *Service) CreateConfigItem(ctx context.Context, req *pbds.CreateConfigItemReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	now := time.Now()
	ci := &table.ConfigItem{
		Spec:       req.Spec.ConfigItemSpec(),
		Attachment: req.Attachment.ConfigItemAttachment(),
		Revision: &table.Revision{
			Creator:   grpcKit.User,
			Reviser:   grpcKit.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	id, err := s.dao.ConfigItem().Create(grpcKit, ci)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.RPCAbortedErr(err)
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// UpdateConfigItem update config item.
func (s *Service) UpdateConfigItem(ctx context.Context, req *pbds.UpdateConfigItemReq) (
	*pbbase.EmptyResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	ci := &table.ConfigItem{
		ID:         req.Id,
		Spec:       req.Spec.ConfigItemSpec(),
		Attachment: req.Attachment.ConfigItemAttachment(),
		Revision: &table.Revision{
			Reviser:   grpcKit.User,
			UpdatedAt: time.Now(),
		},
	}
	if err := s.dao.ConfigItem().Update(grpcKit, ci); err != nil {
		logs.Errorf("update config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteConfigItem delete config item.
func (s *Service) DeleteConfigItem(ctx context.Context, req *pbds.DeleteConfigItemReq) (*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	ci := &table.ConfigItem{
		ID:         req.Id,
		Attachment: req.Attachment.ConfigItemAttachment(),
	}
	if err := s.dao.ConfigItem().Delete(grpcKit, ci); err != nil {
		logs.Errorf("delete config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// GetConfigItem get config item detail
func (s *Service) GetConfigItem(ctx context.Context, req *pbds.GetConfigItemReq) (*pbci.ConfigItem, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	configItem, err := s.dao.ConfigItem().Get(grpcKit, req.Id, req.BizId)
	if err != nil {
		logs.Errorf("get config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	resp := pbci.PbConfigItem(configItem)
	return resp, nil
}

// ListConfigItems list config items by query condition.
func (s *Service) ListConfigItems(ctx context.Context, req *pbds.ListConfigItemsReq) (*pbds.ListConfigItemsResp,
	error) {

	grpcKit := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pn struct to expression failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	query := &types.ListConfigItemsOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.ConfigItem().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.ListConfigItemsResp{
		Count:   details.Count,
		Details: pbci.PbConfigItems(details.Details),
	}
	return resp, nil
}

// queryAppConfigItemList query config item list under specific app.
func (s *Service) queryAppConfigItemList(kit *kit.Kit, bizID, appID uint32) ([]*table.ConfigItem, error) {
	cfgItems := make([]*table.ConfigItem, 0)
	f := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{},
	}

	const step = 200
	for start := uint32(0); ; start += step {
		opt := &types.ListConfigItemsOption{
			BizID:  bizID,
			AppID:  appID,
			Filter: f,
			Page: &types.BasePage{
				Count: false,
				Start: start,
				Limit: step,
				Sort:  "id",
			},
		}

		details, err := s.dao.ConfigItem().List(kit, opt)
		if err != nil {
			return nil, err
		}

		cfgItems = append(cfgItems, details.Details...)

		if len(details.Details) < step {
			break
		}
	}

	return cfgItems, nil
}
