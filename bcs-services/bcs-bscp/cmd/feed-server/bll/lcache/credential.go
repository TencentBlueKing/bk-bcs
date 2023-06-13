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
	pbci "bscp.io/pkg/protocol/core/config-item"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
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
func (s *Credential) CanMatchCI(kt *kit.Kit, bizID uint32, app string, credential string, ci *pbci.ConfigItemSpec) (bool, error) {

	if bizID == 0 {
		return false, fmt.Errorf("invalid biz id")
	}
	if len(app) == 0 {
		return false, fmt.Errorf("invalid app name")
	}
	if len(credential) == 0 {
		return false, fmt.Errorf("invalid credential")
	}
	if ci == nil {
		return false, fmt.Errorf("ci is nil")
	}

	c, hit, err := s.getCredentialFromCache(kt, bizID, credential)
	if err != nil {
		return false, err
	}

	if hit {
		s.mc.hitCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
		for _, s := range c.Scope {
			if tools.MatchAppConfigItem(s, app, ci.Path, ci.Name) {
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

	if err := s.client.SetWithExpire(fmt.Sprintf("%d-%s", bizID, credential), c, time.Second); err != nil {
		logs.Errorf("refresh credential %d-%s cache failed, %s", bizID, credential, err.Error())
		// do not return, ignore th error directly.
	}

	for _, s := range c.Scope {
		if tools.MatchAppConfigItem(s, app, ci.Path, ci.Name) {
			return true, nil
		}
	}

	return false, nil
}

func (s *Credential) getCredentialFromCache(kt *kit.Kit, bizID uint32, credential string) (
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
