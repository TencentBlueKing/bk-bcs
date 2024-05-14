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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

type appPool struct {
	sch  *Scheduler
	lock sync.RWMutex
	pool map[uint32]*appEvent
}

// AddSidecar add a sidecar instance to the app subscriber list
func (ap *appPool) AddSidecar(currentRelease uint32, sn uint64, subSpec *SubscribeSpec) error {

	ap.lock.Lock()
	defer ap.lock.Unlock()
	app, exist := ap.pool[subSpec.InstSpec.AppID]
	if !exist {
		app = newAppEvent(subSpec.InstSpec.BizID, subSpec.InstSpec.AppID, ap.sch)
		ap.pool[subSpec.InstSpec.AppID] = app
	}

	if err := app.AddSidecar(currentRelease, sn, subSpec); err != nil {
		return err
	}

	return nil
}

// RemoveSidecar remove a sidecar instance from the app subscriber list
func (ap *appPool) RemoveSidecar(sn uint64, appID uint32) {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	app, exist := ap.pool[appID]
	if !exist {
		return
	}

	if !app.RemoveSidecar(sn) {
		return
	}

	app.Stop()
	delete(ap.pool, appID)
}

// PushEvent push events to the according app event handler.
func (ap *appPool) PushEvent(appID uint32, es []*types.EventMeta) {
	if len(es) == 0 {
		return
	}

	ap.lock.RLock()
	app, exist := ap.pool[appID]
	ap.lock.RUnlock()
	if !exist {
		return
	}

	app.pushEvents(es)
}

// gc NOTES
// Note: GC
// nolint: unused
func (ap *appPool) gc() {
	panic("implement this")
}
