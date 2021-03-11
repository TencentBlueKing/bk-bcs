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

package storage

import (
	"time"

	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
)

//ChannelProxy Proxy offer particular channel for
//handling data in private goroutine
type ChannelProxy struct {
	clusterID     string
	dataQueue     chan *types.BcsSyncData //queue for async
	actionHandler InfoHandler             //data operator interface
}

//Run ChannelProxy running a dataType handler channel, stop Run() By external context
func (proxy *ChannelProxy) Run(ctx context.Context) {
	// report handler queue length periodically
	go proxy.reportHandlerQueueLength(ctx)

	tick := time.NewTicker(300 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("tick: ChannelProxy(%s) is alive, current task queue(%d/%d)",
				proxy.actionHandler.GetType(), len(proxy.dataQueue), cap(proxy.dataQueue))
			proxy.actionHandler.CheckDirty()

		case <-ctx.Done():
			blog.Info("ChannelProxy(%s) asked to exit, current task queue(%d/%d)",
				proxy.actionHandler.GetType(), len(proxy.dataQueue), cap(proxy.dataQueue))
			return

		case data := <-proxy.dataQueue:
			util.ReportHandlerQueueLengthDec(proxy.clusterID, proxy.actionHandler.GetType())
			if len(proxy.dataQueue)+100 > cap(proxy.dataQueue) {
				blog.Warnf("ChannelProxy(%s) busy, current task queue(%d/%d)",
					proxy.actionHandler.GetType(), len(proxy.dataQueue), cap(proxy.dataQueue))
			} else {
				blog.V(3).Infof("ChannelProxy(%s) receive task, current task queue(%d/%d)",
					proxy.actionHandler.GetType(), len(proxy.dataQueue), cap(proxy.dataQueue))
			}

			switch data.Action {
			case "Add":
				proxy.actionHandler.Add(data.Item)
				break
			case "Delete":
				proxy.actionHandler.Delete(data.Item)
				break
			case "Update":
				proxy.actionHandler.Update(data.Item)
				break
			default:
				blog.Error("CCHandler Get Unknown Action %s", data.Action)
			}
		}
	}
}

//Handle for handling data action like Add, Delete, Update
func (proxy *ChannelProxy) Handle(data *types.BcsSyncData) {
	if data == nil {
		blog.Error("ChannelProxy Get nil BcsSyncData")
		return
	}
	proxy.dataQueue <- data
	util.ReportHandlerQueueLengthInc(proxy.clusterID, proxy.actionHandler.GetType())
}

// HandleWithTimeOut send data to proxy.dataQueue with timeout
func (proxy *ChannelProxy) HandleWithTimeOut(data *types.BcsSyncData, timeout time.Duration) {
	if data == nil {
		blog.Error("ChannelProxy Get nil BcsSyncData")
		return
	}

	select {
	case proxy.dataQueue <- data:
		util.ReportHandlerQueueLengthInc(proxy.clusterID, proxy.actionHandler.GetType())
	case <-time.After(timeout):
		blog.Warn("can't handle data to dataQueue, queue timeout")
		util.ReportHandlerDiscardEvents(proxy.clusterID, proxy.actionHandler.GetType())
	}
}

func (proxy *ChannelProxy) reportHandlerQueueLength(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	blog.Infof("begin to monitor handler[%s] queue length", proxy.actionHandler.GetType())
	for {
		select {
		case <-ctx.Done():
			blog.Warn("external context cancel() %v", ctx.Err())
			return
		case <-ticker.C:
		}

		util.ReportHandlerQueueLength(proxy.clusterID, proxy.actionHandler.GetType(), float64(len(proxy.dataQueue)))
	}
}
