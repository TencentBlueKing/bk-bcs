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

package controller

import (
	"math"
	"math/rand"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

// balancer try to partition specified number to N units.
// it designs for allocating resources into different elastic nodegroups.
type balancer interface {
	distribute(num int) []*nodeGroup
}

type nodeGroup struct {
	storage.GroupInfo
	// partition
	partition int
	// limitation max limit for allocatedNum
	limitation int
}

func newSimpleBalancer(groups []*storage.GroupInfo) balancer {
	// sort nodegroup according their weights
	sort.SliceStable(groups, func(i int, j int) bool {
		return groups[i].Weight < groups[j].Weight
	})

	nodes := make([]*nodeGroup, len(groups))
	max := 0
	for i, group := range groups {
		n := &nodeGroup{
			GroupInfo: *group,
			partition: 0,
		}
		nodes[i] = n
		max += group.Weight
	}
	return &simpleBalancer{
		nodes: nodes,
		max:   max,
	}
}

// simpleBalancer just allocates resources into N(nodes length) units simply.
// convert result to intege by math.Floor() if results are float.
// simpleBalancer only designs for nodegroup scaleup operation.
type simpleBalancer struct {
	// nodes are in ascending order
	nodes []*nodeGroup
	max   int
}

func (s *simpleBalancer) distribute(n int) []*nodeGroup {
	total := 0
	distn := float64(n)
	for _, node := range s.nodes {
		node.partition = int(math.Floor(distn * float64(node.Weight) / float64(s.max)))
		total += node.partition
	}
	// add left resource to max weight node simply
	left := n - total
	if left > 0 {
		s.nodes[len(s.nodes)-1].partition += left
	}
	return s.nodes
}

func newWeightBalancer(groups []*storage.GroupInfo, nodegroups map[string]*storage.NodeGroup) balancer {
	// sort slice
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].Weight < groups[j].Weight
	})
	ruler := make([]int, len(groups))
	nodes := make([]*nodeGroup, len(groups))
	max := 0
	for i, group := range groups {
		// get minSize for limitation
		nodegroup := nodegroups[group.NodeGroupID]
		n := &nodeGroup{
			GroupInfo: *group,
			partition: 0,
			// !controller can only scaleDown such resources.
			// !sometimes nodegroup desiredSize is larger than exist nodes(last status is scaleUp),
			// !if controller scaleDown nodes larger than exist nodes that according to desiredSize,
			// !cluster-autoscaler may be panic, because upComing nodes are still not affective.
			limitation: len(nodegroup.NodeIPs) - nodegroup.MinSize,
		}
		max += n.Weight
		ruler[i] = max
		nodes[i] = n
	}
	balance := &weightBalancer{
		nodes: nodes,
		ruler: ruler,
		max:   max,
	}
	return balance
}

// weightBalancer allocates resources with weight in random mode.
// when nodegroup scales down, its resource may be not enough for releasing.
// so scaledown operation is not balance between all elastic nodegroups,
// controller had to release more node from other specified nodegroups.
type weightBalancer struct {
	nodes []*nodeGroup
	ruler []int
	max   int
}

func (balance *weightBalancer) distribute(n int) []*nodeGroup {
	totalLimit := 0
	for _, node := range balance.nodes {
		totalLimit += node.limitation
	}
	// totalLimit is all that can allocate
	if totalLimit >= n {
		totalLimit = n
	}
	for {
		if totalLimit < 1 {
			// all scaleDown resources are partitioned into nodeGroups
			break
		}
		selected := rand.Intn(balance.max) + 1
		index := sort.SearchInts(balance.ruler, selected)
		node := balance.nodes[index]
		if node.partition+1 > node.limitation {
			// nodegroup scale down resource reach limitation
			// skip assignment
			continue
		}
		node.partition++
		totalLimit--
	}
	return balance.nodes
}
