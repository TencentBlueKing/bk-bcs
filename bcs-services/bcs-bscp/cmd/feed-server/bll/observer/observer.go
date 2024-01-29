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

// Package observer NOTES
package observer

import (
	"errors"
	"fmt"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"

	clientset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// New create an observer instance.
func New(h *Handler, cs *clientset.ClientSet, name string) (Interface, error) {
	ob := &observer{
		cs:           cs,
		mc:           initMetric(name),
		handler:      h,
		pipes:        make([]chan []*types.EventMeta, 0),
		lastCursorID: atomic.NewUint32(0),
		loopInterval: 250 * time.Millisecond,
		isReady:      atomic.NewBool(false),
	}

	if err := ob.run(); err != nil {
		return nil, err
	}

	return ob, nil
}

// Handler works to delete the related local resources.
type Handler struct {
	LocalCache func(kt *kit.Kit, es []*types.EventMeta)
}

// observer is used to watch the event and clean the cached local cache
// if the resource have related event.
type observer struct {
	cs           *clientset.ClientSet
	mc           *metric
	handler      *Handler
	pipes        []chan []*types.EventMeta
	lastCursorID *atomic.Uint32
	loopInterval time.Duration
	isReady      *atomic.Bool
}

func (ob *observer) run() error {
	startCursor, err := ob.doInit()
	if err != nil {
		logs.Errorf("init local cache observer failed, err: %v", err)
		return err
	}

	go ob.doLoop(startCursor)

	return nil
}

func (ob *observer) doInit() (uint32, error) {

	logs.Infof("local cache observer, start init start cursor.")

	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			return 0, errors.New("initialize local cache observer, get current cursor reminder timeout")
		default:
		}

		kt := kit.New()
		resp, err := ob.cs.CS().GetCurrentCursorReminder(kt.RpcCtx(), new(pbbase.EmptyReq))
		if err != nil {
			logs.Errorf("init observer, get cursor reminder failed, retry later, err: %v, rid: %s", err, kt.Rid)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		ob.isReady.Store(true)

		logs.Infof("local cache observer got start event reminder cursor: %d, rid: %s", resp.Cursor, kt.Rid)

		return resp.Cursor, nil
	}

}

// IsReady returns 'true' if the observer is already initialized,
// which means the start cursor is loaded from db.
func (ob *observer) IsReady() bool {
	return ob.isReady.Load()
}

// CurrentCursor returns the latest consumed event's cursor id which is consumed
// by the local cache.
func (ob *observer) CurrentCursor() uint32 {
	return ob.lastCursorID.Load()
}

// Next return a channel, it blocks until a batch of events occurs.
func (ob *observer) Next() <-chan []*types.EventMeta {
	ch := make(chan []*types.EventMeta, 200)
	ob.pipes = append(ob.pipes, ch)
	return ch
}

// LoopInterval return the observer's loop duration to watch the events.
func (ob *observer) LoopInterval() time.Duration {
	return ob.loopInterval
}

func (ob *observer) doLoop(startCursor uint32) {
	logs.Infof("start loop events job with start cursor: %d.", startCursor)

	reminder := time.Now()
	notifier := shutdown.AddNotifier()
	for {
		select {
		case <-notifier.Signal:
			logs.Infof("received shutdown signal, stop local cache observer success.")
			notifier.Done()
			return
		default:
		}

		// loop the events every 250ms.
		time.Sleep(ob.loopInterval)

		kt := kit.New()
		if time.Since(reminder) >= 5*time.Minute {
			logs.Infof("list events with start cursor: %d, rid: %s", startCursor, kt.Rid)
			reminder = time.Now()
		}

		ob.mc.lastCursor.With(prm.Labels{}).Set(float64(startCursor))

		lastCursor, events, err := ob.listEvents(kt, startCursor)
		if err != nil {
			logs.Errorf("list events with cursor: %d failed, retry later, err: %v, rid: %s", startCursor, err, kt.Rid)
			time.Sleep(time.Second)
			continue
		}

		if len(events) == 0 {
			logs.V(2).Infof("list 0 events, skip, rid: %s", kt.Rid)
			continue
		}

		logs.Infof("received %d events with start cursor: %d, rid: %s", len(events), startCursor, kt.Rid)

		ob.handleEvents(kt, events)

		// update the start cursor to the last handled event id as is event cursor
		startCursor = lastCursor

		logs.Infof("handle all events success, rid: %s", kt.Rid)
	}
}

const step = 200

func (ob *observer) listEvents(kt *kit.Kit, startCursor uint32) (uint32, []*types.EventMeta, error) {
	opt := &pbcs.ListEventsReq{
		StartCursor: startCursor,
		Page: &pbbase.BasePage{
			Start: 0,
			Limit: step,
			// order with ascending id.
			Sort:  "id",
			Order: string(types.Ascending),
		},
	}

	resp, err := ob.cs.CS().ListEventsMeta(kt.RpcCtx(), opt)
	if err != nil {
		return 0, nil, fmt.Errorf("list events failed, %v", err)
	}

	if len(resp.List) == 0 {
		return startCursor, make([]*types.EventMeta, 0), nil
	}

	events := make([]*types.EventMeta, len(resp.List))
	for idx := range resp.List {
		events[idx] = resp.List[idx].EventMeta()
	}

	lastCursor := events[len(resp.List)-1].ID

	return lastCursor, events, nil
}

func (ob *observer) handleEvents(kt *kit.Kit, metas []*types.EventMeta) {

	lastCursorID := metas[len(metas)-1].ID

	// update the last cursor id
	ob.lastCursorID.Store(lastCursorID)

	// Firstly, callback local cache's handler to purge the resource's cache
	ob.handler.LocalCache(kt, metas)

	// Secondly, send events with pipes
	for _, pipe := range ob.pipes {
		// send event metadata with pipe channel
		pipe <- metas
	}

}
