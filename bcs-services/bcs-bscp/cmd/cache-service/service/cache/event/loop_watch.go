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

package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

type consumerFunc func(kt *kit.Kit, es []*table.Event) (needRetry bool)

// loopWatch is used to loop watch the events from db with list operation,
// and call the consumer to consume them.
type loopWatch struct {
	ds    daoSet
	state serviced.State

	// consumer is to consume the watched events.
	consumer consumerFunc
	mc       *metric
}

func (lw *loopWatch) run() error {

	if lw.consumer == nil {
		return errors.New("no consumer is set to consume the watched events")
	}

	go lw.loop()
	lw.tickState()

	return nil
}

// loop loopWatch the events and call the consumer to consume the events.
func (lw *loopWatch) loop() {

	// sleep a while before start loop to avoid restart scenario.
	time.Sleep(5 * time.Second)
	logs.Infof("start job to loop events...")

	// currentCursor is the last already consumed event's id.
	currentCursor := uint32(0)
	ob := &observer{state: lw.state, preState: false}
	initialized := false
	timeReminder := time.Now()
	notifier := shutdown.AddNotifier()
	for {
		select {
		case <-notifier.Signal:
			logs.Infof("received shutdown signal, stop loop watch job success...")
			notifier.Done()
			return
		default:
		}

		kt := kit.New()
		reWatch, tryNext := ob.tryNext()
		if reWatch || !initialized {
			// service state have been changed, and need to do re-watch to get the
			// latest event cursor.
			latestCursor, err := lw.ds.event.LatestCursor(kt)
			if err != nil {
				lw.mc.errCounter.With(prm.Labels{}).Inc()
				logs.Errorf("watch event, but get latest cursor failed, retry later, err: %v, rid: %s", err, kt.Rid)
				initialized = false
				time.Sleep(2 * time.Second)
				continue
			}

			logs.Infof("service state has changed, the latest cursor id from db is %d, rid: %s", latestCursor, kt.Rid)
			// update the latest cursor and try next loop
			currentCursor = latestCursor
			initialized = true
			continue
		}

		if !tryNext {
			logs.V(2).Infof("this is slave, do not need to loop, skip. rid: %s", kt.Rid)
			time.Sleep(5 * time.Second)
			continue
		}

		lw.mc.lastCursor.With(prm.Labels{}).Set(float64(currentCursor))

		lastCursor, retry := lw.loopOneStep(kt, currentCursor)
		if retry {
			// this means no events are found or the events have been consumed failed,
			// so use the current latest cursor to do next watch again.
			if time.Since(timeReminder) >= 10*time.Minute {
				logs.Infof("finished one watch loop, will try re-loop with cursor: %d, rid: %s", currentCursor, kt.Rid)
				timeReminder = time.Now()
			}
			time.Sleep(time.Second)
			continue
		}

		// for now, watched events and consumed them success, then update the event cursor.
		if err := lw.ds.event.RecordCursor(kt, lastCursor); err != nil {
			lw.mc.errCounter.With(prm.Labels{}).Inc()
			logs.Errorf("record the last cursor: %d failed, will try later, err: %v, rid: %s", lastCursor, err, kt.Rid)
			time.Sleep(time.Second)
			continue
		}

		logs.Infof("watch event, record the new consumed last cursor: %d success, rid: %s", lastCursor, kt.Rid)

		// update the current cursor to the last cursor and do the next round watch.
		currentCursor = lastCursor
		time.Sleep(time.Second)

	}
}

func (lw *loopWatch) loopOneStep(kt *kit.Kit, currentCursor uint32) (lastCursor uint32, retry bool) {

	const (
		// limit do not > 200, otherwise, the list event operation will be failed, because
		// the page limit is limited.
		limit = 200

		// because the event is fired after all the resource's operation is done but
		// not be committed, so it's needed to delay some time to allow the transaction
		// is committed, so that the resource can be accessed.
		// Note:
		// change this delay time carefully, unless you know what will be effected.
		// it effects how long will the user's client can notice that a new release
		// is published.
		lagSeconds = 5
	)

	start := currentCursor
	for {
		last, atEnd, retry := lw.doOneStep(kt, start, limit, lagSeconds)
		if retry {
			// something unexpected happens, normally is an error.
			return start, true
		}

		if atEnd {
			// if start != currentCursor, it means some events have already been consumed,
			// do not need to retry this loop.
			return start, start == currentCursor
		}

		// if not at end, then update the start and try next loop.
		start = last
		time.Sleep(10 * time.Millisecond)
	}
}

// doOneStep do once loop watch operation from the start cursor with a step.
// because the event's id is generated concurrently, and the inserted event's time is different, this may
// result with the lower event id has been inserted with a larger insert time.
// So we CAN NOT list events with cursor and sort with insert time at the same time. The right and reasonable
// way is to list the event with the start event id and sort it with increasing id order, and then filter out the
// events which are satisfied with the lag seconds continuously(one by one) until the not satisfied one occurred.
func (lw *loopWatch) doOneStep(kt *kit.Kit, start uint32, limit uint, lagSeconds int) (lastCursor uint32, atEnd,
	retry bool) {
	// order these events with event id.
	opt := &types.BasePage{Start: 0, Limit: limit, Sort: "id", Order: types.Ascending}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		logs.Errorf("validate page option failed, err: %v, rid: %s", err, kt.Rid)
		return start, false, true
	}

	details, _, err := lw.ds.event.List(kt, start, opt)
	if err != nil {
		logs.Errorf("list event failed, err: %v, rid: %s", err, kt.Rid)
		return start, false, true
	}

	if len(details) == 0 {
		logs.V(1).Infof("watch events with cursor: %d, limit: %d, but no events found, rid: %s", start, limit, kt.Rid)
		return start, true, false
	}

	// filter out events which is satisfied with the lag seconds.
	lag := time.Now().Unix() - int64(lagSeconds)
	filterOut := make([]*table.Event, 0)
	for _, one := range details {
		if one.State.FinalStatus != table.UnknownFS {
			// if the event do have a final status, it means the event related
			// resource's db transaction has already finished with success or failed.
			// we do not care the event lag time under this circumstance.
			filterOut = append(filterOut, one)
			logs.V(1).Infof("filter out event %s, rid: %s", formatEvent(one), kt.Rid)
			continue
		}
		// here we do not filter out the event with which final status is table.FailedFS(failed),
		// because we want these events related resource's cache can be refreshed forcefully.

		if one.Revision.CreatedAt.Unix() <= lag {
			// event with unknown final status, but event lag time is within the wanted lag time
			// is also acceptable, these events are assumed to be a successful final state.
			filterOut = append(filterOut, one)
			logs.V(1).Infof("filter out event %s, rid: %s", formatEvent(one), kt.Rid)
			continue
		}
		// once the event's lag time is overhead what we wanted, the following events will
		// not be consumed for now, they will be consumed at next loop later.
		break
	}

	if len(filterOut) == 0 {
		logs.V(1).Infof("watch events with cursor: %d, filter out '0' events, do nothing, rid: %s", start, kt.Rid)
		return start, true, false
	}

	logs.Infof("watched %d events with cursor: %d, rid: %s", len(filterOut), start, kt.Rid)
	if logs.V(2) {
		js, _ := json.Marshal(filterOut)
		logs.Infof("watched events details as follows: %s, rid: %s", string(js), kt.Rid)
	}

	now := time.Now()
	// call the consumer to consume the events.
	if retry := lw.consumer(kt, filterOut); retry {
		// consume the events failed, retry this loop again.
		return lastCursor, false, true
	}

	lw.mc.loopLagMS.With(prm.Labels{}).Observe(tools.SinceMS(now))
	for _, one := range filterOut {
		lw.mc.eventCounter.With(prm.Labels{"type": string(one.Spec.Resource), "biz": tools.Itoa(
			one.Attachment.BizID)}).Inc()
	}

	lastCursor = filterOut[len(filterOut)-1].ID
	return lastCursor, false, false
}

func formatEvent(one *table.Event) string {
	return fmt.Sprintf("id: %d, biz: %d, app: %d, resource: %s, op: %s, resource_id: %d, uid: %s, status: %d", one.ID,
		one.Attachment.BizID, one.Attachment.AppID, one.Spec.Resource, one.Spec.OpType, one.Spec.ResourceID,
		one.Spec.ResourceUid, one.State.FinalStatus)
}

func (lw *loopWatch) tickState() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			logs.Infof("loop watch event, tick master state: %v", lw.state.IsMaster())
		}
	}()
}

type observer struct {
	state    serviced.State
	preState bool
}

// tryNext describe whether we can still loop the next  events.
// this is a master slave service. we should re-watch the event from the previous
// event cursor, only when we do this, we can loop the continuous events later which
// is no events is skipped or duplicated.
func (o *observer) tryNext() (reWatch bool, loop bool) {
	current := o.state.IsMaster()

	if o.preState == current {
		if !current {
			// not master, status not changed, and can not loop
			return false, false
		}

		// is master, status not changed, and can loop
		return false, true
	}

	logs.Infof("loop watch, is master status changed from %v to %v.", o.preState, current)

	// update status
	o.preState = current

	// status already changed, and can not continue loop, need to re-watch again.
	return true, false
}
