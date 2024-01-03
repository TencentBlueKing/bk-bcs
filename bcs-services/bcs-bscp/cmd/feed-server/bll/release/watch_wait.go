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

package release

import (
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
)

func initWait() *waitShutdown {
	wait := &waitShutdown{
		wait:      sync.WaitGroup{},
		broadcast: make(chan struct{}),
	}

	go wait.waiting()

	return wait
}

// waitShutdown wait for the service to shut down and broadcast the
// shutdown message to all the watch handlers for each sidecar.
type waitShutdown struct {
	wait      sync.WaitGroup
	broadcast chan struct{}
}

func (ws *waitShutdown) waiting() {

	var start time.Time
	notifier := shutdown.AddNotifier()
	<-notifier.Signal
	start = time.Now()
	logs.Infof("sidecar watch received shutdown message, start broadcast the bounce message now.")

	// broadcast the shutdown message to all the watching sidecar
	close(ws.broadcast)
	ws.wait.Wait()

	cost := time.Since(start).Milliseconds()
	logs.Infof("sidecar watch finished the job to broadcast the bounce message to sidecar, cost: %dms.", cost)

	notifier.Done()
}

func (ws *waitShutdown) signal() <-chan struct{} {
	ws.wait.Add(1)
	return ws.broadcast
}

func (ws *waitShutdown) done() {
	ws.wait.Done()
}
