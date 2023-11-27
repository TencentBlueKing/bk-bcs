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

package utils

import (
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// CloudVpcSlice cloud vpc info
type CloudVpcSlice []*cmproto.CloudVPCResp

// Len len()
func (vpc CloudVpcSlice) Len() int {
	return len(vpc)
}

// Less less()
func (vpc CloudVpcSlice) Less(i, j int) bool {
	return vpc[i].AvailableIPNum > vpc[j].AvailableIPNum
}

// Swap swap()
func (vpc CloudVpcSlice) Swap(i, j int) {
	vpc[i], vpc[j] = vpc[j], vpc[i]
}

// NodeSlice cluster node slice
type NodeSlice []*cmproto.ClusterNode

// Len xxx
func (n NodeSlice) Len() int {
	return len(n)
}

// Less xxx
func (n NodeSlice) Less(i, j int) bool {
	return n[i].NodeName < n[j].NodeName
}

// Swap xxx
func (n NodeSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// NodeGroupSlice cluster nodeGroup slice
type NodeGroupSlice []*cmproto.NodeGroup

// Len xxx
func (n NodeGroupSlice) Len() int {
	return len(n)
}

// Less xxx
func (n NodeGroupSlice) Less(i, j int) bool {
	return n[i].NodeGroupID < n[j].NodeGroupID
}

// Swap xxx
func (n NodeGroupSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// ClusterSlice cluster slice
type ClusterSlice []*cmproto.Cluster

// Len xxx
func (n ClusterSlice) Len() int {
	return len(n)
}

// Less xxx
func (n ClusterSlice) Less(i, j int) bool {
	return n[i].ClusterName < n[j].ClusterName
}

// Swap xxx
func (n ClusterSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// InstanceTypeSlice instanceTypes slice
type InstanceTypeSlice []*cmproto.InstanceType

// Len xxx
func (n InstanceTypeSlice) Len() int {
	return len(n)
}

// Less xxx
func (n InstanceTypeSlice) Less(i, j int) bool {
	return n[i].UnitPrice < n[j].UnitPrice
}

// Swap xxx
func (n InstanceTypeSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// SubnetSlice subnet slice
type SubnetSlice []*cmproto.Subnet

// Len xxx
func (n SubnetSlice) Len() int {
	return len(n)
}

// Less xxx
func (n SubnetSlice) Less(i, j int) bool {
	return n[i].AvailableIPAddressCount > n[j].AvailableIPAddressCount
}

// Swap xxx
func (n SubnetSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
