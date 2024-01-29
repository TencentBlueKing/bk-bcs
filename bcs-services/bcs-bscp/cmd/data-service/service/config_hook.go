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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// UpdateConfigHook update ConfigHook.
func (s *Service) UpdateConfigHook(ctx context.Context, req *pbds.UpdateConfigHookReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)
	tx := s.dao.GenQuery().Begin()

	preHook := &table.ReleasedHook{
		AppID: req.AppId,
		BizID: req.BizId,
		// ReleasedID 0 for editing release
		ReleaseID: 0,
		HookID:    req.PreHookId,
		HookType:  table.PreHook,
	}
	postHook := &table.ReleasedHook{
		AppID: req.AppId,
		BizID: req.BizId,
		// ReleasedID 0 for editing release
		ReleaseID: 0,
		HookID:    req.PostHookId,
		HookType:  table.PostHook,
	}

	if req.PreHookId > 0 {
		hook, err := s.getReleasedHook(kt, preHook)
		if err != nil {
			logs.Errorf("no released releases of the pre-hook, err: %v, rid: %s", err, kt.Rid)
			return nil, errors.New("no released releases of the pre-hook")
		}

		if err = s.dao.ReleasedHook().UpsertWithTx(kt, tx, hook); err != nil {
			logs.Errorf("upsert pre-hook failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	} else {
		if err := s.dao.ReleasedHook().DeleteByUniqueKeyWithTx(kt, tx, preHook); err != nil {
			logs.Errorf("delete pre-hook failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	if req.PostHookId > 0 {
		hook, err := s.getReleasedHook(kt, postHook)
		if err != nil {
			logs.Errorf("get post-hook failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}

		if err = s.dao.ReleasedHook().UpsertWithTx(kt, tx, hook); err != nil {
			logs.Errorf("upsert post-hook failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	} else {
		if err := s.dao.ReleasedHook().DeleteByUniqueKeyWithTx(kt, tx, postHook); err != nil {
			logs.Errorf("delete post-hook failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// getReleasedHook ...
func (s *Service) getReleasedHook(kt *kit.Kit, rh *table.ReleasedHook) (*table.ReleasedHook, error) {

	h, err := s.dao.Hook().GetByID(kt, rh.BizID, rh.HookID)
	if err != nil {
		logs.Errorf("get %s failed, err: %v, rid: %s", rh.HookType.String(), err, kt.Rid)
		return nil, err
	}

	opt := &types.GetByPubStateOption{
		BizID:  rh.BizID,
		HookID: rh.HookID,
		State:  table.HookRevisionStatusDeployed,
	}
	hr, err := s.dao.HookRevision().GetByPubState(kt, opt)
	if err != nil {
		logs.Errorf("no released releases of the %s, err: %v, rid: %s", rh.HookType.String(), err, kt.Rid)
		return nil, fmt.Errorf("no released releases of the %s", rh.HookType.String())
	}

	return &table.ReleasedHook{
		BizID:            rh.BizID,
		AppID:            rh.AppID,
		ReleaseID:        0,
		HookID:           h.ID,
		HookName:         h.Spec.Name,
		HookRevisionID:   hr.ID,
		HookRevisionName: hr.Spec.Name,
		Content:          hr.Spec.Content,
		ScriptType:       h.Spec.Type,
		HookType:         rh.HookType,
		Reviser:          kt.User,
	}, nil
}
