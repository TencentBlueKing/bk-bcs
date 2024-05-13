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

// Package aws is implementation for aws cloud
package aws

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
)

// Client client for aws
type Client struct {
}

// NewClient create new client
func NewClient() (*Client, error) {
	return &Client{}, nil
}

// DescribeSubnet describe subnet
func (c *Client) DescribeSubnet(vpcID, region, subnetID string) (*types.CloudSubnet, error) {
	return nil, nil
}

// DescribeSubnetList describe subnet list
func (c *Client) DescribeSubnetList(vpcID, region string, subnetIDs []string) ([]*types.CloudSubnet, error) {
	return nil, nil
}

// QueryEni query eni
func (c *Client) QueryEni(eniID string) (*types.EniObject, error) {
	return nil, nil
}

// QueryEniList query eni list
func (c *Client) QueryEniList(subnetID string) ([]*types.EniObject, error) {
	return nil, nil
}

// AssignIPToEni assign ip to eni
func (c *Client) AssignIPToEni(ip, eniID string) (string, error) {
	return "", nil
}

// UnassignIPFromEni unassign ip from eni
func (c *Client) UnassignIPFromEni(ip []string, eniID string) error {
	return nil
}

// MigrateIP migrate ip
func (c *Client) MigrateIP(ip, srcEniID, destEniID string) error {
	return nil
}

// GetVMInfo get vm info
func (c *Client) GetVMInfo(instanceIP string) (*cloudv1.VMInfo, error) {
	return nil, nil
}
