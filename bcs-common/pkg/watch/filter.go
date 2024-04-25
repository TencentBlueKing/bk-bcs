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

// Package watch xxx
package watch

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
)

// SelectFunc custom function to verify how filter acts
type SelectFunc func(meta.Object) (bool, error)

// NewSelectWatch wrap watcher with filter func
func NewSelectWatch(w Interface, fn SelectFunc) Interface {
	cxt, canceler := context.WithCancel(context.Background())
	f := &SelectorWatch{
		watch:        w,
		cxt:          cxt,
		selectFn:     fn,
		stopFn:       canceler,
		eventChannel: make(chan Event, DefaultChannelBuffer),
	}
	go f.selectWatchEvent()
	return f
}

// SelectorWatch watcher wraper offer filter function to filter data object if needed
type SelectorWatch struct {
	watch        Interface          // inner watch for original data to filte
	selectFn     SelectFunc         // filter for watch
	cxt          context.Context    // context for stop
	stopFn       context.CancelFunc // stopFn for context
	eventChannel chan Event         // event channel for data already filtered
}

// Stop stop watch channel
func (fw *SelectorWatch) Stop() {
	fw.stopFn()
}

// WatchEvent get watch events
func (fw *SelectorWatch) WatchEvent() <-chan Event {
	return fw.eventChannel
}

// selectWatchEvent handler for filter
func (fw *SelectorWatch) selectWatchEvent() {
	tunnel := fw.watch.WatchEvent()
	if tunnel == nil {
		fw.watch.Stop()
		close(fw.eventChannel)
		return
	}
	defer func() {
		fw.watch.Stop()
		close(fw.eventChannel)
	}()
	for {
		select {
		case event, ok := <-tunnel:
			if !ok {
				return
			}
			matched, err := fw.selectFn(event.Data)
			if err != nil || !matched {
				continue
			}
			fw.eventChannel <- event
		case <-fw.cxt.Done():
			return
		}
	}
}
