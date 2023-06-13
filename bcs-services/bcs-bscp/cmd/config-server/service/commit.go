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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbcommit "bscp.io/pkg/protocol/core/commit"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateCommit create commit with options
func (s *Service) CreateCommit(ctx context.Context, req *pbcs.CreateCommitReq) (*pbcs.CreateCommitResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateCommitResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Commit, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, authRes)
	if err != nil {
		return nil, err
	}

	r := &pbds.CreateCommitReq{
		Attachment: &pbcommit.CommitAttachment{
			BizId:        req.BizId,
			AppId:        req.AppId,
			ConfigItemId: req.ConfigItemId,
		},
		ContentId: req.ContentId,
		Memo:      req.Memo,
	}
	rp, err := s.client.DS.CreateCommit(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateCommitResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListCommits list commit with filter
func (s *Service) ListCommits(ctx context.Context, req *pbcs.ListCommitsReq) (*pbcs.ListCommitsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListCommitsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Commit, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, authRes)
	if err != nil {
		return nil, err
	}

	if req.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is null")
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	r := &pbds.ListCommitsReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListCommits(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list commits failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListCommitsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
