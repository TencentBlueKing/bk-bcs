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

package client

import (
	"fmt"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/keys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// GetReleasedKv get released kv from cache.
func (c *client) GetReleasedKv(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error) {

	kv, hit, err := c.getReleasedKvFromCache(kt, bizID, releaseID)
	if err != nil {
		return "", err
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": releasedKvRes, "biz": tools.Itoa(bizID)}).Inc()
		return kv, nil
	}

	// do not find released kv in the cache, then try it get from db directly.
	state := c.rLock.Acquire(keys.ResKind.ReleasedKV(releaseID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {

		start := time.Now()
		kv, err = c.refreshReleasedKvCache(kt, bizID, releaseID)
		if err != nil {
			state.Release(true)
			return "", err
		}
		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": releasedKvRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return kv, nil
	}

	// released Kv cache has already been refreshed, try get from db directly.
	kv, hit, err = c.getReleasedKvFromCache(kt, bizID, releaseID)
	if err != nil {
		return "", err
	}

	if !hit {
		logs.Errorf("retry to get biz: %d, release: %d Kv cache failed, rid: %s", bizID, releaseID, kt.Rid)
		return "", errf.New(errf.RecordNotFound, fmt.Sprintf("release %d Kv cache not found", releaseID))
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": releasedKvRes, "biz": tools.Itoa(bizID)}).Inc()

	return kv, nil
}

// GetReleasedKvValue get rkv value from cache.
func (c *client) GetReleasedKvValue(kt *kit.Kit, bizID, appID, releaseID uint32, key string) (string, error) {

	start := time.Now()
	r := &pbds.GetReleasedKvReq{
		BizId:     bizID,
		AppId:     appID,
		ReleaseId: releaseID,
		Key:       key,
	}
	rkv, err := c.db.GetReleasedKv(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get biz: %d release: %d Kv from db failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}
	c.mc.refreshLagMS.With(prm.Labels{"rsc": releasedKvValueRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

	js, err := jsoni.Marshal(&types.ReleaseKvValueCache{
		ID:        rkv.Id,
		ReleaseID: rkv.ReleaseId,
		Key:       rkv.Spec.Key,
		Value:     rkv.Spec.Value,
		KvType:    rkv.Spec.KvType,
	})
	if err != nil {
		return "", err
	}

	return string(js), nil
}

func (c *client) getReleasedKvFromCache(kt *kit.Kit, bizID, releaseID uint32) (string, bool, error) {
	val, err := c.bds.Get(kt.Ctx, keys.Key.ReleasedKv(bizID, releaseID))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("released: %d kv not found", releaseID))
	}

	return val, true, nil
}

// refreshReleasedKvCache get a release's all the kv and cached them.
func (c *client) refreshReleasedKvCache(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error) {
	cancel := kt.CtxWithTimeoutMS(500)
	defer cancel()

	releasedKvs, err := c.op.ReleasedKv().ListAllByReleaseIDs(kt, []uint32{releaseID}, bizID)
	if err != nil {
		logs.Errorf("get biz: %d release: %d Kv from db failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}

	rkvKey := keys.Key.ReleasedKv(bizID, releaseID)

	if len(releasedKvs) == 0 {
		logs.Errorf("invalid request, can not find biz: %d, release: %d from db, rid: %s", bizID, releaseID, kt.Rid)

		// set a NULL value to block the illegal request.
		err = c.bds.Set(kt.Ctx, rkvKey, keys.Key.NullValue(), keys.Key.NullKeyTtlSec())
		if err != nil {
			logs.Errorf("set biz: %d, release: %d kv cache to NULL failed, err: %v, rid: %s", bizID, releaseID, err,
				kt.Rid)
		}

		return "", errf.New(errf.RecordNotFound, "release not exist in db")
	}

	js, err := jsoni.Marshal(types.ReleaseKvCaches(releasedKvs))
	if err != nil {
		return "", err
	}

	err = c.bds.Set(kt.Ctx, rkvKey, string(js), keys.Key.ReleasedKvTtlSec(false))
	if err != nil {
		logs.Errorf("refresh biz: %d, release: %d Kv cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}

	c.mc.cacheItemByteSize.With(prm.Labels{"rsc": releasedKvRes, "biz": tools.Itoa(bizID)}).Observe(float64(len(js)))

	// return the array string json.
	return string(js), nil
}
