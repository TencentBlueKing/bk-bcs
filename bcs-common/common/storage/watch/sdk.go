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
	"io"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
)

// Get a new Watcher with empty WatchOptions
func New(client *httpclient.HttpClient) *Watcher {
	return NewWithOption(&operator.WatchOptions{}, client)
}

// Get a new Watcher with provided WatchOptions
func NewWithOption(opts *operator.WatchOptions, client *httpclient.HttpClient) *Watcher {
	return &Watcher{
		opts:   opts,
		client: client,
	}
}

// Watcher maintains the actions watching to storage server
type Watcher struct {
	client *httpclient.HttpClient

	opts       *operator.WatchOptions
	storageUrl []string
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool

	resp  *http.Response
	event *operator.Event
	err   error

	nextSignal    chan struct{}
	receiveSignal chan struct{}
}

// Connect starts the watching
func (w *Watcher) Connect(storageURL []string) (err error) {
	if w.ctx != nil {
		return errors.EventWatchAlreadyConnect
	}

	w.storageUrl = storageURL

	if w.opts == nil {
		w.opts = &operator.WatchOptions{}
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
// EventWatchNoUrlAvailable error.
func (w *Watcher) connect() (err error) {
	body := &bytes.Buffer{}
	if err = codec.EncJsonWriter(w.opts, body); err != nil {
		return
	}

	for _, u := range w.storageUrl {
		r, err := http.NewRequest("POST", u, body)
		if err != nil {
			continue
		}

		if w.resp, err = w.client.GetClient().Do(r); err != nil || w.resp.StatusCode != http.StatusOK {
			continue
		}
		return nil
	}
	return errors.EventWatchNoUrlAvailable
}

// Waiting for flushed response body. If the connection break(EOF) and it is
// not closed, then reconnect automatically. If reconnect failed, then return
// EventWatchNoUrlAvailable error.
func (w *Watcher) watching() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.nextSignal:
			w.event = new(operator.Event)
			if w.err = codec.DecJsonReader(w.resp.Body, w.event); w.err == io.ErrUnexpectedEOF && !w.closed {
				if w.err = w.connect(); w.err != nil {
					w.event = operator.EventWatchBreak
				}
			}
			if w.event.Type == operator.Nop {
				w.Close()
				return
			}
			w.receiveSignal <- struct{}{}
		}
	}
}

// Stop watch and close the connection. It must be stop watch first then close connection,
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

// Get next event
func (w *Watcher) Next() (*operator.Event, error) {
	w.nextSignal <- struct{}{}
	select {
	case <-w.ctx.Done():
		return operator.EventWatchBreak, nil
	case <-w.receiveSignal:
		return w.event, w.err
	}
}
