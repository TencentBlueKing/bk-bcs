/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package extendedresource

import (
	"os"
	"sync"
	"testing"
)

func TestDriver(t *testing.T) {
	os.Remove("./test.db")
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		driver1 := &Driver{
			dataFilePath: "./test.db",
			lockerPath:   "./dblock",
		}
		if err := driver1.Lock(); err != nil {
			t.Errorf("lock err %s", err.Error())
			return
		}
		defer driver1.Unlock()

		if err := driver1.AddRecord("resource1", "container1", []string{"1", "2", "3", "4"}); err != nil {
			t.Errorf("err %s", err.Error())
		}
		if err := driver1.AddRecord("resource2", "container1", []string{"3", "4"}); err != nil {
			t.Errorf("err %s", err.Error())
		}
		if err := driver1.DelRecord("resource1", "container1"); err != nil {
			t.Errorf("err %s", err.Error())
		}
		tmpMap, err := driver1.ListRecordByResourceType("resource2")
		if err != nil {
			t.Errorf("err %s", err.Error())
		}
		t.Logf("%v", tmpMap)

	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		driver2 := &Driver{
			dataFilePath: "./test.db",
			lockerPath:   "./dblock",
		}
		driver2.Lock()
		defer driver2.Unlock()

		if err := driver2.AddRecord("resource1", "container2", []string{"1", "2", "3", "4"}); err != nil {
			t.Errorf("err %s", err.Error())
		}
		if err := driver2.AddRecord("resource2", "container2", []string{"1", "2"}); err != nil {
			t.Errorf("err %s", err.Error())
		}
		if err := driver2.DelRecord("resource1", "container2"); err != nil {
			t.Errorf("err %s", err.Error())
		}
		tmpMap, err := driver2.ListRecordByResourceType("resource2")
		if err != nil {
			t.Errorf("err %s", err.Error())
		}
		t.Logf("%v", tmpMap)
	}(&wg)
	wg.Wait()
	t.Error()
}
