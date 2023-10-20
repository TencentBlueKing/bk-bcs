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

package cidrtree

import (
	"fmt"
	"net"
	"testing"
)

func TestNewCidrManager(t *testing.T) {
	ss := []string{""}
	var subnets []*net.IPNet
	for _, s := range ss {
		_, subnet, err := net.ParseCIDR(s)
		if err != nil {
			continue
		}
		subnets = append(subnets, subnet)
	}
	_, cidrBlock, _ := net.ParseCIDR("")

	man := NewCidrManager(cidrBlock, subnets)
	fmt.Println(man.String())

	frees := man.GetFrees()
	fmt.Printf("frees: %v\n", frees)

	allocated := man.GetAllocated()
	for i := range allocated {
		fmt.Println("allocated", " ", allocated[i].String())
	}
}

func TestGetFrees(t *testing.T) {
	ss := []string{""}
	var subnets []*net.IPNet
	for _, s := range ss {
		_, subnet, err := net.ParseCIDR(s)
		if err != nil {
			continue
		}
		subnets = append(subnets, subnet)
	}
	_, cidrBlock, _ := net.ParseCIDR("")

	man := NewCidrManager(cidrBlock, subnets)
	fmt.Println(man.String())
	frees := man.GetFrees()
	fmt.Printf("frees: %v", frees)
}

func TestAllocate(t *testing.T) {
	ss := []string{""}
	var subnets []*net.IPNet
	for _, s := range ss {
		_, subnet, err := net.ParseCIDR(s)
		if err != nil {
			continue
		}
		subnets = append(subnets, subnet)
	}
	_, cidrBlock, _ := net.ParseCIDR("")

	man := NewCidrManager(cidrBlock, subnets)

	sub, err := man.Allocate(24)
	fmt.Printf("get: %v, error: %v\n", sub, err)
}
