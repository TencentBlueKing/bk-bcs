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

package client

import (
	"errors"
	"fmt"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"

	"bscp.io/cmd/cache-service/service/cache/keys"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// ListCredentialMatchedCI list all config item ids which can be matched by credential.
// return with json string: []uint32
func (c *client) ListCredentialMatchedCI(kt *kit.Kit, bizID uint32, credential string) (string, error) {
	cancel := kt.CtxWithTimeoutMS(300)
	defer cancel()

	list, hit, err := c.listMatchedCIFromCache(kt, bizID, credential)
	if err != nil {
		return "", err
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Inc()
		return list, nil
	}

	// can not get cache from redis, then try get it from db directly
	// and refresh cache at the same time.
	state := c.rLock.Acquire(keys.ResKind.CredentialMatchedCI(bizID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {

		start := time.Now()
		list, err := c.refreshMatchedCIFromCache(kt, bizID, credential)
		if err != nil {
			state.Release(true)
			return "", err
		}

		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return list, nil
	}

	list, hit, err = c.listMatchedCIFromCache(kt, bizID, credential)
	if err != nil {
		return "", err
	}

	if !hit {
		return "", errf.ErrCPSInconsistent
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Inc()

	return list, nil
}

func (c *client) listMatchedCIFromCache(kt *kit.Kit, bizID uint32, credential string) (string, bool, error) {

	val, err := c.bds.Get(kt.Ctx, keys.Key.CredentialMatchedCI(bizID, credential))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("credential matched ci: %s not found", credential))
	}

	return val, true, nil
}

// refreshMatchedCIFromCache get the credential matched ci ids from db and try to refresh to the cache.
func (c *client) refreshMatchedCIFromCache(kt *kit.Kit, bizID uint32, credential string) (string, error) {
	cancel := kt.CtxWithTimeoutMS(200)
	defer cancel()

	list, size, err := c.queryMatchedCIFromCache(kt, bizID, credential)
	if err != nil {
		return "", err
	}

	// refresh app credential matched ci cache.
	if err := c.bds.Set(kt.Ctx, keys.Key.CredentialMatchedCI(bizID, credential),
		list, keys.Key.CredentialMatchedCITtlSec(false)); err != nil {
		return "", fmt.Errorf("set biz: %d, credential: %s, matched ci cache failed, err: %v", bizID, credential, err)
	}

	c.mc.releasedGroupByteSize.With(prm.Labels{"rsc": credential, "biz": tools.Itoa(bizID)}).Observe(float64(size))

	return list, nil
}

// queryMatchedCIFromCache query credential matched ci ids from cache.
// return params:
// 1. credential matched ci ids list.
// 2. credential matched ci ids cache size.
func (c *client) queryMatchedCIFromCache(kt *kit.Kit, bizID uint32, str string) (string, int, error) {

	credential, err := c.op.Credential().GetByCredentialString(kt, bizID, str)
	if err != nil {
		return "", 0, err
	}
	if errors.Is(err, errf.ErrCredentialInvalid) {
		return "", 0, errf.Newf(errf.InvalidParameter, "invalid credential: %s", str)
	}
	if !credential.Spec.Enable {
		return "", 0, errf.Newf(errf.InvalidParameter, "credential: %s is disabled", str)
	}

	// list credential scopes
	scopes, err := c.op.CredentialScope().Get(kt, credential.ID, bizID)

	// list all apps which can be matched by credential.
	appDetails, err := c.op.App().List(kt, &types.ListAppsOption{
		BizID: bizID,
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		},
		Page: &types.BasePage{},
	})
	if err != nil {
		return "", 0, err
	}

	appIDs := make([]uint32, 0, len(appDetails.Details))
	for _, app := range appDetails.Details {
		for _, scope := range scopes.Details {
			match, err := scope.Spec.CredentialScope.MatchApp(app.Spec.Name)
			if err != nil {
				return "", 0, err
			}
			if match {
				appIDs = append(appIDs, app.ID)
			}
		}
	}
	if len(appIDs) == 0 {
		// return early to avoid querying db with empty appIDs which will cause error.
		return "[]", 2, nil
	}
	
	cis := make([]uint32, 0)
	listReleasedCIopt := &types.ListReleasedCIsOption{
		BizID: bizID,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "app_id",
					Op:    filter.In.Factory(),
					Value: appIDs,
				},
			},
		},
		Page: &types.BasePage{},
	}
	CIDetails, err := c.op.ReleasedCI().List(kt, listReleasedCIopt)
	if err != nil {
		return "", 0, err
	}
	for _, ci := range CIDetails.Details {
		for _, scope := range scopes.Details {
			match, err := scope.Spec.CredentialScope.MatchConfigItem(ci.ConfigItemSpec.Path, ci.ConfigItemSpec.Name)
			if err != nil {
				return "", 0, err
			}
			if match {
				cis = append(cis, ci.ID)
			}
		}
	}

	// query all config item ids which can be matched by credential.

	b, err := jsoni.Marshal(cis)
	if err != nil {
		logs.Errorf("marshal credential: %s, matched released config item ids failed, err: %v", str, err)
		return "", 0, err
	}
	return string(b), len(b), nil
}
