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

package reflector

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// ListerWatcher is interface perform list all objects and start a watch
type ListerWatcher interface {
	// List should return a list type object
	List() ([]meta.Object, error)
	// Watch should begin a watch from remote storage
	Watch() (watch.Interface, error)
}

//ListFunc define list function for ListerWatcher
type ListFunc func() ([]meta.Object, error)

//WatchFunc define watch function for ListerWatcher
type WatchFunc func() (watch.Interface, error)

//ListWatch implements ListerWatcher interface
//protect ListFunc or WatchFunc is nil
type ListWatch struct {
	ListFn  ListFunc
	WatchFn WatchFunc
}

// List should return a list type object
func (lw *ListWatch) List() ([]meta.Object, error) {
	if lw.ListFn == nil {
		return nil, nil
	}
	return lw.ListFn()
}

// Watch should begin a watch from remote storage
func (lw *ListWatch) Watch() (watch.Interface, error) {
	if lw.WatchFn == nil {
		return nil, fmt.Errorf("Lost watcher function")
	}
	return lw.WatchFn()
}
