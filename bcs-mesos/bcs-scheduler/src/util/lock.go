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

package util

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

var Lock *ObjectLock

func init() {
	Lock = &ObjectLock{
		locks: make(map[string]*sync.RWMutex),
	}
}

type ObjectLock struct {
	rw    sync.RWMutex
	locks map[string]*sync.RWMutex
}

func (l *ObjectLock) Lock(obj interface{}, key string) {
	k := fmt.Sprintf("%s.%s", reflect.TypeOf(obj).Name(), key)
	l.rw.RLock()
	myLock, ok := l.locks[k]
	l.rw.RUnlock()
	if ok {
		myLock.Lock()
		return
	}

	l.rw.Lock()
	myLock, ok = l.locks[k]
	if !ok {
		blog.Info("create lock(%s), current locknum(%d)", k, len(l.locks))
		l.locks[k] = new(sync.RWMutex)
		myLock, _ = l.locks[k]
	}
	l.rw.Unlock()
	myLock.Lock()
	return
}

func (l *ObjectLock) UnLock(obj interface{}, key string) {
	k := fmt.Sprintf("%s.%s", reflect.TypeOf(obj).Name(), key)
	l.rw.RLock()
	myLock, ok := l.locks[k]
	l.rw.RUnlock()

	if !ok {
		blog.Error("lock(%s) not exist when do unlock", k)
		return
	}
	myLock.Unlock()
}
