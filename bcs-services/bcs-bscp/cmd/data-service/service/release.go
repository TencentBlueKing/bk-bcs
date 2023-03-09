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

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/protocol/core/release"
	"bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateRelease create release.
func (s *Service) CreateRelease(ctx context.Context, req *pbds.CreateReleaseReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	releasedCIs := make([]*table.ReleasedConfigItem, 0)
	// TODO: need to change batch operator to query config item and it's commit.
	// step1: query app's all config items.
	cfgItems, err := s.queryAppConfigItemList(grpcKit, req.Attachment.BizId, req.Attachment.AppId)
	if err != nil {
		logs.Errorf("query app config item list failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// step2: query config item newest commit
	now := time.Now()
	for _, item := range cfgItems {
		commit, err := s.queryCILatestCommit(grpcKit, req.Attachment.BizId, req.Attachment.AppId, item.ID)
		if err != nil {
			logs.Errorf("query config item latest commit failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}

		releasedCIs = append(releasedCIs, &table.ReleasedConfigItem{
			CommitID:       commit.ID,
			CommitSpec:     commit.Spec,
			ConfigItemID:   item.ID,
			ConfigItemSpec: item.Spec,
			Attachment:     item.Attachment,
			Revision:       item.Revision,
		})
	}

	// step3: begin transaction to create release and released config item.
	tx, err := s.dao.BeginTx(grpcKit, req.Attachment.BizId)
	if err != nil {
		return nil, err
	}
	// step4: create release, and create release and released config item need to begin tx.
	release := &table.Release{
		Spec:       req.Spec.ReleaseSpec(),
		Attachment: req.Attachment.ReleaseAttachment(),
		Revision: &table.CreatedRevision{
			Creator:   grpcKit.User,
			CreatedAt: now,
		},
	}
	id, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		tx.Rollback(grpcKit)
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// step5: create released config item.
	for _, rci := range releasedCIs {
		rci.ReleaseID = release.ID
	}

	if err = s.dao.ReleasedCI().BulkCreateWithTx(grpcKit, tx, releasedCIs); err != nil {
		tx.Rollback(grpcKit)
		logs.Errorf("bulk create released config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// step6: commit transaction.
	if err = tx.Commit(grpcKit); err != nil {
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListReleases list releases.
func (s *Service) ListReleases(ctx context.Context, req *pbds.ListReleasesReq) (*pbds.ListReleasesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	query := &types.ListReleasesOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Release().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.ListReleasesResp{
		Count:   details.Count,
		Details: pbrelease.PbReleases(details.Details),
	}
	return resp, nil
}
