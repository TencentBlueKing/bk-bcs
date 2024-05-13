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
	"time"
)

const (
	// EventAdd event type add
	EventAdd = "Add"
	// EventUpdate event type update
	EventUpdate = "Update"
	// EventDelete event type delete
	EventDelete = "Delete"
)

// EventType event type
type EventType string

// ListenerEvent listener event
type ListenerEvent struct {
	Type      EventType
	EventTime time.Time
	Name      string
	Namespace string
}

// NewListenerEvent create listener event
func NewListenerEvent(t EventType, name, ns string) *ListenerEvent {
	return &ListenerEvent{
		Type:      t,
		EventTime: time.Now(),
		Name:      name,
		Namespace: ns,
	}
}

// Key key for listener event
func (le *ListenerEvent) Key() string {
	return le.Namespace + "/" + le.Name
}

// ListenerEventList list for listener event
type ListenerEventList []ListenerEvent

// Len implements sort interface
func (lel ListenerEventList) Len() int {
	return len(lel)
}

// Swap implements sort interface
func (lel ListenerEventList) Swap(i, j int) {
	lel[i], lel[j] = lel[j], lel[i]
}

// Less implements sort interface
func (lel ListenerEventList) Less(i, j int) bool {
	return lel[i].EventTime.Before(lel[j].EventTime)
}
