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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcommit "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateCommit create commit.
func (s *Service) CreateCommit(ctx context.Context, req *pbds.CreateCommitReq) (
	*pbds.CreateResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	content, err := s.dao.Content().Get(grpcKit, req.ContentId, req.Attachment.BizId)
	if err != nil {
		logs.Errorf("get content failed, err: %v, rid: %s", err, grpcKit.Rid)
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
			Creator: grpcKit.User,
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

// GetLatestCommit get latest commit by config item id
func (s *Service) GetLatestCommit(ctx context.Context, req *pbds.GetLatestCommitReq) (*pbcommit.Commit, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	commit, err := s.dao.Commit().GetLatestCommit(grpcKit, req.BizId, req.AppId, req.ConfigItemId)
	if err != nil {
		return nil, err
	}

	resp := pbcommit.PbCommit(commit)
	return resp, nil
}
