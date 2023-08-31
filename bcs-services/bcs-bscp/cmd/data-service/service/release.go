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
	"errors"
	"fmt"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateRelease create release.
func (s *Service) CreateRelease(ctx context.Context, req *pbds.CreateReleaseReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	releasedCIs := make([]*table.ReleasedConfigItem, 0)
	// Note: need to change batch operator to query config item and it's commit.
	// step1: query app's all config items.
	cfgItems, err := s.dao.ConfigItem().ListAllByAppID(grpcKit, req.Attachment.AppId, req.Attachment.BizId)
	if err != nil {
		logs.Errorf("query app config item list failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// if no config item, return directly.
	if len(cfgItems) == 0 {
		return nil, errors.New("app config items is empty")
	}
	// step2: query config item newest commit
	for _, item := range cfgItems {
		commit, e := s.dao.Commit().GetLatestCommit(grpcKit, req.Attachment.BizId, req.Attachment.AppId, item.ID)
		if e != nil {
			logs.Errorf("query config item latest commit failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
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
	if _, e := s.dao.Release().GetByName(grpcKit, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Name); e == nil {
		return nil, fmt.Errorf("release name %s already exists", req.Spec.Name)
	}
	// step3: begin transaction to create release and released config item.
	tx := s.dao.GenQuery().Begin()
	// step4: create release, and create release and released config item need to begin tx.
	release := &table.Release{
		Spec:       req.Spec.ReleaseSpec(),
		Attachment: req.Attachment.ReleaseAttachment(),
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}
	id, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		tx.Rollback()
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// step5: create released hook.
	pre, err := s.dao.ReleasedHook().Get(grpcKit, req.Attachment.BizId, req.Attachment.AppId, 0, table.PreHook)
	if err == nil {
		pre.ID = 0
		pre.ReleaseID = release.ID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, pre); e != nil {
			logs.Errorf("create released pre-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			tx.Rollback()
			return nil, e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released pre-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback()
		return nil, err
	}
	post, err := s.dao.ReleasedHook().Get(grpcKit, req.Attachment.BizId, req.Attachment.AppId, 0, table.PostHook)
	if err == nil {
		post.ID = 0
		post.ReleaseID = release.ID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, post); e != nil {
			logs.Errorf("create released post-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			tx.Rollback()
			return nil, e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released post-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback()
		return nil, err
	}
	// step6: create released config item.
	for _, rci := range releasedCIs {
		rci.ReleaseID = release.ID
	}
	if err = s.dao.ReleasedCI().BulkCreateWithTx(grpcKit, tx, releasedCIs); err != nil {
		tx.Rollback()
		logs.Errorf("bulk create released config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// step7: commit transaction.
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	return &pbds.CreateResp{Id: id}, nil
}

// ListReleases list releases.
func (s *Service) ListReleases(ctx context.Context, req *pbds.ListReleasesReq) (*pbds.ListReleasesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	query := &types.ListReleasesOption{
		BizID: req.BizId,
		AppID: req.AppId,
		Page: &types.BasePage{
			Start: req.Start,
			Limit: uint(req.Limit),
		},
		Deprecated: req.Deprecated,
		SearchKey:  req.SearchKey,
	}
	if req.All {
		query.Page.Start = 0
		query.Page.Limit = 0
	}

	details, err := s.dao.Release().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	releases := pbrelease.PbReleases(details.Details)

	gcrs, err := s.dao.ReleasedGroup().ListAllByAppID(grpcKit, req.AppId, req.BizId)
	if err != nil {
		logs.Errorf("list group current releases failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	groups, err := s.dao.Group().ListAppGroups(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list app groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	for _, release := range releases {
		status, selected := s.queryPublishStatus(gcrs, release.Id)
		releasedGroups := make([]*pbrelease.ReleaseStatus_ReleasedGroup, 0)
		for _, gcr := range selected {
			if gcr.GroupID == 0 {
				releasedGroups = append(releasedGroups, &pbrelease.ReleaseStatus_ReleasedGroup{
					Id:   0,
					Name: "默认分组",
					Mode: table.Default.String(),
				})
			}
			for _, group := range groups {
				if group.ID == gcr.GroupID {
					oldSelector := new(pbstruct.Struct)
					newSelector := new(pbstruct.Struct)
					if gcr.Selector != nil {
						s, err := gcr.Selector.MarshalPB()
						if err != nil {
							return nil, err
						}
						oldSelector = s
					}
					if group.Spec.Selector != nil {
						s, err := group.Spec.Selector.MarshalPB()
						if err != nil {
							return nil, err
						}
						newSelector = s
					}
					releasedGroups = append(releasedGroups, &pbrelease.ReleaseStatus_ReleasedGroup{
						Id:          group.ID,
						Name:        group.Spec.Name,
						Mode:        gcr.Mode.String(),
						OldSelector: oldSelector,
						NewSelector: newSelector,
						Edited:      gcr.Edited,
					})
					break
				}
			}
		}
		release.Status = &pbrelease.ReleaseStatus{
			PublishStatus:  status,
			ReleasedGroups: releasedGroups,
		}
	}

	resp := &pbds.ListReleasesResp{
		Count:   details.Count,
		Details: releases,
	}
	return resp, nil
}

// GetReleaseByName get release by release name.
func (s *Service) GetReleaseByName(ctx context.Context, req *pbds.GetReleaseByNameReq) (*pbrelease.Release, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	release, err := s.dao.Release().GetByName(grpcKit, req.GetBizId(), req.GetAppId(), req.GetReleaseName())
	if err != nil {
		logs.Errorf("get release by name failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, fmt.Errorf("query release by name %s failed", req.GetReleaseName())
	}

	return pbrelease.PbRelease(release), nil
}

func (s *Service) queryPublishStatus(gcrs []*table.ReleasedGroup, releaseID uint32) (
	string, []*table.ReleasedGroup) {
	var includeDefault = false
	var inRelease = make([]*table.ReleasedGroup, 0)
	var outRelease = make([]*table.ReleasedGroup, 0)
	for _, gcr := range gcrs {
		if gcr.ReleaseID == releaseID {
			inRelease = append(inRelease, gcr)
			if gcr.GroupID == 0 {
				includeDefault = true
			}
		} else {
			outRelease = append(outRelease, gcr)
		}
	}

	// len(inRelease) == 0: not released
	if len(inRelease) == 0 {
		return table.NotReleased.String(), inRelease
		// len(inRelease) != 0 && len(outRelease) != 0: gray released
	} else if len(outRelease) != 0 {
		return table.PartialReleased.String(), inRelease
		// len(inRelease) != 0 && len(outRelease) == 0 && includeDefault: full released
	} else if includeDefault {
		return table.FullReleased.String(), inRelease
		// len(inRelease) != 0 && len(outRelease) == 0 && !includeDefault: gray released
	}
	return table.PartialReleased.String(), inRelease
}
