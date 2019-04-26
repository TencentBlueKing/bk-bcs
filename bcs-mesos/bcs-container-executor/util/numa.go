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

// #cgo LDFLAGS: -lnuma
// #include <numa.h>
import "C"
import (
	"errors"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

//ErrNumaNotAvailable error type for NUMA
var ErrNumaNotAvailable = errors.New("No NUMA support available on this system")

//IsNumaAvailable Is Numa available
func IsNumaAvailable() bool {
	if int(C.numa_available()) >= 0 {
		return true
	}
	return false
}

//NUMAMaxNode Get max available node
func NUMAMaxNode() int {
	return int(C.numa_max_node())
}

//NUMAConfiguredCPUs get configure cpu number
func NUMAConfiguredCPUs() int {
	return int(C.numa_num_configured_cpus())
}

//NUMAMemoryOfNode Get Memory of the Numa node
func NUMAMemoryOfNode(node int) (inAll, free uint64) {
	cFree := C.longlong(0)
	cInAll := C.numa_node_size64(C.int(node), &cFree)
	return uint64(cInAll), uint64(cFree)
}

//MemInMB set mem in MB unit
func MemInMB(mem uint64) uint64 {
	return mem >> 20
}

//NUMANodes Get Numa Nodes
func NUMANodes() (nodes []int, err error) {
	if !IsNumaAvailable() {
		return nodes, ErrNumaNotAvailable
	}
	maxnode := int(C.numa_max_node())
	for i := 0; i <= maxnode; i++ {
		if C.numa_bitmask_isbitset(C.numa_nodes_ptr, C.uint(i)) > 0 {
			nodes = append(nodes, i)
		}
	}
	return nodes, nil
}

//NUMACPUsOfNode Get CPU slice from the specified Node
func NUMACPUsOfNode(node int) (cpus []int, err error) {
	if !IsNumaAvailable() {
		return cpus, ErrNumaNotAvailable
	}
	mask := C.numa_allocate_cpumask()
	defer C.numa_free_cpumask(mask)
	rc := C.numa_node_to_cpus(C.int(node), mask)
	maxCpus := NUMAConfiguredCPUs()
	if rc >= 0 {
		for i := 0; i < maxCpus; i++ {
			if C.numa_bitmask_isbitset(mask, C.uint(i)) > 0 {
				cpus = append(cpus, i)
			}
		}
	} else {
		return cpus, ErrNumaNotAvailable
	}
	return cpus, nil
}

//GetBindingCPUs get available binding cpu list
func GetBindingCPUs(cpus int, mem int64) (cpuList, numaList []int) {
	cpuNum := runtime.NumCPU()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if !IsNumaAvailable() {
		//No numa support, select cpu randomly
		return r.Perm(cpuNum)[:cpus], numaList
	}
	//only one numa node, randomly select
	nodeList, _ := NUMANodes()
	nodeNum := len(nodeList)
	if nodeNum == 1 {
		return r.Perm(cpuNum)[:cpus], numaList
	}
	//select one usable numa node simply
	//todo(developerJim): Get cpu list according memory size
	selectdNode := nodeList[r.Int()%nodeNum]
	nodeCPUs, _ := NUMACPUsOfNode(selectdNode)
	if cpus < len(nodeCPUs) {
		selectedIndex := r.Perm(len(nodeCPUs))[:cpus]
		for _, si := range selectedIndex {
			cpuList = append(cpuList, nodeCPUs[si])
		}
	} else {
		cpuList = append(cpuList, nodeCPUs...)
	}
	numaList = append(numaList, selectdNode)
	return cpuList, numaList
}

//ListJoin join list to one string
func ListJoin(l []int) string {
	if len(l) == 0 {
		return ""
	}
	lStr := strconv.Itoa(l[0])
	if len(l) == 1 {
		return lStr
	}
	for _, item := range l[1:] {
		lStr += "," + strconv.Itoa(item)
	}
	return lStr
}
