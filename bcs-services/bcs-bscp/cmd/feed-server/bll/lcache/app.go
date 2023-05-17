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
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/bluele/gcache"
	prm "github.com/prometheus/client_golang/prometheus"

	clientset "bscp.io/cmd/feed-server/bll/client-set"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// newApp create an app meta's cache instance.
func newApp(mc *metric, cs *clientset.ClientSet) *App {
	app := new(App)
	app.mc = mc
	opt := cc.FeedServer().FSLocalCache
	metaClient := gcache.New(int(opt.AppCacheSize)).
		LRU().
		EvictedFunc(app.evictRecorder).
		Expiration(time.Duration(opt.AppCacheTTLSec) * time.Second).
		Build()

	idClient := gcache.New(int(opt.AppCacheSize)).
		LRU().
		EvictedFunc(app.evictRecorder).
		Expiration(time.Duration(opt.AppCacheTTLSec) * time.Second).
		Build()

	app.metaClient = metaClient
	app.idClient = idClient
	app.cs = cs
	app.collectHitRate()

	return app
}

// App is the instance of the app cache.
type App struct {
	mc         *metric
	metaClient gcache.Cache
	idClient   gcache.Cache
	cs         *clientset.ClientSet
}

// IsAppExist validate app if exist.
func (ap *App) IsAppExist(kt *kit.Kit, bizID uint32, appIDs ...uint32) (bool, error) {
	if len(appIDs) == 0 {
		return false, errors.New("appID is required")
	}

	for index := range appIDs {
		_, err := ap.GetMeta(kt, bizID, appIDs[index])
		if err != nil {
			if errf.Error(err).Code == errf.RecordNotFound {
				return false, nil
			}

			return false, err
		}
	}

	return true, nil
}

// GetAppID get app id by app name.
func (ap *App) GetAppID(kt *kit.Kit, bizID uint32, appName string) (uint32, error) {
	key := fmt.Sprintf("%d-%s", bizID, appName)
	val, err := ap.idClient.GetIFPresent(key)
	if err == nil {
		ap.mc.hitCounter.With(prm.Labels{"resource": "app_id", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		appID, yes := val.(uint32)
		if !yes {
			return 0, fmt.Errorf("unsupported app id cache value type: %v", reflect.TypeOf(val).String())
		}
		return appID, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, appName: %s app id from local cache failed, err: %v, rid: %s", bizID, appName,
			err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()
	// get the cache from cache service directly.
	opt := &pbcs.GetAppIDReq{
		BizId:   bizID,
		AppName: appName,
	}

	resp, err := ap.cs.CS().GetAppID(kt.RpcCtx(), opt)
	if err != nil {
		ap.mc.errCounter.With(prm.Labels{"resource": "app_id", "biz": tools.Itoa(bizID)}).Inc()
		return 0, err
	}

	err = ap.idClient.Set(key, resp.AppId)
	if err != nil {
		logs.Errorf("update biz: %d, appName: %s app id cache failed, err: %v, rid: %s", bizID, appName, err, kt.Rid)
		// do not return, ignore the error directly.
	}

	ap.mc.refreshLagMS.With(prm.Labels{"resource": "app_id", "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

	return resp.AppId, nil
}

// GetMeta the app meta cache.
func (ap *App) GetMeta(kt *kit.Kit, bizID uint32, appID uint32) (*types.AppCacheMeta, error) {

	val, err := ap.metaClient.GetIFPresent(appID)
	if err == nil {
		ap.mc.hitCounter.With(prm.Labels{"resource": "app_meta", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		meta, yes := val.(*types.AppCacheMeta)
		if !yes {
			return nil, fmt.Errorf("unsupported app meta cache value type: %v", reflect.TypeOf(val).String())
		}
		return meta, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, app: %d app meta from local cache failed, err: %v, rid: %s", bizID, appID,
			err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()
	// get the cache from cache service directly.
	opt := &pbcs.GetAppMetaReq{
		BizId: bizID,
		AppId: appID,
	}

	resp, err := ap.cs.CS().GetAppMeta(kt.RpcCtx(), opt)
	if err != nil {
		ap.mc.errCounter.With(prm.Labels{"resource": "app_meta", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	meta := new(types.AppCacheMeta)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, meta)
	if err != nil {
		return nil, err
	}

	err = ap.metaClient.Set(appID, meta)
	if err != nil {
		logs.Errorf("update biz: %d, app: %d cache failed, err: %v, rid: %s", bizID, appID, err, kt.Rid)
		// do not return, ignore the error directly.
	}

	ap.mc.refreshLagMS.With(prm.Labels{"resource": "app_meta", "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

	return meta, nil
}

func (ap *App) delete(appID uint32) {
	ap.metaClient.Remove(appID)
}

func (ap *App) evictRecorder(key interface{}, _ interface{}) {
	appID, yes := key.(uint32)
	if !yes {
		return
	}

	ap.mc.evictCounter.With(prm.Labels{"resource": "app_meta"}).Inc()

	if logs.V(2) {
		logs.Infof("evict app meta cache, app: %d", appID)
	}
}

func (ap *App) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			ap.mc.hitRate.With(prm.Labels{"resource": "app_meta"}).Set(ap.metaClient.HitRate())
		}
	}()
}
