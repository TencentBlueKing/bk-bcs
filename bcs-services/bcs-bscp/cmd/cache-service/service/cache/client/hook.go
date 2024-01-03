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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// GetReleasedHook get released hook from cache.
// the returned string is a json array string of []table.ReleasedConfigItem
func (c *client) GetReleasedHook(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error) {
	hook, hit, err := c.getReleasedHookFromCache(kt, bizID, releaseID)
	if err != nil {
		return "", nil
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": releasedHookRes, "biz": tools.Itoa(bizID)}).Inc()
		return hook, err
	}

	// do not find released hook in the cache, then try it get from db directly.
	state := c.rLock.Acquire(keys.ResKind.ReleasedHook(releaseID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {

		start := time.Now()
		hook, err = c.refreshReleasedHookCache(kt, bizID, releaseID)
		if err != nil {
			state.Release(true)
			return "", err
		}
		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": releasedHookRes, "biz": tools.Itoa(bizID)}).
			Observe(tools.SinceMS(start))

		return hook, nil
	}

	// released hook cache has already been refreshed, try get from db directly.
	hook, hit, err = c.getReleasedHookFromCache(kt, bizID, releaseID)
	if err != nil {
		return "", err
	}

	if !hit {
		logs.Errorf("retry to get biz: %d, release: %d hook cache failed, rid: %s", bizID, releaseID, kt.Rid)
		return "", errf.New(errf.RecordNotFound, fmt.Sprintf("release %d hook cache not found", releaseID))
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": releasedHookRes, "biz": tools.Itoa(bizID)}).Inc()

	return hook, nil
}

func (c *client) getReleasedHookFromCache(kt *kit.Kit, bizID uint32, releaseID uint32) (string, bool, error) {
	val, err := c.bds.Get(kt.Ctx, keys.Key.ReleasedHook(bizID, releaseID))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("released: %d hook not found", releaseID))
	}

	return val, true, nil
}

// refreshReleasedHookCache get a release's pre and post hooks and cached them.
func (c *client) refreshReleasedHookCache(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error) {
	cancel := kt.CtxWithTimeoutMS(500)
	defer cancel()

	pre, post, err := c.op.ReleasedHook().GetByReleaseID(kt, bizID, releaseID)
	if err != nil {
		logs.Errorf("get biz: %d release: %d hook from db failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}

	hookKey := keys.Key.ReleasedHook(bizID, releaseID)

	cache := &types.ReleasedHooksCache{
		BizID: bizID,
	}
	if pre != nil {
		cache.PreHook = &types.ReleasedHookCache{
			HookID:         pre.HookID,
			HookRevisionID: pre.HookRevisionID,
			Content:        pre.Content,
			Type:           pre.ScriptType,
		}
	}
	if post != nil {
		cache.PostHook = &types.ReleasedHookCache{
			HookID:         post.HookID,
			HookRevisionID: post.HookRevisionID,
			Content:        post.Content,
			Type:           post.ScriptType,
		}
	}

	js, err := jsoni.Marshal(cache)
	if err != nil {
		return "", err
	}

	err = c.bds.Set(kt.Ctx, hookKey, string(js), keys.Key.ReleasedHookTtlSec(false))
	if err != nil {
		logs.Errorf("refresh biz: %d, release: %d hook cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}

	c.mc.cacheItemByteSize.With(prm.Labels{"rsc": releasedHookRes, "biz": tools.Itoa(bizID)}).Observe(float64(len(js)))

	// return the array string json.
	return string(js), nil
}
