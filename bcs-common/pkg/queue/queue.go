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
 *
 */

package queue

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// Queue integrates all data events to one seqential queue
type Queue interface {
	// Push specified event to local queue
	Push(e *Event)
	// Get event from queue, blocked
	Get() (*Event, error)
	// AGet async get event from queue, not blocked
	AGet() (*Event, error)
	// GetChannel event reading queue
	GetChannel() (<-chan *Event, error)
	// Close close Queue
	Close()
}

// NewQueue create default Queue for local usage
func NewQueue() Queue {
	return &channelQueue{
		localQ: make(chan *Event, watch.DefaultChannelBuffer),
	}
}

// channelQueue default queue using channel
type channelQueue struct {
	localQ chan *Event
}

// Push specified event to local queue
func (cq *channelQueue) Push(e *Event) {
	if e != nil {
		cq.localQ <- e
	}
}

// Get event from queue
func (cq *channelQueue) Get() (*Event, error) {
	e, ok := <-cq.localQ
	if ok {
		return e, nil
	}
	return nil, fmt.Errorf("Queue closed")
}

// AGet async get event from queue, not blocked
func (cq *channelQueue) AGet() (*Event, error) {
	select {
	case e, ok := <-cq.localQ:
		if ok {
			return e, nil
		}
		return nil, fmt.Errorf("Queue closed")
	default:
		return nil, nil
	}
}

// GetChannel event reading queue
func (cq *channelQueue) GetChannel() (<-chan *Event, error) {
	if cq.localQ == nil {
		return nil, fmt.Errorf("lost event queue")
	}
	return cq.localQ, nil
}

// Close event queue
func (cq *channelQueue) Close() {
	close(cq.localQ)
}
