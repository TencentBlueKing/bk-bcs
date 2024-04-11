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
	"path"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbatb "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app-template-binding"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcredential "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/credential"
	pbtset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-set"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateCredential Create Credential
func (s *Service) CreateCredential(ctx context.Context, req *pbds.CreateCredentialReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Credential().GetByName(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("credential name %s already exists", req.Spec.Name)
	}

	credential := &table.Credential{
		Spec:       req.Spec.CredentialSpec(),
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	credential.Spec.ExpiredAt = time.Now().UTC()
	id, err := s.dao.Credential().Create(kt, credential)
	if err != nil {
		logs.Errorf("create credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

// ListCredentials get credentials
func (s *Service) ListCredentials(ctx context.Context, req *pbds.ListCredentialReq) (*pbds.ListCredentialResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	// StrToUint32Slice the comma separated string goes to uint32 slice
	topIds, _ := tools.StrToUint32Slice(req.TopIds)
	details, count, err := s.dao.Credential().List(kt, req.BizId, req.SearchKey, opt, topIds)

	if err != nil {
		logs.Errorf("list credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	credentialScopes := map[uint32][]string{}
	if count > 0 {
		credentialID := []uint32{}
		for _, v := range details {
			credentialID = append(credentialID, v.ID)
		}
		// 获取关联规则
		item, err := s.dao.CredentialScope().ListByCredentialIDs(kt, credentialID, req.BizId)
		if err != nil {
			return nil, err
		}
		for _, v := range item {
			app, scope, err := v.Spec.CredentialScope.Split()
			if err != nil {
				return nil, err
			}
			credentialScopes[v.Attachment.CredentialId] = append(credentialScopes[v.Attachment.CredentialId],
				fmt.Sprintf("%s%s", app, scope))
		}
	}

	data := pbcredential.PbCredentials(details)

	for _, v := range data {
		v.CredentialScopes = credentialScopes[v.Id]
	}

	resp := &pbds.ListCredentialResp{
		Count:   uint32(count),
		Details: data,
	}
	return resp, nil
}

// DeleteCredential delete credential
func (s *Service) DeleteCredential(ctx context.Context, req *pbds.DeleteCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()

	// 查看credential_scopes表中的数据
	if err := s.dao.CredentialScope().DeleteByCredentialIDWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete credential scope by credential id failed, err: %v, rid: %s", err, kt.Rid)
		if e := tx.Rollback(); e != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
		}
		return nil, err
	}

	if err := s.dao.Credential().DeleteWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete credential failed, err: %v, rid: %s", err, kt.Rid)
		if e := tx.Rollback(); e != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("transaction commit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// UpdateCredential update credential
func (s *Service) UpdateCredential(ctx context.Context, req *pbds.UpdateCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	old, err := s.dao.Credential().GetByName(kt, req.Attachment.BizId, req.Spec.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if !errors.Is(gorm.ErrRecordNotFound, err) && old.ID != req.Id {
		return nil, fmt.Errorf("credential name %s already exists", req.Spec.Name)
	}

	credential := &table.Credential{
		ID:         req.Id,
		Spec:       req.Spec.CredentialSpec(),
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if e := s.dao.Credential().Update(kt, credential); e != nil {
		logs.Errorf("update credential failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// CheckCredentialName Check if the credential name exists
func (s *Service) CheckCredentialName(ctx context.Context, req *pbds.CheckCredentialNameReq) (
	*pbds.CheckCredentialNameResp, error) {
	kt := kit.FromGrpcContext(ctx)

	credential, err := s.dao.Credential().GetByName(kt, req.BizId, req.CredentialName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var exist bool
	if credential != nil && credential.ID != 0 {
		exist = true
	}
	return &pbds.CheckCredentialNameResp{
		Exist: exist,
	}, nil
}

// CredentialScopePreview 关联规则预览配置项
func (s *Service) CredentialScopePreview(ctx context.Context, req *pbds.CredentialScopePreviewReq) (
	*pbds.CredentialScopePreviewResp, error) {
	kt := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().GetByName(kt, req.BizId, req.AppName)
	if err != nil {
		return nil, err
	}

	var preview []*pbds.CredentialScopePreviewResp_Detail
	if app.Spec.ConfigType == table.File {
		preview, err = s.getFileConfileItems(kt, app.ID, req.BizId, req.Scope, req.SearchValue)
		if err != nil {
			return nil, err
		}
	} else {
		preview, err = s.getKVConfigItems(kt, app.ID, req.BizId, req.Scope, req.SearchValue)
		if err != nil {
			return nil, err
		}
	}

	startIdx := int(req.Start)
	endIdx := startIdx + int(req.Limit)
	// 检查结束索引是否超出数据范围
	if endIdx >= len(preview) {
		endIdx = len(preview)
	}
	// 检查起始索引是否超出数据范围
	if startIdx >= len(preview) {
		startIdx = len(preview)
	}

	// 获取当前页的数据
	currentPageData := preview[startIdx:endIdx]

	return &pbds.CredentialScopePreviewResp{Details: currentPageData, Count: uint32(len(preview))}, nil
}

// 获取文件配置项
func (s *Service) getFileConfileItems(kt *kit.Kit, appID, bizID uint32, scope, searchValue string) (
	[]*pbds.CredentialScopePreviewResp_Detail, error) {

	status := []string{constant.FileStateAdd, constant.FileStateRevise, constant.FileStateUnchange}
	ci, err := s.ListConfigItems(kt.RpcCtx(), &pbds.ListConfigItemsReq{
		BizId:      bizID,
		AppId:      appID,
		All:        true,
		WithStatus: true,
		Status:     status,
	})
	if err != nil {
		return nil, err
	}

	tci, err := s.getAllUnPublishedTmpConfig(kt, appID, bizID, status)
	if err != nil {
		return nil, err
	}

	allCi := make([]*pbci.ConfigItem, len(ci.Details))
	copy(allCi, ci.Details)
	allCi = append(allCi, tci...)

	preview := []*pbds.CredentialScopePreviewResp_Detail{}
	for _, v := range allCi {
		if ok, _ := tools.MatchConfigItem(scope, v.Spec.Path, v.Spec.Name); !ok {
			continue
		}

		if searchValue != "" && !strings.Contains(path.Join(v.Spec.Path, v.Spec.Name), searchValue) {
			continue
		}
		preview = append(preview, &pbds.CredentialScopePreviewResp_Detail{
			Name: v.Spec.Name,
			Path: v.Spec.Path,
		})
	}
	return preview, nil
}

// 获取kv配置项
func (s *Service) getKVConfigItems(kt *kit.Kit, appID, bizID uint32, scope, searchValue string) (
	[]*pbds.CredentialScopePreviewResp_Detail, error) {
	preview := []*pbds.CredentialScopePreviewResp_Detail{}

	scope = strings.TrimPrefix(scope, "/")
	g, err := glob.Compile(scope)
	if err != nil {
		return preview, err
	}
	// 获取未删除的KV
	kvState := []string{
		string(table.KvStateAdd),
		string(table.KvStateRevise),
		string(table.KvStateUnchange),
	}
	kv, err := s.dao.Kv().ListAllByAppID(kt, appID, bizID, kvState)
	if err != nil {
		return nil, err
	}

	for _, v := range kv {
		if !g.Match(v.Spec.Key) {
			continue
		}
		if searchValue != "" && !strings.Contains(v.Spec.Key, searchValue) {
			continue
		}
		preview = append(preview, &pbds.CredentialScopePreviewResp_Detail{
			Name: v.Spec.Key,
		})
	}

	return preview, nil
}

// 获取未发布下的所有模板配置
func (s *Service) getAllUnPublishedTmpConfig(kt *kit.Kit, appID, bizID uint32,
	status []string) ([]*pbci.ConfigItem, error) {

	tmplSetInfo, err := s.getAllAppTmplSets(kt, bizID, appID)
	if err != nil {
		logs.Errorf("get all app template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var rp *pbds.ListAppBoundTmplRevisionsResp
	rp, err = s.ListAppBoundTmplRevisions(kt.RpcCtx(), &pbds.ListAppBoundTmplRevisionsReq{
		BizId:      bizID,
		AppId:      appID,
		All:        true,
		WithStatus: true,
	})
	if err != nil {
		logs.Errorf("list app template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// group by template set
	tmplSetMap := make(map[uint32][]*pbatb.AppBoundTmplRevision)
	for _, d := range rp.Details {
		tmplSetMap[d.TemplateSetId] = append(tmplSetMap[d.TemplateSetId], d)
	}
	details := make([]*pbatb.AppBoundTmplRevisionGroupBySet, 0)
	for _, tmplSet := range tmplSetInfo {
		group := &pbatb.AppBoundTmplRevisionGroupBySet{
			TemplateSpaceId:   tmplSet.TemplateSpaceId,
			TemplateSpaceName: tmplSet.TemplateSpaceName,
			TemplateSetId:     tmplSet.TemplateSetId,
			TemplateSetName:   tmplSet.TemplateSetName,
		}
		revisions := tmplSetMap[tmplSet.TemplateSetId]
		for _, r := range revisions {
			group.TemplateRevisions = append(group.TemplateRevisions,
				&pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail{
					Name:      r.Name,
					Path:      r.Path,
					FileState: r.FileState,
				})
		}
		// 过滤删除的配置
		sortFileStateInGroup(group, status)
		details = append(details, group)
	}
	tci := []*pbci.ConfigItem{}
	for _, v := range details {
		for _, vv := range v.TemplateRevisions {
			tci = append(tci, &pbci.ConfigItem{
				Spec: &pbci.ConfigItemSpec{
					Name: vv.Name,
					Path: vv.Path,
				},
			})
		}
	}
	return tci, nil
}

func (s *Service) getAllAppTmplSets(grpcKit *kit.Kit, bizID, appID uint32) ([]*pbtset.TemplateSetBriefInfo, error) {
	atbReq := &pbds.ListAppTemplateBindingsReq{
		BizId: bizID,
		AppId: appID,
		All:   true,
	}

	atbRsp, err := s.ListAppTemplateBindings(grpcKit.RpcCtx(), atbReq)
	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if len(atbRsp.Details) == 0 {
		return []*pbtset.TemplateSetBriefInfo{}, nil
	}
	tmplSetIDs := make([]uint32, 0)
	for _, b := range atbRsp.Details[0].Spec.Bindings {
		tmplSetIDs = append(tmplSetIDs, b.TemplateSetId)
	}

	var tsbRsp *pbds.ListTemplateSetBriefInfoByIDsResp
	tsbRsp, err = s.ListTemplateSetBriefInfoByIDs(grpcKit.RpcCtx(), &pbds.ListTemplateSetBriefInfoByIDsReq{
		Ids: tmplSetIDs,
	})
	if err != nil {
		logs.Errorf("list template set brief info by ids failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return tsbRsp.Details, nil
}

// sortFileStateInGroup sort as add > revise > delete > unchange
func sortFileStateInGroup(g *pbatb.AppBoundTmplRevisionGroupBySet, status []string) {
	if len(g.TemplateRevisions) <= 1 {
		return
	}

	result := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	add := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	del := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	revise := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	unchange := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	for _, ci := range g.TemplateRevisions {
		switch ci.FileState {
		case constant.FileStateAdd:
			add = append(add, ci)
		case constant.FileStateDelete:
			del = append(del, ci)
		case constant.FileStateRevise:
			revise = append(revise, ci)
		case constant.FileStateUnchange:
			unchange = append(unchange, ci)
		}
	}

	if len(status) == 0 {
		result = append(result, add...)
		result = append(result, revise...)
		result = append(result, del...)
		result = append(result, unchange...)
	} else {
		for _, v := range status {
			switch strings.ToUpper(v) {
			case constant.FileStateAdd:
				result = append(result, add...)
			case constant.FileStateRevise:
				result = append(result, revise...)
			case constant.FileStateDelete:
				result = append(result, del...)
			case constant.FileStateUnchange:
				result = append(result, unchange...)
			}
		}
	}
	g.TemplateRevisions = result
}
