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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/types"
)

// GetAppMeta get app's basic info.
func (s *Service) GetAppMeta(ctx context.Context, req *pbcs.GetAppMetaReq) (*pbcs.JsonRawResp, error) {
	if req.BizId <= 0 || req.AppId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id or app id")
	}

	kt := kit.FromGrpcContext(ctx)
	meta, err := s.op.GetAppMeta(kt, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	return &pbcs.JsonRawResp{
		JsonRaw: meta,
	}, nil
}

// GetReleasedCI get released config items from cache.
func (s *Service) GetReleasedCI(ctx context.Context, req *pbcs.GetReleasedCIReq) (*pbcs.JsonRawResp, error) {
	if req.BizId <= 0 || req.ReleaseId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id or release id")
	}

	kt := kit.FromGrpcContext(ctx)
	ci, err := s.op.GetReleasedCI(kt, req.BizId, req.ReleaseId)
	if err != nil {
		return nil, err
	}

	return &pbcs.JsonRawResp{
		JsonRaw: ci,
	}, nil
}

// GetAppInstanceRelease get an app instance's specific release if it has.
func (s *Service) GetAppInstanceRelease(ctx context.Context, req *pbcs.GetAppInstanceReleaseReq) (
	*pbcs.GetAppInstanceReleaseResp, error) {

	if req.BizId <= 0 || req.AppId <= 0 || len(req.Uid) == 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id, app id or app instance uid")
	}

	kt := kit.FromGrpcContext(ctx)
	meta, err := s.dao.CRInstance().GetAppCRIMeta(kt, req.BizId, req.AppId, req.Uid)
	if err != nil {
		return nil, err
	}

	return &pbcs.GetAppInstanceReleaseResp{ReleaseId: meta.ReleaseID}, nil
}

// GetAppCpsID get app's latest published strategy id.
func (s *Service) GetAppCpsID(ctx context.Context, req *pbcs.GetAppCpsIDReq) (*pbcs.GetAppCpsIDResp, error) {

	if req.BizId <= 0 || req.AppId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id or app id")
	}

	kt := kit.FromGrpcContext(ctx)

	opt := &types.GetAppCpsIDOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		Namespace: req.Namespace,
	}
	list, err := s.dao.Publish().GetAppCpsID(kt, opt)
	if err != nil {
		return nil, err
	}

	resp := &pbcs.GetAppCpsIDResp{
		CpsId: list,
	}

	return resp, nil
}

// GetAppReleasedStrategy get app's latest published strategies with different rules.
func (s *Service) GetAppReleasedStrategy(ctx context.Context, req *pbcs.GetAppReleasedStrategyReq) (
	*pbcs.JsonArrayRawResp, error) {

	if req.BizId <= 0 || req.AppId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id or app id")
	}

	kt := kit.FromGrpcContext(ctx)
	list, err := s.op.GetAppReleasedStrategies(kt, req.BizId, req.AppId, req.CpsId)
	if err != nil {
		return nil, err
	}

	return &pbcs.JsonArrayRawResp{JsonRaw: list}, nil
}

// ListAppReleasedGroups list app's released groups.
func (s *Service) ListAppReleasedGroups(ctx context.Context, req *pbcs.ListAppReleasedGroupsReq) (
	*pbcs.JsonRawResp, error) {

	if req.BizId <= 0 || req.AppId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id or app id")
	}

	kt := kit.FromGrpcContext(ctx)
	list, err := s.op.ListAppReleasedGroups(kt, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	return &pbcs.JsonRawResp{JsonRaw: list}, nil
}

// ListCredentialMatchedCI list all config item ids which can be matched by credential.
func (s *Service) ListCredentialMatchedCI(ctx context.Context, req *pbcs.ListCredentialMatchedCIReq) (
	*pbcs.JsonRawResp, error) {

	if req.BizId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "invalid biz id")
	}

	if req.Credential == "" {
		return nil, errf.New(errf.InvalidParameter, "invalid credential")
	}

	kt := kit.FromGrpcContext(ctx)
	list, err := s.op.ListCredentialMatchedCI(kt, req.BizId, req.Credential)
	if err != nil {
		return nil, err
	}

	return &pbcs.JsonRawResp{JsonRaw: list}, nil
}

// GetCurrentCursorReminder get the current consumed event's id, which is the cursor reminder's resource id.
func (s *Service) GetCurrentCursorReminder(ctx context.Context, _ *pbbase.EmptyReq) (*pbcs.CurrentCursorReminderResp,
	error) {

	kt := kit.FromGrpcContext(ctx)
	cursor, err := s.dao.Event().LatestCursor(kt)
	if err != nil {
		return nil, err
	}

	return &pbcs.CurrentCursorReminderResp{Cursor: cursor}, nil
}

// ListEventsMeta list event metas with filter
func (s *Service) ListEventsMeta(ctx context.Context, req *pbcs.ListEventsReq) (*pbcs.ListEventsResp, error) {

	if req.Page.Count {
		return nil, errors.New("invalid request, do now allows to count events")
	}

	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if req.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is null")
	}

	opt := &types.ListEventsOption{
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	result, err := s.dao.Event().ListConsumedEvents(kt, opt)
	if err != nil {
		return nil, err
	}

	metas := make([]*types.EventMeta, len(result.Details))
	for idx := range result.Details {
		metas[idx] = &types.EventMeta{
			ID:         result.Details[idx].ID,
			Spec:       result.Details[idx].Spec,
			Attachment: result.Details[idx].Attachment,
		}
	}

	return &pbcs.ListEventsResp{List: pbcs.PbEventMeta(metas)}, nil
}
