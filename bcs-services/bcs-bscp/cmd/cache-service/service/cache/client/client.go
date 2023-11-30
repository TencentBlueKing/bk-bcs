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

// Package client NOTES
package client

import (
	"context"
	"time"

	"github.com/bluele/gcache"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/bedis"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/runtime/lock"
)

// Interface defines all the supported operations to get resource cache.
type Interface interface {
	GetAppID(kt *kit.Kit, bizID uint32, appName string) (uint32, error)
	GetAppMeta(kt *kit.Kit, bizID uint32, appID uint32) (string, error)
	GetReleasedCI(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	GetReleasedHook(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	ListAppReleasedGroups(kt *kit.Kit, bizID uint32, appID uint32) (string, error)
	GetCredential(kt *kit.Kit, bizID uint32, credential string) (string, error)
	RefreshAppCache(kt *kit.Kit, bizID uint32, appID uint32) error
	GetReleasedKv(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	GetReleasedKvValue(kt *kit.Kit, bizID, appID, releaseID uint32, key string) (string, error)
}

// New initialize a cache client.
func New(op dao.Set, bds bedis.Client, db pbds.DataClient) (Interface, error) {

	opt := lock.Option{
		QPS:   2000,
		Burst: 500,
	}
	rLock := lock.New(opt)

	lc := gcache.New(int(cc.CacheService().CSLocalCache.ReleasedKvCacheSize)).
		LRU().
		Expiration(time.Duration(cc.CacheService().CSLocalCache.ReleasedKvCacheTTLSec) * time.Second).
		Build()

	return &client{
		op:    op,
		bds:   bds,
		rLock: rLock,
		mc:    initMetric(),
		lc:    lc,
		db:    db,
	}, nil
}

// client do all the read cache related operations.
type client struct {
	op  dao.Set
	db  pbds.DataClient //nolint:unused
	bds bedis.Client
	// rLock is the resource's lock
	rLock lock.Interface
	mc    *metric
	lc    gcache.Cache
}

// RefreshAppCache refresh app related cache
func (c *client) RefreshAppCache(kt *kit.Kit, bizID uint32, appID uint32) error {
	_, err := c.refreshAppMetaCache(kt, bizID, appID)
	if err != nil {
		logs.Errorf("refresh app meta cache failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// refresh app released group related cache, including itself & released ci in it
	groupsJs, err := c.refreshAppReleasedGroupCache(kt, bizID, appID)
	if err != nil {
		logs.Errorf("refresh app released group cache failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	var releaseGroups []*table.ReleasedGroup
	if err = jsoni.Unmarshal([]byte(groupsJs), &releaseGroups); err != nil {
		logs.Errorf("unmarshal groups %s failed, err: %v, rid: %s", groupsJs, err, kt.Rid)
		return err
	}
	kt.Ctx = context.TODO()
	done := make(map[uint32]bool)
	for _, group := range releaseGroups {
		if done[group.ReleaseID] {
			continue
		}
		if _, err = c.refreshReleasedCICache(kt, bizID, group.ReleaseID); err != nil {
			logs.Errorf("refresh released ci cache failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		done[group.ReleaseID] = true
		kt.Ctx = context.TODO()
	}

	return nil
}
