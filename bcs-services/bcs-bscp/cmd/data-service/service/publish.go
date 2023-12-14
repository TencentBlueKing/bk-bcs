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
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbgroup "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/group"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Publish exec publish strategy.
func (s *Service) Publish(ctx context.Context, req *pbds.PublishReq) (*pbds.PublishResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	groupIDs := make([]uint32, 0)
	tx := s.dao.GenQuery().Begin()

	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if err := publishMode.Validate(); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, err
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, groupID := range req.Groups {
				if groupID == 0 {
					groupIDs = append(groupIDs, groupID)
					continue
				}
				group, e := s.dao.Group().Get(grpcKit, groupID, req.BizId)
				if e != nil {
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
					}
					return nil, fmt.Errorf("group %d not exist", groupID)
				}
				groupIDs = append(groupIDs, group.ID)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, err := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if err != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", err, grpcKit.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
				}
				return nil, err
			}
			groupIDs = append(groupIDs, groupID)
		}
	}

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: req.ReleaseId,
		All:       req.All,
		Default:   req.Default,
		Memo:      req.Memo,
		Groups:    groupIDs,
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}

	pshID, err := s.dao.Publish().PublishWithTx(grpcKit, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.PublishResp{PublishedStrategyHistoryId: pshID}
	return resp, nil
}

// GenerateReleaseAndPublish generate release and publish.
// nolint: funlen
func (s *Service) GenerateReleaseAndPublish(ctx context.Context, req *pbds.GenerateReleaseAndPublishReq) (
	*pbds.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().GetByID(grpcKit, req.AppId)
	if err != nil {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if _, e := s.dao.Release().GetByName(grpcKit, req.BizId, req.AppId, req.ReleaseName); e == nil {
		return nil, fmt.Errorf("release name %s already exists", req.ReleaseName)
	}

	tx := s.dao.GenQuery().Begin()

	groupIDs, err := s.genReleaseAndPublishGroupID(grpcKit, tx, req)
	if err != nil {
		return nil, err
	}

	// create release.
	release := &table.Release{
		Spec: &table.ReleaseSpec{
			Name: req.ReleaseName,
			Memo: req.ReleaseMemo,
		},
		Attachment: &table.ReleaseAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}
	releaseID, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	// create released hook.
	if err = s.createReleasedHook(grpcKit, tx, req.BizId, req.AppId, releaseID); err != nil {
		logs.Errorf("create released hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	switch app.Spec.ConfigType {
	case table.File:

		// Note: need to change batch operator to query config item and it's commit.
		// query app's all config items.
		cfgItems, e := s.getAppConfigItems(grpcKit)
		if e != nil {
			logs.Errorf("query app config item list failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		// get app template revisions which are template config items
		tmplRevisions, e := s.getAppTmplRevisions(grpcKit)
		if e != nil {
			logs.Errorf("get app template revisions failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		// if no config item, return directly.
		if len(cfgItems) == 0 && len(tmplRevisions) == 0 {
			return nil, errors.New("app config items is empty")
		}

		// do template and non-template config item related operations for create release.
		if err = s.doConfigItemOperations(grpcKit, req.Variables, tx, release.ID, tmplRevisions, cfgItems); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			logs.Errorf("do template action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	case table.KV:
		if err = s.doKvOperations(grpcKit, tx, req.AppId, req.BizId, release.ID); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			logs.Errorf("do kv action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	}

	// publish with transaction.
	kt := kit.FromGrpcContext(ctx)

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: releaseID,
		All:       req.All,
		Memo:      req.ReleaseMemo,
		Groups:    groupIDs,
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	pshID, err := s.dao.Publish().PublishWithTx(kt, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	// commit transaction.
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &pbds.PublishResp{PublishedStrategyHistoryId: pshID}, nil
}

func (s *Service) genReleaseAndPublishGroupID(grpcKit *kit.Kit, tx *gen.QueryTx,
	req *pbds.GenerateReleaseAndPublishReq) ([]uint32, error) {

	groupIDs := make([]uint32, 0)

	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if e := publishMode.Validate(); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, e
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, name := range req.Groups {
				group, e := s.dao.Group().GetByName(grpcKit, req.BizId, name)
				if e != nil {
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
					}
					return nil, fmt.Errorf("group %s not exist", name)
				}
				groupIDs = append(groupIDs, group.ID)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, e := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if e != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", e, grpcKit.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
				}
				return nil, e
			}
			groupIDs = append(groupIDs, groupID)
		}
	}

	return groupIDs, nil
}

func (s *Service) getOrCreateGroupByLabels(grpcKit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32, groupName string,
	labels []*structpb.Struct) (uint32, error) {
	elements := make([]selector.Element, 0)
	for _, label := range labels {
		element, err := pbgroup.UnmarshalElement(label)
		if err != nil {
			return 0, fmt.Errorf("unmarshal group label failed, err: %v", err)
		}
		elements = append(elements, *element)
	}
	sel := &selector.Selector{
		LabelsAnd: elements,
	}
	groups, err := s.dao.Group().ListAppValidGroups(grpcKit, bizID, appID)
	if err != nil {
		return 0, err
	}
	exists := make([]*table.Group, 0)
	for _, group := range groups {
		if group.Spec.Selector.Equal(sel) {
			exists = append(exists, group)
		}
	}
	// if same labels group exists, return it's id.
	if len(exists) > 0 {
		return exists[0].ID, nil
	}
	// else create new one.
	if groupName != "" {
		// if group name is not empty, use it as group name.
		_, err = s.dao.Group().GetByName(grpcKit, bizID, groupName)
		// if group name already exists, return error.
		if err == nil {
			return 0, fmt.Errorf("group %s already exists", groupName)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	} else {
		// generate group name by time.
		groupName = time.Now().Format("20060102150405.000")
		groupName = fmt.Sprintf("g_%s", strings.ReplaceAll(groupName, ".", ""))
	}
	group := table.Group{
		Spec: &table.GroupSpec{
			Name:     groupName,
			Public:   false,
			Mode:     table.Custom,
			Selector: sel,
		},
		Attachment: &table.GroupAttachment{
			BizID: bizID,
		},
		Revision: &table.Revision{
			Creator: grpcKit.User,
			Reviser: grpcKit.User,
		},
	}
	groupID, err := s.dao.Group().CreateWithTx(grpcKit, tx, &group)
	if err != nil {
		return 0, err
	}
	if err := s.dao.GroupAppBind().BatchCreateWithTx(grpcKit, tx, []*table.GroupAppBind{
		{
			GroupID: groupID,
			AppID:   appID,
			BizID:   bizID,
		},
	}); err != nil {
		return 0, err
	}
	return groupID, nil
}

func (s *Service) createReleasedHook(grpcKit *kit.Kit, tx *gen.QueryTx, bizID, appID, releaseID uint32) error {
	pre, err := s.dao.ReleasedHook().Get(grpcKit, bizID, appID, 0, table.PreHook)
	if err == nil {
		pre.ID = 0
		pre.ReleaseID = releaseID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, pre); e != nil {
			logs.Errorf("create released pre-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released pre-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	post, err := s.dao.ReleasedHook().Get(grpcKit, bizID, appID, 0, table.PostHook)
	if err == nil {
		post.ID = 0
		post.ReleaseID = releaseID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, post); e != nil {
			logs.Errorf("create released post-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released post-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	return nil
}
