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
)

// GetAppID get app's id by app name.
func (c *client) GetAppID(kt *kit.Kit, bizID uint32, appName string) (uint32, error) {
	// try read from cache at first.
	appID, hit, err := c.getAppIDFromCache(kt, bizID, appName)
	if err != nil {
		return 0, err
	}
	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": aiRes, "biz": tools.Itoa(bizID)}).Inc()
		return appID, nil
	}

	// do not find app in the cache, then try get from db directly.
	state := c.rLock.Acquire(keys.ResKind.AppID(bizID, appName))
	if state.Acquired || (!state.Acquired && state.WithLimit) {
		start := time.Now()
		appID, err = c.refreshAppIDCache(kt, bizID, appName)
		if err != nil {
			state.Release(true)
			return 0, err
		}

		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": aiRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return appID, nil
	}

	// do not acquire the lock, but the cache have already been refreshed,
	// so retry to get app id from cache again.
	appID, hit, err = c.getAppIDFromCache(kt, bizID, appName)
	if err != nil {
		return 0, err
	}

	if !hit {
		return 0, errf.New(errf.RecordNotFound, fmt.Sprintf("app %d-%s cache not found", bizID, appName))
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": aiRes, "biz": tools.Itoa(bizID)}).Inc()

	return appID, nil
}

// GetAppMeta get app's basic meta info.
// return with json string: AppCacheMeta
func (c *client) GetAppMeta(kt *kit.Kit, bizID uint32, appID uint32) (string, error) {

	// try read from cache at first.
	meta, hit, err := c.getAppMetaFromCache(kt, bizID, appID)
	if err != nil {
		return "", err
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": amRes, "biz": tools.Itoa(bizID)}).Inc()
		return meta, nil
	}

	// do not find app in the cache, then try get from db directly.
	state := c.rLock.Acquire(keys.ResKind.AppMeta(appID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {
		var appMeta string
		start := time.Now()
		appMeta, err = c.refreshAppMetaCache(kt, bizID, appID)
		if err != nil {
			state.Release(true)
			return "", err
		}

		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": amRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return appMeta, nil
	}

	// do not acquire the lock, but the cache have already been refreshed,
	// so retry to get app meta from cache again.
	meta, hit, err = c.getAppMetaFromCache(kt, bizID, appID)
	if err != nil {
		return "", err
	}

	if !hit {
		return "", errf.New(errf.RecordNotFound, fmt.Sprintf("app %d cache not found", appID))
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": amRes, "biz": tools.Itoa(bizID)}).Inc()

	return meta, nil
}

// getAppIDFromCache get app id from cache
func (c *client) getAppIDFromCache(kt *kit.Kit, bizID uint32, appName string) (uint32, bool, error) {

	val, err := c.bds.Get(kt.Ctx, keys.Key.AppID(bizID, appName))
	if err != nil {
		return 0, false, err
	}

	if len(val) == 0 {
		return 0, false, nil
	}

	if val == keys.Key.NullValue() {
		return 0, false, errf.New(errf.RecordNotFound, fmt.Sprintf("app %d-%s not found", bizID, appName))
	}

	var id uint32
	if e := jsoni.UnmarshalFromString(val, &id); e != nil {
		return 0, false, e
	}

	return id, true, nil
}

// getAppMetaFromCache get app meta from cache, and return the meta's json string
func (c *client) getAppMetaFromCache(kt *kit.Kit, bizID uint32, appID uint32) (string, bool, error) {

	val, err := c.bds.Get(kt.Ctx, keys.Key.AppMeta(bizID, appID))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("app %d not found", appID))
	}

	return val, true, nil
}

// refreshAppIDCache get the app id from db and try to refresh to the cache.
func (c *client) refreshAppIDCache(kt *kit.Kit, bizID uint32, appName string) (uint32, error) {

	cancel := kt.CtxWithTimeoutMS(200)
	defer cancel()

	app, err := c.op.App().GetByName(kt, bizID, appName)
	if err != nil {
		return 0, err
	}

	b, err := jsoni.Marshal(app.ID)
	if err != nil {
		return 0, err
	}

	// update the app meta to cache.
	err = c.bds.Set(kt.Ctx, keys.Key.AppID(bizID, appName), string(b), keys.Key.AppMetaTtlSec(false))
	if err != nil {
		logs.Errorf("set app: %d-%s cache failed, err: %v, rid: %s", bizID, appName, err, kt.Rid)
		return 0, err
	}

	c.mc.cacheItemByteSize.With(prm.Labels{"rsc": aiRes, "biz": tools.Itoa(bizID)}).Observe(float64(len(b)))

	return app.ID, nil
}

// refreshAppMetaCache get the app meta's from db and try to refresh to the cache.
func (c *client) refreshAppMetaCache(kt *kit.Kit, bizID uint32, appID uint32) (string, error) {

	cancel := kt.CtxWithTimeoutMS(200)
	defer cancel()

	metaMap, err := c.op.App().ListAppMetaForCache(kt, bizID, []uint32{appID})
	if err != nil {
		return "", err
	}

	appMeta, exist := metaMap[appID]
	if !exist {
		// this app is not exist in the db, which means this is an illegal request,
		err = c.bds.Set(kt.Ctx, keys.Key.AppMeta(bizID, appID), keys.Key.NullValue(), keys.Key.NullKeyTtlSec())
		if err != nil {
			logs.Errorf("set app: %d meta to NULL value failed, err: %v, rid: %s", appID, err, kt.Rid)
		}

		return "", errf.New(errf.RecordNotFound, fmt.Sprintf("app %d not exist", appID))
	}

	js, err := jsoni.Marshal(appMeta)
	if err != nil {
		return "", err
	}

	// update the app meta to cache.
	err = c.bds.Set(kt.Ctx, keys.Key.AppMeta(bizID, appID), string(js), keys.Key.AppMetaTtlSec(false))
	if err != nil {
		logs.Errorf("set app: %d cache failed, err: %v, rid: %s", appID, err, kt.Rid)
		return "", err
	}

	c.mc.cacheItemByteSize.With(prm.Labels{"rsc": amRes, "biz": tools.Itoa(bizID)}).Observe(float64(len(js)))

	return string(js), nil
}
