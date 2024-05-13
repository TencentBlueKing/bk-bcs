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

// Package types is types for cloud netservice
package types

import (
	"time"
)

const (
	// SubnetStatusDisabled subnet is disabled
	SubnetStatusDisabled = iota
	// SubnetStatusEnabled subnet is enabled
	SubnetStatusEnabled
)

const (
	// IPStatusReserved ip is reserved
	IPStatusReserved = "reserved"
	// IPStatusENIPrimary ip is eni primary ip
	IPStatusENIPrimary = "eniprimary"
	// IPStatusActive ip is active
	IPStatusActive = "active"
	// IPStatusAvailable ip is available
	IPStatusAvailable = "available"
	// IPStatusFree ip is free
	IPStatusFree = "free"
	// IPStatusApplying ip is applying
	IPStatusApplying = "applying"
	// IPStatusDeleting ip is deleting
	IPStatusDeleting = "deleting"
	// SubnetLeastIPNum least ip num in a available subnet
	SubnetLeastIPNum = 5
)

// CloudSubnet subnet on cloud
type CloudSubnet struct {
	SubnetID       string `json:"subnetID"`
	VpcID          string `json:"vpcID"`
	Region         string `json:"region"`
	Zone           string `json:"zone"`
	SubnetCidr     string `json:"subnetCidr"`
	AvailableIPNum int64  `json:"AvailableIPNum"`
	MinIPNumPerEni int32  `json:"minIPNumPerEni"`
	State          int32  `json:"state"`
	CreateTime     string `json:"createTime"`
	UpdateTime     string `json:"updateTime"`
}

// IPObject object for allocated ip
type IPObject struct {
	ResourceVersion string    `json:"resourceVersion"`
	Address         string    `json:"address"`
	VpcID           string    `json:"vpcID"`
	Region          string    `json:"region"`
	SubnetID        string    `json:"subnetID"`
	SubnetCidr      string    `json:"subnetCidr"`
	Cluster         string    `json:"cluster"`
	Namespace       string    `json:"namespace"`
	PodName         string    `json:"podName"`
	WorkloadName    string    `json:"workloadName"`
	WorkloadKind    string    `json:"workloadKind"`
	ContainerID     string    `json:"containerID"`
	Host            string    `json:"host"`
	EniID           string    `json:"eniID"`
	IsFixed         bool      `json:"isFixed"`
	Status          string    `json:"status"`
	KeepDuration    string    `json:"keepDuration"`
	CreateTime      time.Time `json:"createTime"`
	UpdateTime      time.Time `json:"updateTime"`
}

// EniIPAddr object for ip
type EniIPAddr struct {
	IP        string `json:"ip"`
	IsPrimary bool   `json:"isPrimary"`
}

// EniObject object for elastic network interface
type EniObject struct {
	Region   string       `json:"region"`
	Zone     string       `json:"zone"`
	SubnetID string       `json:"subnetID"`
	VpcID    string       `json:"vpcID"`
	EniID    string       `json:"eniID"`
	EniName  string       `json:"eniName"`
	MacAddr  string       `json:"macAddr"`
	IPs      []*EniIPAddr `json:"ips,omitempty"`
}

// EniRecord eni record for store
type EniRecord struct {
	EniName       string `json:"eniName"`
	InstanceID    string `json:"instanceID"`
	Index         uint64 `json:"index"`
	EniSubnetID   string `json:"eniSubnetID"`
	EniSubnetCidr string `json:"eniSubnetCidr"`
	Region        string `json:"region"`
	Zone          string `json:"zone"`
}

// IPQuota cloud ip quota for certain cluster or namespace
type IPQuota struct {
	Cluster string `json:"cluster"`
	Limit   int64  `json:"limit"`
}
