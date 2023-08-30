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
	"fmt"
	"io/ioutil"
	"sync"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbatv "bscp.io/pkg/protocol/core/app-template-variable"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbtv "bscp.io/pkg/protocol/core/template-variable"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// ExtractAppTemplateVariables extract app template variables.
func (s *Service) ExtractAppTemplateVariables(ctx context.Context, req *pbds.ExtractAppTemplateVariablesReq) (
	*pbds.ExtractAppTemplateVariablesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tmplRevisions, err := s.getAppTmplRevisions(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("extract app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(tmplRevisions) == 0 {
		return &pbds.ExtractAppTemplateVariablesResp{
			Details: []string{},
		}, nil
	}

	contents, err := s.downloadTmplContent(kt, tmplRevisions)
	if err != nil {
		logs.Errorf("extract app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// merge all template content
	allContent := bytes.Join(contents, []byte(" "))
	// extract all template variables
	variables := s.tmplProc.ExtractVariables(allContent)

	return &pbds.ExtractAppTemplateVariablesResp{
		Details: variables,
	}, nil
}

// getAppTmplRevisions get app template revision details
func (s *Service) getAppTmplRevisions(kt *kit.Kit, bizID, appID uint32) ([]*table.TemplateRevision, error) {
	opt := &types.BasePage{All: true}
	details, _, err := s.dao.AppTemplateBinding().List(kt, bizID, appID, opt)
	if err != nil {
		logs.Errorf("get app template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// so far, no any template config item exists for the app
	if len(details) == 0 {
		return nil, nil
	}

	// get template revision details
	tmplRevisions, err := s.dao.TemplateRevision().
		ListByIDs(kt, details[0].Spec.TemplateRevisionIDs)
	if err != nil {
		logs.Errorf("get app template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return tmplRevisions, nil
}

// downloadTmplContent download template config item content from repo.
// the order of elements in slice contents and slice tmplRevisions is consistent
func (s *Service) downloadTmplContent(kt *kit.Kit, tmplRevisions []*table.TemplateRevision) ([][]byte, error) {
	contents := make([][]byte, len(tmplRevisions))
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, r := range tmplRevisions {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, r *table.TemplateRevision) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			k := kt.GetKitForRepoTmpl(r.Attachment.TemplateSpaceID)
			body, _, err := s.repo.Download(k, r.Spec.ContentSpec.Signature)
			if err != nil {
				hitError = fmt.Errorf("download template config content from repo failed, "+
					"template id: %d, name: %s, path: %s, error: %v",
					r.Attachment.TemplateID, r.Spec.Name, r.Spec.Path, err)
				return
			}
			content, err := ioutil.ReadAll(body)
			if err != nil {
				hitError = fmt.Errorf("read template config content from body failed, "+
					"template id: %d, name: %s, path: %s, error: %v",
					r.Attachment.TemplateID, r.Spec.Name, r.Spec.Path, err)
				return
			}

			contents[idx] = content
		}(idx, r)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("download template content failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	return contents, nil
}

// GetAppTemplateVariableReferences get app template variable references.
func (s *Service) GetAppTemplateVariableReferences(ctx context.Context, req *pbds.GetAppTemplateVariableReferencesReq) (
	*pbds.GetAppTemplateVariableReferencesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tmplRevisions, err := s.getAppTmplRevisions(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("get app template variable references failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(tmplRevisions) == 0 {
		return &pbds.GetAppTemplateVariableReferencesResp{
			Details: []*pbatv.AppTemplateVariableReference{},
		}, nil
	}

	contents, err := s.downloadTmplContent(kt, tmplRevisions)
	if err != nil {
		logs.Errorf("get app template variable references failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	allVariables := make([]string, 0)
	revisionVariableMap := make(map[uint32]map[string]struct{}, len(tmplRevisions))
	revisionMap := make(map[uint32]*table.TemplateRevision, len(tmplRevisions))
	for idx, r := range tmplRevisions {
		// extract template variables for one template config item
		variables := s.tmplProc.ExtractVariables(contents[idx])
		allVariables = append(allVariables, variables...)
		revisionMap[r.ID] = r

		revisionVariableMap[r.ID] = map[string]struct{}{}
		for _, v := range variables {
			revisionVariableMap[r.ID][v] = struct{}{}
		}
	}
	allVariables = tools.RemoveDuplicateStrings(allVariables)

	refs := make([]*pbatv.AppTemplateVariableReference, len(allVariables))
	for idx, v := range allVariables {
		ref := &pbatv.AppTemplateVariableReference{
			VariableName: v,
		}
		for rID, variables := range revisionVariableMap {
			if _, ok := variables[v]; ok {
				ref.References = append(ref.References, &pbatv.AppTemplateVariableReferenceReference{
					TemplateId:         revisionMap[rID].Attachment.TemplateID,
					TemplateRevisionId: rID,
					Name:               revisionMap[rID].Spec.Name,
				})
			}
		}
		refs[idx] = ref
	}

	return &pbds.GetAppTemplateVariableReferencesResp{
		Details: refs,
	}, nil
}

// ListAppTemplateVariables get app template variable references.
func (s *Service) ListAppTemplateVariables(ctx context.Context, req *pbds.ListAppTemplateVariablesReq) (
	*pbds.ListAppTemplateVariablesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// extract all variables for current app
	extractRep, err := s.ExtractAppTemplateVariables(ctx, &pbds.ExtractAppTemplateVariablesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	})
	if err != nil {
		logs.Errorf("list app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	allVariables := extractRep.Details
	if len(allVariables) == 0 {
		return &pbds.ListAppTemplateVariablesResp{
			Details: []*pbtv.TemplateVariableSpec{},
		}, nil
	}

	// get app template variables
	appVars, err := s.dao.AppTemplateVariable().ListVariables(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appVarMap := make(map[string]*table.TemplateVariableSpec, len(appVars))
	for _, v := range appVars {
		appVarMap[v.Name] = v
	}

	// get biz template variables
	bizVars, _, err := s.dao.TemplateVariable().List(kt, req.BizId, nil, &types.BasePage{All: true})
	if err != nil {
		logs.Errorf("list app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	bizVarMap := make(map[string]*table.TemplateVariableSpec, len(bizVars))
	for _, v := range bizVars {
		bizVarMap[v.Spec.Name] = v.Spec
	}

	// get final app template variables
	// use app variables first, then use biz variables
	finalVar := make([]*pbtv.TemplateVariableSpec, 0)
	for _, name := range allVariables {
		if v, ok := appVarMap[name]; ok {
			finalVar = append(finalVar, pbtv.PbTemplateVariableSpec(v))
			continue
		}
		if v, ok := bizVarMap[name]; ok {
			finalVar = append(finalVar, pbtv.PbTemplateVariableSpec(v))
		}
	}

	return &pbds.ListAppTemplateVariablesResp{
		Details: finalVar,
	}, nil
}

// UpdateAppTemplateVariables update app template variables.
func (s *Service) UpdateAppTemplateVariables(ctx context.Context, req *pbds.UpdateAppTemplateVariablesReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	appVar := &table.AppTemplateVariable{
		Spec:       req.Spec.AppTemplateVariableSpec(),
		Attachment: req.Attachment.AppTemplateVariableAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	if err := s.dao.AppTemplateVariable().Upsert(kt, appVar); err != nil {
		logs.Errorf("update app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
