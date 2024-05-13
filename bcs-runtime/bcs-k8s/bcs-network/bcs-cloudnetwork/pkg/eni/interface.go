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

package eni

import (
	cloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
)

// Interface interface for eni client
type Interface interface {
	Init() error
	// GetVMInfo get vm info
	GetVMInfo() (*cloud.VMInfo, error)
	// GetMaxENIIndex get max eni index
	GetMaxENIIndex() (int, error)
	// GetENILimit get eni limit
	GetENILimit() (eniNum, ipNum int, err error)
	// CreateENI create eni
	CreateENI(name string, ipNum int) (*cloud.ElasticNetworkInterface, error)
	// AttachENI attach eni
	AttachENI(index int, eniID, instanceID, eniMac string) (*cloud.NetworkInterfaceAttachment, error)
	// DetachENI detach eni
	DetachENI(*cloud.NetworkInterfaceAttachment) error
	// DeleteENI delete eni
	DeleteENI(eniID string) error
	// // ListENIs list enis of a vm
	// ListENIs(instanceID string) ([]*cloud.ElasticNetworkInterface, error)
	// AssignIPToENI(eniID string, ipNum int, ips []string) error
	// UnAssginIPFromENI(eniID string, ips []string) error
}
