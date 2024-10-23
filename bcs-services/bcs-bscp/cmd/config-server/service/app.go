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
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	iamauth "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/view/webannotation"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/natsort"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/space"
)

// CreateApp create app with options
func (s *Service) CreateApp(ctx context.Context, req *pbcs.CreateAppReq) (*pbcs.CreateAppResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := req.Validate(kt); err != nil {
		return nil, err
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Create}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.CreateAppReq{
		BizId: req.BizId,
		Spec: &pbapp.AppSpec{
			Name:        req.Name,
			ConfigType:  req.ConfigType,
			Memo:        req.Memo,
			Alias:       req.Alias,
			DataType:    req.DataType,
			IsApprove:   req.IsApprove,
			ApproveType: req.ApproveType,
			Approver:    req.Approver,
		},
	}
	rp, err := s.client.DS.CreateApp(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if err := s.authorizer.GrantResourceCreatorAction(kt, &client.GrantResourceCreatorActionOption{
		System:  sys.SystemIDBSCP,
		Type:    sys.Application,
		ID:      strconv.Itoa(int(rp.Id)),
		Name:    req.Name,
		Creator: kt.User,
	}); err != nil {
		logs.Errorf("grant app creator action failed, err: %v, rid: %s", err, kt.Rid)
	}

	resp := &pbcs.CreateAppResp{Id: rp.Id}
	return resp, nil
}

// UpdateApp update app with options
func (s *Service) UpdateApp(ctx context.Context, req *pbcs.UpdateAppReq) (*pbapp.App, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.Id}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.UpdateAppReq{
		Id:    req.Id,
		BizId: req.BizId,
		Spec: &pbapp.AppSpec{
			Name:        req.Name,
			Memo:        req.Memo,
			Alias:       req.Alias,
			DataType:    req.DataType,
			IsApprove:   req.IsApprove,
			ApproveType: req.ApproveType,
			Approver:    req.Approver,
		},
	}
	app, err := s.client.DS.UpdateApp(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return app, nil
}

// DeleteApp delete app with options
func (s *Service) DeleteApp(ctx context.Context, req *pbcs.DeleteAppReq) (*pbcs.DeleteAppResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteAppResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Delete, ResourceID: req.Id}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteAppReq{
		Id:    req.Id,
		BizId: req.BizId,
	}
	_, err = s.client.DS.DeleteApp(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp, nil
}

// GetApp get app with app id
func (s *Service) GetApp(ctx context.Context, req *pbcs.GetAppReq) (*pbapp.App, error) {
	kt := kit.FromGrpcContext(ctx)
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetAppReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}
	rp, err := s.client.DS.GetApp(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rp, nil
}

// GetAppByName get app by app name
func (s *Service) GetAppByName(ctx context.Context, req *pbcs.GetAppByNameReq) (*pbapp.App, error) {
	kt := kit.FromGrpcContext(ctx)

	// nolint
	// TODO: 暂不鉴权
	// resp := new(pbapp.App)

	// res := []*meta.ResourceAttribute{
	// 	{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	// 	{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	// }
	// err := s.authorizer.AuthorizeWithApplyDetail(kt, res...)
	// if err != nil {
	// 	return nil, err
	// }

	r := &pbds.GetAppByNameReq{
		BizId:   req.BizId,
		AppName: req.AppName,
	}
	rp, err := s.client.DS.GetAppByName(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rp, nil
}

// ListAppsRest list apps with rest filter
func (s *Service) ListAppsRest(ctx context.Context, req *pbcs.ListAppsRestReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	userSpaceResp, err := s.client.AS.ListUserSpace(kt.RpcCtx(), &pbas.ListUserSpaceReq{})
	if err != nil {
		return nil, err
	}

	if len(userSpaceResp.GetItems()) == 0 {
		return nil, errors.New("use have no spaces")
	}

	spaceMap := map[string]*pbas.Space{}
	spaceIdList := []string{}
	for _, s := range userSpaceResp.GetItems() {
		spaceMap[s.SpaceId] = s
		spaceIdList = append(spaceIdList, s.SpaceId)
	}

	r := &pbds.ListAppsRestReq{
		BizId:    strings.Join(spaceIdList, ","),
		Start:    req.Start,
		Limit:    req.Limit,
		Operator: req.Operator,
		Name:     req.Name,
	}
	rp, err := s.client.DS.ListAppsRest(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 只填写当前页的space
	for _, app := range rp.Details {
		id := strconv.Itoa(int(app.BizId))
		sp, ok := spaceMap[id]
		if !ok {
			app.SpaceId = id
			app.SpaceName = ""
			app.SpaceTypeId = ""
			app.SpaceTypeName = ""
		} else {
			app.SpaceId = id
			app.SpaceName = sp.SpaceName
			app.SpaceTypeId = sp.SpaceTypeId
			app.SpaceTypeName = sp.SpaceTypeName
		}
	}

	resp := &pbcs.ListAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListAppsBySpaceRest list apps with rest filter
func (s *Service) ListAppsBySpaceRest(ctx context.Context,
	req *pbcs.ListAppsBySpaceRestReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListAppsRestReq{
		BizId:    strconv.Itoa(int(req.BizId)),
		Start:    req.Start,
		Limit:    req.Limit,
		Operator: req.Operator,
		Name:     req.Name,
		All:      req.All,
	}
	rp, err := s.client.DS.ListAppsRest(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	spaceUidMap := map[string]struct{}{}
	for _, app := range rp.Details {
		uid := space.BuildSpaceUid(space.BK_CMDB, strconv.Itoa(int(app.BizId)))
		spaceUidMap[uid] = struct{}{}

	}
	querySpaceReq := &pbas.QuerySpaceReq{SpaceUid: []string{}}
	for spaceUid := range spaceUidMap {
		querySpaceReq.SpaceUid = append(querySpaceReq.SpaceUid, spaceUid)
	}

	querySpaceResp, err := s.client.AS.QuerySpace(ctx, querySpaceReq)
	if err != nil {
		return nil, errors.Wrap(err, "QuerySpace")
	}

	spaceMap := map[string]*pbas.Space{}
	for _, s := range querySpaceResp.GetItems() {
		spaceMap[s.SpaceId] = s
	}

	// 只填写当前页的space
	for _, app := range rp.Details {
		id := strconv.Itoa(int(app.BizId))
		sp, ok := spaceMap[id]
		if !ok {
			app.SpaceId = id
			app.SpaceName = ""
			app.SpaceTypeId = ""
			app.SpaceTypeName = ""
		} else {
			app.SpaceId = id
			app.SpaceName = sp.SpaceName
			app.SpaceTypeId = sp.SpaceTypeId
			app.SpaceTypeName = sp.SpaceTypeName
		}
	}

	sort.SliceStable(rp.Details, func(i, j int) bool {
		return natsort.NaturalLess(rp.Details[i].Spec.Name, rp.Details[j].Spec.Name)
	})

	resp := &pbcs.ListAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}

	return resp, nil
}

// ListAppsAnnotation list apps permission annotations
func ListAppsAnnotation(ctx context.Context, kt *kit.Kit,
	authorizer iamauth.Authorizer, msg proto.Message) (*webannotation.Annotation, error) {

	resp, ok := msg.(*pbcs.ListAppsResp)
	if !ok {
		return nil, nil
	}

	perms := map[string]webannotation.Perm{}
	authRes := make([]*meta.ResourceAttribute, 0, len(resp.Details))
	for _, v := range resp.Details {
		bID, _ := strconv.ParseInt(v.SpaceId, 10, 64)
		authRes = append(authRes, &meta.ResourceAttribute{Basic: meta.Basic{
			Type: meta.App, Action: meta.View, ResourceID: v.Id}, BizID: uint32(bID)},
		)
		authRes = append(authRes, &meta.ResourceAttribute{Basic: meta.Basic{
			Type: meta.App, Action: meta.Update, ResourceID: v.Id}, BizID: uint32(bID)},
		)
		authRes = append(authRes, &meta.ResourceAttribute{Basic: meta.Basic{
			Type: meta.App, Action: meta.Delete, ResourceID: v.Id}, BizID: uint32(bID)},
		)
		authRes = append(authRes, &meta.ResourceAttribute{Basic: meta.Basic{
			Type: meta.App, Action: meta.Publish, ResourceID: v.Id}, BizID: uint32(bID)},
		)
		authRes = append(authRes, &meta.ResourceAttribute{Basic: meta.Basic{
			Type: meta.App, Action: meta.GenerateRelease, ResourceID: v.Id}, BizID: uint32(bID)},
		)
	}

	decisions, _, err := authorizer.AuthorizeDecision(kt, authRes...)
	if err != nil {
		return nil, err
	}

	dMap := meta.DecisionsMap(decisions)

	for _, res := range authRes {
		if _, ok := perms[strconv.Itoa(int(res.ResourceID))]; !ok {
			perms[strconv.Itoa(int(res.ResourceID))] = webannotation.Perm{
				string(res.Action): dMap[*res],
			}
		} else {
			perms[strconv.Itoa(int(res.ResourceID))][string(res.Action)] = dMap[*res]
		}
	}
	return &webannotation.Annotation{Perms: perms}, nil
}
