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

	"go.uber.org/atomic"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// SubscribeSpec defines the metadata to watch the event.
type SubscribeSpec struct {
	InstSpec *sfs.InstanceSpec
	Receiver *Receiver
}

// Validate the SubscribeSpec is valid or not.
func (s SubscribeSpec) Validate() error {
	if s.InstSpec == nil {
		return errors.New("instance spec is nil")
	}

	if err := s.InstSpec.Validate(); err != nil {
		return err
	}

	if s.Receiver == nil {
		return errors.New("receiver is nil")
	}

	return nil
}

// InitReceiver initial a new receiver instance.
func InitReceiver(notify func(event *Event, sn uint64) bool, closeWatch context.CancelFunc) *Receiver {
	return &Receiver{
		state:      atomic.NewBool(true),
		closeWatch: closeWatch,
		notify:     notify,
	}
}

// Receiver is defined for the subscriber to control its working state
// and receive the event messages.
type Receiver struct {
	// State defines the receiver's working state.
	// 1. 'true' means this receiver is working and accept the subscribed
	// event message;
	// 2. 'false' means this receiver is already not working and stop to
	// receive any event messages.
	state *atomic.Bool
	// notify the event to the subscriber, if it is needed to retry send the event
	// then return true, otherwise, return false.
	notify func(event *Event, sn uint64) bool
	// closeWatch close sidecar and feed server watch stream.
	closeWatch context.CancelFunc
}

// SetState set the receiver's state
func (r *Receiver) SetState(state bool) {
	r.state.Store(state)
}

// State return the current state of receiver
func (r *Receiver) State() bool {
	return r.state.Load()
}

// Notify send the event to the subscriber.
func (r *Receiver) Notify(event *Event, uid string, sn uint64) bool {
	if !r.state.Load() {
		logs.Errorf("notify app: %d, uid: %s event failed, the subscriber has already closed, skip.",
			event.Change.AppID, uid)
		// skip the error, nothing can be done for the upper logic.
		return false
	}

	return r.notify(event, sn)
}

// CloseWatch close sidecar and feed server watch stream.
func (r *Receiver) CloseWatch() {
	r.closeWatch()
}

// Event defines the event details.
type Event struct {
	Change   *sfs.ReleaseEventMetaV1
	Instance *sfs.InstanceSpec
	CursorID uint32
}

type member struct {
	*SubscribeSpec
	sn uint64
}

func formatEvent(meta *types.EventMeta) string {
	return fmt.Sprintf("id: %d, biz: %d, app: %d, resource: %s, op: %s, resource_id: %d, uid: %s", meta.ID,
		meta.Attachment.BizID, meta.Attachment.AppID, meta.Spec.Resource, meta.Spec.OpType, meta.Spec.ResourceID,
		meta.Spec.ResourceUid)
}

// cursor is used to manage the app's current working(last handled) event cursor id.
type cursor struct {
	lo              sync.Mutex
	workingCursorID uint32
}

// ID return the current working cursor id
func (c *cursor) ID() uint32 {
	c.lo.Lock()
	defer c.lo.Unlock()

	return c.workingCursorID
}

// Set update the current working cursor id
func (c *cursor) Set(cursorID uint32) {
	c.lo.Lock()
	defer c.lo.Unlock()

	c.workingCursorID = cursorID
}

func initEventQueue() *eventQueue {
	return &eventQueue{
		lock:  sync.Mutex{},
		queue: make([]*types.EventMeta, 0),
		// the initial signal size is >=2 is ok.
		signal: make(chan struct{}, 3),
	}
}

// eventQueue is used to manage the app's all the events.
type eventQueue struct {
	lock   sync.Mutex
	queue  []*types.EventMeta
	signal chan struct{}
}

func (eq *eventQueue) push(es []*types.EventMeta) {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	eq.queue = append(eq.queue, es...)

	select {
	case eq.signal <- struct{}{}:
	default:
	}
}

func (eq *eventQueue) popAll() []*types.EventMeta {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	copied := make([]*types.EventMeta, len(eq.queue))
	copy(copied, eq.queue)

	// reset the queue.
	eq.queue = make([]*types.EventMeta, 0)

	return copied
}

func (eq *eventQueue) notifier() <-chan struct{} {
	return eq.signal
}
