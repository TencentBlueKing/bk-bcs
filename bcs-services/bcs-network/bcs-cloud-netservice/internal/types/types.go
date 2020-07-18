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

package types

const (
	// StateSubnetDisabled subnet is disabled
	StateSubnetDisabled = iota
	// StateSubnetEnabled subnet is enabled
	StateSubnetEnabled
)

const (
	// StatusIPActive ip is active
	StatusIPActive = "active"
	// StatusIPAvailable ip is available
	StatusIPAvailable = "available"
	// LeastIPNum least ip num in a available subnet
	LeastIPNum = 5
)

// CloudSubnet subnet on cloud
type CloudSubnet struct {
	SubnetID       string `json:"subnetID"`
	VpcID          string `json:"vpcID"`
	Region         string `json:"region"`
	Zone           string `json:"zone"`
	SubnetCidr     string `json:"subnetCidr"`
	AvailableIPNum int64  `json:"AvailableIPNum"`
	State          int32  `json:"state"`
	CreateTime     string `json:"createTime"`
	UpdateTime     string `json:"updateTime"`
}

// IPObject object for allocated ip
type IPObject struct {
	Address      string `json:"address"`
	VpcID        string `json:"vpcID"`
	Region       string `json:"region"`
	SubnetID     string `json:"subnetID"`
	SubnetCidr   string `json:"subnetCidr"`
	Cluster      string `json:"cluster"`
	Namespace    string `json:"namespace"`
	PodName      string `json:"podName"`
	WorkloadName string `json:"workloadName"`
	WorkloadKind string `json:"workloadKind"`
	ContainerID  string `json:"containerID"`
	Host         string `json:"host"`
	EniID        string `json:"eniID"`
	IsFixed      bool   `json:"isFixed"`
	Status       string `json:"status"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
}

// EniObject object for elastic network interface
type EniObject struct {
	Region   string `json:"region"`
	Zone     string `json:"zone"`
	SubnetID string `json:"subnetID"`
	VpcID    string `json:"vpcID"`
	EniID    string `json:"eniID"`
	EniName  string `json:"eniName"`
	MacAddr  string `json:"macAddr"`
}
