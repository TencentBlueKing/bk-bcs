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
	"strconv"
	"sync"

	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

func initConsumer(mc *metric) *consumer {
	return &consumer{
		lo:   new(sync.Mutex),
		list: make(map[uint64]*member),
		mc:   mc,
	}
}

// consumer holds one app's all the consumer.
type consumer struct {
	lo *sync.Mutex
	// do not attempt to use app instance's uid as the map's key,
	// this can avoid different app's instance use the same uid.
	// map[sn]*member
	list map[uint64]*member
	mc   *metric
}

// Add one member to the instances.
func (in *consumer) Add(sn uint64, subSpec *SubscribeSpec) *member {

	one := &member{
		SubscribeSpec: subSpec,
		sn:            sn,
	}

	in.lo.Lock()
	defer in.lo.Unlock()

	in.list[sn] = one
	in.mc.consumerCount.With(prm.Labels{"biz": tools.Itoa(subSpec.InstSpec.BizID),
		"app": tools.Itoa(subSpec.InstSpec.AppID)}).Inc()

	return one
}

// Delete one member form the instances.
// it returns true if all the app consumers(sidecar) is empty.
func (in *consumer) Delete(sn uint64) bool {
	in.lo.Lock()
	defer in.lo.Unlock()

	in.mc.consumerCount.With(prm.Labels{"biz": tools.Itoa(in.list[sn].InstSpec.BizID),
		"app": strconv.Itoa(int(in.list[sn].InstSpec.AppID))}).Dec()
	delete(in.list, sn)

	return len(in.list) == 0
}

// Members list app members which works at namespace mode.
func (in *consumer) Members() []*member {
	in.lo.Lock()
	defer in.lo.Unlock()

	members := make([]*member, len(in.list))
	idx := 0
	for _, mem := range in.list {
		members[idx] = mem
		idx++
	}

	return members
}

// MemberWithUid get members with uid.
// normally, only one can be found, but user may configure the several sidecar
// with the same uid, so we match it with a list.
func (in *consumer) MemberWithUid(uid string) []*member {
	in.lo.Lock()
	defer in.lo.Unlock()

	matched := make([]*member, 0)
	for _, m := range in.list {
		if m.InstSpec.Uid == uid {
			matched = append(matched, m)
		}
	}

	return matched

}
