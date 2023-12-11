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

	clientset "bscp.io/cmd/feed-server/bll/client-set"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// newReleasedKv create released kv cache instance.
func newReleasedKv(mc *metric, cs *clientset.ClientSet) *ReleasedKv {
	kv := new(ReleasedKv)
	kv.mc = mc
	opt := cc.FeedServer().FSLocalCache
	client := gcache.New(int(opt.ReleasedKvCacheSize)).
		LRU().
		EvictedFunc(kv.evictRecorder).
		Expiration(time.Duration(opt.ReleasedKvCacheTTLSec) * time.Second).
		Build()

	kv.client = client
	kv.cs = cs
	kv.collectHitRate()

	return kv
}

// ReleasedKv is the instance of the released kv cache.
type ReleasedKv struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// GetKvValue Get the rkv value cache.
func (kv *ReleasedKv) GetKvValue(kt *kit.Kit, bizID, appID, releaseID uint32, key string) (*types.ReleaseKvValueCache,
	error) {
	cacheKey := fmt.Sprintf("%d-%d-%d-%s", bizID, appID, releaseID, key)
	val, err := kv.client.GetIFPresent(cacheKey)
	if err == nil {
		kv.mc.hitCounter.With(prm.Labels{"resource": "released_kv_value", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		meta, yes := val.(*types.ReleaseKvValueCache)
		if !yes {
			return nil, fmt.Errorf("unsupported released KV cache value type: %v", reflect.TypeOf(val).String())
		}
		return meta, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, release: %d KV cache from local cache failed, err: %v, rid: %s", bizID,
			releaseID, err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.GetReleasedKvValueReq{
		BizId:     bizID,
		ReleaseId: releaseID,
		AppId:     appID,
		Key:       key,
	}

	resp, err := kv.cs.CS().GetReleasedKvValue(kt.RpcCtx(), opt)
	if err != nil {
		kv.mc.errCounter.With(prm.Labels{"resource": "released_kv", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	rkv := new(types.ReleaseKvValueCache)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &rkv)
	if err != nil {
		return nil, err
	}

	if err = kv.client.Set(cacheKey, rkv); err != nil {
		logs.Errorf("refresh biz: %d, release: %d KV cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		// do not return, ignore the error directly.
	}

	kv.mc.refreshLagMS.With(prm.Labels{"resource": "released_kv", "biz": tools.Itoa(bizID)}).Observe(
		tools.SinceMS(start))

	return rkv, nil

}

// Get the released kvs cache.
func (kv *ReleasedKv) Get(kt *kit.Kit, bizID uint32, releaseID uint32) ([]*types.ReleaseKvCache, error) {
	cacheKey := fmt.Sprintf("%d-%d", bizID, releaseID)
	val, err := kv.client.GetIFPresent(cacheKey)
	if err == nil {
		kv.mc.hitCounter.With(prm.Labels{"resource": "released_kv", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		meta, yes := val.([]*types.ReleaseKvCache)
		if !yes {
			return nil, fmt.Errorf("unsupported released KV cache value type: %v", reflect.TypeOf(val).String())
		}
		return meta, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, release: %d KV cache from local cache failed, err: %v, rid: %s", bizID, releaseID,
			err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.GetReleasedKvReq{
		BizId:     bizID,
		ReleaseId: releaseID,
	}

	resp, err := kv.cs.CS().GetReleasedKv(kt.RpcCtx(), opt)
	if err != nil {
		kv.mc.errCounter.With(prm.Labels{"resource": "released_kv", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	rkv := make([]*types.ReleaseKvCache, 0)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &rkv)
	if err != nil {
		return nil, err
	}

	maxRKVSizeKB := 10

	if len(resp.JsonRaw) <= maxRKVSizeKB {
		// only cache the released kv which is less than maxRKVSizeKB
		err = kv.client.Set(cacheKey, rkv)
		if err != nil {
			logs.Errorf("refresh biz: %d, release: %d KV cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
			// do not return, ignore the error directly.
		}
	}

	kv.mc.refreshLagMS.With(prm.Labels{"resource": "released_kv", "biz": tools.Itoa(bizID)}).Observe(
		tools.SinceMS(start))

	return rkv, nil
}

func (kv *ReleasedKv) evictRecorder(key interface{}, _ interface{}) {
	releaseID, yes := key.(uint32)
	if !yes {
		return
	}

	kv.mc.evictCounter.With(prm.Labels{"resource": "released_kv"}).Inc()

	if logs.V(2) {
		logs.Infof("evict released KV cache, release: %d", releaseID)
	}
}

func (kv *ReleasedKv) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			kv.mc.hitRate.With(prm.Labels{"resource": "released_kv"}).Set(kv.client.HitRate())
		}
	}()
}
