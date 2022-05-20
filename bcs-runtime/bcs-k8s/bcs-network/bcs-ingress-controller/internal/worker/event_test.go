/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package worker

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// TestListenerEvent test listener event
func TestListenerEvent(t *testing.T) {
	var arr []ListenerEvent
	var wg sync.WaitGroup
	lock := &sync.Mutex{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			name := "name" + strconv.Itoa(rand.Intn(100))
			ns := "ns" + strconv.Itoa(rand.Intn(5))
			newEvent := NewListenerEvent(
				EventAdd,
				name,
				ns,
				&networkextensionv1.Listener{})
			if newEvent.Key() != ns+"/"+name {
				t.Errorf("failed")
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			lock.Lock()
			arr = append(arr, *newEvent)
			lock.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()

	sort.Sort(ListenerEventList(arr))

	for i := 0; i < 9; i++ {
		if arr[i].EventTime.After(arr[i+1].EventTime) {
			t.Errorf("failed")
		}
	}
}
