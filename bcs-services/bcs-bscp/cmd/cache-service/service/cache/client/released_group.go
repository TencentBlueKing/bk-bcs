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

// ListAppReleasedGroups get app's released groups.
// return with json string: types.ReleasedGroupCache
func (c *client) ListAppReleasedGroups(kt *kit.Kit, bizID uint32, appID uint32) (string, error) {

	cancel := kt.CtxWithTimeoutMS(300)
	defer cancel()

	list, hit, err := c.getReleasedGroupsFromCache(kt, bizID, appID)
	if err != nil {
		return "", err
	}

	if hit {
		c.mc.hitCounter.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Inc()
		return list, nil
	}

	// can not get cache from redis, then try get it from db directly
	// and refresh cache at the same time.
	state := c.rLock.Acquire(keys.ResKind.ReleasedGroup(appID))
	if state.Acquired || (!state.Acquired && state.WithLimit) {

		start := time.Now()
		list, err = c.refreshAppReleasedGroupCache(kt, bizID, appID)
		if err != nil {
			state.Release(true)
			return "", err
		}

		state.Release(false)

		c.mc.refreshLagMS.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))

		return list, nil
	}

	list, hit, err = c.getReleasedGroupsFromCache(kt, bizID, appID)
	if err != nil {
		return "", err
	}

	if !hit {
		return "", errf.ErrCPSInconsistent
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Inc()

	return list, nil
}

func (c *client) getReleasedGroupsFromCache(kt *kit.Kit, bizID uint32, appID uint32) (string, bool, error) {

	val, err := c.bds.Get(kt.Ctx, keys.Key.ReleasedGroup(bizID, appID))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("released groups: %d not found", appID))
	}

	return val, true, nil
}

// refreshAppReleasedGroupCache get the app released groups from db and try to refresh to the cache.
// if not released group found in db, will return ErrGroupNotFound.
func (c *client) refreshAppReleasedGroupCache(kt *kit.Kit, bizID uint32, appID uint32) (string, error) {
	cancel := kt.CtxWithTimeoutMS(200)
	defer cancel()

	list, size, err := c.queryAppReleasedGroups(kt, bizID, appID)
	if err != nil {
		return "", err
	}

	// refresh app released groups cache.
	if err := c.bds.Set(kt.Ctx, keys.Key.ReleasedGroup(bizID, appID), list, keys.Key.ReleasedGroupTtlSec(false)); err != nil {
		return "", fmt.Errorf("set biz: %d, app: %d, released group cache failed, err: %v", bizID, appID, err)
	}

	c.mc.releasedGroupByteSize.With(prm.Labels{"rsc": releasedGroupRes, "biz": tools.Itoa(bizID)}).Observe(float64(size))

	return list, nil
}

// queryAppReleasedGroups query app released group.
// return params:
// 1. app's released group list.
// 2. app's all released group cache size.
func (c *client) queryAppReleasedGroups(kt *kit.Kit, bizID uint32, appID uint32) (string, int, error) {

	opts := &types.ListReleasedGroupsOption{
		BizID: bizID,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "app_id",
					Op:    filter.Equal.Factory(),
					Value: appID,
				},
			},
		},
	}
	groups, err := c.op.ReleasedGroup().List(kt, opts)
	if err != nil {
		return "", 0, err
	}

	b, err := jsoni.Marshal(groups)
	if err != nil {
		logs.Errorf("marshal app: %d, released group list failed, err: %v", appID, err)
		return "", 0, err
	}
	return string(b), len(b), nil
}
