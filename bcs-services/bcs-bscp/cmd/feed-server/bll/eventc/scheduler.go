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

package eventc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/samber/lo"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/lcache"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/observer"
	btyp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbct "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbhook "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/hook"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Option defines options to create a scheduler instance.
type Option struct {
	Observer observer.Interface
	Cache    *lcache.Cache
}

// Handler all the call back handles, used to handle schedule jobs.
type Handler struct {
	// GetMatchedRelease get the specified app instance's current release id.
	// Note: this function's match pipeline should not use cache data,
	GetMatchedRelease func(kt *kit.Kit, meta *btyp.AppInstanceMeta) (uint32, error)
}

// NewScheduler create a new scheduler instance.
// And, scheduler start accept subscribe and unsubscribe operations, but still not works for
// events processing, which means scheduler do not match the subscribed instance's release.
func NewScheduler(opt *Option, name string) (*Scheduler, error) {

	provider, err := repository.NewProvider(cc.FeedServer().Repository)
	if err != nil {
		return nil, fmt.Errorf("schduler init repository provider failed, err: %v", err)
	}

	mc := initMetric(name)
	sch := &Scheduler{
		ob:            opt.Observer,
		lc:            opt.Cache,
		retry:         newRetryList(mc),
		serialNumber:  atomic.NewUint64(0),
		notifyLimiter: semaphore.NewWeighted(int64(cc.FeedServer().Downstream.NotifyMaxLimit)),
		mc:            mc,
		provider:      provider,
	}

	sch.appPool = &appPool{
		sch:  sch,
		lock: sync.RWMutex{},
		pool: make(map[uint32]*appEvent),
	}

	go sch.watchRetry()

	return sch, nil
}

// Scheduler works at all the events handling jobs.
// 1. it accepts subscribe from sidecar and unsubscribe when the sidecar close the connection.
// 2. it sends events to all the subscribers and will retry to send event if it fails.
type Scheduler struct {
	appPool *appPool
	ob      observer.Interface
	lc      *lcache.Cache
	//nolint:unused
	csm          *consumer
	retry        *retryList
	handler      *Handler
	serialNumber *atomic.Uint64
	provider     repository.Provider
	// notifyLimiter controls the concurrent of sending the event messages to the
	// event subscribers.
	notifyLimiter *semaphore.Weighted
	mc            *metric
}

// Run start the scheduler's job
func (sch *Scheduler) Run(h *Handler) error {
	if h == nil {
		return errors.New("handler not set")
	}

	sch.handler = h

	// start watch events from the observer, and if events happens, then
	// match these related app's instance release with call back.
	go sch.loopWatch()
	return nil
}

// Subscribe register an app instance to subscribe the release event for it.
// it returns a serial number(as is sn) which represent this app instance's watch identity id.
func (sch *Scheduler) Subscribe(currentRelease uint32, currentCursorID uint32, subSpec *SubscribeSpec) (uint64, error) {
	if err := subSpec.Validate(); err != nil {
		return 0, err
	}

	if err := sch.waitForObserverReady(currentCursorID); err != nil {
		return 0, err
	}

	sn := sch.nextSN()
	if err := sch.appPool.AddSidecar(currentRelease, sn, subSpec); err != nil {
		return 0, err
	}

	return sn, nil
}

// waitForObserverReady check the cursor id now, should be <= scheduler's local cursor id,
// if not, then wait until it is.
func (sch *Scheduler) waitForObserverReady(cursorID uint32) error {

	interval := sch.ob.LoopInterval()

	after := time.After(10 * interval)
	for {
		time.Sleep(interval / 2)

		select {
		case <-after:
			return errors.New("wait for observer to be ready timeout")

		default:
		}

		if !sch.ob.IsReady() {
			continue
		}

		// the request sidecar's cursor id should 'less equal' than the current
		// observer's cursor id, this can ensure what it to be matching current
		// release is correct, because it can avoid this instance got a mistaken
		// matched released because of the inconsistently local cache.
		if cursorID <= sch.ob.CurrentCursor() {
			return nil
		}
	}

}

// Unsubscribe for the app to unsubscribe the event.
func (sch *Scheduler) Unsubscribe(appID uint32, sn uint64, uid string) {
	// remove it from consumer
	sch.appPool.RemoveSidecar(sn, appID)

	// remove it from retry list if it exists.
	sch.retry.DeleteInstance(sn)

	logs.Infof("unsubscribe watch event success, app: %d, uid: %s, sn: %d", appID, uid, sn)
}

// nextSN generate next serial number.
func (sch *Scheduler) nextSN() uint64 {
	return sch.serialNumber.Add(1)
}

// loopWatch start watch the events from the observer and handle these watched events.
func (sch *Scheduler) loopWatch() {

	notifier := shutdown.AddNotifier()
	next := sch.ob.Next()
	for {
		select {
		case <-notifier.Signal:
			logs.Infof("watch scheduler received shutdown signal, stop loop watch scheduler successfully.")
			notifier.Done()
			return

		case events := <-next:
			logs.Infof("received %d events from observer", len(events))

			sch.handleOneBatch(events)
		}
	}

}

func (sch *Scheduler) handleOneBatch(events []*types.EventMeta) {

	arrangedApps := make(map[uint32][]*types.EventMeta)
	for _, one := range events {
		_, exist := arrangedApps[one.Attachment.AppID]
		if !exist {
			arrangedApps[one.Attachment.AppID] = make([]*types.EventMeta, 0)
		}

		arrangedApps[one.Attachment.AppID] = append(arrangedApps[one.Attachment.AppID], one)
	}

	for appID, events := range arrangedApps {
		sch.appPool.PushEvent(appID, events)
	}

}

func (sch *Scheduler) notifyEvent(kt *kit.Kit, cursorID uint32, members []*member) {
	if len(members) == 0 {
		return
	}

	cnt := 0
	wg := sync.WaitGroup{}
	for idx := range members {
		cnt++

		if err := sch.notifyLimiter.Acquire(kt.Ctx, 1); err != nil {
			sch.retry.Add(cursorID, members[idx])
			logs.Errorf("acquire notify semaphore failed, inst: %s, err: %v, rid: %s", members[idx].InstSpec.Format(),
				err, kt.Rid)
			continue
		}

		wg.Add(1)

		go func(one *member) {
			sch.notifyOne(kt, cursorID, one)
			sch.notifyLimiter.Release(1)
			wg.Done()
		}(members[idx])
	}

	wg.Wait()
}

func (sch *Scheduler) notifyOne(kt *kit.Kit, cursorID uint32, one *member) {
	// Note: optimize this when a mount of instances have the same labels with same release id.
	inst := one.InstSpec
	meta := &btyp.AppInstanceMeta{
		BizID:  inst.BizID,
		AppID:  inst.AppID,
		App:    inst.App,
		Uid:    inst.Uid,
		Labels: inst.Labels,
	}
	releaseID, e := sch.handler.GetMatchedRelease(kt, meta)
	if e != nil {
		sch.retry.Add(cursorID, one)
		logs.Errorf("get %s [sn: %d] matched strategy failed, err: %v, rid: %s", inst.Format(), one.sn, e, kt.Rid)
		return
	}

	event := new(Event) //nolint:ineffassign

	switch inst.ConfigType {
	case table.KV:
		kvList, err := sch.lc.ReleasedKv.Get(kt, inst.BizID, releaseID)
		if err != nil {
			logs.Errorf("get %s [sn: %d] released[%d] Kv failed, err: %v, rid: %s", inst.Format(), one.sn, releaseID, err,
				kt.Rid)
			sch.retry.Add(cursorID, one)
			return
		}
		if len(kvList) == 0 {
			return
		}
		event = sch.buildEventForRkv(inst, kvList, releaseID, cursorID)

	case table.File:
		ciList, err := sch.lc.ReleasedCI.Get(kt, inst.BizID, releaseID)
		if err != nil {
			logs.Errorf("get %s [sn: %d] released[%d] CI failed, err: %v, rid: %s", inst.Format(), one.sn, releaseID, err,
				kt.Rid)
			sch.retry.Add(cursorID, one)
			return
		}
		preHook, postHook, err := sch.lc.ReleasedHook.Get(kt, inst.BizID, releaseID)
		if err != nil {
			logs.Errorf("get %s [sn: %d] released[%d] hook failed, err: %v, rid: %s", inst.Format(), one.sn, releaseID, err,
				kt.Rid)
			sch.retry.Add(cursorID, one)
			return
		}
		if len(ciList) == 0 {
			return
		}
		event = sch.buildEvent(inst, ciList, preHook, postHook, releaseID, cursorID)

	default:
		logs.Errorf("Unsupported application type (%s), rid: %s", inst.Format(), kt.Rid)
		return
	}

	if one.Receiver.Notify(event, inst.Uid, one.sn) {
		logs.Warnf("notify app instance event failed, need retry, biz: %d, app: %d, uid: %s, sn: %d, rid: %s",
			inst.BizID, inst.AppID, inst.Uid, one.sn, kt.Rid)
		sch.retry.Add(cursorID, one)
	}
}

func (sch *Scheduler) buildEvent(inst *sfs.InstanceSpec, ciList []*types.ReleaseCICache,
	pre *types.ReleasedHookCache, post *types.ReleasedHookCache, releaseID uint32, cursorID uint32) *Event {
	uriD := sch.provider.URIDecorator(inst.BizID)
	ciMeta := make([]*sfs.ConfigItemMetaV1, 0)
	for _, one := range ciList {
		cis := one.ConfigItemSpec
		// filter out mismatched config items
		// if inst.Match is empty, then all the config items are matched.
		if len(inst.Match) > 0 {
			isMatch := lo.SomeBy(inst.Match, func(scope string) bool {
				ok, _ := tools.MatchConfigItem(scope, cis.Path, cis.Name)
				return ok
			})
			if !isMatch {
				continue
			}
		}
		m := &sfs.ConfigItemMetaV1{
			ID:       one.ID,
			CommitID: one.CommitID,
			ContentSpec: &pbct.ContentSpec{
				Signature: one.CommitSpec.Signature,
				ByteSize:  one.CommitSpec.ByteSize,
				Md5:       one.CommitSpec.Md5,
			},
			ConfigItemSpec: &pbci.ConfigItemSpec{
				Name:     cis.Name,
				Path:     cis.Path,
				FileType: string(cis.FileType),
				FileMode: string(cis.FileMode),
				// Memo is useless for sidecar, so remove it.
				Memo: "",
				Permission: &pbci.FilePermission{
					User:      cis.Permission.User,
					UserGroup: cis.Permission.UserGroup,
					Privilege: cis.Permission.Privilege,
				},
			},
			ConfigItemAttachment: &pbci.ConfigItemAttachment{
				BizId: one.Attachment.BizID,
				AppId: one.Attachment.AppID,
			},
			ConfigItemRevision: &pbbase.Revision{
				Creator:  one.Revision.Creator,
				Reviser:  one.Revision.Reviser,
				CreateAt: one.Revision.CreatedAt.Format(time.RFC3339),
				UpdateAt: one.Revision.UpdatedAt.Format(time.RFC3339),
			},
			RepositoryPath: uriD.Path(one.CommitSpec.Signature),
		}
		ciMeta = append(ciMeta, m)
	}
	var releaseName string
	if len(ciList) > 0 {
		releaseName = ciList[0].ReleaseName
	}
	var preHook, postHook *pbhook.HookSpec
	if pre != nil {
		preHook = &pbhook.HookSpec{
			Type:    pre.Type.String(),
			Content: pre.Content,
		}
	}
	if post != nil {
		postHook = &pbhook.HookSpec{
			Type:    post.Type.String(),
			Content: post.Content,
		}
	}

	return &Event{
		Change: &sfs.ReleaseEventMetaV1{
			App:         inst.App,
			AppID:       inst.AppID,
			ReleaseID:   releaseID,
			ReleaseName: releaseName,
			CIMetas:     ciMeta,
			Repository: &sfs.RepositoryV1{
				Root: uriD.Root(),
				Url:  uriD.Url(),
			},
			PreHook:  preHook,
			PostHook: postHook,
		},
		Instance: inst,
		CursorID: cursorID,
	}
}

func (sch *Scheduler) watchRetry() {
	limiter := rate.NewLimiter(2, 1)
	notifier := shutdown.AddNotifier()
	retrySignal := sch.retry.Signal()
	for {

		select {
		case <-notifier.Signal:
			logs.Infof("event scheduler retry job received shutdown signal, stop retry job success.")
			notifier.Done()
			return

		case <-retrySignal:
			// apps or instances need to be retried to send events.
		}

		_ = limiter.Wait(context.TODO())

		kt := kit.New()
		logs.Infof("scheduler received retry send event signal, rid: %s", kt.Rid)

		instCount, members := sch.retry.Purge()
		for _, one := range members {
			sch.notifyEvent(kt, one.cursorID, []*member{one.member})
		}

		logs.Infof("finished scheduler retry send event job, instance count: %d, rid: %s", instCount, kt.Rid)
	}

}

func (sch *Scheduler) buildEventForRkv(inst *sfs.InstanceSpec, kvList []*types.ReleaseKvCache, releaseID uint32,
	cursorID uint32) *Event {

	kvMeta := make([]*sfs.KvMetaV1, 0)
	for _, one := range kvList {
		// filter out mismatched config items
		if !tools.MatchPattern(one.Key, inst.Match) {
			continue
		}
		m := &sfs.KvMetaV1{
			ID:     one.ID,
			Key:    one.Key,
			KvType: one.KvType,
			Revision: &pbbase.Revision{
				Creator:  one.Revision.Creator,
				Reviser:  one.Revision.Reviser,
				CreateAt: one.Revision.CreatedAt.Format(time.RFC3339),
				UpdateAt: one.Revision.UpdatedAt.Format(time.RFC3339),
			},
			KvAttachment: &pbkv.KvAttachment{
				BizId: one.Attachment.BizID,
				AppId: one.Attachment.AppID,
			},
			ContentSpec: pbct.PbContentSpec(one.ContentSpec),
		}
		kvMeta = append(kvMeta, m)
	}

	return &Event{
		Change: &sfs.ReleaseEventMetaV1{
			App:       inst.App,
			AppID:     inst.AppID,
			ReleaseID: releaseID,
			KvMetas:   kvMeta,
		},
		Instance: inst,
		CursorID: cursorID,
	}
}
