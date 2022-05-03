/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package filewatcher

import (
	"time"

	"go.uber.org/zap"
)

// Handler event handler
type Handler interface {
	OnEvent(event Event) error
}

// Watcher watchs event for configs
type Watcher struct {
	checkPeriod time.Duration
	handlers    []Handler
	md5sumMap   map[string]string
	stopCh      chan struct{}

	lister Lister
}

// NewWatcher create new watcher
func NewWatcher(lister Lister, checkPeriod time.Duration) *Watcher {
	return &Watcher{
		checkPeriod: checkPeriod,
		lister:      lister,
		md5sumMap:   make(map[string]string),
		stopCh:      make(chan struct{}),
	}
}

// RegisterHandler register event handler
func (w *Watcher) RegisterHandler(h Handler) {
	w.handlers = append(w.handlers, h)
}

// WatchLoop start watch loop
func (w *Watcher) WatchLoop() {
	w.doNotify()

	ticker := time.NewTicker(w.checkPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.doNotify()
		case <-w.stopCh:
			zap.L().Info("config watcher exit")
			return
		}
	}
}

func (w *Watcher) doNotify() {
	newMd5sumMap, err := w.listFiles()
	if err != nil {
		zap.L().Error("list files failed", zap.Error(err))
		return
	}
	adds, updates, dels := w.diffFiles(w.md5sumMap, newMd5sumMap)
	if w.sendEvent(EventDelete, dels, newMd5sumMap) {
		for _, del := range dels {
			delete(w.md5sumMap, del)
		}
	}
	if w.sendEvent(EventAdd, adds, newMd5sumMap) {
		for _, add := range adds {
			w.md5sumMap[add] = newMd5sumMap[add]
		}
	}
	if w.sendEvent(EventUpdate, updates, newMd5sumMap) {
		for _, update := range updates {
			w.md5sumMap[update] = newMd5sumMap[update]
		}
	}
}

// Stop stops watch loop
func (w *Watcher) Stop() {
	w.stopCh <- struct{}{}
}

func (w *Watcher) listFiles() (map[string]string, error) {
	return w.lister.List()
}

func (w *Watcher) diffFiles(oldMap, newMap map[string]string) ([]string, []string, []string) {
	var adds []string
	var updates []string
	var dels []string
	for k, v := range newMap {
		oldv, ok := oldMap[k]
		if !ok {
			adds = append(adds, k)
		} else if oldv != v {
			updates = append(updates, k)
		}
	}
	for k := range oldMap {
		_, ok := newMap[k]
		if !ok {
			dels = append(dels, k)
		}
	}
	return adds, updates, dels
}

func (w *Watcher) sendEvent(t EventType, filenames []string, contentMap map[string]string) bool {
	isSuccessful := true
	for _, filename := range filenames {
		for _, handler := range w.handlers {
			newEvent := Event{
				Type:     t,
				Filename: filename,
			}
			if t != EventDelete {
				newEvent.Content = contentMap[filename]
			}
			if err := handler.OnEvent(newEvent); err != nil {
				zap.L().Error("trigger event failed", zap.Error(err), zap.Any("event", newEvent))
				isSuccessful = false
			}
		}
	}
	return isSuccessful
}
