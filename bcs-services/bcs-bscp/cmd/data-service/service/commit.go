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
	pbcommit "bscp.io/pkg/protocol/core/commit"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateCommit create commit.
func (s *Service) CreateCommit(ctx context.Context, req *pbds.CreateCommitReq) (
	*pbds.CreateResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	opt := &queryContentOption{
		ID:    req.ContentId,
		BizID: req.Attachment.BizId,
		AppID: req.Attachment.AppId,
	}
	content, err := s.queryContent(grpcKit, opt)
	if err != nil {
		logs.Errorf("query content failed, opt: %v, err: %v, rid: %s", opt, err, grpcKit.Rid)
		return nil, err
	}

	commit := &table.Commit{
		Spec: &table.CommitSpec{
			ContentID: req.ContentId,
			Content:   content.Spec,
			Memo:      req.Memo,
		},
		Attachment: req.Attachment.CommitAttachment(),
		Revision: &table.CreatedRevision{
			Creator:   grpcKit.User,
			CreatedAt: time.Now(),
		},
	}
	id, err := s.dao.Commit().Create(grpcKit, commit)
	if err != nil {
		logs.Errorf("create commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListCommits list commits by query condition.
func (s *Service) ListCommits(ctx context.Context, req *pbds.ListCommitsReq) (*pbds.ListCommitsResp,
	error) {

	grpcKit := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	query := &types.ListCommitsOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Commit().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.ListCommitsResp{
		Count:   details.Count,
		Details: pbcommit.PbCommits(details.Details),
	}
	return resp, nil
}

// GetLatestCommit get latest commit by config item id
func (s *Service) GetLatestCommit(ctx context.Context, req *pbds.GetLatestCommitReq) (*pbcommit.Commit, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	commit, err := s.queryCILatestCommit(grpcKit, req.BizId, req.AppId, req.ConfigItemId)
	if err != nil {
		return nil, err
	}

	resp := pbcommit.PbCommit(commit)
	return resp, nil
}

// queryCILatestCommit query config item latest commit.
func (s *Service) queryCILatestCommit(kit *kit.Kit, bizID, appID, ciID uint32) (*table.Commit, error) {
	opt := &types.ListCommitsOption{
		BizID: bizID,
		AppID: appID,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "config_item_id",
					Op:    filter.Equal.Factory(),
					Value: ciID,
				},
			},
		},
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: 1,
			Order: types.Descending,
		},
	}

	details, err := s.dao.Commit().List(kit, opt)
	if err != nil {
		return nil, err
	}

	if len(details.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "commit not exist")
	}

	return details.Details[0], nil
}
