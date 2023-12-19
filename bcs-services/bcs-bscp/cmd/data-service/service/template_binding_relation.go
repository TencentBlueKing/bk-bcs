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
	"strconv"
	"strings"
	"sync"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbtbr "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-binding-relation"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ListTmplBoundCounts list template bound counts.
func (s *Service) ListTmplBoundCounts(ctx context.Context, req *pbds.ListTmplBoundCountsReq) (
	*pbds.ListTmplBoundCountsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}

	var hitError error
	details := make([]*pbtbr.TemplateBoundCounts, len(req.TemplateIds))
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, tmplID := range req.TemplateIds {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, tmplID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			var (
				unnamedAppCnt, namedAppCnt, tmplSetCnt uint32
				err                                    error
			)

			if unnamedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateBoundUnnamedAppCount(kt, req.BizId, tmplID); err != nil {
				hitError = err
				return
			}
			if namedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateBoundNamedAppCount(kt, req.BizId, tmplID); err != nil {
				hitError = err
				return
			}
			if tmplSetCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateBoundTemplateSetCount(kt, req.BizId, tmplID); err != nil {
				hitError = err
				return
			}

			// save the result with index
			details[idx] = &pbtbr.TemplateBoundCounts{
				TemplateId:            tmplID,
				BoundUnnamedAppCount:  unnamedAppCnt,
				BoundNamedAppCount:    namedAppCnt,
				BoundTemplateSetCount: tmplSetCnt,
			}
		}(idx, tmplID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("list template bound counts failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	resp := &pbds.ListTmplBoundCountsResp{
		Details: details,
	}
	return resp, nil
}

// ListTmplRevisionBoundCounts list template revision bound counts.
func (s *Service) ListTmplRevisionBoundCounts(ctx context.Context, req *pbds.ListTmplRevisionBoundCountsReq) (
	*pbds.ListTmplRevisionBoundCountsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplRevisionsExist(kt, req.TemplateRevisionIds); err != nil {
		return nil, err
	}

	var hitError error
	details := make([]*pbtbr.TemplateRevisionBoundCounts, len(req.TemplateRevisionIds))
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, tmplRevisionID := range req.TemplateRevisionIds {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, tmplRevisionID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			var (
				unnamedAppCnt, namedAppCnt uint32
				err                        error
			)

			if unnamedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateRevisionBoundUnnamedAppCount(kt, req.BizId, tmplRevisionID); err != nil {
				hitError = err
				return
			}
			if namedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateRevisionBoundNamedAppCount(kt, req.BizId, tmplRevisionID); err != nil {
				hitError = err
				return
			}

			// save the result with index
			details[idx] = &pbtbr.TemplateRevisionBoundCounts{
				TemplateRevisionId:   tmplRevisionID,
				BoundUnnamedAppCount: unnamedAppCnt,
				BoundNamedAppCount:   namedAppCnt,
			}
		}(idx, tmplRevisionID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("list template revision bound counts failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	resp := &pbds.ListTmplRevisionBoundCountsResp{
		Details: details,
	}
	return resp, nil
}

// ListTmplSetBoundCounts list template bound counts.
func (s *Service) ListTmplSetBoundCounts(ctx context.Context, req *pbds.ListTmplSetBoundCountsReq) (
	*pbds.ListTmplSetBoundCountsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	var hitError error
	details := make([]*pbtbr.TemplateSetBoundCounts, len(req.TemplateSetIds))
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, tmplSetID := range req.TemplateSetIds {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, tmplSetID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			var (
				unnamedAppCnt, namedAppCnt uint32
				err                        error
			)

			if unnamedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateSetBoundUnnamedAppCount(kt, req.BizId, tmplSetID); err != nil {
				hitError = err
				return
			}
			if namedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateSetBoundNamedAppCount(kt, req.BizId, tmplSetID); err != nil {
				hitError = err
				return
			}

			// save the result with index
			details[idx] = &pbtbr.TemplateSetBoundCounts{
				TemplateSetId:        tmplSetID,
				BoundUnnamedAppCount: unnamedAppCnt,
				BoundNamedAppCount:   namedAppCnt,
			}
		}(idx, tmplSetID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("list template set bound counts failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	resp := &pbds.ListTmplSetBoundCountsResp{
		Details: details,
	}
	return resp, nil
}

// ListTmplBoundUnnamedApps list template bound unnamed app details.
//
//nolint:funlen
func (s *Service) ListTmplBoundUnnamedApps(ctx context.Context,
	req *pbds.ListTmplBoundUnnamedAppsReq) (
	*pbds.ListTmplBoundUnnamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTmplBoundUnnamedApps(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get template revision details of the template
	tmplRevisions, _, err := s.dao.TemplateRevision().
		List(kt, req.BizId, req.TemplateId, nil, &types.BasePage{All: true})
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplRevisionMap := make(map[uint32]*table.TemplateRevision, len(tmplRevisions))
	for _, r := range tmplRevisions {
		tmplRevisionMap[r.ID] = r
	}

	// get app details
	appIDs := make([]uint32, len(relations))
	for i, r := range relations {
		appIDs[i] = r.AppID
	}
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// combine resp details
	details := make([]*pbtbr.TemplateBoundUnnamedAppDetail, 0)
	for _, r := range relations {
		for _, id := range r.TemplateRevisionIDs {
			// if app doesn't exist, ignore the invalid record
			if _, ok := appMap[r.AppID]; !ok {
				continue
			}
			// the template revision must belong to the target template
			if _, ok := tmplRevisionMap[id]; !ok {
				continue
			}
			details = append(details, &pbtbr.TemplateBoundUnnamedAppDetail{
				TemplateRevisionId:   id,
				TemplateRevisionName: tmplRevisionMap[id].Spec.RevisionName,
				AppId:                r.AppID,
				AppName:              appMap[r.AppID].Spec.Name,
			})

		}
	}

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TmplBoundUnnamedApp)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		newDetails := make([]*pbtbr.TemplateBoundUnnamedAppDetail, 0)
		for _, detail := range details {
			if (fieldsMap["app_name"] && strings.Contains(detail.AppName, req.SearchValue)) ||
				(fieldsMap["template_revision_name"] && strings.Contains(detail.TemplateRevisionName,
					req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplBoundUnnamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplBoundUnnamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTmplBoundNamedApps list template bound named app details.
// Deprecated: not in use currently
// if use it, consider to add column app_name, release_name on table released_app_templates in case of app is deleted.
//
//nolint:funlen
func (s *Service) ListTmplBoundNamedApps(ctx context.Context,
	req *pbds.ListTmplBoundNamedAppsReq) (*pbds.ListTmplBoundNamedAppsResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTmplBoundNamedApps(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get template revision details of the template
	tmplRevisions, _, err := s.dao.TemplateRevision().
		List(kt, req.BizId, req.TemplateId, nil, &types.BasePage{All: true})
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplRevisionMap := make(map[uint32]*table.TemplateRevision, len(tmplRevisions))
	for _, r := range tmplRevisions {
		tmplRevisionMap[r.ID] = r
	}

	// get app and release details
	appIDs := make([]uint32, len(relations))
	releaseIDs := make([]uint32, len(relations))
	for i, r := range relations {
		appIDs[i] = r.AppID
		releaseIDs[i] = r.ReleaseID
	}
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}
	releaseIDs = tools.RemoveDuplicates(releaseIDs)

	releases, err := s.dao.Release().ListAllByIDs(kt, releaseIDs, req.BizId)
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	releaseMap := make(map[uint32]*table.Release, len(releases))
	for _, r := range releases {
		releaseMap[r.ID] = r
	}

	// combine resp details
	details := make([]*pbtbr.TemplateBoundNamedAppDetail, 0)
	for _, r := range relations {
		// the template revision must belong to the target template
		if _, ok := tmplRevisionMap[r.TemplateRevisionID]; !ok {
			continue
		}
		appName := ""
		if _, ok := appMap[r.AppID]; ok {
			appName = appMap[r.AppID].Spec.Name
		}
		details = append(details, &pbtbr.TemplateBoundNamedAppDetail{
			TemplateRevisionId:   r.TemplateRevisionID,
			TemplateRevisionName: tmplRevisionMap[r.TemplateRevisionID].Spec.RevisionName,
			AppId:                r.AppID,
			AppName:              appName,
			ReleaseId:            r.ReleaseID,
			ReleaseName:          releaseMap[r.ReleaseID].Spec.Name,
		})
	}

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TmplBoundNamedApp)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		newDetails := make([]*pbtbr.TemplateBoundNamedAppDetail, 0)
		for _, detail := range details {
			if (fieldsMap["app_name"] && strings.Contains(detail.AppName, req.SearchValue)) ||
				(fieldsMap["template_revision_name"] && strings.Contains(detail.TemplateRevisionName,
					req.SearchValue)) ||
				(fieldsMap["release_name"] && strings.Contains(detail.ReleaseName, req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplBoundNamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplBoundNamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTmplBoundTmplSets list template bound template set details.
func (s *Service) ListTmplBoundTmplSets(ctx context.Context,
	req *pbds.ListTmplBoundTmplSetsReq) (
	*pbds.ListTmplBoundTmplSetsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	tmplSetIDs, err := s.dao.TemplateBindingRelation().
		ListTmplBoundTmplSets(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template bound template set details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, tmplSetIDs)
	if err != nil {
		logs.Errorf("list template bound template set details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSetMap := make(map[uint32]*table.TemplateSet, len(tmplSets))
	for _, t := range tmplSets {
		tmplSetMap[t.ID] = t
	}

	// combine resp details
	details := make([]*pbtbr.TemplateBoundTemplateSetDetail, 0)
	for _, id := range tmplSetIDs {
		details = append(details, &pbtbr.TemplateBoundTemplateSetDetail{
			TemplateSetId:   id,
			TemplateSetName: tmplSetMap[id].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplBoundTmplSetsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplBoundTmplSetsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListMultiTmplBoundTmplSets list template bound template set details.
//
//nolint:funlen
func (s *Service) ListMultiTmplBoundTmplSets(ctx context.Context,
	req *pbds.ListMultiTmplBoundTmplSetsReq) (
	*pbds.ListMultiTmplBoundTmplSetsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	templateIDs := tools.RemoveDuplicates(req.TemplateIds)
	if err := s.dao.Validator().ValidateTemplatesExist(kt, templateIDs); err != nil {
		return nil, err
	}

	allTmplSetIDs := make([]uint32, 0)
	tmplTmplSetsMap := make(map[uint32][]uint32)
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	for _, tmplID := range templateIDs {
		wg.Add(1)

		pipe <- struct{}{}
		go func(tmplID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			tmplSetIDs, err := s.dao.TemplateBindingRelation().
				ListTmplBoundTmplSets(kt, req.BizId, tmplID)
			if err != nil {
				hitError = err
				return
			}
			lock.Lock()
			tmplTmplSetsMap[tmplID] = tmplSetIDs
			allTmplSetIDs = append(allTmplSetIDs, tmplSetIDs...)
			lock.Unlock()
		}(tmplID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("list multiple template bound template set details failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}
	allTmplSetIDs = tools.RemoveDuplicates(allTmplSetIDs)

	// get template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, allTmplSetIDs)
	if err != nil {
		logs.Errorf("list multiple template bound template set details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSetMap := make(map[uint32]*table.TemplateSet, len(tmplSets))
	for _, t := range tmplSets {
		tmplSetMap[t.ID] = t
	}

	// get template details
	tmpls, err := s.dao.Template().ListByIDs(kt, templateIDs)
	if err != nil {
		logs.Errorf("list multiple template bound template set details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplMap := make(map[uint32]*table.Template, len(tmpls))
	for _, t := range tmpls {
		tmplMap[t.ID] = t
	}

	// combine resp details
	details := make([]*pbtbr.MultiTemplateBoundTemplateSetDetail, 0)
	for tmplID, tmplSetIDs := range tmplTmplSetsMap {
		for _, tmplSetID := range tmplSetIDs {
			details = append(details, &pbtbr.MultiTemplateBoundTemplateSetDetail{
				TemplateId:      tmplID,
				TemplateName:    tmplMap[tmplID].Spec.Name,
				TemplateSetId:   tmplSetID,
				TemplateSetName: tmplSetMap[tmplSetID].Spec.Name,
			})
		}

	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListMultiTmplBoundTmplSetsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListMultiTmplBoundTmplSetsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTmplRevisionBoundUnnamedApps list template revision bound unnamed app details.
func (s *Service) ListTmplRevisionBoundUnnamedApps(ctx context.Context,
	req *pbds.ListTmplRevisionBoundUnnamedAppsReq) (
	*pbds.ListTmplRevisionBoundUnnamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplRevisionExist(kt, req.TemplateRevisionId); err != nil {
		return nil, err
	}

	appIDs, err := s.dao.TemplateBindingRelation().
		ListTmplRevisionBoundUnnamedApps(kt, req.BizId, req.TemplateRevisionId)
	if err != nil {
		logs.Errorf("list template revision bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get app details
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template revision bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// combine resp details
	details := make([]*pbtbr.TemplateRevisionBoundUnnamedAppDetail, 0)
	for _, id := range appIDs {
		// if app doesn't exist, ignore the invalid record
		if _, ok := appMap[id]; !ok {
			continue
		}
		details = append(details, &pbtbr.TemplateRevisionBoundUnnamedAppDetail{
			AppId:   id,
			AppName: appMap[id].Spec.Name,
		})
	}

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TmplRevisionBoundUnnamedApp)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		newDetails := make([]*pbtbr.TemplateRevisionBoundUnnamedAppDetail, 0)
		for _, detail := range details {
			if fieldsMap["app_name"] && strings.Contains(detail.AppName, req.SearchValue) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplRevisionBoundUnnamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplRevisionBoundUnnamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTmplRevisionBoundNamedApps list template revision bound named app details.
// Deprecated: not in use currently
// if use it, consider to add column app_name, release_name on table released_app_templates in case of app is deleted.
//
//nolint:funlen
func (s *Service) ListTmplRevisionBoundNamedApps(ctx context.Context,
	req *pbds.ListTmplRevisionBoundNamedAppsReq) (
	*pbds.ListTmplRevisionBoundNamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplRevisionExist(kt, req.TemplateRevisionId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTmplRevisionBoundNamedApps(kt, req.BizId, req.TemplateRevisionId)
	if err != nil {
		logs.Errorf("list template revision bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get app and release details
	appIDs := make([]uint32, len(relations))
	releaseIDs := make([]uint32, len(relations))
	for i, r := range relations {
		appIDs[i] = r.AppID
		releaseIDs[i] = r.ReleaseID
	}
	appIDs = tools.RemoveDuplicates(appIDs)
	releaseIDs = tools.RemoveDuplicates(releaseIDs)

	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template revision bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}
	releases, err := s.dao.Release().ListAllByIDs(kt, releaseIDs, req.BizId)
	if err != nil {
		logs.Errorf("list template revision bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	releaseMap := make(map[uint32]*table.Release, len(releases))
	for _, r := range releases {
		releaseMap[r.ID] = r
	}

	// combine resp details
	details := make([]*pbtbr.TemplateRevisionBoundNamedAppDetail, 0)
	for _, r := range relations {
		appName := ""
		if _, ok := appMap[r.AppID]; ok {
			appName = appMap[r.AppID].Spec.Name
		}
		details = append(details, &pbtbr.TemplateRevisionBoundNamedAppDetail{
			AppId:       r.AppID,
			AppName:     appName,
			ReleaseId:   r.ReleaseID,
			ReleaseName: releaseMap[r.ReleaseID].Spec.Name,
		})
	}

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TmplRevisionBoundNamedApp)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		newDetails := make([]*pbtbr.TemplateRevisionBoundNamedAppDetail, 0)
		for _, detail := range details {
			if (fieldsMap["app_name"] && strings.Contains(detail.AppName, req.SearchValue)) ||
				(fieldsMap["release_name"] && strings.Contains(detail.ReleaseName, req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplRevisionBoundNamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplRevisionBoundNamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTmplSetBoundUnnamedApps list template set bound unnamed app details.
func (s *Service) ListTmplSetBoundUnnamedApps(ctx context.Context,
	req *pbds.ListTmplSetBoundUnnamedAppsReq) (
	*pbds.ListTmplSetBoundUnnamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplSetExist(kt, req.TemplateSetId); err != nil {
		return nil, err
	}

	appIDs, err := s.dao.TemplateBindingRelation().
		ListTmplSetBoundUnnamedApps(kt, req.BizId, req.TemplateSetId)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get app details
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// combine resp details
	details := make([]*pbtbr.TemplateSetBoundUnnamedAppDetail, 0)
	for _, id := range appIDs {
		// if app doesn't exist, ignore the invalid record
		if _, ok := appMap[id]; !ok {
			continue
		}
		details = append(details, &pbtbr.TemplateSetBoundUnnamedAppDetail{
			AppId:   id,
			AppName: appMap[id].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplSetBoundUnnamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplSetBoundUnnamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListMultiTmplSetBoundUnnamedApps list template set bound unnamed app details.
//
//nolint:funlen
func (s *Service) ListMultiTmplSetBoundUnnamedApps(ctx context.Context,
	req *pbds.ListMultiTmplSetBoundUnnamedAppsReq) (

	*pbds.ListMultiTmplSetBoundUnnamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	templateSetIDs := tools.RemoveDuplicates(req.TemplateSetIds)
	if err := s.dao.Validator().ValidateTmplSetsExist(kt, templateSetIDs); err != nil {
		return nil, err
	}

	allAppIDs := make([]uint32, 0)
	tmplSetAppsMap := make(map[uint32][]uint32)
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	for _, tmplSetID := range templateSetIDs {
		wg.Add(1)

		pipe <- struct{}{}
		go func(tmplSetID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			appIDs, err := s.dao.TemplateBindingRelation().
				ListTmplSetBoundUnnamedApps(kt, req.BizId, tmplSetID)
			if err != nil {
				hitError = err
				return
			}
			lock.Lock()
			tmplSetAppsMap[tmplSetID] = appIDs
			allAppIDs = append(allAppIDs, appIDs...)
			lock.Unlock()
		}(tmplSetID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("list multiple template set bound unnamed app details failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}
	allAppIDs = tools.RemoveDuplicates(allAppIDs)

	// get app details
	apps, err := s.dao.App().ListAppsByIDs(kt, allAppIDs)
	if err != nil {
		logs.Errorf("list multiple template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// get template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, templateSetIDs)
	if err != nil {
		logs.Errorf("list multiple template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSetMap := make(map[uint32]*table.TemplateSet, len(tmplSets))
	for _, t := range tmplSets {
		tmplSetMap[t.ID] = t
	}

	// combine resp details
	details := make([]*pbtbr.MultiTemplateSetBoundUnnamedAppDetail, 0)
	for tmplSetID, appIDs := range tmplSetAppsMap {
		for _, id := range appIDs {
			// if app doesn't exist, ignore the invalid record
			if _, ok := appMap[id]; !ok {
				continue
			}
			details = append(details, &pbtbr.MultiTemplateSetBoundUnnamedAppDetail{
				TemplateSetId:   tmplSetID,
				TemplateSetName: tmplSetMap[tmplSetID].Spec.Name,
				AppId:           id,
				AppName:         appMap[id].Spec.Name,
			})
		}

	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListMultiTmplSetBoundUnnamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListMultiTmplSetBoundUnnamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTmplSetBoundNamedApps list template set bound named app details.
// Deprecated: not in use currently
// if use it, consider to add column app_name, release_name on table released_app_templates in case of app is deleted
//
//nolint:funlen
func (s *Service) ListTmplSetBoundNamedApps(ctx context.Context,
	req *pbds.ListTmplSetBoundNamedAppsReq) (
	*pbds.ListTmplSetBoundNamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplSetExist(kt, req.TemplateSetId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTmplSetBoundNamedApps(kt, req.BizId, req.TemplateSetId)
	if err != nil {
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	uniqueRelations := make([]*types.TmplSetBoundNamedAppDetail, 0)
	uniqueRelationMap := make(map[string]bool)
	// get app and release details
	appIDs := make([]uint32, len(relations))
	releaseIDs := make([]uint32, len(relations))
	for i, r := range relations {
		key := strconv.FormatUint(uint64(r.AppID), 10) + "_" + strconv.FormatUint(uint64(r.ReleaseID), 10)
		if !uniqueRelationMap[key] {
			uniqueRelationMap[key] = true
			uniqueRelations = append(uniqueRelations, r)
		}
		appIDs[i] = r.AppID
		releaseIDs[i] = r.ReleaseID
	}
	appIDs = tools.RemoveDuplicates(appIDs)
	releaseIDs = tools.RemoveDuplicates(releaseIDs)

	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}
	releases, err := s.dao.Release().ListAllByIDs(kt, releaseIDs, req.BizId)
	if err != nil {
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	releaseMap := make(map[uint32]*table.Release, len(releases))
	for _, r := range releases {
		releaseMap[r.ID] = r
	}

	// combine resp details
	details := make([]*pbtbr.TemplateSetBoundNamedAppDetail, 0)
	for _, r := range uniqueRelations {
		appName := ""
		if _, ok := appMap[r.AppID]; ok {
			appName = appMap[r.AppID].Spec.Name
		}
		details = append(details, &pbtbr.TemplateSetBoundNamedAppDetail{
			AppId:       r.AppID,
			AppName:     appName,
			ReleaseId:   r.ReleaseID,
			ReleaseName: releaseMap[r.ReleaseID].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplSetBoundNamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTmplSetBoundNamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListLatestTmplBoundUnnamedApps list the latest template bound unnamed app details.
//
//nolint:funlen
func (s *Service) ListLatestTmplBoundUnnamedApps(ctx context.Context,
	req *pbds.ListLatestTmplBoundUnnamedAppsReq) (
	*pbds.ListLatestTmplBoundUnnamedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	atbs, err := s.dao.TemplateBindingRelation().
		ListLatestTmplBoundUnnamedApps(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list the latest template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(atbs) == 0 {
		return &pbds.ListLatestTmplBoundUnnamedAppsResp{
			Count:   0,
			Details: []*pbtbr.LatestTemplateBoundUnnamedAppDetail{},
		}, nil
	}

	var (
		appIDs, tmplSetIDs []uint32
		appTmplSetMap      = make(map[uint32]uint32)
	)
	for _, atb := range atbs {
		appIDs = append(appIDs, atb.Attachment.AppID)
		hitTmplID := false
		for _, b := range atb.Spec.Bindings {
			for _, r := range b.TemplateRevisions {
				if r.TemplateID == req.TemplateId {
					tmplSetIDs = append(tmplSetIDs, b.TemplateSetID)
					appTmplSetMap[atb.Attachment.AppID] = b.TemplateSetID
					hitTmplID = true
					break
				}
			}
			if hitTmplID {
				break
			}
		}
	}
	tmplSetIDs = tools.RemoveDuplicates(tmplSetIDs)

	// get app details
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list the latest template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// get template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, tmplSetIDs)
	if err != nil {
		logs.Errorf("list the latest template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSetMap := make(map[uint32]*table.TemplateSet, len(tmplSets))
	for _, t := range tmplSets {
		tmplSetMap[t.ID] = t
	}

	// combine resp details
	details := make([]*pbtbr.LatestTemplateBoundUnnamedAppDetail, 0)
	for _, appID := range appIDs {
		// if app doesn't exist, ignore the invalid record
		if _, ok := appMap[appID]; !ok {
			continue
		}
		details = append(details, &pbtbr.LatestTemplateBoundUnnamedAppDetail{
			TemplateSetId:   appTmplSetMap[appID],
			TemplateSetName: tmplSetMap[appTmplSetMap[appID]].Spec.Name,
			AppId:           appID,
			AppName:         appMap[appID].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListLatestTmplBoundUnnamedAppsResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListLatestTmplBoundUnnamedAppsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}
