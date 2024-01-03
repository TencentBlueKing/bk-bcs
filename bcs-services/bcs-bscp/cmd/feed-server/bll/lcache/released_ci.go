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

// newReleasedCI create released config item's cache instance.
func newReleasedCI(mc *metric, cs *clientset.ClientSet) *ReleasedCI {
	ci := new(ReleasedCI)
	ci.mc = mc
	opt := cc.FeedServer().FSLocalCache
	client := gcache.New(int(opt.ReleasedCICacheSize)).
		LRU().
		EvictedFunc(ci.evictRecorder).
		Expiration(time.Duration(opt.ReleasedCICacheTTLSec) * time.Second).
		Build()

	ci.client = client
	ci.cs = cs
	ci.collectHitRate()

	return ci
}

const (
	// maxRCISizeKB defines the max size of the released config item's cache.
	// the minimum size of one CI's json raw size is about 0.5 KB.
	// the maximum size of one CI'S json raw size is about 1.1 KB.
	// one app's maximum number of CI is 50.
	// so the released CI cache should cover the most scenario of the user cases.
	maxRCISizeKB = 30 * 0.5 * 1024 // 15KB
)

// ReleasedCI is the instance of the released ci cache.
type ReleasedCI struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// Get the released config item's cache.
func (ci *ReleasedCI) Get(kt *kit.Kit, bizID uint32, releaseID uint32) ([]*types.ReleaseCICache, error) {
	val, err := ci.client.GetIFPresent(releaseID)
	if err == nil {
		ci.mc.hitCounter.With(prm.Labels{"resource": "released_ci", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		meta, yes := val.([]*types.ReleaseCICache)
		if !yes {
			return nil, fmt.Errorf("unsupported released CI cache value type: %v", reflect.TypeOf(val).String())
		}
		return meta, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, release: %d CI cache from local cache failed, err: %v, rid: %s", bizID, releaseID,
			err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.GetReleasedCIReq{
		BizId:     bizID,
		ReleaseId: releaseID,
	}

	resp, err := ci.cs.CS().GetReleasedCI(kt.RpcCtx(), opt)
	if err != nil {
		ci.mc.errCounter.With(prm.Labels{"resource": "released_ci", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	rci := make([]*types.ReleaseCICache, 0)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &rci)
	if err != nil {
		return nil, err
	}

	if len(resp.JsonRaw) <= maxRCISizeKB {
		// only cache the released ci which is less than maxRCISizeKB
		err = ci.client.Set(releaseID, rci)
		if err != nil {
			logs.Errorf("refresh biz: %d, release: %d CI cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
			// do not return, ignore the error directly.
		}
	}

	ci.mc.refreshLagMS.With(prm.Labels{"resource": "released_ci", "biz": tools.Itoa(bizID)}).Observe(
		tools.SinceMS(start))

	return rci, nil
}

func (ci *ReleasedCI) evictRecorder(key interface{}, _ interface{}) {
	releaseID, yes := key.(uint32)
	if !yes {
		return
	}

	ci.mc.evictCounter.With(prm.Labels{"resource": "released_ci"}).Inc()

	if logs.V(2) {
		logs.Infof("evict released CI cache, release: %d", releaseID)
	}
}

func (ci *ReleasedCI) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			ci.mc.hitRate.With(prm.Labels{"resource": "released_ci"}).Set(ci.client.HitRate())
		}
	}()
}
