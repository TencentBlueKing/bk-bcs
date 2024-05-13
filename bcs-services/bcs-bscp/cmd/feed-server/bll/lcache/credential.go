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

// newCredential credential's local cache instance.
func newCredential(mc *metric, cs *clientset.ClientSet) *Credential {
	stg := new(Credential)
	stg.cs = cs
	opt := cc.FeedServer().FSLocalCache

	stg.client = gcache.New(int(opt.CredentialCacheSize)).
		LRU().
		EvictedFunc(stg.evictRecorder).
		Expiration(time.Duration(opt.CredentialCacheTTLSec) * time.Second).
		Build()
	stg.mc = mc
	stg.collectHitRate()

	return stg
}

// Credential is the instance of the credential local cache.
type Credential struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// CanMatchCI the credential's local cache.
// CanMatchCI 支持文件类型和KV类型
func (s *Credential) CanMatchCI(kt *kit.Kit, bizID uint32, app string, credential string,
	path string, name string) (bool, error) {
	if bizID == 0 {
		return false, fmt.Errorf("invalid biz id")
	}
	if len(app) == 0 {
		return false, fmt.Errorf("invalid app name")
	}
	if len(credential) == 0 {
		return false, fmt.Errorf("invalid credential")
	}

	c, hit, err := s.getCredentialFromCache(kt, bizID, credential)
	if err != nil {
		return false, err
	}

	if hit {
		s.mc.hitCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
		if !c.Enabled {
			return false, nil
		}
		for _, s := range c.Scope {
			ok, _ := tools.MatchAppConfigItem(s, app, path, name)
			if ok {
				return true, nil
			}
		}
		return false, nil
	}

	// get the cache from cache service directly.
	opt := &pbcs.GetCredentialReq{
		BizId:      bizID,
		Credential: credential,
	}
	resp, err := s.cs.CS().GetCredential(kt.RpcCtx(), opt)
	if err != nil {
		s.mc.errCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
		return false, err
	}

	err = jsoni.UnmarshalFromString(resp.JsonRaw, &c)
	if err != nil {
		return false, err
	}

	if err := s.client.SetWithExpire(fmt.Sprintf("%d-%s", bizID, credential), c, 10*time.Second); err != nil {
		logs.Errorf("refresh credential %d-%s cache failed, %s", bizID, credential, err.Error())
		// do not return, ignore th error directly.
	}

	for _, s := range c.Scope {
		if !c.Enabled {
			return false, nil
		}
		ok, _ := tools.MatchAppConfigItem(s, app, path, name)
		if ok {
			return true, nil
		}
	}

	return false, nil
}

// GetCred 获取凭证, 并缓存
func (s *Credential) GetCred(kt *kit.Kit, bizID uint32, credential string) (*types.CredentialCache, error) {
	c, hit, err := s.getCredentialFromCache(kt, bizID, credential)
	if err != nil {
		return nil, err
	}

	if hit {
		s.mc.hitCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
	}

	// get the cache from cache service directly.
	opt := &pbcs.GetCredentialReq{
		BizId:      bizID,
		Credential: credential,
	}
	resp, err := s.cs.CS().GetCredential(kt.RpcCtx(), opt)
	if err != nil {
		s.mc.errCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	err = jsoni.UnmarshalFromString(resp.JsonRaw, &c)
	if err != nil {
		return nil, err
	}

	if err := s.client.SetWithExpire(fmt.Sprintf("%d-%s", bizID, credential), c, 10*time.Second); err != nil {
		logs.Errorf("refresh credential %d-%s cache failed, %s", bizID, credential, err.Error())
	}

	return &c, nil
}

func (s *Credential) getCredentialFromCache(_ *kit.Kit, bizID uint32, credential string) (
	c types.CredentialCache, hit bool, err error) {

	c = types.CredentialCache{}

	key := fmt.Sprintf("%d-%s", bizID, credential)
	val, err := s.client.GetIFPresent(key)
	if err != nil {
		if err != gcache.KeyNotFoundError {
			return c, false, err
		}

		return c, false, nil
	}

	c, yes := val.(types.CredentialCache)
	if !yes {
		return c, false, fmt.Errorf("unsupported credential value type: %v",
			reflect.TypeOf(val).String())
	}

	return c, true, nil
}

func (s *Credential) evictRecorder(key interface{}, _ interface{}) {
	s.mc.evictCounter.With(prm.Labels{"resource": "credential"}).Inc()

	if logs.V(3) {
		logs.Infof("evict credential cache, key: %v", key)
	}
}

func (s *Credential) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			s.mc.hitRate.With(prm.Labels{"resource": "credential"}).Set(s.client.HitRate())
		}
	}()
}
