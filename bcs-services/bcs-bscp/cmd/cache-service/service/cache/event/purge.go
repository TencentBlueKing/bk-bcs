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

package event

import (
	"context"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// purger is used to purge the outdated events from db.
type purger struct {
	ds    daoSet
	state serviced.State
}

var sleepTime = 5 * time.Minute

// purge the outdated events that is two weeks ago, we do this everyday at a fixed time.
func (p *purger) purge() {
	// sleep a while before start loop to avoid restart scenario.
	time.Sleep(sleepTime)
	logs.Infof("start purge history event job")

	notifier := shutdown.AddNotifier()
	var lastPurgeDay int
	for {
		kt := kit.New()
		ctx, cancel := context.WithCancel(kt.Ctx)
		kt.Ctx = ctx

		select {
		case <-notifier.Signal:
			logs.Infof("stop purge history event job success")
			cancel()
			notifier.Done()
			return
		default:
		}

		if time.Now().Hour() != 1 {
			time.Sleep(sleepTime)
			continue
		}

		if lastPurgeDay == time.Now().Day() {
			time.Sleep(sleepTime)
			continue
		}
		if !p.state.IsMaster() {
			logs.V(2).Infof("this is slave, do not need to purge, skip. rid: %s", kt.Rid)
			time.Sleep(sleepTime)
			continue
		}

		logs.Infof("start purge history event, rid: %s", kt.Rid)

		if err := p.ds.event.Purge(kt, 14); err != nil {
			logs.Errorf("purge events failed, err: %v", err)
			time.Sleep(sleepTime)
			continue
		}

		logs.Infof("purge history event success, rid: %s", kt.Rid)

		lastPurgeDay = time.Now().Day()
	}
}
