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
	"bytes"
	"context"
	"errors"
	"fmt"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/tools"
	pbstruct "github.com/golang/protobuf/ptypes/struct"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateRelease create release.
func (s *Service) CreateRelease(ctx context.Context, req *pbds.CreateReleaseReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	// the url path doesn't include appID, set the appID which used by create released rendered template config items
	grpcKit.AppID = req.Attachment.AppId
	// Note: need to change batch operator to query config item and its commit.
	// step1: query app's all config items.
	cfgItems, err := s.dao.ConfigItem().ListAllByAppID(grpcKit, req.Attachment.AppId, req.Attachment.BizId)
	if err != nil {
		logs.Errorf("query app config item list failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// get app template revisions which are template config items
	tmplRevisions, err := s.getAppTmplRevisions(grpcKit, req.Attachment.BizId, req.Attachment.AppId)
	if err != nil {
		logs.Errorf("get app template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// if no config item, return directly.
	if len(cfgItems) == 0 && len(tmplRevisions) == 0 {
		return nil, errors.New("app config items is empty")
	}

	if _, e := s.dao.Release().GetByName(grpcKit, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Name); e == nil {
		return nil, fmt.Errorf("release name %s already exists", req.Spec.Name)
	}
	// begin transaction to create release and released config item.
	tx := s.dao.GenQuery().Begin()
	// 1. create release, and create release and released config item need to begin tx.
	release := &table.Release{
		Spec:       req.Spec.ReleaseSpec(),
		Attachment: req.Attachment.ReleaseAttachment(),
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}
	id, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		tx.Rollback()
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 2. create released hook.
	pre, err := s.dao.ReleasedHook().Get(grpcKit, req.Attachment.BizId, req.Attachment.AppId, 0, table.PreHook)
	if err == nil {
		pre.ID = 0
		pre.ReleaseID = release.ID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, pre); e != nil {
			logs.Errorf("create released pre-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			tx.Rollback()
			return nil, e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released pre-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback()
		return nil, err
	}
	post, err := s.dao.ReleasedHook().Get(grpcKit, req.Attachment.BizId, req.Attachment.AppId, 0, table.PostHook)
	if err == nil {
		post.ID = 0
		post.ReleaseID = release.ID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, post); e != nil {
			logs.Errorf("create released post-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			tx.Rollback()
			return nil, e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released post-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback()
		return nil, err
	}

	// 3. do config item action for create release.
	if err = s.doConfigItemActionForCreateRelease(grpcKit, req, tx, release.ID, cfgItems); err != nil {
		tx.Rollback()
		logs.Errorf("do config item action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 4: do template action for create release.
	if err = s.doTemplateActionForCreateRelease(grpcKit, req, tx, release.ID, tmplRevisions); err != nil {
		tx.Rollback()
		logs.Errorf("do template action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 5: commit transaction.
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	return &pbds.CreateResp{Id: id}, nil
}

// doConfigItemActionForCreateRelease do config item action for create release
func (s *Service) doConfigItemActionForCreateRelease(kt *kit.Kit, req *pbds.CreateReleaseReq, tx *gen.QueryTx,
	releaseID uint32, cfgItems []*table.ConfigItem) error {
	releasedCIs := make([]*table.ReleasedConfigItem, 0)
	if len(cfgItems) == 0 {
		return nil
	}

	// query config item newest commit
	for _, item := range cfgItems {
		commit, e := s.dao.Commit().GetLatestCommit(kt, req.Attachment.BizId, req.Attachment.AppId, item.ID)
		if e != nil {
			logs.Errorf("query config item latest commit failed, err: %v, rid: %s", e, kt.Rid)
			return e
		}
		releasedCIs = append(releasedCIs, &table.ReleasedConfigItem{
			CommitID:       commit.ID,
			CommitSpec:     commit.Spec,
			ConfigItemID:   item.ID,
			ConfigItemSpec: item.Spec,
			Attachment:     item.Attachment,
			Revision:       item.Revision,
		})
	}

	// create released config item
	for _, rci := range releasedCIs {
		rci.ReleaseID = releaseID
	}
	if err := s.dao.ReleasedCI().BulkCreateWithTx(kt, tx, releasedCIs); err != nil {
		logs.Errorf("bulk create released config item failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// doTemplateActionForCreateRelease do template action for create release.
/*
1.下载服务的所有模版文件内容，提取服务模版变量
2.获取入参变量和业务变量，判断是否缺少变量，缺少则报错
3.使用变量渲染模版文件，上传渲染后的内容
4.创建已生成版本服务的模版配置项
5.创建已生成版本服务的服务模版详情
6.创建已生成版本服务的模版变量
7.将当前使用变量更新到未命名版本的服务模版变量
*/
func (s *Service) doTemplateActionForCreateRelease(kt *kit.Kit, req *pbds.CreateReleaseReq, tx *gen.QueryTx,
	releaseID uint32, tmplRevisions []*table.TemplateRevision) error {
	if len(tmplRevisions) == 0 {
		return nil
	}

	// validate input variables and get the map
	inputVarMap := make(map[string]*table.TemplateVariableSpec)
	for _, v := range req.Variables {
		if v == nil {
			continue
		}
		if err := v.TemplateVariableSpec().ValidateCreate(); err != nil {
			logs.Errorf("validate template variables failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		inputVarMap[v.Name] = v.TemplateVariableSpec()
	}

	contents, err := s.downloadTmplContent(kt, tmplRevisions)
	if err != nil {
		logs.Errorf("download template content failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// merge all template content
	allContent := bytes.Join(contents, []byte(" "))
	// extract all template variables
	allVars := s.tmplProc.ExtractVariables(allContent)

	// get biz template variables
	bizVars, _, err := s.dao.TemplateVariable().List(kt, req.Attachment.BizId, nil, &types.BasePage{All: true})
	if err != nil {
		logs.Errorf("list template variables failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	bizVarMap := make(map[string]*table.TemplateVariableSpec, len(bizVars))
	for _, v := range bizVars {
		bizVarMap[v.Spec.Name] = v.Spec
	}

	// get variables which are used to render the template
	usedVars := make([]*table.TemplateVariableSpec, 0)
	renderKV := make(map[string]interface{})
	var missingVars []string
	for _, name := range allVars {
		if _, ok := inputVarMap[name]; ok {
			usedVars = append(usedVars, inputVarMap[name])
			renderKV[name] = inputVarMap[name].DefaultVal
			continue
		}
		if _, ok := bizVarMap[name]; ok {
			usedVars = append(usedVars, bizVarMap[name])
			renderKV[name] = bizVarMap[name].DefaultVal
			continue
		}
		missingVars = append(missingVars, name)
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("variable name in %v is missing for render the app's template config", missingVars)
	}

	// get rendered content map which is template revision id => rendered content
	renderedContentMap := make(map[uint32][]byte, len(tmplRevisions))
	signatureMap := make(map[uint32]string, len(tmplRevisions))
	revisionMap := make(map[uint32]*table.TemplateRevision, len(tmplRevisions))
	for idx, r := range tmplRevisions {
		revisionMap[r.ID] = r
		renderedContentMap[r.ID] = s.tmplProc.Render(contents[idx], renderKV)
		signatureMap[r.ID] = tools.ByteSHA256(renderedContentMap[r.ID])
	}

	// upload rendered template content
	if err := s.uploadRenderedTmplContent(kt, renderedContentMap, signatureMap, revisionMap); err != nil {
		logs.Errorf("upload rendered template failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := s.createReleasedRenderedTemplateCIs(kt, tx, releaseID, tmplRevisions, renderedContentMap, signatureMap); err != nil {
		logs.Errorf("create released rendered template config items failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := s.createReleasedAppTemplates(kt, tx, releaseID, renderedContentMap, signatureMap); err != nil {
		logs.Errorf("create released rendered template config items failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := s.createReleasedAppTemplateVariable(kt, tx, releaseID, usedVars); err != nil {
		logs.Errorf("create released app template variable failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := s.updateAppTemplateVariable(kt, tx, usedVars); err != nil {
		logs.Errorf("update app template variable failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// createReleasedRenderedTemplateCIs create released rendered templates config items.
func (s *Service) createReleasedRenderedTemplateCIs(kt *kit.Kit, tx *gen.QueryTx, releaseID uint32,
	tmplRevisions []*table.TemplateRevision, renderedContentMap map[uint32][]byte, signatureMap map[uint32]string) error {
	releasedCIs := make([]*table.ReleasedConfigItem, len(tmplRevisions))
	for idx, r := range tmplRevisions {
		releasedCIs[idx] = &table.ReleasedConfigItem{
			ReleaseID: releaseID,
			CommitSpec: &table.CommitSpec{
				ContentID: 0,
				Content: &table.ContentSpec{
					Signature: signatureMap[r.ID],
					ByteSize:  uint64(len(renderedContentMap[r.ID])),
				},
				Memo: "",
			},
			ConfigItemSpec: &table.ConfigItemSpec{
				Name:       r.Spec.Name,
				Path:       r.Spec.Path,
				FileType:   r.Spec.FileType,
				FileMode:   r.Spec.FileMode,
				Memo:       r.Spec.RevisionMemo,
				Permission: r.Spec.Permission,
			},
			Attachment: &table.ConfigItemAttachment{
				BizID: kt.BizID,
				AppID: kt.AppID,
			},
			Revision: &table.Revision{
				Creator: kt.User,
				Reviser: kt.User,
			},
		}
	}
	if err := s.dao.ReleasedCI().BulkCreateWithTx(kt, tx, releasedCIs); err != nil {
		logs.Errorf("bulk create released rendered template config item failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// createReleasedAppTemplates create released app templates.
func (s *Service) createReleasedAppTemplates(kt *kit.Kit, tx *gen.QueryTx, releaseID uint32,
	renderedContentMap map[uint32][]byte, signatureMap map[uint32]string) error {
	revisionsResp, err := s.ListAppBoundTemplateRevisions(kt.Ctx, &pbds.ListAppBoundTemplateRevisionsReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
		All:   true,
	})
	if err != nil {
		logs.Errorf("list app bound template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	revisions := revisionsResp.Details

	releasedATs := make([]*table.ReleasedAppTemplate, len(revisions))
	for idx, r := range revisions {
		releasedATs[idx] = &table.ReleasedAppTemplate{
			Spec: &table.ReleasedAppTemplateSpec{
				ReleaseID:            releaseID,
				TemplateSpaceID:      r.TemplateSpaceId,
				TemplateSpaceName:    r.TemplateSpaceName,
				TemplateSetID:        r.TemplateSetId,
				TemplateSetName:      r.TemplateSetName,
				TemplateID:           r.TemplateId,
				Name:                 r.Name,
				Path:                 r.Path,
				TemplateRevisionID:   r.TemplateRevisionId,
				IsLatest:             r.IsLatest,
				TemplateRevisionName: r.TemplateRevisionName,
				TemplateRevisionMemo: r.TemplateRevisionMemo,
				FileType:             r.FileType,
				FileMode:             r.FileMode,
				User:                 r.User,
				UserGroup:            r.UserGroup,
				Privilege:            r.Privilege,
				Signature:            r.Signature,
				ByteSize:             r.ByteSize,
				RenderedSignature:    signatureMap[r.TemplateRevisionId],
				RenderedByteSize:     uint64(len(renderedContentMap[r.TemplateRevisionId])),
			},
			Attachment: &table.ReleasedAppTemplateAttachment{
				BizID: kt.BizID,
				AppID: kt.AppID,
			},
			Revision: &table.CreatedRevision{
				Creator: kt.User,
			},
		}
	}
	if err = s.dao.ReleasedAppTemplate().BulkCreateWithTx(kt, tx, releasedATs); err != nil {
		logs.Errorf("bulk create released app template config item failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// createReleasedAppTemplates create released app template variable.
func (s *Service) createReleasedAppTemplateVariable(kt *kit.Kit, tx *gen.QueryTx, releaseID uint32,
	usedVars []*table.TemplateVariableSpec) error {
	releasedAppVar := &table.ReleasedAppTemplateVariable{
		Spec: &table.ReleasedAppTemplateVariableSpec{
			ReleaseID: releaseID,
			Variables: usedVars,
		},
		Attachment: &table.ReleasedAppTemplateVariableAttachment{
			BizID: kt.BizID,
			AppID: kt.AppID,
		},
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	if _, err := s.dao.ReleasedAppTemplateVariable().CreateWithTx(kt, tx, releasedAppVar); err != nil {
		logs.Errorf("create released app template variable failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// updateAppTemplateVariable update app template variable.
func (s *Service) updateAppTemplateVariable(kt *kit.Kit, tx *gen.QueryTx, usedVars []*table.TemplateVariableSpec) error {
	appVar := &table.AppTemplateVariable{
		Spec: &table.AppTemplateVariableSpec{
			Variables: usedVars,
		},
		Attachment: &table.AppTemplateVariableAttachment{
			BizID: kt.BizID,
			AppID: kt.AppID,
		},
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	if err := s.dao.AppTemplateVariable().UpsertWithTx(kt, tx, appVar); err != nil {
		logs.Errorf("upsert app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// ListReleases list releases.
func (s *Service) ListReleases(ctx context.Context, req *pbds.ListReleasesReq) (*pbds.ListReleasesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	query := &types.ListReleasesOption{
		BizID: req.BizId,
		AppID: req.AppId,
		Page: &types.BasePage{
			Start: req.Start,
			Limit: uint(req.Limit),
		},
		Deprecated: req.Deprecated,
		SearchKey:  req.SearchKey,
	}
	if req.All {
		query.Page.Start = 0
		query.Page.Limit = 0
	}

	details, err := s.dao.Release().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	releases := pbrelease.PbReleases(details.Details)

	gcrs, err := s.dao.ReleasedGroup().ListAllByAppID(grpcKit, req.AppId, req.BizId)
	if err != nil {
		logs.Errorf("list group current releases failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	groups, err := s.dao.Group().ListAppGroups(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list app groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	for _, release := range releases {
		status, selected := s.queryPublishStatus(gcrs, release.Id)
		releasedGroups := make([]*pbrelease.ReleaseStatus_ReleasedGroup, 0)
		for _, gcr := range selected {
			if gcr.GroupID == 0 {
				releasedGroups = append(releasedGroups, &pbrelease.ReleaseStatus_ReleasedGroup{
					Id:   0,
					Name: "默认分组",
					Mode: table.Default.String(),
				})
			}
			for _, group := range groups {
				if group.ID == gcr.GroupID {
					oldSelector := new(pbstruct.Struct)
					newSelector := new(pbstruct.Struct)
					if gcr.Selector != nil {
						s, err := gcr.Selector.MarshalPB()
						if err != nil {
							return nil, err
						}
						oldSelector = s
					}
					if group.Spec.Selector != nil {
						s, err := group.Spec.Selector.MarshalPB()
						if err != nil {
							return nil, err
						}
						newSelector = s
					}
					releasedGroups = append(releasedGroups, &pbrelease.ReleaseStatus_ReleasedGroup{
						Id:          group.ID,
						Name:        group.Spec.Name,
						Mode:        gcr.Mode.String(),
						OldSelector: oldSelector,
						NewSelector: newSelector,
						Edited:      gcr.Edited,
					})
					break
				}
			}
		}
		release.Status = &pbrelease.ReleaseStatus{
			PublishStatus:  status,
			ReleasedGroups: releasedGroups,
		}
	}

	resp := &pbds.ListReleasesResp{
		Count:   details.Count,
		Details: releases,
	}
	return resp, nil
}

// GetReleaseByName get release by release name.
func (s *Service) GetReleaseByName(ctx context.Context, req *pbds.GetReleaseByNameReq) (*pbrelease.Release, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	release, err := s.dao.Release().GetByName(grpcKit, req.GetBizId(), req.GetAppId(), req.GetReleaseName())
	if err != nil {
		logs.Errorf("get release by name failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, fmt.Errorf("query release by name %s failed", req.GetReleaseName())
	}

	return pbrelease.PbRelease(release), nil
}

func (s *Service) queryPublishStatus(gcrs []*table.ReleasedGroup, releaseID uint32) (
	string, []*table.ReleasedGroup) {
	var includeDefault = false
	var inRelease = make([]*table.ReleasedGroup, 0)
	var outRelease = make([]*table.ReleasedGroup, 0)
	for _, gcr := range gcrs {
		if gcr.ReleaseID == releaseID {
			inRelease = append(inRelease, gcr)
			if gcr.GroupID == 0 {
				includeDefault = true
			}
		} else {
			outRelease = append(outRelease, gcr)
		}
	}

	// len(inRelease) == 0: not released
	if len(inRelease) == 0 {
		return table.NotReleased.String(), inRelease
		// len(inRelease) != 0 && len(outRelease) != 0: gray released
	} else if len(outRelease) != 0 {
		return table.PartialReleased.String(), inRelease
		// len(inRelease) != 0 && len(outRelease) == 0 && includeDefault: full released
	} else if includeDefault {
		return table.FullReleased.String(), inRelease
		// len(inRelease) != 0 && len(outRelease) == 0 && !includeDefault: gray released
	}
	return table.PartialReleased.String(), inRelease
}
