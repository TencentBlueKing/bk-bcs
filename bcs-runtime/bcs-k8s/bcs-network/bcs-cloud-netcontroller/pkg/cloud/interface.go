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

package cloud

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/internal/metric"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
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
	// QueryENI query eni
	QueryENI(eniID string) (*cloudv1.ElasticNetworkInterface, error)
	// CreateENI create eni
	CreateENI(name, subnetID, addr string, ipNum int) (*cloudv1.ElasticNetworkInterface, error)
	// AttachENI attach eni
	AttachENI(index int, eniID, instanceID, eniMac string) (*cloudv1.NetworkInterfaceAttachment, error)
	// DetachENI detach eni
	DetachENI(*cloudv1.NetworkInterfaceAttachment) error
	// DeleteENI delete eni
	DeleteENI(eniID string) error
}

// CloudWithMetric cloud client with metric
type CloudWithMetric struct {
	cloud Interface
}

// NewCloudWithMetic create new cloud with metric
func NewCloudWithMetic(cloud Interface) *CloudWithMetric {
	return &CloudWithMetric{
		cloud: cloud,
	}
}

// Init implements Cloud interface
func (cwm *CloudWithMetric) Init() error {
	return cwm.cloud.Init()
}

// GetVMInfo get vm info
func (cwm *CloudWithMetric) GetVMInfo(instanceIP string) (*cloudv1.VMInfo, error) {
	start := time.Now()
	mf := func(result string) {
		metric.StatClientRequest("cloud", "GetVMInfo", 0, result, start, time.Now())
	}
	vmInfo, err := cwm.cloud.GetVMInfo(instanceIP)
	if err != nil {
		mf(metric.RequestResultFailed)
		return nil, err
	}
	mf(metric.RequestResultSuccess)
	return vmInfo, nil
}

// GetMaxENIIndex get max eni index
func (cwm *CloudWithMetric) GetMaxENIIndex(instanceIP string) (int, error) {
	return cwm.cloud.GetMaxENIIndex(instanceIP)
}

// GetENILimit get eni limit
func (cwm *CloudWithMetric) GetENILimit(instanceIP string) (eniNum, ipNum int, err error) {
	return cwm.cloud.GetENILimit(instanceIP)
}

// QueryENI query eni
func (cwm *CloudWithMetric) QueryENI(eniID string) (*cloudv1.ElasticNetworkInterface, error) {
	start := time.Now()
	mf := func(result string) {
		metric.StatClientRequest("cloud", "QueryENI", 0, result, start, time.Now())
	}
	interf, err := cwm.cloud.QueryENI(eniID)
	if err != nil {
		mf(metric.RequestResultFailed)
		return nil, err
	}
	mf(metric.RequestResultSuccess)
	return interf, nil
}

// CreateENI create eni
func (cwm *CloudWithMetric) CreateENI(
	name, subnetID, addr string, ipNum int) (*cloudv1.ElasticNetworkInterface, error) {
	start := time.Now()
	mf := func(result string) {
		metric.StatClientRequest("cloud", "CreateENI", 0, result, start, time.Now())
	}
	interf, err := cwm.cloud.CreateENI(name, subnetID, addr, ipNum)
	if err != nil {
		mf(metric.RequestResultFailed)
		return nil, err
	}
	mf(metric.RequestResultSuccess)
	return interf, nil
}

// AttachENI attach eni
func (cwm *CloudWithMetric) AttachENI(
	index int, eniID, instanceID, eniMac string) (*cloudv1.NetworkInterfaceAttachment, error) {
	start := time.Now()
	mf := func(result string) {
		metric.StatClientRequest("cloud", "AttachENI", 0, result, start, time.Now())
	}
	attachment, err := cwm.cloud.AttachENI(index, eniID, instanceID, eniMac)
	if err != nil {
		mf(metric.RequestResultFailed)
		return nil, err
	}
	mf(metric.RequestResultSuccess)
	return attachment, nil
}

// DetachENI detach eni
func (cwm *CloudWithMetric) DetachENI(attachment *cloudv1.NetworkInterfaceAttachment) error {
	start := time.Now()
	mf := func(result string) {
		metric.StatClientRequest("cloud", "DetachENI", 0, result, start, time.Now())
	}
	err := cwm.cloud.DetachENI(attachment)
	if err != nil {
		mf(metric.RequestResultFailed)
		return err
	}
	mf(metric.RequestResultSuccess)
	return nil
}

// DeleteENI delete eni
func (cwm *CloudWithMetric) DeleteENI(eniID string) error {
	start := time.Now()
	mf := func(result string) {
		metric.StatClientRequest("cloud", "DeleteENI", 0, result, start, time.Now())
	}
	err := cwm.cloud.DeleteENI(eniID)
	if err != nil {
		mf(metric.RequestResultFailed)
		return err
	}
	mf(metric.RequestResultSuccess)
	return nil
}
