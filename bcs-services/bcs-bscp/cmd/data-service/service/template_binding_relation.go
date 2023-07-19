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

// ListTemplateReleaseBoundCounts list template release bound counts.
func (s *Service) ListTemplateReleaseBoundCounts(ctx context.Context, req *pbds.ListTemplateReleaseBoundCountsReq) (
	*pbds.ListTemplateReleaseBoundCountsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateReleasesExist(kt, req.TemplateReleaseIds); err != nil {
		return nil, err
	}

	var hitError error
	details := make([]*pbtbr.TemplateReleaseBoundCounts, len(req.TemplateReleaseIds))
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, tmplReleaseID := range req.TemplateReleaseIds {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, tmplReleaseID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			var (
				unnamedAppCnt, namedAppCnt uint32
				err                        error
			)

			if unnamedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateReleaseBoundUnnamedAppCount(kt, req.BizId, tmplReleaseID); err != nil {
				hitError = err
				return
			}
			if namedAppCnt, err = s.dao.TemplateBindingRelation().
				GetTemplateReleaseBoundNamedAppCount(kt, req.BizId, tmplReleaseID); err != nil {
				hitError = err
				return
			}

			// save the result with index
			details[idx] = &pbtbr.TemplateReleaseBoundCounts{
				TemplateReleaseId:    tmplReleaseID,
				BoundUnnamedAppCount: unnamedAppCnt,
				BoundNamedAppCount:   namedAppCnt,
			}
		}(idx, tmplReleaseID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("list template release bound counts failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	resp := &pbds.ListTemplateReleaseBoundCountsResp{
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

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateBoundUnnamedAppDetails(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get template release details of the template
	tmplReleases, _, err := s.dao.TemplateRelease().
		List(kt, req.BizId, req.TemplateId, "", &types.BasePage{Start: 0, Limit: types.DefaultMaxPageLimit})
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplReleaseMap := make(map[uint32]*table.TemplateRelease, len(tmplReleases))
	for _, r := range tmplReleases {
		tmplReleaseMap[r.ID] = r
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
		for _, id := range r.TemplateReleaseIDs {
			if _, ok := tmplReleaseMap[id]; ok {
				details = append(details, &pbtbr.TemplateBoundUnnamedAppDetail{
					TemplateReleaseId:   id,
					TemplateReleaseName: tmplReleaseMap[id].Spec.ReleaseName,
					AppId:               r.AppID,
					AppName:             appMap[r.AppID].Spec.Name,
				})
			}
		}
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
		Count:   uint32(len(details)),
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

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateBoundNamedAppDetails(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get template release details of the template
	tmplReleases, _, err := s.dao.TemplateRelease().
		List(kt, req.BizId, req.TemplateId, "", &types.BasePage{Start: 0, Limit: types.DefaultMaxPageLimit})
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplReleaseMap := make(map[uint32]*table.TemplateRelease, len(tmplReleases))
	for _, r := range tmplReleases {
		tmplReleaseMap[r.ID] = r
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
		for _, id := range r.TemplateReleaseIDs {
			if _, ok := tmplReleaseMap[id]; ok {
				details = append(details, &pbtbr.TemplateBoundNamedAppDetail{
					TemplateReleaseId:   id,
					TemplateReleaseName: tmplReleaseMap[id].Spec.ReleaseName,
					AppId:               r.AppID,
					AppName:             appMap[r.AppID].Spec.Name,
					ReleaseId:           r.ReleaseID,
					ReleaseName:         releaseMap[r.ReleaseID].Spec.Name,
				})
			}
		}
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
		Count:   uint32(len(details)),
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

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
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

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTemplateBoundTemplateSetDetailsResp{
		Count:   uint32(len(details)),
		Details: details,
	}
	return resp, nil
}

// ListTemplateReleaseBoundUnnamedAppDetails list template release bound unnamed app details.
func (s *Service) ListTemplateReleaseBoundUnnamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateReleaseBoundUnnamedAppDetailsReq) (
	*pbds.ListTemplateReleaseBoundUnnamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateReleaseExist(kt, req.TemplateReleaseId); err != nil {
		return nil, err
	}

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	appIDs, err := s.dao.TemplateBindingRelation().
		ListTemplateReleaseBoundUnnamedAppDetails(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template release bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get app details
	apps, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		logs.Errorf("list template release bound unnamed app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}

	// combine resp details
	details := make([]*pbtbr.TemplateReleaseBoundUnnamedAppDetail, 0)
	for _, id := range appIDs {
		details = append(details, &pbtbr.TemplateReleaseBoundUnnamedAppDetail{
			AppId:   id,
			AppName: appMap[id].Spec.Name,
		})
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTemplateReleaseBoundUnnamedAppDetailsResp{
		Count:   uint32(len(details)),
		Details: details,
	}
	return resp, nil
}

// ListTemplateReleaseBoundNamedAppDetails list template release bound named app details.
func (s *Service) ListTemplateReleaseBoundNamedAppDetails(ctx context.Context,
	req *pbds.ListTemplateReleaseBoundNamedAppDetailsReq) (
	*pbds.ListTemplateReleaseBoundNamedAppDetailsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateReleaseExist(kt, req.TemplateReleaseId); err != nil {
		return nil, err
	}

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	relations, err := s.dao.TemplateBindingRelation().
		ListTemplateReleaseBoundNamedAppDetails(kt, req.BizId, req.TemplateId)
	if err != nil {
		logs.Errorf("list template release bound named app details failed, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("list template release bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	appMap := make(map[uint32]*table.App, len(apps))
	for _, a := range apps {
		appMap[a.ID] = a
	}
	releases, err := s.dao.Release().ListAllByIDs(kt, appIDs, req.BizId)
	if err != nil {
		logs.Errorf("list template release bound named app details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	releaseMap := make(map[uint32]*table.Release, len(releases))
	for _, r := range releases {
		releaseMap[r.ID] = r
	}

	// combine resp details
	details := make([]*pbtbr.TemplateReleaseBoundNamedAppDetail, 0)
	for _, r := range relations {
		details = append(details, &pbtbr.TemplateReleaseBoundNamedAppDetail{
			AppId:       r.AppID,
			AppName:     appMap[r.AppID].Spec.Name,
			ReleaseId:   r.ReleaseID,
			ReleaseName: releaseMap[r.ReleaseID].Spec.Name,
		})
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTemplateReleaseBoundNamedAppDetailsResp{
		Count:   uint32(len(details)),
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

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
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

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTemplateSetBoundUnnamedAppDetailsResp{
		Count:   uint32(len(details)),
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

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
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

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	resp := &pbds.ListTemplateSetBoundNamedAppDetailsResp{
		Count:   uint32(len(details)),
		Details: details,
	}
	return resp, nil
}
