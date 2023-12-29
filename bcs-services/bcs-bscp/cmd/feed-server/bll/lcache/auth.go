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
	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// newAuth create an auth cache instance.
func newAuth(mc *metric, authorizer auth.Authorizer) *Auth {
	auth := new(Auth)
	auth.mc = mc
	opt := cc.FeedServer().FSLocalCache
	client := gcache.New(int(opt.AuthCacheSize)).
		LRU().
		EvictedFunc(auth.evictRecorder).
		Expiration(time.Duration(opt.AuthCacheTTLSec) * time.Second).
		Build()

	auth.client = client
	auth.authorizer = authorizer
	auth.collectHitRate()

	return auth
}

// Auth is the instance of the auth cache.
type Auth struct {
	mc         *metric
	client     gcache.Cache
	authorizer auth.Authorizer
}

// Authorize if user has permission to the bscp resource.
func (au *Auth) Authorize(kt *kit.Kit, res *meta.ResourceAttribute) (bool, error) {
	key := au.generateBizAuthKey(kt.User, res)

	val, err := au.client.GetIFPresent(key)
	if err == nil {
		au.mc.hitCounter.With(prometheus.Labels{"resource": "auth", "biz": tools.Itoa(res.BizID)}).Inc()

		// hit from cache.
		authorized, yes := val.(bool)
		if !yes {
			return false, fmt.Errorf("unsupported auth cache value type: %v", reflect.TypeOf(val).String())
		}
		return authorized, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, auth key: %s authorized from local cache failed, err: %v, rid: %s", res.BizID, key,
			err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()

	_, authorized, err := au.authorizer.AuthorizeDecision(kt, res)
	if err != nil {
		au.mc.errCounter.With(prometheus.Labels{"resource": "auth", "biz": tools.Itoa(res.BizID)}).Inc()
		return false, err
	}

	if authorized {
		err = au.client.Set(key, authorized)
		if err != nil {
			logs.Errorf("update biz: %d, auth key: %s cache failed, err: %v, rid: %s", res.BizID, key, err, kt.Rid)
			// do not return, ignore the error directly.
		}
	} else {
		// keep no permission result in cache for 30s
		err = au.client.SetWithExpire(key, authorized, 30*time.Second)
		if err != nil {
			logs.Errorf("update biz: %d, auth key: %s cache failed, err: %v, rid: %s", res.BizID, key, err, kt.Rid)
			// do not return, ignore the error directly.
		}
	}

	au.mc.refreshLagMS.With(prometheus.Labels{"resource": "auth", "biz": tools.Itoa(res.BizID)}).
		Observe(tools.SinceMS(start))

	return authorized, nil
}

func (au *Auth) generateBizAuthKey(user string, res *meta.ResourceAttribute) string {
	return fmt.Sprintf("%s-%s-%s-%d-%d", user, res.Type, res.Action, res.ResourceID, res.BizID)
}

func (au *Auth) evictRecorder(key interface{}, _ interface{}) {
	authKey, yes := key.(string)
	if !yes {
		return
	}

	au.mc.evictCounter.With(prometheus.Labels{"resource": "auth"}).Inc()

	if logs.V(2) {
		logs.Infof("evict auth cache, auth key: %s", authKey)
	}
}

func (au *Auth) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			au.mc.hitRate.With(prometheus.Labels{"resource": "auth"}).Set(au.client.HitRate())
		}
	}()
}
