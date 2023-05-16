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

package randhostport

import (
	"container/heap"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// PortEntry entry of port
type PortEntry struct {
	Port     uint64
	Quantity uint64
	Index    int
}

// PortHeap min-heap to choose the smallest value
type PortHeap []*PortEntry

// Len implements golang heap interface
func (ph PortHeap) Len() int { return len(ph) }

// Less implements golang heap interface
func (ph PortHeap) Less(i, j int) bool { return ph[i].Quantity < ph[j].Quantity }

// Swap implements golang heap interface
func (ph PortHeap) Swap(i, j int) {
	ph[i], ph[j] = ph[j], ph[i]
	ph[i].Index = i
	ph[j].Index = j
}

// Push implements golang heap interface
func (ph *PortHeap) Push(x interface{}) {
	entry, ok := x.(*PortEntry)
	if !ok {
		blog.Errorf("PortHeap Push x %v is not PortEntry", x)
		return
	}
	entry.Index = len(*ph)
	*ph = append(*ph, entry)
}

// Pop implements golang heap interface
func (ph *PortHeap) Pop() interface{} {
	old := *ph
	n := len(old)
	if n == 0 {
		return nil
	}
	x := old[n-1]
	*ph = old[0 : n-1]
	return x
}

// PortCache cache for host port
type PortCache struct {
	h    *PortHeap
	m    map[uint64]*PortEntry
	lock sync.Mutex
}

// NewPortCache create port cache
func NewPortCache() *PortCache {
	h := &PortHeap{}
	heap.Init(h)
	m := make(map[uint64]*PortEntry)
	return &PortCache{
		h:    h,
		m:    m,
		lock: sync.Mutex{},
	}
}

// PushPortEntry add port with quantity
func (pc *PortCache) PushPortEntry(entry *PortEntry) {
	if oldEntry, ok := pc.m[entry.Port]; ok {
		oldEntry.Quantity = entry.Quantity
		heap.Fix(pc.h, oldEntry.Index)
		return
	}
	pc.m[entry.Port] = entry
	heap.Push(pc.h, entry)
}

// GetPortEntry get port entry by port
func (pc *PortCache) GetPortEntry(port uint64) *PortEntry {
	if entry, ok := pc.m[port]; ok {
		return &PortEntry{
			Port:     entry.Port,
			Quantity: entry.Quantity,
			Index:    entry.Index,
		}
	}
	return nil
}

// Lock lock cache
func (pc *PortCache) Lock() {
	pc.lock.Lock()
}

// Unlock unlock cache
func (pc *PortCache) Unlock() {
	pc.lock.Unlock()
}

// IncPortQuantity increase quantity for certain port
func (pc *PortCache) IncPortQuantity(port uint64) error {
	if entry, ok := pc.m[port]; ok {
		entry.Quantity = entry.Quantity + 1
		heap.Fix(pc.h, entry.Index)
	}
	return fmt.Errorf("entry for port %d not found when do increase", port)
}

// DecPortQuantity decrease quantity for certain port
func (pc *PortCache) DecPortQuantity(port uint64) error {
	if entry, ok := pc.m[port]; ok {
		entry.Quantity = entry.Quantity - 1
		heap.Fix(pc.h, entry.Index)
		return nil
	}
	return fmt.Errorf("entry for port %d not found when do descrease", port)
}

// PopPortEntry pop port entry
func (pc *PortCache) PopPortEntry() *PortEntry {
	el := heap.Pop(pc.h)
	if portEntry, ok := el.(*PortEntry); ok {
		delete(pc.m, portEntry.Port)
		return portEntry
	}
	return nil
}

// PopPortEntryByPort pop port entry by port
func (pc *PortCache) PopPortEntryByPort(port uint64) *PortEntry {
	if oldEntry, ok := pc.m[port]; ok {
		delete(pc.m, port)
		heap.Remove(pc.h, oldEntry.Index)
		return oldEntry
	}
	return nil
}
