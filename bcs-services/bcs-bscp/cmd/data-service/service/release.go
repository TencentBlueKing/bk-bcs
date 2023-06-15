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
	"time"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
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
	// if no config item, return directly.
	if len(cfgItems) == 0 {
		return nil, errors.New("app config items is empty")
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

	if _, err := s.dao.Release().GetByName(grpcKit, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("release name %s already exists", req.Spec.Name)
	}

	// step3: begin transaction to create release and released config item.
	tx, err := s.dao.BeginTx(grpcKit, req.Attachment.BizId)
	if err != nil {
		return nil, err
	}
	// step4: create release, and create release and released config item need to begin tx.
	hook, err := s.dao.ConfigHook().GetByAppID(grpcKit, req.Attachment.BizId, req.Attachment.AppId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get configHook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {

		req.Spec.Hook = &pbrelease.Hook{
			PreHookId:         hook.Spec.PreHookID,
			PreHookReleaseId:  hook.Spec.PreHookReleaseID,
			PostHookId:        hook.Spec.PostHookID,
			PostHookReleaseId: hook.Spec.PostHookReleaseID,
		}
	}

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
	ft, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	query := &types.ListReleasesOption{
		BizID:      req.BizId,
		AppID:      req.AppId,
		Filter:     ft,
		Page:       req.Page.BasePage(),
		Deprecated: req.Deprecated,
	}

	details, err := s.dao.Release().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	releases := pbrelease.PbReleases(details.Details)

	gcrs, err := s.dao.ReleasedGroup().List(grpcKit, &types.ListReleasedGroupsOption{
		BizID: req.BizId,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{
					Field: "app_id",
					Op:    filter.Equal.Factory(),
					Value: req.AppId,
				},
			},
		},
	})
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
