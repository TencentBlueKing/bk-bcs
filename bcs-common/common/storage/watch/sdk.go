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

package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

type storageError struct {
	Message string
	Code    int
}

func (se *storageError) Error() string {
	return se.Message
}

var (
	eventWatchAlreadyConnect = &storageError{Code: common.AdditionErrorCode + 6317, Message: "already connected"}
	eventWatchNoURLAvailable = &storageError{Code: common.AdditionErrorCode + 6318, Message: "no url available"}
)

// New get a new Watcher with empty WatchOptions
func New(client *httpclient.HttpClient) *Watcher {
	return NewWithOption(&types.WatchOptions{}, client)
}

// EventType event type
type EventType int32

const (
	// Nop no operation event
	Nop EventType = iota
	// Add add event
	Add
	// Del delete event
	Del
	// Chg change event
	Chg
	// SChg self change event
	SChg
	// Brk event
	Brk EventType = -1
)

// Event event of watch
type Event struct {
	Type  EventType              `json:"type"`
	Value map[string]interface{} `json:"value"`
}

var (
	// EventWatchBreak watch break event
	EventWatchBreak = &Event{Type: Brk, Value: nil}
	// EventWatchBreakBytes watch break event content
	EventWatchBreakBytes, _ = json.Marshal(EventWatchBreak)
)

// NewWithOption get a new Watcher with provided WatchOptions
func NewWithOption(opts *types.WatchOptions, client *httpclient.HttpClient) *Watcher {
	return &Watcher{
		opts:   opts,
		client: client,
	}
}

// Watcher maintains the actions watching to storage server
type Watcher struct {
	client *httpclient.HttpClient

	opts       *types.WatchOptions
	storageURL []string
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool

	resp  *http.Response
	event *Event
	err   error

	nextSignal    chan struct{}
	receiveSignal chan struct{}
}

// Connect starts the watching
func (w *Watcher) Connect(storageURL []string) (err error) {
	if w.ctx != nil {
		return eventWatchAlreadyConnect
	}

	w.storageURL = storageURL

	if w.opts == nil {
		w.opts = &types.WatchOptions{}
	}

	if err = w.connect(); err != nil {
		return
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.closed = false
	w.nextSignal, w.receiveSignal = make(chan struct{}), make(chan struct{})
	go w.watching()
	return
}

// connect: Try to connect the url list, if they are all unreachable, then return
// eventWatchNoURLAvailable error.
func (w *Watcher) connect() (err error) {
	body := &bytes.Buffer{}
	if err = codec.EncJsonWriter(w.opts, body); err != nil {
		return
	}

	for _, u := range w.storageURL {
		r, err := http.NewRequest("POST", u, body)
		if err != nil {
			continue
		}

		if w.resp, err = w.client.GetClient().Do(r); err != nil || w.resp.StatusCode != http.StatusOK {
			continue
		}
		return nil
	}
	return eventWatchNoURLAvailable
}

// Waiting for flushed response body. If the connection break(EOF) and it is
// not closed, then reconnect automatically. If reconnect failed, then return
// eventWatchNoUrlAvailable error.
func (w *Watcher) watching() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.nextSignal:
			w.event = new(Event)
			if w.err = codec.DecJsonReader(w.resp.Body, w.event); w.err == io.ErrUnexpectedEOF && !w.closed {
				if w.err = w.connect(); w.err != nil {
					w.event = EventWatchBreak
				}
			}
			if w.event.Type == Nop {
				w.Close()
				return
			}
			w.receiveSignal <- struct{}{}
		}
	}
}

// Close stop watch and close the connection. It must be stop watch first then close connection,
// because watching() will check if the watch is stop after get a EOF error and if not it will reconnect.
func (w *Watcher) Close() {
	w.cancel()
	w.closed = true
	if w.resp != nil {
		if body := w.resp.Body; body != nil {
			body.Close()
		}
	}
}

// Next get next event
func (w *Watcher) Next() (*Event, error) {
	w.nextSignal <- struct{}{}
	select {
	case <-w.ctx.Done():
		return EventWatchBreak, nil
	case <-w.receiveSignal:
		return w.event, w.err
	}
}
