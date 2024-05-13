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
	"reflect"
	"testing"
	"time"
)

var (
	event1 = ListenerEvent{
		EventTime: time.Now(),
		Name:      "event1",
		Namespace: "test",
	}
	event2 = ListenerEvent{
		EventTime: time.Now(),
		Name:      "event2",
		Namespace: "test",
	}
)

// TestEventCacheCreate test cache create function
func TestEventCacheCreate(t *testing.T) {
	cache := NewEventCache()
	cache.Set("event1", &event1)
	event1Get, isFound := cache.Get("event1")
	if !isFound || !reflect.DeepEqual(&event1, event1Get) {
		t.Errorf("test failed")
	}
	_, isFound = cache.Get("noexisted")
	if isFound {
		t.Errorf("test failed")
	}
}

// TestEventCacheList test cache list function
func TestEventCacheList(t *testing.T) {
	cache := NewEventCache()
	cache.Set("event1", &event1)
	cache.Set("event2", &event2)
	retList := cache.List()
	if len(retList) != 2 {
		t.Errorf("test failed")
	}
}

// TestEventCacheDelete test cache delete function
func TestEventCacheDelete(t *testing.T) {
	cache := NewEventCache()
	cache.Set("event1", &event1)
	cache.Set("event2", &event2)
	_, found := cache.Delete("event1")
	if !found {
		t.Errorf("test failed")
	}
	_, found = cache.Delete("event2")
	if !found {
		t.Errorf("test failed")
	}
	_, found = cache.Delete("notexisted")
	if found {
		t.Errorf("test failed")
	}

	list := cache.List()
	if len(list) != 0 {
		t.Errorf("test failed")
	}
}

// TestEventCacheDrain test cache drain function
func TestEventCacheDrain(t *testing.T) {
	cache := NewEventCache()
	cache.Set("event1", &event1)
	cache.Set("event2", &event2)
	newCache := NewEventCache()
	cache.Drain(newCache)

	list1 := cache.List()
	list2 := newCache.List()

	if len(list1) != 0 || len(list2) != 2 {
		t.Errorf("test failed")
	}
}
