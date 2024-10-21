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
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbtr "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-revision"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateTemplateRevision create template revision.
func (s *Service) CreateTemplateRevision(ctx context.Context,
	req *pbds.CreateTemplateRevisionReq) (*pbds.CreateResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateRevision().GetByUniqueKey(kt, req.Attachment.BizId, req.Attachment.TemplateId,
		req.Spec.RevisionName); err == nil {
		return nil, fmt.Errorf("template revision's same revision name %s already exists", req.Spec.RevisionName)
	}

	template, err := s.dao.Template().GetByID(kt, req.Attachment.BizId, req.Attachment.TemplateId)
	if err != nil {
		logs.Errorf("get template by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	// 1. create template revision
	spec := req.Spec.TemplateRevisionSpec()
	// if no revision name is specified, generate it by system
	if spec.RevisionName == "" {
		spec.RevisionName = tools.GenerateRevisionName()
	}

	// keep the revision's name and path same with template
	spec.Name = template.Spec.Name
	spec.Path = template.Spec.Path
	templateRevision := &table.TemplateRevision{
		Spec:       spec,
		Attachment: req.Attachment.TemplateRevisionAttachment(),
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	id, err := s.dao.TemplateRevision().CreateWithTx(kt, tx, templateRevision)
	if err != nil {
		logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 2. update app template bindings if necessary
	atbs, err := s.dao.TemplateBindingRelation().
		ListLatestTmplBoundUnnamedApps(kt, req.Attachment.BizId, req.Attachment.TemplateId)
	if err != nil {
		logs.Errorf("list latest template bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(atbs) > 0 {
		for _, atb := range atbs {
			if e := s.CascadeUpdateATB(kt, tx, atb); e != nil {
				logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", e, kt.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, e
			}
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}
	return &pbds.CreateResp{Id: id}, nil
}

// ListTemplateRevisions list template revision.
func (s *Service) ListTemplateRevisions(ctx context.Context,
	req *pbds.ListTemplateRevisionsReq) (*pbds.ListTemplateRevisionsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TemplateRevision)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateRevision().List(kt, req.BizId, req.TemplateId, searcher, opt)
	if err != nil {
		logs.Errorf("list template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details, err = s.listTplRevPerms(kt, req.BizId, details)
	if err != nil {
		return nil, err
	}

	resp := &pbds.ListTemplateRevisionsResp{
		Count:   uint32(count),
		Details: pbtr.PbTemplateRevisions(details),
	}

	return resp, nil
}

// DeleteTemplateRevision delete template revision.
func (s *Service) DeleteTemplateRevision(ctx context.Context,
	req *pbds.DeleteTemplateRevisionReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	templateRevision := &table.TemplateRevision{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateRevisionAttachment(),
	}
	if err := s.dao.TemplateRevision().Delete(kt, templateRevision); err != nil {
		logs.Errorf("delete template revision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// ListTemplateRevisionsByIDs list template revision by ids.
func (s *Service) ListTemplateRevisionsByIDs(ctx context.Context, req *pbds.ListTemplateRevisionsByIDsReq) (*pbds.
	ListTemplateRevisionsByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplRevisionsExist(kt, req.Ids); err != nil {
		return nil, err
	}

	details, err := s.dao.TemplateRevision().ListByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details, err = s.listTplRevPerms(kt, req.BizId, details)
	if err != nil {
		return nil, err
	}

	resp := &pbds.ListTemplateRevisionsByIDsResp{
		Details: pbtr.PbTemplateRevisions(details),
	}
	return resp, nil
}

// ListTmplRevisionNamesByTmplIDs list template revision by ids.
func (s *Service) ListTmplRevisionNamesByTmplIDs(ctx context.Context,
	req *pbds.ListTmplRevisionNamesByTmplIDsReq) (
	*pbds.ListTmplRevisionNamesByTmplIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}

	tmplRevisions, err := s.dao.TemplateRevision().ListByTemplateIDs(kt, req.BizId, req.TemplateIds)
	if err != nil {
		logs.Errorf("list template revision names by template ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(tmplRevisions) == 0 {
		return &pbds.ListTmplRevisionNamesByTmplIDsResp{}, nil
	}

	// get the map of template id => the latest template revision id
	latestRevisionMap := getLatestTmplRevisions(tmplRevisions)
	// get the map of template id => template revision detail
	tmplRevisionMap := make(map[uint32]*pbtr.TemplateRevisionNamesDetail)
	for _, t := range tmplRevisions {
		if _, ok := tmplRevisionMap[t.Attachment.TemplateID]; !ok {
			tmplRevisionMap[t.Attachment.TemplateID] = &pbtr.TemplateRevisionNamesDetail{}
		}
		tmplRevisionMap[t.Attachment.TemplateID].TemplateRevisions = append(
			tmplRevisionMap[t.Attachment.TemplateID].TemplateRevisions,
			&pbtr.TemplateRevisionNamesDetailRevisionNames{
				TemplateRevisionId:   t.ID,
				TemplateRevisionName: t.Spec.RevisionName,
				TemplateRevisionMemo: t.Spec.RevisionMemo,
			})
	}

	tmpls, err := s.dao.Template().ListByIDs(kt, req.TemplateIds)
	if err != nil {
		logs.Errorf("list template sets of biz failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]*pbtr.TemplateRevisionNamesDetail, 0)
	for _, t := range tmpls {
		details = append(details, &pbtr.TemplateRevisionNamesDetail{
			TemplateId:               t.ID,
			TemplateName:             t.Spec.Name,
			LatestTemplateRevisionId: latestRevisionMap[t.ID].ID,
			LatestRevisionName:       latestRevisionMap[t.ID].Spec.RevisionName,
			LatestSignature:          latestRevisionMap[t.ID].Spec.ContentSpec.Signature,
			LatestByteSize:           latestRevisionMap[t.ID].Spec.ContentSpec.ByteSize,
			TemplateRevisions:        tmplRevisionMap[t.ID].TemplateRevisions,
		})
	}

	resp := &pbds.ListTmplRevisionNamesByTmplIDsResp{
		Details: details,
	}
	return resp, nil
}

// getLatestTmplRevisions get the map: tmplID => latest tmplRevision
func getLatestTmplRevisions(tmplRevisions []*table.TemplateRevision) map[uint32]*table.TemplateRevision {
	latestRevisionMap := make(map[uint32]*table.TemplateRevision)
	for _, t := range tmplRevisions {
		if _, ok := latestRevisionMap[t.Attachment.TemplateID]; !ok {
			latestRevisionMap[t.Attachment.TemplateID] = t
		} else if t.ID > latestRevisionMap[t.Attachment.TemplateID].ID {
			latestRevisionMap[t.Attachment.TemplateID] = t
		}
	}

	return latestRevisionMap
}

// 创建模板版本
func (s *Service) doCreateTemplateRevisions(kt *kit.Kit, tx *gen.QueryTx, data []*table.TemplateRevision) error {

	for i := range data {
		data[i].Spec.RevisionName = tools.GenerateRevisionName()
	}

	if err := s.dao.TemplateRevision().BatchCreateWithTx(kt, tx, data); err != nil {
		logs.Errorf("batch create template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, fmt.Sprintf("batch create template revisions failed, err: %v, rid: %s", err, kt.Rid)))
	}
	return nil
}

// GetTemplateRevision 根据版本号获取 TemplateRevisions
func (s *Service) GetTemplateRevision(ctx context.Context, req *pbds.GetTemplateRevisionReq) (
	*pbds.GetTemplateRevisionResp, error) {
	kt := kit.FromGrpcContext(ctx)

	var revision *table.TemplateRevision
	var err error
	var isLatest bool
	if req.RevisionName == "" {
		revision, err = s.dao.TemplateRevision().GetLatestTemplateRevision(kt, req.GetBizId(), req.GetTemplateId())
		isLatest = true
	} else {
		revision, err = s.dao.TemplateRevision().GetByUniqueKey(kt, req.GetBizId(), req.GetTemplateId(),
			req.GetRevisionName())
		latestRevision, errL := s.dao.TemplateRevision().GetLatestTemplateRevision(kt, req.GetBizId(), req.GetTemplateId())
		if errL != nil {
			return nil, errL
		}
		if latestRevision.ID <= revision.ID {
			isLatest = true
		}
	}
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, fmt.Sprintf("get template revision failed, err: %v", err)))
	}

	revision, err = s.getTplRevPerm(kt, req.BizId, revision)
	if err != nil {
		return nil, err
	}

	return &pbds.GetTemplateRevisionResp{
		Detail: &pbds.GetTemplateRevisionResp_TemplateRevision{
			TemplateId:           revision.Attachment.TemplateID,
			Name:                 revision.Spec.Name,
			Path:                 revision.Spec.Path,
			TemplateRevisionId:   revision.ID,
			TemplateRevisionName: revision.Spec.RevisionName,
			TemplateRevisionMemo: revision.Spec.RevisionMemo,
			FileType:             string(revision.Spec.FileType),
			FileMode:             string(revision.Spec.FileMode),
			User:                 revision.Spec.Permission.User,
			UserGroup:            revision.Spec.Permission.UserGroup,
			Privilege:            revision.Spec.Permission.Privilege,
			Signature:            revision.Spec.ContentSpec.Signature,
			ByteSize:             revision.Spec.ContentSpec.ByteSize,
			Creator:              revision.Revision.Creator,
			CreateAt:             revision.Revision.CreatedAt.Format(time.RFC3339),
			Md5:                  revision.Spec.ContentSpec.Md5,
			IsLatest:             isLatest,
			Uid:                  revision.Spec.Permission.Uid,
			Gid:                  revision.Spec.Permission.Gid,
		},
	}, nil
}

// UpdateTemplateRevision implements pbds.DataServer.
// nolint:funlen
func (s *Service) UpdateTemplateRevision(ctx context.Context, req *pbds.UpdateTemplateRevisionReq) (
	*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateRevision().GetByUniqueKey(kt, req.Attachment.BizId, req.Attachment.TemplateId,
		req.Spec.RevisionName); err == nil {
		return nil, fmt.Errorf("template revision's same revision name %s already exists", req.Spec.RevisionName)
	}

	template, err := s.dao.Template().GetByID(kt, req.Attachment.BizId, req.Attachment.TemplateId)
	if err != nil {
		logs.Errorf("get template by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 获取对应已存在的版本
	revision, err := s.dao.TemplateRevision().GetTemplateRevisionById(kt,
		req.GetAttachment().GetBizId(), req.GetTemplateRevisionId())
	if err != nil {
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()
	// 1. create template revision
	spec := req.Spec.TemplateRevisionSpec()
	spec.RevisionMemo = ""
	// if no revision name is specified, generate it by system
	if spec.RevisionName == "" {
		spec.RevisionName = tools.GenerateRevisionName()
	}

	// keep the revision's name and path same with template
	spec.Name = template.Spec.Name
	spec.Path = template.Spec.Path
	templateRevision := &table.TemplateRevision{
		Spec:       spec,
		Attachment: req.Attachment.TemplateRevisionAttachment(),
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: time.Now().UTC(),
		},
	}

	var id uint32
	//  如果文件权限和内容没变化不更新模板版本数据
	if !reflect.DeepEqual(revision.Spec.ContentSpec, spec.ContentSpec) ||
		!reflect.DeepEqual(revision.Spec.Permission, spec.Permission) {
		id, err = s.dao.TemplateRevision().CreateWithTx(kt, tx, templateRevision)
		template.Revision.Reviser = kt.User
		template.Revision.UpdatedAt = time.Now().UTC()
	}
	if err != nil {
		logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// update template
	err = s.dao.Template().UpdateWithTx(kt, tx, &table.Template{
		ID: template.ID,
		Spec: &table.TemplateSpec{
			Memo: req.Spec.RevisionMemo,
		},
		Attachment: template.Attachment,
		Revision:   template.Revision,
	})
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	userPrivilege := &table.UserPrivilege{
		Spec: &table.UserPrivilegeSpec{
			User: req.GetSpec().Permission.User,
		},
		Attachment: &table.UserPrivilegeAttachment{
			BizID:           req.GetAttachment().BizId,
			TemplateSpaceID: req.Attachment.TemplateSpaceId,
			Uid:             req.GetSpec().GetPermission().GetUid(),
		},
	}

	userGroupPrivilege := &table.UserGroupPrivilege{
		Spec: &table.UserGroupPrivilegeSpec{
			UserGroup: req.GetSpec().Permission.UserGroup,
		},
		Attachment: &table.UserGroupPrivilegeAttachment{
			BizID:           req.GetAttachment().BizId,
			TemplateSpaceID: req.Attachment.TemplateSpaceId,
			Gid:             req.GetSpec().GetPermission().GetGid(),
		},
	}

	// 2. 更新权限组
	err = s.upsertSinglePermissions(kt, tx, req.Attachment.BizId, 0, req.Attachment.TemplateSpaceId,
		userPrivilege, userGroupPrivilege, true)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}
	return &pbds.CreateResp{Id: id}, nil
}
