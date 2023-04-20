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

package lcache

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/bluele/gcache"
	prm "github.com/prometheus/client_golang/prometheus"

	clientset "bscp.io/cmd/feed-server/bll/client-set"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// newStrategy create strategy's local cache instance.
func newStrategy(mc *metric, cs *clientset.ClientSet) *Strategy {
	stg := new(Strategy)
	stg.cs = cs
	opt := cc.FeedServer().FSLocalCache

	stg.client = gcache.New(int(opt.PublishedStrategyCacheSize)).
		LRU().
		EvictedFunc(stg.evictRecorder).
		Expiration(time.Duration(opt.PublishedStrategyCacheTTLSec) * time.Second).
		Build()
	stg.mc = mc
	stg.collectHitRate()

	return stg
}

// Strategy is the instance of the strategy local cache.
type Strategy struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// Get the strategy's local cache.
func (s *Strategy) Get(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) (
	[]*types.PublishedStrategyCache, error) {

	list, hit, err := s.getStrategyFromCache(kt, bizID, appID, cpsID)
	if err != nil {
		return nil, err
	}

	if hit {
		s.mc.hitCounter.With(prm.Labels{"resource": "strategy", "biz": tools.Itoa(bizID)}).Inc()
		return list, nil
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.GetAppReleasedStrategyReq{
		BizId: bizID,
		AppId: appID,
		CpsId: cpsID,
	}
	resp, err := s.cs.CS().GetAppReleasedStrategy(kt.RpcCtx(), opt)
	if err != nil {
		s.mc.errCounter.With(prm.Labels{"resource": "strategy", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	sList := make([]*types.PublishedStrategyCache, len(resp.JsonRaw))
	for idx := range resp.JsonRaw {
		psc := new(types.PublishedStrategyCache)
		err = jsoni.UnmarshalFromString(resp.JsonRaw[idx], psc)
		if err != nil {
			return nil, fmt.Errorf("unmarshal released strategy cache failed, err: %v", err)
		}

		sList[idx] = psc
	}

	// 1. Sort the published strategies with strategy id, which is unique, so that
	// these strategies can be matched with a stable order.
	// 2. If the strategy have configured with multiple strategies, but an instance
	// can be matched with multiple strategies, this sort can avoid an app's
	// instance matched different strategy with different request, which may cause
	// this app instance change configures frequently. Normally, this scenario should
	// not happen, but we use this to reduce the influence.
	sort.Slice(sList, func(i, j int) bool {
		return sList[i].StrategyID > sList[j].StrategyID
	})

	for _, sty := range sList {
		if err := s.client.Set(sty.ID, sty); err != nil {
			logs.Errorf("refresh biz: %d, app: %d, cpsID: %d client strategy cache failed, err: %v",
				bizID, appID, sty.ID, err)
		}
	}

	s.mc.refreshLagMS.With(prm.Labels{"resource": "strategy", "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

	return sList, nil
}

func (s *Strategy) getStrategyFromCache(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) (
	[]*types.PublishedStrategyCache, bool, error) {

	result := make([]*types.PublishedStrategyCache, 0)
	for _, id := range cpsID {
		val, err := s.client.GetIFPresent(id)
		if err != nil {
			if err != gcache.KeyNotFoundError {
				return nil, false, err
			}

			return nil, false, nil
		}

		strategy, yes := val.(*types.PublishedStrategyCache)
		if !yes {
			return nil, false, fmt.Errorf("unsupported client strategy cache value type: %v",
				reflect.TypeOf(val).String())
		}

		result = append(result, strategy)
	}

	if len(result) != len(cpsID) {
		return nil, false, nil
	}

	return result, true, nil
}

func (s *Strategy) evictRecorder(key interface{}, _ interface{}) {
	s.mc.evictCounter.With(prm.Labels{"resource": "strategy"}).Inc()

	if logs.V(3) {
		logs.Infof("evict strategy cache, key: %v", key)
	}
}

func (s *Strategy) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			s.mc.hitRate.With(prm.Labels{"resource": "strategy"}).Set(s.client.HitRate())
		}
	}()
}
