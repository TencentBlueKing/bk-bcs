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
	"reflect"
	"testing"
	"time"
)

// eventCollector for unit test
type eventCollector struct {
	events []Event
}

// OnEvent implements Handler interface
func (ec *eventCollector) OnEvent(event Event) error {
	ec.events = append(ec.events, event)
	return nil
}

// TestChecker tests config add, update and delete event
func TestChecker(t *testing.T) {
	testCases := []struct {
		cachedMd5Map map[string]string
		md5Map       map[string]string
		events       []Event
	}{
		{
			cachedMd5Map: map[string]string{
				"file1": "md5-1",
				"file2": "md5-2",
			},
			md5Map: map[string]string{
				"file1": "md5-1",
				"file2": "md5-2",
				"file3": "md5-3",
			},
			events: []Event{
				{
					Type:     EventAdd,
					Filename: "file3",
					Content:  "md5-3",
				},
			},
		},
		{
			cachedMd5Map: map[string]string{
				"file1": "md5-1",
				"file2": "md5-2222",
				"file3": "md5-3",
			},
			md5Map: map[string]string{
				"file1": "md5-1",
				"file2": "md5-2",
				"file3": "md5-3",
			},
			events: []Event{
				{
					Type:     EventUpdate,
					Filename: "file2",
					Content:  "md5-2",
				},
			},
		},
		{
			cachedMd5Map: map[string]string{
				"file1": "md5-1",
				"file2": "md5-2",
				"file3": "md5-3",
			},
			md5Map: map[string]string{
				"file1": "md5-1",
				"file3": "md5-3",
			},
			events: []Event{
				{
					Type:     EventDelete,
					Filename: "file2",
					Content:  "",
				},
			},
		},
	}

	for _, test := range testCases {
		newCollector := &eventCollector{}
		newWatcher := &Watcher{
			checkPeriod: 1 * time.Second,
			lister: &MockLister{
				FileMd5Map: test.md5Map,
			},
			md5sumMap: test.cachedMd5Map,
			stopCh:    make(chan struct{}),
		}
		newWatcher.RegisterHandler(newCollector)
		newWatcher.doNotify()
		if !reflect.DeepEqual(newCollector.events, test.events) {
			t.Errorf("expect %v but get %v", test.events, newCollector.events)
		}
	}
}
