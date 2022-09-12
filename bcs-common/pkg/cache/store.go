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

package cache

import (
	"fmt"
)

// Store is storage interface
type Store interface {
	Add(obj interface{}) error
	Update(obj interface{}) error
	Delete(obj interface{}) error
	List() []interface{}
	ListKeys() []string
	Get(obj interface{}) (item interface{}, exists bool, err error)
	GetByKey(key string) (item interface{}, exists bool, err error)
	Num() int
	Clear()
	Replace([]interface{}) error
}

// ObjectKeyFunc define make object to a uniq key
type ObjectKeyFunc func(obj interface{}) (string, error)

// KeyError wrapper error return from ObjectKeyFunc
type KeyError struct {
	Obj interface{}
	Err error
}

// Error return string info for KeyError
func (k KeyError) Error() string {
	return fmt.Sprintf("create key for object %+v failed: %v", k.Obj, k.Err)
}

// DataNoExist return when No data in Store
type DataNoExist struct {
	Obj interface{}
}

// Error return string info
func (k DataNoExist) Error() string {
	return fmt.Sprintf("no data object %+v in Store", k.Obj)
}
