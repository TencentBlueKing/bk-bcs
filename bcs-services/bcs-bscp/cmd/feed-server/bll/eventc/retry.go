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

package eventc

import (
	"sync"

	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

func newRetryList(mc *metric) *retryList {
	return &retryList{
		lo:     sync.Mutex{},
		signal: make(chan struct{}, 5),
		list:   make(map[uint64]*retryMember),
		mc:     mc,
	}
}

// retryList holds all the resources which are needed to retry send events.
type retryList struct {
	lo     sync.Mutex
	signal chan struct{}

	// list hold all the need retry send event's consumer.
	// map[sn]*retryMember
	list map[uint64]*retryMember
	mc   *metric
}

type retryMember struct {
	member   *member
	cursorID uint32
}

// Add instance member to the retry list.
func (rl *retryList) Add(cursorID uint32, m *member) {
	rl.lo.Lock()
	defer rl.lo.Unlock()

	rm, exist := rl.list[m.sn]
	if exist {
		// aggregate the duplicated members with same serial number.
		if rm.cursorID < cursorID {
			// use the larger cursor id as the new retry cursor id.
			rm.cursorID = cursorID

			rl.notify()
			return
		}

		rl.notify()
		return
	}

	rl.list[m.sn] = &retryMember{
		member:   m,
		cursorID: cursorID,
	}

	rl.notify()
}

func (rl *retryList) notify() {
	select {
	case rl.signal <- struct{}{}:
	default:
	}
}

// Signal return a channel which can read retry list change signal
func (rl *retryList) Signal() <-chan struct{} {
	return rl.signal
}

// DeleteInstance remove an instance from retry list
func (rl *retryList) DeleteInstance(sn uint64) {
	rl.lo.Lock()
	defer rl.lo.Unlock()

	delete(rl.list, sn)
}

// DeleteAppAllInstance remove an app's all instance from retry list.
func (rl *retryList) DeleteAppAllInstance(appID uint32) {
	rl.lo.Lock()
	defer rl.lo.Unlock()

	for sn, val := range rl.list {
		if val.member.InstSpec.AppID == appID {
			delete(rl.list, sn)
		}
	}
}

// Purge list all the apps instances in the retry list
// it will finally purge these instances from retry list.
func (rl *retryList) Purge() (int, []*retryMember) {
	rl.lo.Lock()
	defer rl.lo.Unlock()

	if len(rl.list) == 0 {
		return 0, make([]*retryMember, 0)
	}

	cnt := len(rl.list)
	copied := make([]*retryMember, 0, cnt)
	for sn := range rl.list {
		rl.mc.retryCounter.With(prm.Labels{"biz": tools.Itoa(rl.list[sn].member.InstSpec.BizID),
			"app": tools.Itoa(rl.list[sn].member.InstSpec.AppID)}).Inc()
		copied = append(copied, rl.list[sn])
	}

	// purge the instance list.
	rl.list = make(map[uint64]*retryMember)

	return cnt, copied
}
