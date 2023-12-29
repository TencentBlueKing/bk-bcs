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
	"reflect"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbgroup "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/group"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateGroup create group.
func (s *Service) CreateGroup(ctx context.Context, req *pbds.CreateGroupReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.GroupSpec()
	if err != nil {
		logs.Errorf("get group spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if !req.Spec.Public && len(req.Spec.BindApps) == 0 {
		logs.Errorf("group must bind apps when public is set to false, rid: %s", kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "group must bind apps when public is set to false")
	}
	if req.Spec.Public && len(req.Spec.BindApps) > 0 {
		logs.Errorf("group must not bind apps when public is set to true, rid: %s", kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "group must not bind apps when public is set to true")
	}

	group := &table.Group{
		Spec:       spec,
		Attachment: req.Attachment.GroupAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	tx := s.dao.GenQuery().Begin()
	id, err := s.dao.Group().CreateWithTx(kt, tx, group)
	if err != nil {
		logs.Errorf("create group failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}
	if len(req.Spec.BindApps) != 0 {
		groupApps := make([]*table.GroupAppBind, len(req.Spec.BindApps))
		for idx, app := range req.Spec.BindApps {
			groupApps[idx] = &table.GroupAppBind{
				GroupID: id,
				AppID:   app,
				BizID:   req.Attachment.BizId,
			}
		}
		if e := s.dao.GroupAppBind().BatchCreateWithTx(kt, tx, groupApps); e != nil {
			logs.Errorf("create group app failed, err: %v, rid: %s", e, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, e
		}
	}
	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListAllGroups list all groups in biz.
func (s *Service) ListAllGroups(ctx context.Context, req *pbds.ListAllGroupsReq) (*pbds.ListAllGroupsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	resp := new(pbds.ListAllGroupsResp)

	details, err := s.dao.Group().ListAll(kt, req.BizId)
	if err != nil {
		logs.Errorf("list group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(details) == 0 {
		return resp, nil
	}

	groups, err := pbgroup.PbGroups(details)
	if err != nil {
		logs.Errorf("get pb group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	groupIDs := make([]uint32, len(groups))
	for idx, group := range groups {
		groupIDs[idx] = group.Id
	}

	list, err := s.dao.GroupAppBind().BatchListByGroupIDs(kt, req.BizId, groupIDs)
	if err != nil {
		logs.Errorf("list group app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, group := range groups {
		for _, app := range list {
			if group.Id == app.GroupID {
				group.Spec.BindApps = append(group.Spec.BindApps, app.AppID)
			}
		}
	}

	resp.Details = groups
	return resp, nil
}

// ListAppGroups list groups in app.
func (s *Service) ListAppGroups(ctx context.Context, req *pbds.ListAppGroupsReq) (*pbds.ListAppGroupsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	groups, err := s.dao.Group().ListAppGroups(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list app groups failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	groups = append(groups, &table.Group{
		ID: 0,
		Spec: &table.GroupSpec{
			Name:     "默认分组",
			Public:   true,
			Mode:     table.Default,
			Selector: new(selector.Selector),
			UID:      "",
		},
	})

	gcrs, err := s.dao.ReleasedGroup().ListAllByAppID(kt, req.AppId, req.BizId)
	if err != nil {
		logs.Errorf("list released group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	releaseMap := make(map[uint32]*table.Release, 0)
	for _, gcr := range gcrs {
		releaseMap[gcr.ReleaseID] = nil
	}
	releaseIDs := make([]uint32, 0)
	for releaseID := range releaseMap {
		releaseIDs = append(releaseIDs, releaseID)
	}

	releases, err := s.dao.Release().ListAllByIDs(kt, releaseIDs, req.BizId)
	if err != nil {
		logs.Errorf("list app releases failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, release := range releases {
		if _, ok := releaseMap[release.ID]; ok {
			releaseMap[release.ID] = release
		}
	}
	details := make([]*pbds.ListAppGroupsResp_ListAppGroupsData, 0)
	for _, group := range groups {
		oldSelector := new(structpb.Struct)
		newSelector := new(structpb.Struct)
		if group.Spec.Selector != nil {
			newSelector, err = group.Spec.Selector.MarshalPB()
			if err != nil {
				logs.Errorf("marshal selector failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
		data := &pbds.ListAppGroupsResp_ListAppGroupsData{
			GroupId:     group.ID,
			GroupName:   group.Spec.Name,
			ReleaseId:   0,
			ReleaseName: "",
			OldSelector: oldSelector,
			NewSelector: newSelector,
			Edited:      false,
		}
		for _, gcr := range gcrs {
			if group.ID == gcr.GroupID {
				data.ReleaseId = gcr.ReleaseID
				data.Edited = gcr.Edited
				if gcr.Selector != nil {
					oldSelector, err = gcr.Selector.MarshalPB()
					if err != nil {
						logs.Errorf("marshal selector failed, err: %v, rid: %s", err, kt.Rid)
						return nil, err
					}
					data.NewSelector = oldSelector
				}
				if release, ok := releaseMap[gcr.ReleaseID]; ok && release != nil {
					data.ReleaseName = release.Spec.Name
				}
				break
			}
		}
		details = append(details, data)
	}
	return &pbds.ListAppGroupsResp{
		Details: details,
	}, nil
}

// GetGroupByName get group by group name.
func (s *Service) GetGroupByName(ctx context.Context, req *pbds.GetGroupByNameReq) (*pbgroup.Group, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	group, err := s.dao.Group().GetByName(grpcKit, req.GetBizId(), req.GetGroupName())
	if err != nil {
		logs.Errorf("get group by name failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, fmt.Errorf("query group by name %s failed", req.GetGroupName())
	}

	return pbgroup.PbGroup(group)
}

// UpdateGroup update group.
// nolint: funlen
func (s *Service) UpdateGroup(ctx context.Context, req *pbds.UpdateGroupReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.GroupSpec()
	if err != nil {
		logs.Errorf("get group spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if !req.Spec.Public && len(req.Spec.BindApps) == 0 {
		logs.Errorf("group must bind apps when public is set to false, rid: %s", kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "group must bind apps when public is set to false")
	}
	if req.Spec.Public && len(req.Spec.BindApps) > 0 {
		logs.Errorf("group must not bind apps when public is set to true, rid: %s", kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "group must not bind apps when public is set to true")
	}

	n := &table.Group{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.GroupAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}

	old, err := s.dao.Group().Get(kt, req.Id, req.Attachment.BizId)
	if err != nil {
		return nil, err
	}

	if !n.Spec.Public {
		// check if the reduced app was already released.
		apps, e := s.queryReducedApps(kt, old, req)
		if e != nil {
			return nil, e
		}
		published, e := s.dao.ReleasedGroup().ListAllByGroupID(kt, req.Id, req.Attachment.BizId)
		if e != nil {
			return nil, e
		}

		for _, app := range apps {
			for _, p := range published {
				if app.ID == p.AppID {
					return nil, errf.New(errf.ErrGroupAlreadyPublished,
						fmt.Sprintf("group has already published in app [%s]", app.Spec.Name))
				}
			}
		}
	}

	tx := s.dao.GenQuery().Begin()
	if e := s.dao.Group().UpdateWithTx(kt, tx, n); e != nil {
		logs.Errorf("update group failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if e := s.dao.GroupAppBind().BatchDeleteByGroupIDWithTx(kt, tx, req.Id, req.Attachment.BizId); e != nil {
		logs.Errorf("delete group app failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if !n.Spec.Public {
		groupApps := make([]*table.GroupAppBind, len(req.Spec.BindApps))
		for idx, app := range req.Spec.BindApps {
			groupApps[idx] = &table.GroupAppBind{
				GroupID: req.Id,
				AppID:   app,
				BizID:   req.Attachment.BizId,
			}
		}
		if e := s.dao.GroupAppBind().BatchCreateWithTx(kt, tx, groupApps); e != nil {
			logs.Errorf("create group app failed, err: %v, rid: %s", e, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, e
		}
	}

	var edited = false
	if old.Spec.UID != n.Spec.UID {
		edited = true
	}

	if !edited && !reflect.DeepEqual(old.Spec.Selector, n.Spec.Selector) {
		edited = true
	}

	if edited {
		if e := s.dao.ReleasedGroup().UpdateEditedStatusWithTx(kt, tx,
			edited, req.Id, req.Attachment.BizId); e != nil {
			logs.Errorf("update group current release failed, err: %v, rid: %s", e, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, e
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return &pbbase.EmptyResp{}, nil
}

// DeleteGroup delete group.
func (s *Service) DeleteGroup(ctx context.Context, req *pbds.DeleteGroupReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	group := &table.Group{
		ID:         req.Id,
		Attachment: req.Attachment.GroupAttachment(),
	}

	// check if the group was already released in any app.
	published, err := s.dao.ReleasedGroup().ListAllByGroupID(kt, req.Id, req.Attachment.BizId)
	if err != nil {
		return nil, err
	}
	publishedApps := make([]uint32, len(published))
	for idx, app := range published {
		publishedApps[idx] = app.AppID
	}
	if len(published) > 0 {
		return nil, errf.New(errf.ErrGroupAlreadyPublished,
			fmt.Sprintf("group has already published in apps [%s]", tools.JoinUint32(publishedApps, ",")))
	}
	tx := s.dao.GenQuery().Begin()
	if e := s.dao.Group().DeleteWithTx(kt, tx, group); e != nil {
		logs.Errorf("delete group failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if e := s.dao.GroupAppBind().BatchDeleteByGroupIDWithTx(kt, tx, req.Id, req.Attachment.BizId); e != nil {
		logs.Errorf("delete group app failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// ListGroupReleasedApps list group's published apps and their release.
func (s *Service) ListGroupReleasedApps(ctx context.Context, req *pbds.ListGroupReleasedAppsReq) (
	*pbds.ListGroupReleasedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	resp, err := s.dao.Group().ListGroupReleasedApps(kt, &types.ListGroupReleasedAppsOption{
		BizID:     req.BizId,
		GroupID:   req.GroupId,
		SearchKey: req.SearchKey,
		Start:     req.Start,
		Limit:     req.Limit,
	})
	if err != nil {
		logs.Errorf("list groups published apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	data := make([]*pbds.ListGroupReleasedAppsResp_ListGroupReleasedAppsData, len(resp.Details))
	for idx, detail := range resp.Details {
		data[idx] = &pbds.ListGroupReleasedAppsResp_ListGroupReleasedAppsData{
			AppId:       detail.AppID,
			AppName:     detail.AppName,
			ReleaseId:   detail.ReleaseID,
			ReleaseName: detail.ReleaseName,
			Edited:      detail.Edited,
		}
	}

	return &pbds.ListGroupReleasedAppsResp{
		Count:   resp.Count,
		Details: data,
	}, nil
}

func (s *Service) queryReducedApps(kt *kit.Kit, old *table.Group, new *pbds.UpdateGroupReq) ([]*table.App, error) {
	reduced := make([]*table.App, 0)
	if new.Spec.Public {
		return reduced, nil
	}
	apps, err := s.dao.App().ListAppsByGroupID(kt, old.ID, old.Attachment.BizID)
	if err != nil {
		return reduced, err
	}
	for _, app := range apps {
		exists := false
		for _, newApp := range new.Spec.BindApps {
			if app.ID == newApp {
				exists = true
				break
			}
		}
		if !exists {
			reduced = append(reduced, app)
		}
	}
	return reduced, nil
}
