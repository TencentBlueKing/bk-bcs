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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/keys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbclient "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client"
	pbce "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client-event"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/lock"
)

// Interface defines all the supported operations to get resource cache.
type Interface interface {
	GetAppID(kt *kit.Kit, bizID uint32, appName string, refresh bool) (uint32, error)
	GetAppMeta(kt *kit.Kit, bizID uint32, appID uint32) (string, error)
	GetReleasedCI(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	GetReleasedHook(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	ListAppReleasedGroups(kt *kit.Kit, bizID uint32, appID uint32) (string, error)
	GetCredential(kt *kit.Kit, bizID uint32, credential string) (string, error)
	RefreshAppCache(kt *kit.Kit, bizID uint32, appID uint32) error
	GetReleasedKv(kt *kit.Kit, bizID uint32, releaseID uint32) (string, error)
	GetReleasedKvValue(kt *kit.Kit, bizID, appID, releaseID uint32, key string) (string, error)
	SetClientMetric(kt *kit.Kit, bizID, appID uint32, payload []byte) error
	BatchUpsertClientMetrics(kt *kit.Kit, clientData []*pbclient.Client, clientEventData []*pbce.ClientEvent) error
	BatchUpdateLastConsumedTime(kt *kit.Kit, bizID uint32, appIDs []uint32) error
	GetPublishTime(kt *kit.Kit, publishTime int64) (map[uint32]PublishInfo, error)
	SetPublishTime(kt *kit.Kit, bizID, appID, strategyID uint32, publishTime int64) (int64, error)
}

// New initialize a cache client.
func New(op dao.Set, bds bedis.Client, db pbds.DataClient) (Interface, error) {

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

// SetClientMetric set client metric data
func (c *client) SetClientMetric(kt *kit.Kit, bizID, appID uint32, payload []byte) error {
	if err := c.bds.RPush(kt.Ctx, keys.Key.ClientMetricKey(bizID, appID), payload); err != nil {
		return err
	}
	return nil
}

// BatchUpsertClientMetrics batch upsert client metrics data
func (c *client) BatchUpsertClientMetrics(kt *kit.Kit, clientData []*pbclient.Client,
	clientEventData []*pbce.ClientEvent) error {
	in := &pbds.BatchUpsertClientMetricsReq{
		ClientItems:      clientData,
		ClientEventItems: clientEventData,
	}
	_, err := c.db.BatchUpsertClientMetrics(kt.Ctx, in)
	if err != nil {
		return err
	}
	return nil
}

// BatchUpdateLastConsumedTime 批量更新服务拉取时间
func (c *client) BatchUpdateLastConsumedTime(kit *kit.Kit, bizID uint32, appIDs []uint32) error {

	if _, err := c.db.BatchUpdateLastConsumedTime(kit.Ctx, &pbds.BatchUpdateLastConsumedTimeReq{
		BizId:  bizID,
		AppIds: appIDs,
	}); err != nil {
		return err
	}

	return nil
}
