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

package worker

import (
	"sort"
	"sync"
)

// EventCache cache for listener event
type EventCache struct {
	lock  sync.Mutex
	items map[string]*ListenerEvent
}

// NewEventCache create event cache
func NewEventCache() *EventCache {
	return &EventCache{
		items: make(map[string]*ListenerEvent),
	}
}

// Set set listener event
func (c *EventCache) Set(key string, e *ListenerEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[key] = e
}

// Get get listener event by key
func (c *EventCache) Get(key string) (*ListenerEvent, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	item, found := c.items[key]
	if found {
		return item, true
	}
	return nil, false
}

// Delete delete listener event by key
func (c *EventCache) Delete(key string) (*ListenerEvent, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	item, found := c.items[key]
	if found {
		delete(c.items, key)
		return item, true
	}
	return nil, false
}

// List list cache
func (c *EventCache) List() []ListenerEvent {
	c.lock.Lock()
	defer c.lock.Unlock()
	list := ListenerEventList{}
	for _, v := range c.items {
		list = append(list, *v)
	}
	sort.Sort(list)
	return list
}

// Drain drain to another cache
func (c *EventCache) Drain(recvCache *EventCache) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for key, item := range c.items {
		delete(c.items, key)
		recvCache.Set(key, item)
	}
}

// Clean clean cache
func (c *EventCache) Clean() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items = make(map[string]*ListenerEvent)
}
