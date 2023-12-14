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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/keys"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// GetReleasedCI get released configure items from cache.
// the returned string is a json array string of []table.ReleasedConfigItem
func (c *client) GetReleasedCI(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error) {
	ci, hit, err := c.getReleasedCIFromCache(kt, bizID, releaseID)
	if err != nil {
		return "", nil
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": releasedCIRes, "biz": tools.Itoa(bizID)}).Inc()
		return ci, err
	}

	// do not find released CI in the cache, then try it get from db directly.
	state := c.rLock.Acquire(keys.ResKind.ReleasedCI(releaseID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {

		start := time.Now()
		ci, err = c.refreshReleasedCICache(kt, bizID, releaseID)
		if err != nil {
			state.Release(true)
			return "", err
		}
		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": releasedCIRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return ci, nil
	}

	// released CI cache has already been refreshed, try get from db directly.
	ci, hit, err = c.getReleasedCIFromCache(kt, bizID, releaseID)
	if err != nil {
		return "", err
	}

	if !hit {
		logs.Errorf("retry to get biz: %d, release: %d CI cache failed, rid: %s", bizID, releaseID, kt.Rid)
		return "", errf.New(errf.RecordNotFound, fmt.Sprintf("release %d CI cache not found", releaseID))
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": releasedCIRes, "biz": tools.Itoa(bizID)}).Inc()

	return ci, nil
}

func (c *client) getReleasedCIFromCache(kt *kit.Kit, bizID uint32, releaseID uint32) (string, bool, error) {
	val, err := c.bds.Get(kt.Ctx, keys.Key.ReleasedCI(bizID, releaseID))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("released: %d ci not found", releaseID))
	}

	return val, true, nil
}

// refreshReleasedCICache get a release's all the config items and cached them.
func (c *client) refreshReleasedCICache(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error) {
	cancel := kt.CtxWithTimeoutMS(500)
	defer cancel()

	releasedCIs, err := c.op.ReleasedCI().ListAllByReleaseIDs(kt, []uint32{releaseID}, bizID)
	if err != nil {
		logs.Errorf("get biz: %d release: %d CI from db failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}

	ciKey := keys.Key.ReleasedCI(bizID, releaseID)

	if len(releasedCIs) == 0 {
		logs.Errorf("invalid request, can not find biz: %d, release: %d from db, rid: %s", bizID, releaseID, kt.Rid)

		// set a NULL value to block the illegal request.
		err = c.bds.Set(kt.Ctx, ciKey, keys.Key.NullValue(), keys.Key.NullKeyTtlSec())
		if err != nil {
			logs.Errorf("set biz: %d, release: %d CI cache to NULL failed, err: %v, rid: %s", bizID, releaseID, err,
				kt.Rid)
		}

		return "", errf.New(errf.RecordNotFound, "release not exist in db")
	}

	js, err := jsoni.Marshal(types.ReleaseCICaches(releasedCIs))
	if err != nil {
		return "", err
	}

	err = c.bds.Set(kt.Ctx, ciKey, string(js), keys.Key.ReleasedCITtlSec(false))
	if err != nil {
		logs.Errorf("refresh biz: %d, release: %d CI cache failed, err: %v, rid: %s", bizID, releaseID, err, kt.Rid)
		return "", err
	}

	c.mc.cacheItemByteSize.With(prm.Labels{"rsc": releasedCIRes, "biz": tools.Itoa(bizID)}).Observe(float64(len(js)))

	// return the array string json.
	return string(js), nil
}
