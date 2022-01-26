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

package utils

import (
	"sync"
	"time"
)

type Cache interface {
	SetData(interface{})
	GetData() interface{}
}

type ResourceCache struct {
	sync.RWMutex
	lasteUpdateTime *time.Time
	timeout         time.Duration
	data            interface{}
}

func NewResourceCache(timeout time.Duration) Cache {
	return &ResourceCache{timeout: timeout}
}

func (rc *ResourceCache) SetData(data interface{}) {
	rc.Lock()
	defer rc.Unlock()
	rc.data = data
	now := time.Now()
	rc.lasteUpdateTime = &now
}

func (rc *ResourceCache) GetData() interface{} {
	rc.RLock()
	defer rc.RUnlock()
	if rc.needRenew() {
		return nil
	}
	return rc.data
}

func (rc *ResourceCache) needRenew() bool {
	if rc.lasteUpdateTime == nil {
		rc.data = nil
		return true
	}
	if time.Now().Sub(*rc.lasteUpdateTime) >= rc.timeout {
		rc.data = nil
		return true
	}
	return false
}
