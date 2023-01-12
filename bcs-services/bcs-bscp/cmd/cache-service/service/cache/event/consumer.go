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

package event

import (
	"encoding/json"
	"fmt"
	"sync"

	"bscp.io/cmd/cache-service/service/cache/keys"
	"bscp.io/pkg/dal/bedis"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/runtime/jsoni"
	"bscp.io/pkg/types"
)

// consumer is used to consume the produced events.
type consumer struct {
	bds bedis.Client
	op  dao.Set
}

// consume the events.
func (c *consumer) consume(kt *kit.Kit, es []*table.Event) (needRetry bool) {
	if len(es) == 0 {
		return false
	}

	insertEvents := make([]*table.Event, 0)
	updateEvents := make([]*table.Event, 0)
	deleteEvents := make([]*table.Event, 0)
	eventList := make([]uint32, 0)
	for _, one := range es {
		eventList = append(eventList, one.ID)

		switch one.Spec.OpType {
		case table.InsertOp:
			insertEvents = append(insertEvents, one)
		case table.UpdateOp:
			updateEvents = append(updateEvents, one)
		case table.DeleteOp:
			deleteEvents = append(deleteEvents, one)
		default:
			logs.Errorf("unsupported event op type: %s, id: %s, rid: %s", one.Spec.OpType, one.ID, kt.Rid)
			continue
		}
	}

	if len(insertEvents) != 0 {
		if err := c.consumeInsertEvent(kt, insertEvents); err != nil {
			logs.Errorf("consume insert event failed, err: %v, rid: %s", err, kt.Rid)
			return true
		}
	}

	if len(updateEvents) != 0 {
		if err := c.consumeUpdateEvent(kt, updateEvents); err != nil {
			logs.Errorf("consume update event failed, err: %v, rid: %s", err, kt.Rid)
			return true
		}
	}

	if len(deleteEvents) != 0 {
		if err := c.consumeDeleteEvent(kt, deleteEvents); err != nil {
			logs.Errorf("consume delete event failed, err: %v, rid: %s", err, kt.Rid)
			return true
		}
	}

	logs.Infof("consume event success, id list: %v, rid: %s", eventList, kt.Rid)

	return false
}

func (c *consumer) consumeInsertEvent(kt *kit.Kit, events []*table.Event) error {
	publishEvent := make([]*table.Event, 0)
	insertAppEvent := make([]*table.Event, 0)

	for _, event := range events {
		switch event.Spec.Resource {
		case table.PublishInstance, table.PublishStrategy:
			publishEvent = append(publishEvent, event)
		case table.Application:
			insertAppEvent = append(insertAppEvent, event)
		default:
			logs.Errorf("unsupported insert event resource: %s, id: %s, rid: %s", event.Spec.Resource, event.ID, kt.Rid)
			continue
		}
	}

	if len(publishEvent) != 0 {
		if err := c.refreshAllCache(kt, publishEvent); err != nil {
			logs.Errorf("refresh publish cache failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(insertAppEvent) != 0 {
		if err := c.refreshAppMetaCache(kt, insertAppEvent); err != nil {
			logs.Errorf("refresh app meta cache failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// refreshAllCache refresh all cache for match config release, include instancePublish、
// strategyPublish、releaseCI、appMeta.
func (c *consumer) refreshAllCache(kt *kit.Kit, events []*table.Event) error {
	appBizID := make(map[uint32]uint32)
	for _, one := range events {
		appBizID[one.Attachment.AppID] = one.Attachment.BizID
	}

	// step1: refresh app meta cache.
	if err := c.refreshAppMetaCache(kt, events); err != nil {
		return err
	}

	// step2: refresh strategy publish cache.
	stgReleaseID, err := c.cacheAppStrategy(kt, appBizID)
	if err != nil {
		return err
	}

	// step3: query all instance release id.
	instReleaseID, err := c.queryInstReleaseID(kt, appBizID)
	if err != nil {
		return err
	}

	// step4: refresh released config item cache.
	releaseID := instReleaseID
	for k, v := range stgReleaseID {
		releaseID[k] = v
	}
	err = c.cacheReleasedCI(kt, releaseID)
	if err != nil {
		return err
	}

	return nil
}

// queryInstReleaseID query all instance publish release id under app. when release matching, although the instance
// publish is through db query, but instance publish's release ci is static data that can be cached and prewarmed
// first.
func (c *consumer) queryInstReleaseID(kt *kit.Kit, appBizID map[uint32]uint32) (map[uint32]uint32, error) {
	releaseBizID := make(map[uint32]uint32, 0)
	for appID, bizID := range appBizID {
		meta, err := c.op.CRInstance().ListAppCRIMeta(kt, bizID, appID)
		if err != nil {
			return nil, err
		}

		for _, one := range meta {
			releaseBizID[one.ReleaseID] = bizID
		}
	}

	return releaseBizID, nil
}

// consumeUpdateEvent consume update event.
func (c *consumer) consumeUpdateEvent(kt *kit.Kit, events []*table.Event) error {
	updateAppEvents := make([]*table.Event, 0)
	for _, event := range events {
		switch event.Spec.Resource {
		case table.Application:
			updateAppEvents = append(updateAppEvents, event)
		default:
			logs.Errorf("unsupported update event resource: %s, id: %s, rid: %s", event.Spec.Resource, event.ID, kt.Rid)
			continue
		}
	}

	if len(updateAppEvents) != 0 {
		if err := c.refreshAppMetaCache(kt, updateAppEvents); err != nil {
			logs.Errorf("refresh app meta cache failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// consumeDeleteEvent delete strategy publish and instance publish cache.
func (c *consumer) consumeDeleteEvent(kt *kit.Kit, events []*table.Event) error {
	delPublishEvents := make([]*table.Event, 0)
	delAppEvents := make([]*table.Event, 0)
	for _, event := range events {
		switch event.Spec.Resource {
		case table.PublishStrategy, table.PublishInstance:
			delPublishEvents = append(delPublishEvents, event)
		case table.Application:
			delAppEvents = append(delAppEvents, event)
		default:
			logs.Errorf("unsupported delete event resource: %s, id: %s, rid: %s", event.Spec.Resource, event.ID, kt.Rid)
			continue
		}
	}

	if len(delPublishEvents) != 0 {
		if err := c.refreshAllCache(kt, delPublishEvents); err != nil {
			return err
		}
	}

	if len(delAppEvents) != 0 {
		if err := c.deleteAppMetaCache(kt, delAppEvents); err != nil {
			return err
		}
	}

	return nil
}

// deleteAppMetaCache delete app meta cache from event.
func (c *consumer) deleteAppMetaCache(kt *kit.Kit, events []*table.Event) error {

	appKeys := make([]string, 0)
	for _, one := range events {
		appKeys = append(appKeys, keys.Key.AppMeta(one.Attachment.BizID, one.Spec.ResourceID))
	}

	if err := c.bds.Delete(kt.Ctx, appKeys...); err != nil {
		logs.Errorf("delete app meta cache failed, keys: %v, err: %v, rid: %s", appKeys, err, kt.Rid)
		return err
	}

	return nil
}

// cacheReleasedCI cache the all publish related release's configure items.
func (c *consumer) cacheReleasedCI(kt *kit.Kit, releaseBizID map[uint32]uint32) error {
	reminder := make(map[uint32][]uint32, 0)
	for rlID, bizID := range releaseBizID {
		// remove useless revision info
		reminder[bizID] = append(reminder[bizID], rlID)
	}

	for bizID, releaseIDs := range reminder {
		releasedCI, err := c.listReleasedCI(kt, bizID, releaseIDs)
		if err != nil {
			logs.Errorf("list released ci failed, bizID: %d, releaseIDs: %v, err: %v, rid: %s", bizID, releaseIDs,
				err, kt.Rid)
			return err
		}

		if len(releasedCI) == 0 {
			logs.Infof("list released ci with bizID: %d, releaseIDs: %v, but got nothing, skip caching, rid: %s",
				bizID, releaseIDs, kt.Rid)
			return nil
		}

		ciList := make(map[string][]*table.ReleasedConfigItem)
		for _, one := range releasedCI {
			// remove useless revision info
			one.Revision = nil
			key := keys.Key.ReleasedCI(one.Attachment.BizID, one.ReleaseID)
			ciList[key] = append(ciList[key], one)
		}

		kv := make(map[string]string)
		var js []byte
		for k, list := range ciList {
			if len(list) == 0 {
				continue
			}
			js, err = json.Marshal(types.ReleaseCICaches(list))
			if err != nil {
				logs.Errorf("marshal ci list failed, skip, list: %+v, err: %v, rid: %s", list, err, kt.Rid)
				continue
			}
			kv[k] = string(js)
		}

		err = c.bds.SetWithTxnPipe(kt.Ctx, kv, keys.Key.ReleasedCITtlSec(false))
		if err != nil {
			logs.Errorf("create released ci cache failed, bizID: %d, releaseIDs: %v,err: %v, rid: %s", bizID,
				releaseIDs, err, kt.Rid)
			return err
		}
	}

	logs.Infof("event cache released ci success detail: biz[release_id]: %v , rid: %s", reminder, kt.Rid)

	return nil
}

// cacheAppStrategy cache all the event strategy's related app's all the strategies.
func (c *consumer) cacheAppStrategy(kt *kit.Kit, appBizID map[uint32]uint32) (map[uint32]uint32, error) {
	pipe := make(chan struct{}, 10)
	releaseBizID := newReleaseBizID()
	wg := sync.WaitGroup{}
	var hitErr error
	// get app's all the published strategies and cache them.
	for appID, bizID := range appBizID {
		pipe <- struct{}{}
		wg.Add(1)

		go func(bizID, appID uint32) {
			defer func() {
				<-pipe
				wg.Done()
			}()

			// in the namespace mode, an app has at most for 200 strategies,
			// so we get strategies with app one by one.
			rlID, err := c.cacheOneAppStrategy(kt, bizID, appID)

			if err != nil {
				hitErr = err
				return
			}

			for releaseID, bID := range rlID {
				releaseBizID.Put(releaseID, bID)
			}
			logs.Infof("event cache biz: %d, app: %d, strategies success, rid: %s", bizID, appID, kt.Rid)

		}(bizID, appID)
	}

	wg.Wait()

	return releaseBizID.GetMap(), hitErr
}

// cacheOneAppStrategy cache one app's all strategy.
// Because considering that if the redis is inserted in batches, there may be some failures.
// So here is the plan to query it in batches, then insert the redis in full.
func (c *consumer) cacheOneAppStrategy(kt *kit.Kit, bizID, appID uint32) (map[uint32]uint32, error) {
	opts := &types.GetAppCPSOption{
		BizID: bizID,
		AppID: appID,
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: types.GetCPSMaxPageLimit,
		},
	}
	releaseBizID := make(map[uint32]uint32, 0)
	kv := make(map[string]string)
	for start := uint32(0); ; start += types.GetCPSMaxPageLimit {
		opts.Page.Start = start
		appStrategies, err := c.op.Publish().GetAppCPStrategies(kt, opts)
		if err != nil {
			logs.Errorf("get biz: %d, app: %d all the CPS failed, err: %v, rid: %s", bizID, appID, err, kt.Rid)
			return nil, err
		}

		if len(appStrategies) == 0 {
			break
		}

		mode := appStrategies[0].Mode
		for _, one := range appStrategies {
			if mode != one.Mode {
				logs.Errorf("biz: %d, app: %d, got multiple mode, rid: %s", bizID, appID, kt.Rid)
				return nil, fmt.Errorf("biz: %d, app: %d, got multiple mode", bizID, appID)
			}

			// record publish strategy's release id, these used to add released config item cache.
			releaseBizID[one.ReleaseID] = bizID
			if one.Scope != nil && one.Scope.SubStrategy != nil && !one.Scope.SubStrategy.IsEmpty() &&
				one.Scope.SubStrategy.Spec.ReleaseID > 0 {
				releaseBizID[one.Scope.SubStrategy.Spec.ReleaseID] = bizID
			}

			js, err := jsoni.Marshal(one)
			if err != nil {
				logs.Errorf("biz: %d, marshal strategy: %d failed, err: %v, rid: %s", bizID, one.StrategyID, err, kt.Rid)
				return nil, fmt.Errorf("biz: %d, mrashal strategy: %d failed, err: %v", bizID, one.StrategyID, err)
			}

			kv[keys.Key.CPStrategy(bizID, one.ID)] = string(js)
		}

		if len(appStrategies) < types.GetCPSMaxPageLimit {
			break
		}
	}

	if len(kv) == 0 {
		return nil, nil
	}

	if err := c.bds.SetWithTxnPipe(kt.Ctx, kv, keys.Key.CPStrategyTtlSec(false)); err != nil {
		logs.Errorf("set biz: %d, app: %d, strategies cache failed, err: %v, rid: %s", bizID, appID, err, kt.Rid)
		return nil, err
	}

	return releaseBizID, nil
}

func (c *consumer) listReleasedCI(kt *kit.Kit, bizID uint32, releaseIDs []uint32) ([]*table.ReleasedConfigItem, error) {
	opts := &types.ListReleasedCIsOption{
		BizID: bizID,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "release_id",
					Op:    filter.In.Factory(),
					Value: releaseIDs,
				},
			},
		},
		// use unlimited page.
		Page: &types.BasePage{Start: 0, Limit: 0},
	}

	resp, err := c.op.ReleasedCI().List(kt, opts)
	if err != nil {
		logs.Errorf("list biz: %d released ci failed, err: %v, rid: %s", bizID, err, kt.Rid)
		return nil, err
	}

	return resp.Details, nil
}

func (c *consumer) refreshAppMetaCache(kt *kit.Kit, events []*table.Event) error {
	if len(events) == 0 {
		return nil
	}

	bizApps := make(map[uint32][]uint32, 0)
	for _, event := range events {
		list, exist := bizApps[event.Attachment.BizID]
		if !exist {
			list = make([]uint32, 0)
		}

		list = append(list, event.Spec.ResourceID)
		bizApps[event.Attachment.BizID] = list
	}

	var hitErr error
	for bizID, appIDs := range bizApps {
		if err := c.refreshOneBizAppMetaCache(kt, bizID, appIDs); err != nil {
			logs.Errorf("refresh one biz app meta cache failed, err: %v, rid: %s", err, kt.Rid)
			hitErr = err
			continue
		}
	}

	if hitErr != nil {
		return hitErr
	}

	return nil
}

func (c *consumer) refreshOneBizAppMetaCache(kt *kit.Kit, bizID uint32, appIDs []uint32) error {
	metaMap, err := c.op.App().ListAppMetaForCache(kt, bizID, appIDs)
	if err != nil {
		return err
	}

	for appID, appMeta := range metaMap {
		js, err := jsoni.Marshal(appMeta)
		if err != nil {
			return err
		}

		// update the app meta to cache.
		err = c.bds.Set(kt.Ctx, keys.Key.AppMeta(bizID, appID), string(js), keys.Key.AppMetaTtlSec(false))
		if err != nil {
			logs.Errorf("set app: %d cache failed, err: %v, rid: %s", appID, err, kt.Rid)
			return err
		}

		logs.V(1).Infof("refresh app: %d app meta: %s successfully, rid: %s", appID, js, kt.Rid)
	}

	return nil
}

type releaseBizID struct {
	lock *sync.RWMutex
	data map[ /*releaseID*/ uint32] /*bizID*/ uint32
}

func newReleaseBizID() *releaseBizID {
	return &releaseBizID{
		lock: new(sync.RWMutex),
		data: make(map[uint32]uint32, 0),
	}
}

// Put new release data to releaseBizID with write lock.
func (r *releaseBizID) Put(key uint32, value uint32) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.data[key] = value
}

// GetMap get releaseBizID's data with write lock.
func (r *releaseBizID) GetMap() map[uint32]uint32 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.data
}
