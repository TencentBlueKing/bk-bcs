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
	pbcontent "bscp.io/pkg/protocol/core/content"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateContent create content.
func (s *Service) CreateContent(ctx context.Context, req *pbds.CreateContentReq) (*pbds.CreateResp, error) {
	kit := kit.FromGrpcContext(ctx)

	content := &table.Content{
		Spec:       req.Spec.ContentSpec(),
		Attachment: req.Attachment.ContentAttachment(),
		Revision: &table.CreatedRevision{
			Creator:   kit.User,
			CreatedAt: time.Now(),
		},
	}
	id, err := s.dao.Content().Create(kit, content)
	if err != nil {
		logs.Errorf("create content failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListContents list contents by query condition.
func (s *Service) ListContents(ctx context.Context, req *pbds.ListContentsReq) (*pbds.ListContentsResp, error) {
	kit := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	query := &types.ListContentsOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Content().List(kit, query)
	if err != nil {
		logs.Errorf("list content failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	resp := &pbds.ListContentsResp{
		Count:   details.Count,
		Details: pbcontent.PbContents(details.Details),
	}
	return resp, nil
}

// queryContentOption query content option.
type queryContentOption struct {
	// ID content id.
	ID uint32
	// BizID content attachment biz id.
	BizID uint32
	// AppID content attachment app id.
	AppID uint32
}

// queryContent query content by option.
func (s *Service) queryContent(kit *kit.Kit, opt *queryContentOption) (*table.Content, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "query content option is nil")
	}

	// build query option.
	query := &types.ListContentsOption{
		BizID: opt.BizID,
		AppID: opt.AppID,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "id",
					Op:    filter.Equal.Factory(),
					Value: opt.ID,
				},
			},
		},
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: 1,
		},
	}
	details, err := s.dao.Content().List(kit, query)
	if err != nil {
		return nil, err
	}

	// if query data is nil by query option, return err.
	if len(details.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "requested content not exist")
	}

	return details.Details[0], nil
}
