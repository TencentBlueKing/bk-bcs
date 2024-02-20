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

// Package cidrtree xxx
package cidrtree

import (
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/pkg/errors"
)

// ErrNoEnoughFreeSubnet xxx
var ErrNoEnoughFreeSubnet = errors.New("no enough free subnet")

// Mananger define interface to manage cidr
type Mananger interface {
	String() string
	GetFrees() []*net.IPNet
	GetAllocated() []*net.IPNet
	Allocate(mask int) (*net.IPNet, error)
}

// NewCidrManager create cidrTree that implement Manager interface
func NewCidrManager(cidrBlock *net.IPNet, subnets []*net.IPNet) Mananger {
	tree := New(cidrBlock)
	for _, subnet := range subnets {
		tree = Insert(tree, subnet)
	}
	return tree
}

const (
	// NODE_UNUSED unused
	NODE_UNUSED byte = 0
	// NODE_USED use
	NODE_USED byte = 1
	// NODE_SPLIT split
	NODE_SPLIT byte = 2
	// NODE_FULL fill
	NODE_FULL byte = 3 // nolint
)

type node struct {
	*net.IPNet
	status byte
}

// IPNetEqual reports whether node's IPNet and x are the same IPNet.
func (n *node) IPNetEqual(x *net.IPNet) bool {
	nOnes, _ := n.IPNet.Mask.Size()
	iOnes, _ := x.Mask.Size()
	if n.IPNet.IP.Equal(x.IP) && nOnes == iOnes {
		return true
	}
	return false
}

// MaskSize get node ipnet's mask size
func (n *node) MaskSize() int {
	ones, _ := n.IPNet.Mask.Size()
	return ones
}

// IPNetContains report whether node's IPNet contains x
func (n *node) IPNetContains(x *net.IPNet) bool {
	first, last := cidr.AddressRange(x)
	if n.IPNet.Contains(first) && n.IPNet.Contains(last) {
		return true
	}
	return false
}

// SetStatus set node status
func (n *node) SetStatus(status byte) {
	n.status = status
}

type cidrTree struct {
	Left  *cidrTree
	Value *node
	Right *cidrTree
}

// New create a new cidrTree
// nolint
func New(cidrBlock *net.IPNet) *cidrTree {
	return &cidrTree{nil, &node{cidrBlock, NODE_UNUSED}, nil}
}

// Insert node to cidrTree
// nolint
func Insert(t *cidrTree, subnet *net.IPNet) *cidrTree {
	if cidr.VerifyNoOverlap([]*net.IPNet{subnet}, t.Value.IPNet) != nil {
		return t
	}
	if t.Value.IPNetEqual(subnet) {
		t.Value.SetStatus(NODE_USED)
		return t
	}
	if t.Left == nil {
		left, err := cidr.Subnet(t.Value.IPNet, 1, 0)
		if err != nil {
			return t
		}
		t.Left = &cidrTree{nil, &node{left, NODE_UNUSED}, nil}
		t.Value.SetStatus(NODE_SPLIT)
	}
	if t.Right == nil {
		right, err := cidr.Subnet(t.Value.IPNet, 1, 1)
		if err != nil {
			return t
		}
		t.Right = &cidrTree{nil, &node{right, NODE_UNUSED}, nil}
		t.Value.SetStatus(NODE_SPLIT)
	}
	if t.Left.Value.IPNetContains(subnet) {
		t.Left = Insert(t.Left, subnet)
	} else {
		t.Right = Insert(t.Right, subnet)
	}
	return t
}

// Walk traverse the tree
func Walk(t *cidrTree, ch chan *node) {
	defer close(ch)
	var walk func(t *cidrTree)
	walk = func(t *cidrTree) {
		if t == nil {
			return
		}
		walk(t.Left)
		ch <- t.Value
		walk(t.Right)
	}
	walk(t)
}

// String printf the tree
func (t *cidrTree) String() string {
	if t == nil {
		return "()"
	}
	s := ""
	if t.Left != nil {
		s += t.Left.String() + " "
	}
	s += fmt.Sprint(t.Value.IPNet.String())
	if t.Right != nil {
		s += " " + t.Right.String()
	}
	return "(" + s + ")"
}

// GetAllocated return used nodes
func (t *cidrTree) GetAllocated() []*net.IPNet {
	ch := make(chan *node)
	var allocated []*net.IPNet
	go Walk(t, ch)
	for node := range ch {
		if node.status == NODE_USED {
			allocated = append(allocated, node.IPNet)
		}
	}
	return allocated
}

// GetFrees return unused nodes
func (t *cidrTree) GetFrees() []*net.IPNet {
	ch := make(chan *node)
	var frees []*net.IPNet
	go Walk(t, ch)
	for node := range ch {
		if node.status == NODE_UNUSED {
			frees = append(frees, node.IPNet)
		}
	}
	return frees
}

// Allocate a node with mask
func (t *cidrTree) Allocate(mask int) (*net.IPNet, error) {
	ch := make(chan *node)
	var toSplit *node
	go Walk(t, ch)
	for node := range ch {
		if node.status == NODE_UNUSED {
			if node.MaskSize() == mask {
				return node.IPNet, nil
			}
			if node.MaskSize() > mask {
				continue
			}
			if node.MaskSize() < mask {
				if toSplit == nil || toSplit.MaskSize() < node.MaskSize() {
					toSplit = node
				}
			}
		}
	}
	if toSplit == nil {
		return nil, ErrNoEnoughFreeSubnet
	}
	sub, err := cidr.Subnet(toSplit.IPNet, mask-toSplit.MaskSize(), 0)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// IsIPnetEqual report whether two IPNet equal
func IsIPnetEqual(a, b *net.IPNet) bool {
	if a == nil || b == nil {
		return false
	}
	nOnes, _ := a.Mask.Size()
	iOnes, _ := b.Mask.Size()
	if a.IP.Equal(b.IP) && nOnes == iOnes {
		return true
	}
	return false
}

// IPNetMaskSize return IPNet mask size
func IPNetMaskSize(a *net.IPNet) int {
	ones, _ := a.Mask.Size()
	return ones
}

// AllocateFromFrees allocate IPNet of mask from free IPNets
func AllocateFromFrees(mask int, frees []*net.IPNet) (*net.IPNet, error) {
	var toSplit *net.IPNet
	for _, free := range frees {
		if IPNetMaskSize(free) == mask {
			return free, nil
		}
		if IPNetMaskSize(free) > mask {
			continue
		}
		if IPNetMaskSize(free) < mask {
			if toSplit == nil || IPNetMaskSize(toSplit) < IPNetMaskSize(free) {
				toSplit = free
			}
		}
	}
	if toSplit == nil {
		return nil, ErrNoEnoughFreeSubnet
	}
	sub, err := cidr.Subnet(toSplit, mask-IPNetMaskSize(toSplit), 0)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
