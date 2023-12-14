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

// Package eventc eventc ...
package eventc

import (
	"context"
	"errors"

	btyp "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

func newAppEvent(bizID, appID uint32, sch *Scheduler) *appEvent {
	ctx, cancel := context.WithCancel(context.Background())
	ae := &appEvent{
		cancel: cancel,
		bizID:  bizID,
		appID:  appID,
		cursor: new(cursor),
		csm:    initConsumer(sch.mc),
		eQueue: initEventQueue(),
		sch:    sch,
		mc:     sch.mc,
	}

	// Note: remove this go-routine, mounts of app may cost unacceptable size of memory.
	go ae.watchEvents(ctx)
	return ae
}

type appEvent struct {
	appID  uint32
	bizID  uint32
	cancel context.CancelFunc
	eQueue *eventQueue
	cursor *cursor
	csm    *consumer
	sch    *Scheduler
	mc     *metric
}

// AddSidecar add a sidecar instance to the subscriber.
func (ae *appEvent) AddSidecar(currentRelease uint32, sn uint64, subSpec *SubscribeSpec) (hitErr error) {

	// add this sidecar to the consumer list at first, in case the event handling is working
	me := ae.csm.Add(sn, subSpec)
	defer func() {
		if hitErr != nil {
			// if hit an error, then this sidecar should be removed form the consumer list.
			ae.csm.Delete(sn)
		}
	}()

	kt := kit.New()
	matchedRelease, matchedCursor, err := ae.doFirstMatch(kt, subSpec)
	if err != nil {
		return err
	}

	if matchedRelease != currentRelease {
		// release has already changed, notify immediately.
		ae.sch.notifyEvent(kt, matchedCursor, []*member{me})
	}

	return nil
}

// RemoveSidecar remove one sidecar from the consumer list.
// it returns true if all the app's sidecar instances is empty.
func (ae *appEvent) RemoveSidecar(sn uint64) bool {
	return ae.csm.Delete(sn)
}

// Stop this app event handler.
func (ae *appEvent) Stop() {
	ae.cancel()
}

// doFirstMatch do the first release match when the sidecar is added to this app at first.
func (ae *appEvent) doFirstMatch(kt *kit.Kit, subSpec *SubscribeSpec) (uint32, uint32, error) { //nolint:unparam

	cursor := ae.cursor.ID()

	meta := &btyp.AppInstanceMeta{
		BizID:  subSpec.InstSpec.BizID,
		AppID:  subSpec.InstSpec.AppID,
		App:    subSpec.InstSpec.App,
		Uid:    subSpec.InstSpec.Uid,
		Labels: subSpec.InstSpec.Labels,
	}

	matchedRelease, err := ae.sch.handler.GetMatchedRelease(kt, meta)
	if err != nil {
		// filter out the no matched strategies error and handle it specially.
		// so that sidecar do not retry repeatedly.
		if errf.Error(err).Code == errf.RecordNotFound {
			return 0, 0, nil
		}
		if errors.Is(err, errf.ErrAppInstanceNotMatchedRelease) {
			return 0, 0, nil
		}
	}

	return matchedRelease, cursor, nil
}

func (ae *appEvent) pushEvents(events []*types.EventMeta) {
	ae.eQueue.push(events)
}

func (ae *appEvent) watchEvents(ctx context.Context) {
	notifier := ae.eQueue.notifier()
	for {
		select {
		case <-ctx.Done():
			logs.Warnf("biz[%d], app[%d] event handler stop watch events", ae.bizID, ae.appID)
			return
		case <-notifier:
		}

		es := ae.eQueue.popAll()
		if len(es) == 0 {
			continue
		}

		ae.eventHandler(es)
	}
}

func (ae *appEvent) eventHandler(events []*types.EventMeta) {

	// the event should be handled one by one.
	// Note: do not try to aggregate these events, otherwise some sidecars
	// may get 'unexpected' release because of the inconsistent of cache,
	// which is unacceptable.
	for _, one := range events {

		kt := kit.New()
		switch one.Spec.Resource {
		case table.Publish:
			logs.Infof("start do biz: %d, app: %d publish broadcast to all sidecars, event id: %d, rid: %s", ae.bizID,
				ae.appID, one.ID, kt.Rid)

			// app level publish operation, all the sidecar instance should be notified.
			ae.notifyWithApp(kt, one.ID)

		case table.Application:
			logs.Infof("start handle biz: %d, app: %d app event, event id: %d, rid: %s", ae.bizID, ae.appID,
				one.ID, kt.Rid)

			// app delete, all the sidecar instance should be notified.
			ae.handleAppEvent(kt, one)

		default:
			logs.V(2).Infof("received unused event for scheduler, skip, detail: %s, rid: %s", formatEvent(one), kt.Rid)
		}

		// update the current handled event cursor id.
		ae.cursor.Set(one.ID)
	}

}

// notifyWithApp notify events to all the app's consumer with requested app list.
func (ae *appEvent) notifyWithApp(kt *kit.Kit, cursorID uint32) {
	ae.sch.notifyEvent(kt, cursorID, ae.csm.Members())
}

// nolint: unused
func (ae *appEvent) notifyWithInstance(kt *kit.Kit, cursorID uint32, uid string) {

	one := ae.csm.MemberWithUid(uid)
	if len(one) == 0 {
		logs.Infof("notify app: %d, sidecar with uid[%s] not exist, skip, rid: %s", ae.appID, uid, kt.Rid)
		return
	}

	ae.sch.notifyEvent(kt, cursorID, one)
}

// handleAppEvent handle app delete event, to delete current app's retry notify from retry list.
func (ae *appEvent) handleAppEvent(kt *kit.Kit, event *types.EventMeta) {
	switch event.Spec.OpType {
	case table.DeleteOp:
		ae.sch.retry.DeleteAppAllInstance(event.Attachment.AppID)

		members := ae.csm.Members()
		count := 0
		for _, one := range members {
			if one.InstSpec.AppID == event.Attachment.AppID {
				one.Receiver.CloseWatch()
				count++
			}
		}

		logs.Infof("success handle biz: %d app: %d delete event, close sidecar watch stream number is %d",
			event.Attachment.BizID, event.Attachment.AppID, count)

	default:
		logs.V(2).Infof("received unused app event, skip, detail: %s, rid: %s", formatEvent(event), kt.Rid)
	}
}
