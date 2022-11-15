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

package client

import (
	"fmt"
	"time"

	"bscp.io/cmd/cache-service/service/cache/keys"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"

	prm "github.com/prometheus/client_golang/prometheus"
)

// GetAppReleasedStrategies get cpsID's strategy info.
// return with json string: types.PublishedStrategyCache
func (c *client) GetAppReleasedStrategies(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) (
	[]string, error) {

	cancel := kt.CtxWithTimeoutMS(300)
	defer cancel()

	list, hit, err := c.getStrategyFromCache(kt, bizID, appID, cpsID)
	if err != nil {
		return nil, err
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": strategyRes, "biz": tools.Itoa(bizID)}).Inc()
		return list, nil
	}

	// can not get cache from redis, then try get it from db directly
	// and refresh cache at the same time.
	state := c.rLock.Acquire(keys.ResKind.AppStrategy(appID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {

		start := time.Now()
		list, err := c.refreshAppStrategyCache(kt, bizID, appID, cpsID)
		if err != nil {
			state.Release(true)
			return nil, err
		}

		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": strategyRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return list, nil
	}

	list, hit, err = c.getStrategyFromCache(kt, bizID, appID, cpsID)
	if err != nil {
		return nil, err
	}

	if !hit {
		return nil, errf.ErrCPSInconsistent
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": strategyRes, "biz": tools.Itoa(bizID)}).Inc()

	return list, nil
}

func (c *client) getStrategyFromCache(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) ([]string, bool, error) {
	cacheKeys := make([]string, 0)
	for _, cpsID := range cpsID {
		cacheKeys = append(cacheKeys, keys.Key.CPStrategy(bizID, cpsID))
	}

	list, err := c.bds.MGet(kt.Ctx, cacheKeys...)
	if err != nil {
		return nil, false, err
	}

	if len(list) != len(cacheKeys) {
		return nil, false, nil
	}

	for _, val := range list {
		if val == keys.Key.NullValue() {
			return nil, false, errf.ErrCPSInconsistent
		}
	}

	return list, true, nil
}

// refreshAppStrategyCache get the app strategies from db and try to refresh to the cache.if not strategy found in db,
// will return ErrStrategyNotFound.
// Because considering that if the redis is inserted in batches, there may be some failures.
// So here is the plan to query it in batches, then insert the redis in full.
func (c *client) refreshAppStrategyCache(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) ([]string, error) {
	cancel := kt.CtxWithTimeoutMS(200)
	defer cancel()

	list, kv, notFoundCpsID, stgSize, err := c.queryAppStrategy(kt, bizID, appID, cpsID)
	if err != nil {
		return nil, err
	}

	// refresh app strategy cache.
	if err := c.bds.SetWithTxnPipe(kt.Ctx, kv, keys.Key.CPStrategyTtlSec(false)); err != nil {
		return nil, fmt.Errorf("set biz: %d, app: %d, strategies cache failed, err: %v", bizID, appID, err)
	}

	// if cps id's strategy not found in db, set them a null value to avoid cache penetration.
	if len(notFoundCpsID) != 0 {
		notFoundKV := make(map[string]string, 0)
		for _, id := range notFoundCpsID {
			notFoundKV[keys.Key.CPStrategy(bizID, id)] = keys.Key.NullValue()
		}

		if err := c.bds.SetWithTxnPipe(kt.Ctx, notFoundKV, keys.Key.NullKeyTtlSec()); err != nil {
			logs.Errorf("set biz: %d, app: %d the CPS[%v] NULL cache failed, err: %v, rid: %s", bizID, appID,
				notFoundCpsID, err, kt.Rid)
			return nil, err
		}

		return nil, errf.ErrCPSInconsistent
	}

	c.mc.strategyByteSize.With(prm.Labels{"rsc": releasedCIRes, "biz": tools.Itoa(bizID)}).Observe(float64(stgSize))

	return list, nil
}

// queryAppStrategy query app strategy.
// return params:
// 1. cpsID's strategy list.
// 2. app all strategy cache kv from db.
// 3. cpsID that not found strategy in db.
// 4. app all strategy cache size.
func (c *client) queryAppStrategy(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) ([]string,
	map[string]string, []uint32, int, error) {

	var stgSize int
	list := make([]string, 0)
	kv := make(map[string]string)
	opts := &types.GetAppCPSOption{
		BizID: bizID,
		AppID: appID,
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: types.GetCPSMaxPageLimit,
		},
	}

	cpsMap := make(map[uint32]bool, 0)
	for _, id := range cpsID {
		cpsMap[id] = true
	}

	for start := uint32(0); ; start += types.GetCPSMaxPageLimit {
		opts.Page.Start = start
		appStrategies, err := c.op.Publish().GetAppCPStrategies(kt, opts)
		if err != nil {
			logs.Errorf("refresh biz: %d, app: %d all the CPS failed, err: %v, rid: %s", bizID, appID, err, kt.Rid)
			return nil, nil, nil, 0, err
		}

		if len(appStrategies) == 0 {
			break
		}

		mode := appStrategies[0].Mode
		for _, stg := range appStrategies {
			if mode != stg.Mode {
				return nil, nil, nil, 0, fmt.Errorf("biz: %d, app: %d, got multiple mode", bizID, appID)
			}

			js, err := jsoni.Marshal(stg)
			if err != nil {
				return nil, nil, nil, 0, fmt.Errorf("biz: %d, marshal strategy: %d failed, err: %v", bizID,
					stg.StrategyID, err)
			}

			if cpsMap[stg.ID] {
				list = append(list, string(js))
				delete(cpsMap, stg.ID)
			}

			stgSize += len(js)
			kv[keys.Key.CPStrategy(bizID, stg.ID)] = string(js)
		}

		if len(appStrategies) < types.GetCPSMaxPageLimit {
			break
		}
	}

	notFoundCps := make([]uint32, 0)
	for cpsID, _ := range cpsMap {
		notFoundCps = append(notFoundCps, cpsID)
	}

	return list, kv, notFoundCps, stgSize, nil
}
