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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
)

//EventType definition for watch
type EventType string

const (
	//EventSync sync event, reserved for force synchronization
	EventSync EventType = "SYNC"
	//EventAdded add event
	EventAdded EventType = "ADDED"
	//EventUpdated updated/modified event
	EventUpdated EventType = "UPDATED"
	//EventDeleted deleted event
	EventDeleted EventType = "DELETED"
	//EventErr error event for watch, error occured, but watch still works
	EventErr EventType = "ERROR"
	//DefaultChannelBuffer buffer for watch event channel
	DefaultChannelBuffer = 128
)

//Interface define watch channel
type Interface interface {
	//stop watch channel
	Stop()
	//get watch events, if watch stopped/error, watch must close
	// channel and exit, watch user must read channel like
	// e, ok := <-channel
	WatchEvent() <-chan Event
}

//Event holding event info for data object
type Event struct {
	Type EventType   `json:"type"`
	Data meta.Object `json:"data"`
}
