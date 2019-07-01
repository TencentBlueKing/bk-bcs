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

package operator

import (
	"encoding/json"
	"time"
)

type WatchOptions struct {
	// Only watch the node itself, including children added, children removed and node value change.
	// Will not receive existing children's event.
	SelfOnly bool `json:"selfOnly"`

	// Max time of events will received. Watch will be ended after the last event. 0 for infinity.
	MaxEvents uint `json:"maxEvents"`

	// The max waiting time of each event. Watch will be ended after timeout. 0 for no limit.
	Timeout time.Duration `json:"timeout"`

	// The value-change event will be checked if it's different from last status. If not then this event
	// will be ignored. And it will not trigger timeout reset.
	MustDiff string `json:"mustDiff"`
}

type EventType int32

const (
	Nop EventType = iota
	Add
	Del
	Chg
	SChg
	Brk EventType = -1
)

func (et EventType) String() string {
	return eventTypeNames[et]
}

var (
	eventTypeNames = map[EventType]string{
		Nop:  "EventNop",
		Add:  "EventAdd",
		Del:  "EventDelete",
		Chg:  "EventChange",
		SChg: "EventSelfChange",
		Brk:  "EventWatchBreak",
	}
)

type Event struct {
	Type  EventType `json:"type"`
	Value M         `json:"value"`
}

var (
	EventWatchBreak         = &Event{Type: Brk, Value: nil}
	EventWatchBreakBytes, _ = json.Marshal(EventWatchBreak)
)
