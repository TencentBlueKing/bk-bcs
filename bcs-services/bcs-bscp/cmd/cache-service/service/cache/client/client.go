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

// Package client NOTES
package client

import (
	"context"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/bedis"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/runtime/lock"
	"bscp.io/pkg/types"
)

// Interface defines all the supported operations to get resource cache.
type Interface interface {
	GetAppID(kt *kit.Kit, bizID uint32, appName string) (uint32, error)
	GetAppMeta(kt *kit.Kit, bizID uint32, appID uint32) (string, error)
	GetReleasedCI(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	GetAppReleasedStrategies(kt *kit.Kit, bizID uint32, appID uint32, cpsID []uint32) ([]string, error)
	ListAppReleasedGroups(kt *kit.Kit, bizID uint32, appID uint32) (string, error)
	ListCredentialMatchedCI(kt *kit.Kit, bizID uint32, credential string) (string, error)
	GetCredential(kt *kit.Kit, bizID uint32, credential string) (string, error)
	RefreshAppCache(kt *kit.Kit, bizID uint32, appID uint32) error
}

// New initialize a cache client.
func New(op dao.Set, bds bedis.Client) (Interface, error) {

	opt := lock.Option{
		QPS:   2000,
		Burst: 500,
	}
	rLock := lock.New(opt)

	return &client{
		op:    op,
		bds:   bds,
		rLock: rLock,
		mc:    initMetric(),
	}, nil
}

// client do all the read cache related operations.
type client struct {
	op  dao.Set
	db  pbds.DataClient
	bds bedis.Client
	// rLock is the resource's lock
	rLock lock.Interface
	mc    *metric
}

// RefreshAppCache refresh app related cache
func (c *client) RefreshAppCache(kt *kit.Kit, bizID uint32, appID uint32) error {
	_, err := c.refreshAppMetaCache(kt, bizID, appID)
	if err != nil {
		logs.Errorf("refresh app meta cache failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	kt.Ctx = context.TODO()

	opt := &types.GetAppCpsIDOption{
		BizID: bizID,
		AppID: appID,
	}
	cpsIDs, err := c.op.Publish().GetAppCpsID(kt, opt)
	if err != nil {
		logs.Errorf("query app cps id list failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// refresh app strategy related cache, including itself & released ci in it
	strategies, err := c.refreshAppStrategyCache(kt, bizID, appID, cpsIDs)
	if err != nil && err != errf.ErrCPSInconsistent {
		logs.Errorf("refresh app strategy cache failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	kt.Ctx = context.TODO()

	for _, strategyJs := range strategies {
		strategy := new(types.PublishedStrategyCache)
		if err = jsoni.Unmarshal([]byte(strategyJs), strategy); err != nil {
			logs.Errorf("unmarshal strategy %s failed, err: %v, rid: %s", strategyJs, err, kt.Rid)
			return err
		}

		if _, err = c.refreshReleasedCICache(kt, bizID, strategy.ReleaseID); err != nil {
			logs.Errorf("refresh released ci cache failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		kt.Ctx = context.TODO()
	}

	_, err = c.refreshAppReleasedGroupCache(kt, bizID, appID)
	if err != nil {
		logs.Errorf("refresh app released group cache failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}
