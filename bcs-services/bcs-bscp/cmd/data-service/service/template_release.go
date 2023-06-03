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

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbtr "bscp.io/pkg/protocol/core/template-release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateTemplateRelease create template release.
func (s *Service) CreateTemplateRelease(ctx context.Context, req *pbds.CreateTemplateReleaseReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateRelease().GetByUniqueKey(kt, req.Attachment.BizId, req.Attachment.TemplateId,
		req.Spec.ReleaseName); err == nil {
		return nil, fmt.Errorf("template release's same release name %s already exists", req.Spec.ReleaseName)
	}

	template, err := s.dao.Template().GetByID(kt, req.Attachment.BizId, req.Attachment.TemplateId)
	if err != nil {
		logs.Errorf("get template by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	spec := req.Spec.TemplateReleaseSpec()
	// keep the release's name and path same with template
	spec.Name = template.Spec.Name
	spec.Path = template.Spec.Path
	TemplateRelease := &table.TemplateRelease{
		Spec:       spec,
		Attachment: req.Attachment.TemplateReleaseAttachment(),
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	id, err := s.dao.TemplateRelease().Create(kt, TemplateRelease)
	if err != nil {
		logs.Errorf("create template release failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateReleases list template release.
func (s *Service) ListTemplateReleases(ctx context.Context, req *pbds.ListTemplateReleasesReq) (*pbds.ListTemplateReleasesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateRelease().List(kt, req.BizId, req.TemplateId, opt)

	if err != nil {
		logs.Errorf("list template releases failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateReleasesResp{
		Count:   uint32(count),
		Details: pbtr.PbTemplateReleases(details),
	}
	return resp, nil
}

// DeleteTemplateRelease delete template release.
func (s *Service) DeleteTemplateRelease(ctx context.Context, req *pbds.DeleteTemplateReleaseReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	TemplateRelease := &table.TemplateRelease{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateReleaseAttachment(),
	}
	if err := s.dao.TemplateRelease().Delete(kt, TemplateRelease); err != nil {
		logs.Errorf("delete template release failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
