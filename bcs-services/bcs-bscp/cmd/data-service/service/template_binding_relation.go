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
	"sync"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbtbr "bscp.io/pkg/protocol/core/template-binding-relation"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// ListTemplateBoundCounts list template bound counts.
func (s *Service) ListTemplateBoundCounts(ctx context.Context, req *pbds.ListTemplateBoundCountsReq) (
	*pbds.ListTemplateBoundCountsResp, error) {
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

	resp := &pbds.ListTemplateBoundCountsResp{
		Details: details,
	}
	return resp, nil
}

// ListTemplateRevisionBoundCounts list template revision bound counts.
func (s *Service) ListTemplateRevisionBoundCounts(ctx context.Context, req *pbds.ListTemplateRevisionBoundCountsReq) (
	*pbds.ListTemplateRevisionBoundCountsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateRevisionsExist(kt, req.TemplateRevisionIds); err != nil {
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

	resp := &pbds.ListTemplateRevisionBoundCountsResp{
		Details: details,
	}
	return resp, nil
}

// ListTemplateSetBoundCounts list template bound counts.
func (s *Service) ListTemplateSetBoundCounts(ctx context.Context, req *pbds.ListTemplateSetBoundCountsReq) (
	*pbds.ListTemplateSetBoundCountsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateSetsExist(kt, req.TemplateSetIds); err != nil {
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

	resp := &pbds.ListTemplateSetBoundCountsResp{
		Details: details,
	}
	return resp, nil
}

// ListTemplateBoundUnnamedAppDetails list template bound unnamed app details.
func (s *Service) ListTemplateBoundUnnamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateBoundUnnamedAppDetailsReq) (
	*pbds.ListTemplateBoundUnnamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateBoundUnnamedAppDetails(kt, req.BizId, req.TemplateId)
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
			if _, ok := tmplRevisionMap[id]; ok {
				details = append(details, &pbtbr.TemplateBoundUnnamedAppDetail{
					TemplateRevisionId:   id,
					TemplateRevisionName: tmplRevisionMap[id].Spec.RevisionName,
					AppId:                r.AppID,
					AppName:              appMap[r.AppID].Spec.Name,
				})
			}
		}
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplateBoundUnnamedAppDetailsResp{
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

	resp := &pbds.ListTemplateBoundUnnamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTemplateBoundNamedAppDetails list template bound named app details.
func (s *Service) ListTemplateBoundNamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateBoundNamedAppDetailsReq) (
	*pbds.ListTemplateBoundNamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateBoundNamedAppDetails(kt, req.BizId, req.TemplateId)
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
	releases, err := s.dao.Release().ListAllByIDs(kt, appIDs, req.BizId)
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
		for _, id := range r.TemplateRevisionIDs {
			if _, ok := tmplRevisionMap[id]; ok {
				details = append(details, &pbtbr.TemplateBoundNamedAppDetail{
					TemplateRevisionId:   id,
					TemplateRevisionName: tmplRevisionMap[id].Spec.RevisionName,
					AppId:                r.AppID,
					AppName:              appMap[r.AppID].Spec.Name,
					ReleaseId:            r.ReleaseID,
					ReleaseName:          releaseMap[r.ReleaseID].Spec.Name,
				})
			}
		}
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplateBoundNamedAppDetailsResp{
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

	resp := &pbds.ListTemplateBoundNamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTemplateBoundTemplateSetDetails list template bound template set details.
func (s *Service) ListTemplateBoundTemplateSetDetails(ctx context.Context,
	req *pbds.ListTemplateBoundTemplateSetDetailsReq) (
	*pbds.ListTemplateBoundTemplateSetDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}

	tmplSetIDs, err := s.dao.TemplateBindingRelation().
		ListTemplateBoundTemplateSetDetails(kt, req.BizId, req.TemplateId)
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
		return &pbds.ListTemplateBoundTemplateSetDetailsResp{
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

	resp := &pbds.ListTemplateBoundTemplateSetDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTemplateRevisionBoundUnnamedAppDetails list template revision bound unnamed app details.
func (s *Service) ListTemplateRevisionBoundUnnamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateRevisionBoundUnnamedAppDetailsReq) (
	*pbds.ListTemplateRevisionBoundUnnamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateRevisionExist(kt, req.TemplateRevisionId); err != nil {
		return nil, err
	}

	appIDs, err := s.dao.TemplateBindingRelation().
		ListTemplateRevisionBoundUnnamedAppDetails(kt, req.BizId, req.TemplateId)
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
		details = append(details, &pbtbr.TemplateRevisionBoundUnnamedAppDetail{
			AppId:   id,
			AppName: appMap[id].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplateRevisionBoundUnnamedAppDetailsResp{
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

	resp := &pbds.ListTemplateRevisionBoundUnnamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTemplateRevisionBoundNamedAppDetails list template revision bound named app details.
func (s *Service) ListTemplateRevisionBoundNamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateRevisionBoundNamedAppDetailsReq) (
	*pbds.ListTemplateRevisionBoundNamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateRevisionExist(kt, req.TemplateRevisionId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateRevisionBoundNamedAppDetails(kt, req.BizId, req.TemplateId)
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
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template revision bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}
	releases, err := s.dao.Release().ListAllByIDs(kt, appIDs, req.BizId)
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
		details = append(details, &pbtbr.TemplateRevisionBoundNamedAppDetail{
			AppId:       r.AppID,
			AppName:     appMap[r.AppID].Spec.Name,
			ReleaseId:   r.ReleaseID,
			ReleaseName: releaseMap[r.ReleaseID].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplateRevisionBoundNamedAppDetailsResp{
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

	resp := &pbds.ListTemplateRevisionBoundNamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTemplateSetBoundUnnamedAppDetails list template set bound unnamed app details.
func (s *Service) ListTemplateSetBoundUnnamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateSetBoundUnnamedAppDetailsReq) (
	*pbds.ListTemplateSetBoundUnnamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateSetExist(kt, req.TemplateSetId); err != nil {
		return nil, err
	}

	appIDs, err := s.dao.TemplateBindingRelation().
		ListTemplateSetBoundUnnamedAppDetails(kt, req.BizId, req.TemplateSetId)
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
		details = append(details, &pbtbr.TemplateSetBoundUnnamedAppDetail{
			AppId:   id,
			AppName: appMap[id].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplateSetBoundUnnamedAppDetailsResp{
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

	resp := &pbds.ListTemplateSetBoundUnnamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListMultiTemplateSetBoundUnnamedAppDetails list template set bound unnamed app details.
func (s *Service) ListMultiTemplateSetBoundUnnamedAppDetails(ctx context.Context,
	req *pbds.ListMultiTemplateSetBoundUnnamedAppDetailsReq) (
	*pbds.ListMultiTemplateSetBoundUnnamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	templateSetIDs := tools.RemoveDuplicates(req.TemplateSetIds)
	if err := s.dao.Validator().ValidateTemplateSetsExist(kt, templateSetIDs); err != nil {
		return nil, err
	}

	allAppIDs := make([]uint32, 0)
	tmplSetAppsMap := make(map[uint32][]uint32)
	for _, templateSetID := range templateSetIDs {
		appIDs, err := s.dao.TemplateBindingRelation().
			ListTemplateSetBoundUnnamedAppDetails(kt, req.BizId, templateSetID)
		if err != nil {
			logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		tmplSetAppsMap[templateSetID] = appIDs
		allAppIDs = append(allAppIDs, appIDs...)
	}
	allAppIDs = tools.RemoveDuplicates(allAppIDs)

	// get app details
	apps, err := s.dao.App().ListAppsByIDs(kt, allAppIDs)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// get template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, templateSetIDs)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
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
		return &pbds.ListMultiTemplateSetBoundUnnamedAppDetailsResp{
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

	resp := &pbds.ListMultiTemplateSetBoundUnnamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListTemplateSetBoundNamedAppDetails list template set bound named app details.
func (s *Service) ListTemplateSetBoundNamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateSetBoundNamedAppDetailsReq) (
	*pbds.ListTemplateSetBoundNamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateSetExist(kt, req.TemplateSetId); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateSetBoundNamedAppDetails(kt, req.BizId, req.TemplateSetId)
	if err != nil {
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
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
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}
	releases, err := s.dao.Release().ListAllByIDs(kt, appIDs, req.BizId)
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
	for _, r := range relations {
		details = append(details, &pbtbr.TemplateSetBoundNamedAppDetail{
			AppId:       r.AppID,
			AppName:     appMap[r.AppID].Spec.Name,
			ReleaseId:   r.ReleaseID,
			ReleaseName: releaseMap[r.ReleaseID].Spec.Name,
		})
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplateSetBoundNamedAppDetailsResp{
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

	resp := &pbds.ListTemplateSetBoundNamedAppDetailsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}
