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

package randhostport

import (
	"container/heap"
	"math/rand"
	"testing"
)

// TestPortHeap test PortHeap
func TestPortHeap(t *testing.T) {
	h := &PortHeap{
		&PortEntry{
			Port:     31000,
			Quantity: 20,
		},
		&PortEntry{
			Port:     31001,
			Quantity: 100,
		},
	}
	heap.Init(h)
	heap.Push(h, &PortEntry{
		Port:     31002,
		Quantity: 50,
	})
	heap.Push(h, &PortEntry{
		Port:     31003,
		Quantity: 88,
	})
	// test heap Pop function
	entry := heap.Pop(h)
	portEntry, ok := entry.(*PortEntry)
	if !ok {
		t.Errorf("entry %v is not *PortEntry", entry)
	}
	if portEntry.Port != 31000 || portEntry.Quantity != 20 {
		t.Errorf("expect port %d quantity %d, but get port %d, quantiry %d",
			31000, 20, portEntry.Port, portEntry.Quantity)
	}
	for index, e := range *h {
		if index != e.Index {
			t.Errorf("index %d is not %d", index, e.Index)
		}
	}
}

// TestPortCache test port cache
func TestPortCache(t *testing.T) {
	pc := NewPortCache()
	portEntryList := []*PortEntry{
		{
			Port:     10000,
			Quantity: 100,
		},
		{
			Port:     10001,
			Quantity: 99,
		},
		{
			Port:     10002,
			Quantity: 120,
		},
	}
	orderedList := []*PortEntry{
		{
			Port:     10001,
			Quantity: 99,
			Index:    0,
		},
		{
			Port:     10000,
			Quantity: 100,
			Index:    1,
		},

		{
			Port:     10002,
			Quantity: 120,
			Index:    2,
		},
	}
	for _, e := range portEntryList {
		pc.PushPortEntry(e)
	}
	for i := 0; i < len(portEntryList); i++ {
		tmp := (*pc.h)[i]
		if tmp.Port != orderedList[i].Port ||
			tmp.Index != orderedList[i].Index ||
			tmp.Quantity != orderedList[i].Quantity {
			t.Errorf("expect %v but get %v", orderedList[i], tmp)
		}
	}

	testList := make([]*PortEntry, 0)
	for i := 0; i < 2000; i++ {
		port := rand.Intn(3000) + 32000
		testList = append(testList, &PortEntry{
			Port:     uint64(port),
			Quantity: uint64(rand.Intn(1000)),
		})
	}
	for i := 0; i < 1000; i++ {
		index := i % len(testList)
		pc.PushPortEntry(testList[index])
	}
	// pop some entry by port
	for i := 0; i < 100; i++ {
		port := rand.Intn(3000) + 32000
		entry := pc.PopPortEntryByPort(uint64(port))
		t.Logf("pop by port %d, entry %+v", port, entry)
	}
	minQuantity := uint64(0)
	length := len(*pc.h)
	for i := 0; i < length; i++ {
		entry := pc.PopPortEntry()
		if entry == nil {
			t.Errorf("entry is nil")
		}
		if entry.Quantity < minQuantity {
			t.Errorf("entry quantity should be bigger than %d", minQuantity)
		}
		t.Logf("round: %d, get entry %+v", i, entry)
		minQuantity = entry.Quantity
	}
}

// BenchmarkPortCache test PortCache performance
func BenchmarkPortCache(b *testing.B) {
	testList := make([]*PortEntry, 0)
	for i := 0; i < 2000; i++ {
		port := rand.Intn(3000) + 32000
		testList = append(testList, &PortEntry{
			Port:     uint64(port),
			Quantity: uint64(rand.Intn(1000)),
		})
	}
	pc := NewPortCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := i % len(testList)
		pc.PushPortEntry(testList[index])
	}
}
