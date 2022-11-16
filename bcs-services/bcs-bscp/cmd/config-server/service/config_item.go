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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/protocol/core/config-item"
	"bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateConfigItem create config item with option
func (s *Service) CreateConfigItem(ctx context.Context, req *pbcs.CreateConfigItemReq) (
	*pbcs.CreateConfigItemResp, error) {

	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateConfigItemResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ConfigItem, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateConfigItemReq{
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbci.ConfigItemSpec{
			Name:     req.Name,
			Path:     req.Path,
			FileType: req.FileType,
			FileMode: req.FileMode,
			Memo:     req.Memo,
			Permission: &pbci.FilePermission{
				User:      req.User,
				UserGroup: req.UserGroup,
				Privilege: req.Privilege,
			},
		},
	}
	rp, err := s.client.DS.CreateConfigItem(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("create config item failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateConfigItemResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// UpdateConfigItem update config item with option
func (s *Service) UpdateConfigItem(ctx context.Context, req *pbcs.UpdateConfigItemReq) (
	*pbcs.UpdateConfigItemResp, error) {

	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateConfigItemResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ConfigItem, Action: meta.Update,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.UpdateConfigItemReq{
		Id: req.Id,
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbci.ConfigItemSpec{
			Name:     req.Name,
			Path:     req.Path,
			FileType: req.FileType,
			FileMode: req.FileMode,
			Memo:     req.Memo,
			Permission: &pbci.FilePermission{
				User:      req.User,
				UserGroup: req.UserGroup,
				Privilege: req.Privilege,
			},
		},
	}
	_, err = s.client.DS.UpdateConfigItem(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("update config item failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// DeleteConfigItem delete config item with option
func (s *Service) DeleteConfigItem(ctx context.Context, req *pbcs.DeleteConfigItemReq) (
	*pbcs.DeleteConfigItemResp, error) {

	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteConfigItemResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ConfigItem, Action: meta.Delete,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.DeleteConfigItemReq{
		Id: req.Id,
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err = s.client.DS.DeleteConfigItem(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("delete config item failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// ListConfigItems list config item with filter
func (s *Service) ListConfigItems(ctx context.Context, req *pbcs.ListConfigItemsReq) (
	*pbcs.ListConfigItemsResp, error) {

	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListConfigItemsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ConfigItem, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(kit, resp)
		return resp, nil
	}

	if err := req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(kit, resp)
		return resp, nil
	}

	r := &pbds.ListConfigItemsReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListConfigItems(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("list config items failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListConfigItemsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
