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

package release

import (
	"context"
	"errors"
	"fmt"

	"bscp.io/cmd/feed-server/bll/types"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	ptypes "bscp.io/pkg/types"
)

// GetMatchedRelease get the app instance's matched release id.
func (rs *ReleasedService) GetMatchedRelease(kt *kit.Kit, meta *types.AppInstanceMeta) (uint32, error) {

	ctx, _ := context.WithTimeout(context.TODO(), rs.matchReleaseWaitTime)
	if err := rs.limiter.Wait(ctx); err != nil {
		return 0, err
	}

	am, err := rs.cache.App.GetMeta(kt, meta.BizID, meta.AppID)
	if err != nil {
		return 0, err
	}

	if am.ConfigType != table.File {
		// only support file app
		return 0, errf.New(errf.InvalidParameter, "app's configure type is not file")
	}

	switch am.Mode {
	case table.Namespace:
		if len(meta.Namespace) == 0 {
			return 0, errf.New(errf.InvalidParameter, "app works at namespace mode, but request namespace is empty")
		}
	case table.Normal:
		if len(meta.Namespace) != 0 {
			return 0, errf.New(errf.InvalidParameter, "app works at normal mode, but namespace is set")
		}
	default:
		return 0, errf.Newf(errf.InvalidParameter, "unsupported app mode: %s", am.Mode)
	}

	// check if this app instance has already been configured a special release.
	releaseID, hit, err := rs.getAppInstanceRelease(kt, meta.BizID, meta.AppID, meta.Uid)
	if err != nil {
		return 0, err
	}

	if hit {
		return releaseID, nil
	}

	// this app instance does not be configured with a special release.
	// check its app strategy for now.
	strategyList, err := rs.getStrategy(kt, meta)
	if err != nil {
		return 0, err
	}

	matched, err := rs.matchOneStrategyWithLabels(kt, am.Mode, strategyList, meta)
	if err != nil {
		return 0, err
	}

	if logs.V(2) {
		logs.Infof("biz: %d, app: %d, instance(uid: %s) matched CP strategy: %d, release id: %d, rid: %s", meta.BizID,
			meta.AppID, meta.Uid, matched.StrategyID, matched.ReleaseID, kt.Rid)
	}

	return matched.ReleaseID, nil
}

// getStrategy get current published strategy, if not exist return error.
func (rs *ReleasedService) getStrategy(kt *kit.Kit, meta *types.AppInstanceMeta) ([]*ptypes.PublishedStrategyCache,
	error) {

	req := &pbcs.GetAppCpsIDReq{
		BizId:     meta.BizID,
		AppId:     meta.AppID,
		Namespace: meta.Namespace,
	}
	resp, err := rs.cs.CS().GetAppCpsID(kt.RpcCtx(), req)
	if err != nil {
		return nil, fmt.Errorf("query current published strategy id failed, err: %v", err)
	}

	if len(resp.CpsId) == 0 {
		return nil, errf.New(errf.RecordNotFound, errf.ErrCPSNotFound.Error())
	}

	list, err := rs.cache.Strategy.Get(kt, meta.BizID, meta.AppID, resp.CpsId)
	if err != nil {
		return nil, fmt.Errorf("get current published strategy failed, err: %v", err)
	}

	return list, nil
}

// getAppInstanceRelease get the app's instance releases if the specific instance
// has already been configured with a special release which is may not same with
// its strategy.
// it returns this app instance's release id if it has been configured.
func (rs *ReleasedService) getAppInstanceRelease(kt *kit.Kit, bizID uint32, appID uint32, uid string) (
	uint32, bool, error) {

	req := pbcs.GetAppInstanceReleaseReq{
		BizId: bizID,
		AppId: appID,
		Uid:   uid,
	}
	resp, err := rs.cs.CS().GetAppInstanceRelease(kt.RpcCtx(), &req)
	if err != nil {
		return 0, false, err
	}

	if resp.ReleaseId > 0 {
		// if release id > 0, it means this app instance with uid have the specific release.
		return resp.ReleaseId, true, nil
	}

	return 0, false, nil
}

type matchedMeta struct {
	StrategyID uint32
	ReleaseID  uint32
}

// matchOneStrategyWithLabels match at most only one strategy with app instance labels.
func (rs *ReleasedService) matchOneStrategyWithLabels(
	kt *kit.Kit,
	mode table.AppMode,
	list []*ptypes.PublishedStrategyCache,
	meta *types.AppInstanceMeta) (*matchedMeta, error) {

	switch mode {
	case table.Namespace:
		return rs.matchNamespacedStrategyWithLabels(kt, list, meta)

	case table.Normal:
		return rs.matchNormalStrategyWithLabels(kt, list, meta)

	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported strategy type: "+string(mode))
	}

}

// matchNamespacedStrategyWithLabels match at most only one strategy with app instance labels
// when the strategy works at namespace mode.
func (rs *ReleasedService) matchNamespacedStrategyWithLabels(kt *kit.Kit, list []*ptypes.PublishedStrategyCache,
	meta *types.AppInstanceMeta) (*matchedMeta, error) {

	if len(list) == 0 {
		return nil, errf.New(errf.AppInstanceNotMatchedStrategy, "no strategy is published for this app")
	}

	// at most 2 strategies in the list, one is the namespaced strategies,
	// and the other one is the default strategy if it has.
	if len(list) > 2 {
		logs.Errorf("biz: %d, app: %d, namespaced strategy got > 2 strategies, should be at most 2, rid: %s",
			meta.BizID, meta.AppID, kt.Rid)
		return nil, errf.New(errf.Aborted, "namespaced strategy got > 2 strategies, should be at most 2")
	}

	var defaultStrategy *ptypes.PublishedStrategyCache
	for _, one := range list {

		if one == nil {
			// this is a compatible policy. it should not happen normally.
			logs.Warnf("biz: %d, app: %d strategy got nil strategy, rid: %s", meta.BizID, meta.AppID, kt.Rid)
			continue
		}

		if one.AsDefault {
			// default strategy is matched when no other strategy is matched.
			defaultStrategy = one
			continue
		}

		if one.Namespace != meta.Namespace {
			logs.Errorf("got mismatch strategy's namespace(%s) with app instance's namespace(%s), rid: %s",
				one.Namespace, meta.Namespace, kt.Rid)
			return nil, errf.New(errf.Aborted, "got mismatch strategy's namespace with app's namespace")
		}

		if one.Scope == nil {
			logs.Errorf("queried strategy:(%d) spec or spec.scope is nil, rid: %s", one.StrategyID, kt.Rid)
			return nil, errf.New(errf.Aborted, fmt.Sprintf("queried strategy:(%d) spec or spec.scope is nil",
				one.StrategyID))
		}

		// namespaced strategy's scope selector is matched all, ignore it automatically,
		// but the strategy may have the sub strategy. so test if the label can match the
		// sub strategy or not now.
		subStrategy := one.Scope.SubStrategy
		matched, err := rs.isMatchSubStrategy(subStrategy, meta.Labels)
		if err != nil {
			return nil, errf.New(errf.Aborted, fmt.Sprintf("match strattegy's(%d) sub strategy failed, err: %v",
				one.StrategyID, err))
		}

		if matched {
			return &matchedMeta{
				StrategyID: one.StrategyID,
				// use sub-strategy's release id
				ReleaseID: subStrategy.Spec.ReleaseID,
			}, nil
		}

		// this app instance does not match the sub strategy, then use the main strategy directly.
		// because this is a namespaced strategy, app instance's namespace is same with the strategy's
		// namespace.
		return &matchedMeta{
			StrategyID: one.StrategyID,
			ReleaseID:  one.ReleaseID,
		}, nil
	}

	// this app instance does not have the namespaced strategy, validate
	// whether it has the been configured a default strategy.
	if defaultStrategy == nil {
		return nil, errf.New(errf.AppInstanceNotMatchedStrategy, "no strategy can match this app instance")
	}

	if defaultStrategy.StrategyID <= 0 {
		return nil, errf.New(errf.Aborted, "got invalid default strategy")
	}

	// use default strategy as this app instance's matched strategy.
	return &matchedMeta{
		StrategyID: defaultStrategy.StrategyID,
		ReleaseID:  defaultStrategy.ReleaseID,
	}, nil
}

// isMatchSubStrategy test if a label can match the sub-strategy.
func (rs *ReleasedService) isMatchSubStrategy(subStrategy *table.SubStrategy, labels map[string]string) (bool, error) {
	if subStrategy == nil || (subStrategy != nil && subStrategy.IsEmpty()) {
		// no sub strategy is configured, then this app's instance matched this
		// strategy directly

		return false, nil
	}

	// this strategy has a sub strategy, try match it.
	if subStrategy.Spec == nil ||
		(subStrategy.Spec != nil && subStrategy.Spec.Scope == nil) ||
		(subStrategy.Spec != nil && subStrategy.Spec.Scope.Selector == nil) {
		// this is an invalid sub strategy

		return false, errors.New("sub strategy is invalid")
	}

	matched, err := subStrategy.Spec.Scope.Selector.MatchLabels(labels)
	if err != nil {
		return false, fmt.Errorf("match label with sub-strategy failed, err: %v", err)
	}

	return matched, nil
}

// matchNormalStrategyWithLabels match at most only one strategy with app instance labels
// when the strategy works at normal mode.
func (rs *ReleasedService) matchNormalStrategyWithLabels(kt *kit.Kit, list []*ptypes.PublishedStrategyCache,
	meta *types.AppInstanceMeta) (*matchedMeta, error) {

	var defaultStrategy *ptypes.PublishedStrategyCache
	for _, one := range list {

		if one == nil {
			// this is a compatible policy. it should not happen normally.
			logs.Warnf("biz: %d, app: %d strategy got nil strategy, rid: %s", meta.BizID, meta.AppID, kt.Rid)
			continue
		}

		if one.AsDefault {
			// default strategy is matched when no other strategy is matched.
			defaultStrategy = one
			continue
		}

		mainMatched, err := rs.isMatchMainScope(kt, one, meta.Labels)
		if err != nil {
			return nil, err
		}

		if !mainMatched {
			// main scope selector does not match this app instance, so skip this strategy and try next strategy.
			continue
		}

		// this app instance has already matched the main scope,
		// it's time to test if it can match the sub strategy.
		subStrategy := one.Scope.SubStrategy
		if subStrategy == nil || (subStrategy != nil && subStrategy.IsEmpty()) {
			// main scope selector is matched, but sub-strategy is not configured,
			// then use main scope release as matched release.
			return &matchedMeta{StrategyID: one.StrategyID, ReleaseID: one.ReleaseID}, nil
		}

		// main scope selector has already matched, and sub strategy is configured,
		// it's time to test if this app instance can match the sub strategy or not.
		subMatched, err := rs.isMatchSubStrategy(subStrategy, meta.Labels)
		if err != nil {
			return nil, errf.New(errf.Aborted, fmt.Sprintf("match strattegy(%d) sub strategy failed, err: %v",
				one.StrategyID, err))
		}

		if subMatched {
			// main scope selector has already matched the app instance, but sub strategy matched this
			// app instance now, then use strategy's release id as its release.
			return &matchedMeta{StrategyID: one.StrategyID, ReleaseID: subStrategy.Spec.ReleaseID}, nil
		}

		// sub strategy has not matched this app instance. but the main scope has
		// already matched it. use main scope's release id.
		return &matchedMeta{StrategyID: one.StrategyID, ReleaseID: one.ReleaseID}, nil
	}

	// no other strategy is matched for now.
	if defaultStrategy == nil {
		return nil, errf.New(errf.AppInstanceNotMatchedStrategy, "no strategy can match this app instance")
	}

	if defaultStrategy.StrategyID <= 0 {
		return nil, errf.New(errf.Aborted, "got invalid default strategy with invalid strategy id")
	}

	// use default strategy as this app instance's matched strategy.
	return &matchedMeta{StrategyID: defaultStrategy.StrategyID, ReleaseID: defaultStrategy.ReleaseID}, nil
}

// isMatchMainScope test if a label can match the strategy.
func (rs *ReleasedService) isMatchMainScope(kt *kit.Kit, one *ptypes.PublishedStrategyCache,
	labels map[string]string) (bool, error) {

	if one.Scope == nil {
		logs.Errorf("queried strategy:(%d) spec or spec.scope is nil, rid: %s", one.StrategyID, kt.Rid)
		return false, errf.New(errf.Aborted, fmt.Sprintf("queried strategy:(%d) spec or spec.scope is nil",
			one.StrategyID))
	}

	scope := one.Scope
	if scope.Selector == nil || (scope != nil && scope.Selector.IsEmpty()) {
		// each strategy's scope can not be empty.
		return false, errf.New(errf.Aborted, fmt.Sprintf("queried strategy:(%d) scope.selector is nil",
			one.StrategyID))
	}

	mainMatched, err := scope.Selector.MatchLabels(labels)
	if err != nil {
		return false, errf.New(errf.Aborted, fmt.Sprintf("match strattegy(%d) main scope failed, err: %v",
			one.StrategyID, err))
	}

	return mainMatched, nil
}
