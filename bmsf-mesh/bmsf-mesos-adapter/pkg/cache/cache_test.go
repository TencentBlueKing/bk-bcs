/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
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
