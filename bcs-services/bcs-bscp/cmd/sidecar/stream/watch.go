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

package stream

import (
	"context"
	"errors"
	"fmt"
	"io"

	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbfs "bscp.io/pkg/protocol/feed-server"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"
)

// StartWatch start watch events from upstream server
func (s *stream) StartWatch(onChange *types.OnChange) error {
	w := &watch{
		stream:   s,
		onChange: onChange,
	}
	s.watch = w

	return w.runWatch()
}

type watch struct {
	// vas.Ctx is used to cancel/close the watch stream.
	vas *kit.Vas
	// cancelFunc is the cancel function combined with the vas.Ctx
	cancelFunc context.CancelFunc

	stream   *stream
	onChange *types.OnChange
}

func (w *watch) runWatch() error {
	vas, cancel := w.stream.vasBuilder()
	w.vas = vas
	w.cancelFunc = cancel

	apps := make([]sfs.SideAppMeta, 0)
	for _, one := range w.stream.settings.AppSpec.Applications {
		releaseID, cursorID := uint32(0), uint32(0)
		cRelease, cCursor, exist := w.onChange.CurrentRelease(one.AppID)
		if exist {
			releaseID = cRelease
			cursorID = cCursor
		}

		apps = append(apps, sfs.SideAppMeta{
			AppID:            one.AppID,
			Namespace:        one.Namespace,
			Uid:              one.Uid,
			Labels:           one.Labels,
			CurrentReleaseID: releaseID,
			CurrentCursorID:  cursorID,
		})
	}

	payload := sfs.SideWatchPayload{
		BizID:        w.stream.settings.AppSpec.BizID,
		Applications: apps,
	}

	bytes, err := jsoni.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode watch payload failed, err: %v", err)
	}

	watchStream, err := w.stream.client.Watch(vas, bytes)
	if err != nil {
		return fmt.Errorf("watch upstream server with payload failed, err: %v, rid: %s", err, vas.Rid)
	}

	logs.Infof("watch stream success, and start watch events from upstream server, rid: %s", vas.Rid)

	go w.loopReceiveWatchedEvent(vas, watchStream)

	return nil
}

func (w *watch) loopReceiveWatchedEvent(vas *kit.Vas, wStream pbfs.Upstream_WatchClient) {
	for {
		select {
		case <-vas.Ctx.Done():
			logs.Warnf("watch will closed because of %v", vas.Ctx.Err())

			if err := wStream.CloseSend(); err != nil {
				logs.Errorf("close watch failed, err: %v, watch rid: %s", err, vas.Rid)
				return
			}

			logs.Infof("watch is closed successfully, watch rid: %s", vas.Rid)
			return

		default:
		}

		event, err := wStream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logs.Errorf("watch stream has been closed by remote upstream stream server, need to re-connect again")
				w.stream.NotifyReconnect(types.ReconnectSignal{Reason: "connection is closed " +
					"by remote upstream server"})
				return
			}

			logs.Errorf("watch stream is corrupted because of %v, rid: %s", err, vas.Rid)
			w.stream.NotifyReconnect(types.ReconnectSignal{Reason: "watch stream corrupted"})
			return
		}

		logs.Infof("received upstream event, apiVer: %s, payload: %s, rid: %s", event.ApiVersion.Format(),
			event.Payload, event.Rid)

		if !sfs.IsAPIVersionMatch(event.ApiVersion) {
			// 此处是不是不应该做版本兼容的校验？
			// TODO: set sidecar unhealthy, offline and exit.
			logs.Errorf("watch stream received incompatible event version: %s, rid: %s", event.ApiVersion.Format(),
				event.Rid)
			break
		}

		switch sfs.FeedMessageType(event.Type) {
		case sfs.Bounce:
			logs.Infof("received upstream bounce request, need to reconnect upstream server, rid: %s", event.Rid)
			w.stream.NotifyReconnect(types.ReconnectSignal{Reason: "received bounce request"})
			return

		case sfs.PublishRelease:
			change := &sfs.ReleaseChangeEvent{
				Rid:        event.Rid,
				APIVersion: event.ApiVersion,
				Payload:    event.Payload,
			}

			w.onChange.OnReleaseChange(change)
			continue

		default:
			logs.Errorf("watch stream received unsupported event type: %s, skip, rid: %s", event.Type, event.Rid)
			continue
		}
	}

	// TODO: version is not compatible, do something

}

// CloseWatch the current watch stream.
func (w *watch) CloseWatch() {
	if w.cancelFunc == nil {
		return
	}

	w.cancelFunc()
}
