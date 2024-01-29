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

package release

import (
	"context"
	"fmt"

	"go.uber.org/atomic"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/eventc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/lcache"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/feed-server"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
)

// Watch handle watch messages delivered from sidecar.
func (rs *ReleasedService) Watch(im *sfs.IncomingMeta, payload *sfs.SideWatchPayload,
	fws pbfs.Upstream_WatchServer) error {

	ctx, cancel := context.WithCancel(context.Background())
	wh := &watchHandler{
		counter:     atomic.NewInt32(0),
		stream:      fws,
		im:          im,
		sidePayload: payload,
		sideMeta:    im.Meta,
		cache:       rs.cache,
		watcher:     rs.watcher,
		snList:      make(map[uint64]*appReminder),
		wait:        rs.wait,
		ctx:         ctx,
		cancelCtx:   cancel,
	}

	if err := wh.subscribe(); err != nil {
		return err
	}

	wh.waitForFinalize()
	return nil
}

type appReminder struct {
	appID    uint32
	uid      string
	receiver *eventc.Receiver
}

type watchHandler struct {
	// counter is used to count the message numbers which send back to sidecar.
	counter *atomic.Int32
	// snList stores the sidecar's registered app's SN returned by watcher's register.
	snList      map[uint64]*appReminder
	stream      pbfs.Upstream_WatchServer
	im          *sfs.IncomingMeta
	cache       *lcache.Cache
	watcher     eventc.Watcher
	sidePayload *sfs.SideWatchPayload
	sideMeta    *sfs.SidecarMetaHeader
	wait        *waitShutdown
	ctx         context.Context
	cancelCtx   context.CancelFunc
}

func (wh *watchHandler) subscribe() error {

	for _, one := range wh.sidePayload.Applications {

		meta, err := wh.cache.App.GetMeta(wh.im.Kit, wh.sidePayload.BizID, one.AppID)
		if err != nil {
			return fmt.Errorf("get app(%d) meta failed, err: %v", one.AppID, err)
		}
		spec := &eventc.SubscribeSpec{
			InstSpec: &sfs.InstanceSpec{
				BizID:      wh.sidePayload.BizID,
				App:        one.App,
				AppID:      one.AppID,
				Uid:        one.Uid,
				Labels:     one.Labels,
				ConfigType: meta.ConfigType,
			},
			Receiver: eventc.InitReceiver(wh.eventReceiver, wh.cancelCtx),
		}

		sn, err := wh.watcher.Subscribe(one.CurrentReleaseID, one.CurrentCursorID, spec)
		if err != nil {
			return fmt.Errorf("subscribe app: %d event failed, err: %v", one.AppID, err)
		}

		wh.snList[sn] = &appReminder{
			appID:    one.AppID,
			uid:      one.Uid,
			receiver: spec.Receiver,
		}
	}

	return nil
}

func (wh *watchHandler) eventReceiver(event *eventc.Event, sn uint64) bool {

	rid := wh.nextRid()
	releasePayload := &sfs.ReleaseChangePayload{
		ReleaseMeta: event.Change,
		Instance:    event.Instance,
		CursorID:    event.CursorID,
	}

	payload, err := releasePayload.Encode()
	if err != nil {
		logs.Errorf("received release change event, but marshal it failed, skip, fingerprint: %s, err: %v, rid: %s",
			wh.im.Meta.Fingerprint, err, rid)
		return false
	}

	wm := &pbfs.FeedWatchMessage{
		ApiVersion: sfs.CurrentAPIVersion,
		Rid:        rid,
		Type:       uint32(releasePayload.MessageType()),
		Payload:    payload,
	}
	if err := wh.stream.Send(wm); err != nil {
		logs.Errorf("send release message to sidecar failed, fingerprint: %s, sn: %d, err: %v, rid: %s",
			wh.im.Meta.Fingerprint, sn, err, rid)

		// Note: 新增判断机制，判断是否还需要重试，避免大量、高频无效重试。
		// 可考虑增加server端主动关链的操作，强制sidecar进行重链，修复链路。
		// if status.Convert(err).Code() == codes.Unavailable {
		//	logs.Errorf("downstream sidecar is unavailable, stop send event, rid: %s", rid)
		//	return false
		// }
		return true
	}

	return false
}

// waitForFinalize do the watch handler's clean up job.
func (wh *watchHandler) waitForFinalize() {
	// deregister this watch handler wait job finally.
	defer wh.wait.done()

	var reason string
	bounce := false
	select {
	case <-wh.stream.Context().Done():
		reason = "sidecar watch stream error, " + wh.stream.Context().Err().Error()
		bounce = false

	case <-wh.wait.signal():
		reason = "feed server shutting down"
		bounce = true

	case <-wh.ctx.Done():
		reason = "feed server initiative close watch stream"
		bounce = true
	}

	for sn, reminder := range wh.snList {
		// set the receiver's state to not working status to stop receive the
		// watched events at first, so that the events can not send to here again.
		reminder.receiver.SetState(false)

		// unsubscribe the registration
		wh.watcher.Unsubscribe(reminder.appID, sn, reminder.uid)
	}

	if !bounce {
		logs.Infof("finish deregister sidecar's watch job because of %s, fingerprint: %s, rid: %s", reason,
			wh.im.Meta.Fingerprint, wh.im.Kit.Rid)
		return
	}

	// send the bounce message to tell the sidecar bounce to another
	// feed server automatically.
	wm := &pbfs.FeedWatchMessage{
		ApiVersion: sfs.CurrentAPIVersion,
		Rid:        wh.nextRid(),
		Type:       uint32(sfs.Bounce),
		Payload:    nil,
	}
	if err := wh.stream.Send(wm); err != nil {
		logs.Errorf("send 'bounce' message to sidecar failed, err: %v, fingerprint: %s, rid: %s", err,
			wh.im.Meta.Fingerprint, wh.im.Kit.Rid)
		return
	}

	logs.V(1).Infof("send 'bounce' message to sidecar success, rid: %s", wh.im.Kit.Rid)

	logs.Infof("finish deregister sidecar's watch job because of %s, fingerprint: %s, rid: %s", reason,
		wh.im.Meta.Fingerprint, wh.im.Kit.Rid)
}

// nextRid generate the next rid based on the incoming rid with rules.
func (wh *watchHandler) nextRid() string {
	wh.counter.Inc()
	return fmt.Sprintf("%s-fd-%d", wh.im.Kit.Rid, wh.counter.Load())
}
