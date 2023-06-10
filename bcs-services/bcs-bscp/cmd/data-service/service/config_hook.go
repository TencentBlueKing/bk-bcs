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

	"gorm.io/gorm"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbch "bscp.io/pkg/protocol/core/config-hook"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateConfigHook create configHook.
func (s *Service) CreateConfigHook(ctx context.Context, req *pbds.CreateConfigHookReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.ConfigHook().GetByAppID(kt, req.Attachment.BizId, req.Attachment.AppId); err == nil {
		return nil, fmt.Errorf("configHook app_id %d already exists", req.Attachment.AppId)
	}

	if req.Spec.PreHookId > 0 {
		opt := &types.GetByPubStateOption{
			BizID:  req.Attachment.BizId,
			HookID: req.Spec.PreHookId,
			State:  table.DeployedHookReleased,
		}
		hr, err := s.dao.HookRelease().GetByPubState(kt, opt)
		if err != nil {
			logs.Errorf("no released releases of the pre-hook, err: %v, rid: %s", err, kt.Rid)
			return nil, errors.New("no released releases of the pre-hook")
		}
		req.Spec.PreHookReleaseId = hr.ID
	}

	if req.Spec.PostHookId > 0 {
		opt := &types.GetByPubStateOption{
			BizID:  req.Attachment.BizId,
			HookID: req.Spec.PostHookId,
			State:  table.DeployedHookReleased,
		}
		hr, err := s.dao.HookRelease().GetByPubState(kt, opt)
		if err != nil {
			logs.Errorf("no released releases of the post-hook, err: %v, rid: %s", err, kt.Rid)
			return nil, errors.New("no released releases of the post-hook")
		}
		req.Spec.PostHookReleaseId = hr.ID
	}

	spec, err := req.Spec.ConfigHookSpec()
	if err != nil {
		logs.Errorf("get configHook spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	hook := &table.ConfigHook{
		Spec:       spec,
		Attachment: req.Attachment.ConfigHookAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.ConfigHook().Create(kt, hook)
	if err != nil {
		logs.Errorf("create configHook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

// UpdateConfigHook update ConfigHook.
func (s *Service) UpdateConfigHook(ctx context.Context, req *pbds.UpdateConfigHookReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if req.Spec.PreHookId > 0 {
		opt := &types.GetByPubStateOption{
			BizID:  req.Attachment.BizId,
			HookID: req.Spec.PreHookId,
			State:  table.DeployedHookReleased,
		}
		hr, err := s.dao.HookRelease().GetByPubState(kt, opt)
		if err != nil {
			logs.Errorf("no released releases of the pre-hook, err: %v, rid: %s", err, kt.Rid)
			return nil, errors.New("no released releases of the pre-hook")
		}
		req.Spec.PreHookReleaseId = hr.ID
	} else {
		req.Spec.PreHookId = 0
		req.Spec.PreHookReleaseId = 0
	}

	if req.Spec.PostHookId > 0 {
		opt := &types.GetByPubStateOption{
			BizID:  req.Attachment.BizId,
			HookID: req.Spec.PostHookId,
			State:  table.DeployedHookReleased,
		}
		hr, err := s.dao.HookRelease().GetByPubState(kt, opt)
		if err != nil {
			logs.Errorf("no released releases of the post-hook, err: %v, rid: %s", err, kt.Rid)
			return nil, errors.New("no released releases of the post-hook")
		}
		req.Spec.PostHookReleaseId = hr.ID
	} else {
		req.Spec.PostHookId = 0
		req.Spec.PostHookReleaseId = 0
	}

	spec, e := req.Spec.ConfigHookSpec()
	if e != nil {
		logs.Errorf("get ConfigHookSpec spec from pb failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}
	hook := &table.ConfigHook{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.ConfigHookAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.ConfigHook().Update(kt, hook); err != nil {
		logs.Errorf("update ConfigHook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// GetConfigHook get a configHook
func (s *Service) GetConfigHook(ctx context.Context, req *pbds.GetConfigHookReq) (*pbch.ConfigHook, error) {

	kt := kit.FromGrpcContext(ctx)

	hook, err := s.dao.ConfigHook().GetByAppID(kt, req.BizId, req.AppId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pbch.PbConfigHook(genNilConfigHook()), nil
	}
	if err != nil {
		logs.Errorf("get ConfigHook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return pbch.PbConfigHook(hook), err

}

func genNilConfigHook() *table.ConfigHook {
	return &table.ConfigHook{
		Spec:       &table.ConfigHookSpec{},
		Attachment: &table.ConfigHookAttachment{},
		Revision:   &table.Revision{},
	}
}
