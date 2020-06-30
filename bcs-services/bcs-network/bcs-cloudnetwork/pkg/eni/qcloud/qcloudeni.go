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

package qcloud

import (
	cloud "github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis/cloud/v1"
)

// Client qcloud client
type Client struct{}

// New create client
func New() *Client {
	return &Client{}
}

// Init client
func (c *Client) Init() error {
	return nil
}

// GetVMInfo get vm info
func (c *Client) GetVMInfo() (*cloud.VMInfo, error) {
	return nil, nil
}

// GetMaxENIIndex get max eni index
func (c *Client) GetMaxENIIndex() (int, error) {
	return 0, nil
}

// GetENILimit get eni limit
func (c *Client) GetENILimit() (eniNum, ipNum int, err error) {
	return 0, 0, nil
}

// CreateENI create eni
func (c *Client) CreateENI(name string, ipNum int) (*cloud.ElasticNetworkInterface, error) {
	return nil, nil
}

// AttachENI attach eni
func (c *Client) AttachENI(index int, eniID, instanceID, eniMac string) (*cloud.NetworkInterfaceAttachment, error) {
	return nil, nil
}

// DetachENI detach eni
func (c *Client) DetachENI(*cloud.NetworkInterfaceAttachment) error {
	return nil
}

// DeleteENI delete eni
func (c *Client) DeleteENI(eniID string) error {
	return nil
}
