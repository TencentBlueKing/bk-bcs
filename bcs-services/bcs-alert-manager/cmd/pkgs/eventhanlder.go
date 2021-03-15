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

package pkgs

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/handler/eventhandler"
)

const (
	// alert-system interface concurrency
	EventHandleConcurrencyNum = 100
	// alert-system handle batch EventNum
	EventHandleAlertEventNum = 100
	// alert-system chan QueueLen
	EventHandleChanQueueLen = 1024
	// alert-system batchAggregation switch
	EventHandleBatchAggregation = false
)

var (
	eventHandlerOnce sync.Once
	eventHandler     *eventhandler.SyncEventHandler
)

// GetEventSyncHandler get eventSyncHandler consumer
func GetEventSyncHandler(options *config.AlertManagerOptions) *eventhandler.SyncEventHandler {
	eventHandlerOnce.Do(func() {
		eventHandler = eventhandler.NewSyncEventHandler(eventhandler.Options{
			AlertEventBatchNum: func() int {
				if options.HandlerConfig.AlertEventNum <= 0 {
					return EventHandleAlertEventNum
				}
				return options.HandlerConfig.AlertEventNum
			}(),
			ConcurrencyNum: func() int {
				if options.HandlerConfig.ConcurrencyNum <= 0 {
					return EventHandleConcurrencyNum
				}
				return options.HandlerConfig.ConcurrencyNum
			}(),
			ChanQueueNum: func() int {
				if options.HandlerConfig.ChanQueueNum <= 0 {
					return EventHandleChanQueueLen
				}
				return options.HandlerConfig.ChanQueueNum
			}(),
			IsBatchAggregation: func() bool {
				if options.HandlerConfig.IsBatchAggregation {
					return options.HandlerConfig.IsBatchAggregation
				}
				return EventHandleBatchAggregation
			}(),
			Client: GetAlertClient(options),
		})
		if eventHandler == nil {
			panic("init NewSyncEventHandler failed")
		}
		blog.Infof("init EventSyncHandler successful")
	})

	return eventHandler
}
