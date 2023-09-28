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

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
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
		HookType:  table.PreHook,
	}
	postHook := &table.ReleasedHook{
		AppID: req.AppId,
		BizID: req.BizId,
		// ReleasedID 0 for editing release
		ReleaseID: 0,
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
			if err = tx.Rollback(); err != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", err, kt.Rid)
			}
			return nil, err
		}
	} else {
		if err := s.dao.ReleasedHook().DeleteByUniqueKeyWithTx(kt, tx, preHook); err != nil {
			logs.Errorf("delete pre-hook failed, err: %v, rid: %s", err, kt.Rid)
			if err = tx.Rollback(); err != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", err, kt.Rid)
			}
			return nil, err
		}
	}

	if req.PostHookId > 0 {
		hook, err := s.getReleasedHook(kt, postHook)
		if err != nil {
			logs.Errorf("get pre-hook failed, err: %v, rid: %s", err, kt.Rid)
			if err = tx.Rollback(); err != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", err, kt.Rid)
			}
			return nil, err
		}

		if err = s.dao.ReleasedHook().UpsertWithTx(kt, tx, hook); err != nil {
			logs.Errorf("upsert post-hook failed, err: %v, rid: %s", err, kt.Rid)
			if err = tx.Rollback(); err != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", err, kt.Rid)
			}
			return nil, err
		}
	} else {
		if err := s.dao.ReleasedHook().DeleteByUniqueKeyWithTx(kt, tx, postHook); err != nil {
			logs.Errorf("delete post-hook failed, err: %v, rid: %s", err, kt.Rid)
			if err = tx.Rollback(); err != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", err, kt.Rid)
			}
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		if err = tx.Rollback(); err != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", err, kt.Rid)
		}
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// getReleasedHook ...
func (s *Service) getReleasedHook(kt *kit.Kit, rh *table.ReleasedHook) (*table.ReleasedHook, error) {

	h, err := s.dao.Hook().GetByID(kt, rh.BizID, rh.HookID)
	if err != nil {
		logs.Errorf("get pre-hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	opt := &types.GetByPubStateOption{
		BizID:  rh.BizID,
		HookID: rh.HookID,
		State:  table.HookRevisionStatusDeployed,
	}
	hr, err := s.dao.HookRevision().GetByPubState(kt, opt)
	if err != nil {
		logs.Errorf("no released releases of the pre-hook, err: %v, rid: %s", err, kt.Rid)
		return nil, errors.New("no released releases of the pre-hook")
	}

	rh.HookID = h.ID
	rh.HookName = h.Spec.Name
	rh.HookRevisionID = hr.ID
	rh.HookRevisionName = hr.Spec.Name
	rh.Content = hr.Spec.Content
	rh.ScriptType = h.Spec.Type
	rh.Reviser = kt.User

	return rh, nil
}
