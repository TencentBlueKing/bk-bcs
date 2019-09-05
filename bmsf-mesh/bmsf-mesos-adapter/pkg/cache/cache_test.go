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
	"testing"
)

type data struct {
	name      string
	namespace string
}

//errData for test err DataKey
type errData struct {
	name string
	age  int
}

func testDatakey(obj interface{}) (string, error) {
	d, ok := obj.(*data)
	if !ok {
		return "", fmt.Errorf("data convert failed")
	}
	return fmt.Sprintf("%s.%s", d.namespace, d.name), nil
}

var testCache Store

func init() {
	testCache = NewCache(testDatakey)
	//initial test data
	data1 := &data{namespace: "team", name: "jim"}
	data2 := &data{namespace: "team", name: "tom"}
	data3 := &data{namespace: "team", name: "kim"}
	testCache.Add(data1)
	testCache.Add(data2)
	testCache.Add(data3)
}

func TestCacheKeyFunc(t *testing.T) {
	errData := &errData{name: "jim", age: 100}
	err := testCache.Add(errData)
	if err != nil {
		t.Logf("KeyFunc testing success")
	}
}

func TestCacheAdd(t *testing.T) {
	tdata := &data{namespace: "team", name: "jack"}
	count := testCache.Num()
	//test no data
	_, exist, _ := testCache.Get(tdata)
	if exist {
		t.Error("Cache Get failed in cacheAdd")
	}
	testCache.Add(tdata)
	_, ok, _ := testCache.Get(tdata)
	if !ok {
		t.Error("Cache Add failed! Lost adding data")
	}
	num := testCache.Num()
	if count != num-1 {
		t.Errorf("Num error, Need: %d, Cur: %d", count+1, num)
	}
}

func TestCacheList(t *testing.T) {
	all := testCache.List()
	if len(all) != testCache.Num() {
		t.Error("List Num != Num()")
	}
}

func TestCacheGet(t *testing.T) {
	tdata := &data{namespace: "team", name: "jim"}
	_, exist, _ := testCache.Get(tdata)
	if !exist {
		t.Error("Get object from cache failed")
	}
}

func TestCacheDelete(t *testing.T) {
	tdata := &data{namespace: "team", name: "jim"}
	if err := testCache.Delete(tdata); err != nil {
		t.Error("Delete object from cache failed")
	}
	if err := testCache.Delete(tdata); err == nil {
		t.Error("delete err, data is expected Non-existent")
	}
}

func TestCacheClear(t *testing.T) {
	testCache.Clear()
	if 0 != testCache.Num() {
		t.Errorf("Clear Cache failed, need 0, current: %d", testCache.Num())
	}
	data5 := &data{namespace: "team", name: "jim"}
	testCache.Add(data5)
	_, exist, _ := testCache.Get(data5)
	if !exist {
		t.Error("Add Error after clear all")
	}
}
