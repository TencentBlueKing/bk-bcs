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

package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/atomic"
)

func initBlocker() *blocker {
	return &blocker{
		state: atomic.NewBool(false),
	}
}

// blocker is a tool to block the request if it is needed.
type blocker struct {
	lo sync.Mutex
	// signal is used to broadcast unblock message for all the waiters.
	signal chan struct{}
	// state describe whether the blocker is blocked or not for now.
	state *atomic.Bool
}

// TryBlock try to block the blocker
// if it blocks success, it returns true.
func (bl *blocker) TryBlock() bool {
	bl.lo.Lock()
	defer bl.lo.Unlock()

	if bl.state.Load() {
		// the blocker is already blocked, do not need to block again.
		return false
	}

	// set the blocker state to be blocked.
	bl.state.Store(true)

	// init the signal channel, which is used to broadcast unblock messages.
	bl.signal = make(chan struct{}, 1)

	return true
}

// Unblock the blocker if the blocker is blocked.
func (bl *blocker) Unblock() {
	bl.lo.Lock()
	defer bl.lo.Unlock()

	if !bl.state.Load() {
		// the blocker is not  blocked, do not need to unblock.
		return
	}

	// set the blocker state to unblocked.
	bl.state.Store(false)

	// broadcast unblock messages to all the waiters
	close(bl.signal)
}

// WaitMS until the block is released if the block is already blocked
// or the timeout time in millisecond is arrived.
// if timeoutMS <= 0, it means block without timeout.
func (bl *blocker) WaitMS(timeoutMS int64) error {

	if !bl.state.Load() {
		// the blocker is not blocked, do not need to wait.
		return nil
	}

	if timeoutMS > 0 {
		// wait with timeout
		du := time.Duration(timeoutMS) * time.Millisecond
		select {
		case <-time.After(du):
			return fmt.Errorf("wait for blocker unblock, but timeout after %s", du.String())
		case <-bl.signal:
		}

		return nil
	}

	// wait without timeout
	select {
	case <-bl.signal:
	}

	return nil
}

// WaitWithContext until the block is released if the block is already blocked
// or the context is canceled or done. it returns an error if the context is
// canceled or done.
func (bl *blocker) WaitWithContext(ctx context.Context) error {

	if !bl.state.Load() {
		// the blocker is not blocked, do not need to wait.
		return nil
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("wait blocker unblock failed because of %s", ctx.Err().Error())

	case <-bl.signal:
	}

	return nil
}
