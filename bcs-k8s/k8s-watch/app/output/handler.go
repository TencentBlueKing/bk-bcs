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

package output

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/k8s-watch/app/output/action"
)

// =================== interface & struct ===================

type Action interface {
	Add(syncData *action.SyncData)
	Delete(syncData *action.SyncData)
	Update(syncData *action.SyncData)
}

type Handler struct {
	dataType string
	queue    chan *action.SyncData
	action   Action
}

// =================== Handle: in ===================

func (handler *Handler) Handle(data *action.SyncData) {
	handler.queue <- data
}

// =================== Run: out ===================

func (handler *Handler) Run() {
	wait.Until(handler.consume, time.Second, wait.NeverStop)

}

// consume handler.queue To Action func
func (handler *Handler) consume() {
	for {
		select {
		case syncData := <-handler.queue:
			currentQueueLen := len(handler.queue)
			if currentQueueLen != 0 && currentQueueLen%10 == 0 {
				glog.Infof("Data in handler %s's queue: %d", handler.dataType, currentQueueLen)
			}

			switch syncData.Action {
			case "Add":
				handler.action.Add(syncData)
				break
			case "Delete":
				handler.action.Delete(syncData)
				break
			case "Update":
				handler.action.Update(syncData)
				break
			default:
				glog.Errorf("Writer handler got unknown Action: %s", syncData.Action)
			}

		}

	}

}
