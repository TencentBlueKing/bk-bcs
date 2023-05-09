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
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
)

// newCredential credential's local cache instance.
func newCredential(mc *metric, cs *clientset.ClientSet) *Credential {
	stg := new(Credential)
	stg.cs = cs
	opt := cc.FeedServer().FSLocalCache

	stg.client = gcache.New(int(opt.AuthCacheSize)).
		LRU().
		EvictedFunc(stg.evictRecorder).
		Expiration(time.Duration(opt.AuthCacheTTLSec) * time.Second).
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
func (s *Credential) CanMatchCI(kt *kit.Kit, bizID uint32, credential string, ciID uint32) (bool, error) {

	if bizID == 0 || ciID == 0 {
		return false, fmt.Errorf("invalid biz id or config item id")
	}

	can, hit, err := s.canMatchCIFromCache(kt, credential, ciID)
	if err != nil {
		return false, err
	}

	if hit {
		s.mc.hitCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
		return can, err
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.ListCredentialMatchedCIReq{
		BizId:      bizID,
		Credential: credential,
	}
	resp, err := s.cs.CS().ListCredentialMatchedCI(kt.RpcCtx(), opt)
	if err != nil {
		s.mc.errCounter.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Inc()
		return false, err
	}

	rci := make([]uint32, 0)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &rci)
	if err != nil {
		return false, err
	}

	var match bool
	for _, id := range rci {
		if id == ciID {
			match = true
		}
		key := fmt.Sprintf("%s-%d", credential, id)
		if err := s.client.Set(key, true); err != nil {
			logs.Errorf("refresh biz: %d, credential: %s can matched CI failed, err: %v, rid: %s",
				bizID, credential, err, kt.Rid)
			// do not return, ignore the error directly.
		}
	}
	if !match {
		key := fmt.Sprintf("%s-%d", credential, ciID)
		if err := s.client.Set(key, false); err != nil {
			logs.Errorf("refresh biz: %d, credential: %s can matched CI failed, err: %v, rid: %s",
				bizID, credential, err, kt.Rid)
		}
	}

	s.mc.refreshLagMS.With(prm.Labels{"resource": "credential", "biz": tools.Itoa(bizID)}).Observe(
		tools.SinceMS(start))

	return match, nil
}

func (s *Credential) canMatchCIFromCache(kt *kit.Kit, credential string, ciID uint32) (
	can, hit bool, err error) {

	key := fmt.Sprintf("%s-%d", credential, ciID)
	val, err := s.client.GetIFPresent(key)
	if err != nil {
		if err != gcache.KeyNotFoundError {
			return false, false, err
		}

		return false, false, nil
	}

	can, yes := val.(bool)
	if !yes {
		return false, false, fmt.Errorf("unsupported client can match ci value type: %v",
			reflect.TypeOf(val).String())
	}

	return can, true, nil
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
