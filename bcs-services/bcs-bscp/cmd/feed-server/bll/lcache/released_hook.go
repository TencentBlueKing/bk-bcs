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

// newReleasedHook create released hook's cache instance.
func newReleasedHook(mc *metric, cs *clientset.ClientSet) *ReleasedHook {
	rh := new(ReleasedHook)
	rh.mc = mc
	opt := cc.FeedServer().FSLocalCache
	client := gcache.New(int(opt.ReleasedHookCacheSize)).
		LRU().
		EvictedFunc(rh.evictRecorder).
		Expiration(time.Duration(opt.ReleasedHookCacheTTLSec) * time.Second).
		Build()

	rh.client = client
	rh.cs = cs
	rh.collectHitRate()

	return rh
}

const (
	// maxReleasedHookSize defines the max size of the released hook's cache.
	//nolint:unused
	maxReleasedHookSize = 1024 * 1024 // 1MB
)

// ReleasedHook is the instance of the released hook cache.
type ReleasedHook struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// Get the released hook's cache.
func (r *ReleasedHook) Get(kt *kit.Kit, bizID uint32, releaseID uint32) (
	*types.ReleasedHookCache, *types.ReleasedHookCache, error) {
	val, err := r.client.GetIFPresent(releaseID)
	if err == nil {
		r.mc.hitCounter.With(prm.Labels{"resource": "released_hook", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		meta, yes := val.(*types.ReleasedHooksCache)
		if !yes {
			return nil, nil, fmt.Errorf("unsupported released hook cache value type: %v", reflect.TypeOf(val).String())
		}
		return meta.PreHook, meta.PostHook, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, release: %d hook cache from local cache failed, err: %v, rid: %s", bizID, releaseID,
			err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.GetReleasedHookReq{
		BizId:     bizID,
		ReleaseId: releaseID,
	}

	resp, err := r.cs.CS().GetReleasedHook(kt.RpcCtx(), opt)
	if err != nil {
		r.mc.errCounter.With(prm.Labels{"resource": "released_hook", "biz": tools.Itoa(bizID)}).Inc()
		return nil, nil, err
	}

	hooks := &types.ReleasedHooksCache{}
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &hooks)
	if err != nil {
		return nil, nil, err
	}

	if len(resp.JsonRaw) <= maxRCISizeKB {
		// only cache the released hook which is less than maxRCISizeKB
		err = r.client.Set(releaseID, hooks)
		if err != nil {
			logs.Errorf("refresh biz: %d, release: %d hook cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
			// do not return, ignore the error directly.
		}
	}

	r.mc.refreshLagMS.With(prm.Labels{"resource": "released_hook", "biz": tools.Itoa(bizID)}).Observe(
		tools.SinceMS(start))

	return hooks.PreHook, hooks.PostHook, nil
}

func (r *ReleasedHook) evictRecorder(key interface{}, _ interface{}) {
	releaseID, yes := key.(uint32)
	if !yes {
		return
	}

	r.mc.evictCounter.With(prm.Labels{"resource": "released_hook"}).Inc()

	if logs.V(2) {
		logs.Infof("evict released hook cache, release: %d", releaseID)
	}
}

func (r *ReleasedHook) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			r.mc.hitRate.With(prm.Labels{"resource": "released_hook"}).Set(r.client.HitRate())
		}
	}()
}
