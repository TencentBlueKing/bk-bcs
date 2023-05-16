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
	"fmt"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbcontent "bscp.io/pkg/protocol/core/content"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/thirdparty/repo"
	"bscp.io/pkg/types"
	"bscp.io/pkg/version"
)

// CreateContent create content with options
func (s *Service) CreateContent(ctx context.Context, req *pbcs.CreateContentReq) (*pbcs.CreateContentResp, error) {
	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateContentResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Content, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return nil, err
	}

	if err := s.validateRepoNodeExist(kit, req); err != nil {
		logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	r := &pbds.CreateContentReq{
		Attachment: &pbcontent.ContentAttachment{
			BizId:        req.BizId,
			AppId:        req.AppId,
			ConfigItemId: req.ConfigItemId,
		},
		Spec: &pbcontent.ContentSpec{
			Signature: req.Sign,
			ByteSize:  req.ByteSize,
		},
	}
	rp, err := s.client.DS.CreateContent(kit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create content failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateContentResp{
		Id: rp.Id,
	}
	return resp, nil
}

func (s *Service) validateRepoNodeExist(kt *kit.Kit, req *pbcs.CreateContentReq) error {
	// build version is debug mode, not need to validate repo node if exist.
	if version.Debug() {
		return nil
	}

	// validate repo file if upload.
	opt := &repo.NodeOption{
		Project: s.client.Repo.ProjectID(),
		BizID:   req.BizId,
		Sign:    req.Sign,
	}
	path, err := repo.GenNodePath(opt)
	if err != nil {
		return err
	}

	exist, err := s.client.Repo.IsNodeExist(kt.ContextWithRid(), path)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("file content %s not upload", req.Sign))
	}

	return nil
}

// ListContents list contents with filter.
func (s *Service) ListContents(ctx context.Context, req *pbcs.ListContentsReq) (*pbcs.ListContentsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListContentsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Content, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, authRes)
	if err != nil {
		return nil, err
	}

	if req.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is null")
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	r := &pbds.ListContentsReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListContents(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list contents failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListContentsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
