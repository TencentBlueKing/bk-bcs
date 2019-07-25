/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package queue

import (
	"bk-bcs/bcs-common/pkg/watch"
	"fmt"
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
