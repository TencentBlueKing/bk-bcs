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

package cloud

import (
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
)

// Interface interface for eni client
type Interface interface {
	// Init do init
	Init() error
	// GetVMInfo get vm info
	GetVMInfo(instanceIP string) (*cloudv1.VMInfo, error)
	// GetMaxENIIndex get max eni index
	GetMaxENIIndex(instanceIP string) (int, error)
	// GetENILimit get eni limit
	GetENILimit(instanceIP string) (eniNum, ipNum int, err error)
	// CreateENI create eni
	CreateENI(name, subnetID string, ipNum int) (*cloudv1.ElasticNetworkInterface, error)
	// AttachENI attach eni
	AttachENI(index int, eniID, instanceID, eniMac string) (*cloudv1.NetworkInterfaceAttachment, error)
	// DetachENI detach eni
	DetachENI(*cloudv1.NetworkInterfaceAttachment) error
	// DeleteENI delete eni
	DeleteENI(eniID string) error
}
