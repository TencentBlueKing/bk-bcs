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

package lcache

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/bluele/gcache"
	prm "github.com/prometheus/client_golang/prometheus"

	clientset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// newReleasedGroup create released group's local cache instance.
func newReleasedGroup(mc *metric, cs *clientset.ClientSet) *ReleasedGroup {
	stg := new(ReleasedGroup)
	stg.cs = cs
	opt := cc.FeedServer().FSLocalCache

	stg.client = gcache.New(int(opt.ReleasedGroupCacheSize)).
		LRU().
		EvictedFunc(stg.evictRecorder).
		Expiration(time.Duration(opt.ReleasedGroupCacheTTLSec) * time.Second).
		Build()
	stg.mc = mc
	stg.collectHitRate()

	return stg
}

// ReleasedGroup is the instance of the ReleasedGroup local cache.
type ReleasedGroup struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// Get the released group's local cache.
func (s *ReleasedGroup) Get(kt *kit.Kit, bizID uint32, appID uint32) (
	[]*types.ReleasedGroupCache, error) {

	list, hit, err := s.getReleasedGroupFromCache(kt, bizID, appID)
	if err != nil {
		return nil, err
	}

	if hit {
		s.mc.hitCounter.With(prm.Labels{"resource": "released_group", "biz": tools.Itoa(bizID)}).Inc()
		return list, nil
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.ListAppReleasedGroupsReq{
		BizId: bizID,
		AppId: appID,
	}
	resp, err := s.cs.CS().ListAppReleasedGroups(kt.RpcCtx(), opt)
	if err != nil {
		s.mc.errCounter.With(prm.Labels{"resource": "released_group", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	groupList := make([]*types.ReleasedGroupCache, 0)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &groupList)
	if err != nil {
		return nil, fmt.Errorf("unmarshal released group cache failed, err: %v", err)
	}

	// sort the released group by strategy id, the larger the strategy id, the latest the group released.
	sort.Slice(groupList, func(i, j int) bool {
		return groupList[i].StrategyID > groupList[j].StrategyID
	})

	if e := s.client.Set(appID, groupList); e != nil {
		logs.Errorf("refresh biz: %d, app: %d, client released group cache failed, err: %v",
			bizID, appID, e)
	}

	s.mc.refreshLagMS.With(prm.Labels{"resource": "released_group", "biz": tools.Itoa(bizID)}).
		Observe(tools.SinceMS(start))

	return groupList, nil
}

func (s *ReleasedGroup) getReleasedGroupFromCache(_ *kit.Kit, _ uint32, appID uint32) (
	[]*types.ReleasedGroupCache, bool, error) {

	val, err := s.client.GetIFPresent(appID)
	if err != nil {
		if err != gcache.KeyNotFoundError {
			return nil, false, err
		}

		return nil, false, nil
	}

	result, yes := val.([]*types.ReleasedGroupCache)
	if !yes {
		return nil, false, fmt.Errorf("unsupported client released group cache value type: %v",
			reflect.TypeOf(val).String())
	}

	return result, true, nil
}

func (s *ReleasedGroup) evictRecorder(key interface{}, _ interface{}) {
	s.mc.evictCounter.With(prm.Labels{"resource": "released_group"}).Inc()

	if logs.V(3) {
		logs.Infof("evict released group cache, key: %v", key)
	}
}

func (s *ReleasedGroup) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			s.mc.hitRate.With(prm.Labels{"resource": "released_group"}).Set(s.client.HitRate())
		}
	}()
}
