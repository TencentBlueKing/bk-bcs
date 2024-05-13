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

// Package tencentcloud is is implementation for tencentcloud cloud
package tencentcloud

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/metric"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// Client client for tencentcloud
type Client struct {
	// tencentcloud vpc domain
	vpcDomain string
	// tencentcloud cvm domain
	cvmDomain string
	// tencentcloud region
	region string
	// secret id for tencent cloud
	secretID string
	// secret key for tencent cloud
	secretKey string
	// security group ids
	securityGroups []string
	// vpcClient client for tencent cloud vpc
	vpcClient *vpc.Client
	// cvmClient client for tencent cloud cvm
	cvmClient *cvm.Client
}

// NewClient create new client
func NewClient() (*Client, error) {
	c := &Client{}
	err := c.loadEnv()
	if err != nil {
		return nil, err
	}
	// create new credential
	credential := common.NewCredential(
		c.secretID,
		c.secretKey,
	)
	// create new client profile
	cpf := profile.NewClientProfile()
	if len(c.vpcDomain) != 0 {
		cpf.HttpProfile.Endpoint = c.vpcDomain
	}
	// create new vpc client
	vpcClient, err := vpc.NewClient(credential, c.region, cpf)
	if err != nil {
		blog.Errorf("new vpc client failed, err %s", err.Error())
		return nil, fmt.Errorf("new vpc client failed, err %s", err.Error())
	}
	c.vpcClient = vpcClient
	// create new cvm client
	cvmClient, err := cvm.NewClient(credential, c.region, cpf)
	if err != nil {
		blog.Errorf("new cvm client failed, err %s", err.Error())
		return nil, fmt.Errorf("new cvm client failed, err %s", err.Error())
	}
	c.cvmClient = cvmClient
	// return client
	return c, nil
}

func (c *Client) loadEnv() error {
	// load tencent cloud vpc domain
	c.vpcDomain = os.Getenv(EnvNameTencentCloudVpcDomain)
	// load tencent cloud cvm domain
	c.cvmDomain = os.Getenv(EnvNameTencentCloudCvmDomain)
	// load tencent cloud region
	c.region = os.Getenv(EnvNameTencentCloudRegion)
	// load tencent cloud secret id
	c.secretID = os.Getenv(EnvNameTencentCloudAccessKeyID)
	// load tencent cloud secret key
	secretKey := os.Getenv(EnvNameTencentCloudAccessKey)

	// security groups
	sGroupsStr := os.Getenv(EnvNameTencentCloudSecurityGroups)
	if len(sGroupsStr) != 0 {
		strings.Replace(sGroupsStr, ";", ",", -1)
		sGroups := strings.Split(sGroupsStr, ",")
		c.securityGroups = sGroups
	}

	// decrypt secret key
	decryptSecretKey, err := encrypt.DesDecryptFromBase([]byte(secretKey))
	if err != nil {
		blog.Errorf("descrpt access secret key failed, err %s", err.Error())
		return fmt.Errorf("descrpt access secret key failed, err %s", err.Error())
	}
	c.secretKey = string(decryptSecretKey)
	return nil
}

// DescribeSubnet describe subnet
func (c *Client) DescribeSubnet(vpcID, region, subnetID string) (*types.CloudSubnet, error) {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "DescribeSubnet", result, startTime, time.Now())
	}
	tSubnets, err := c.describeSubnets([]string{subnetID})
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		return nil, err
	}
	// expect just one subnet
	if len(tSubnets) != 1 {
		mf(metric.CloudOperationResultFailed)
		return nil, fmt.Errorf("describeSubnet expect 1 subnet")
	}
	tSubnet := tSubnets[0]

	if vpcID != *tSubnet.VpcId || region != c.region {
		mf(metric.CloudOperationResultFailed)
		blog.Infof("vpcID or region of subnet does not match, expect (%s, %s), get (%s, %s)",
			vpcID, region, *tSubnet.VpcId, c.region)
		return nil, fmt.Errorf("vpcID or region of subnet does not match, expect (%s, %s), get (%s, %s)",
			vpcID, region, *tSubnet.VpcId, c.region)
	}

	mf(metric.CloudOperationResultSuccess)
	return &types.CloudSubnet{
		SubnetID:       subnetID,
		SubnetCidr:     *tSubnet.CidrBlock,
		VpcID:          *tSubnet.VpcId,
		Region:         region,
		Zone:           *tSubnet.Zone,
		AvailableIPNum: int64(*tSubnet.AvailableIpAddressCount),
	}, nil
}

// DescribeSubnetList describe subnet list
func (c *Client) DescribeSubnetList(vpcID, region string, subnetIDs []string) ([]*types.CloudSubnet, error) {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "DescribeSubnetList", result, startTime, time.Now())
	}
	tSubnets, err := c.describeSubnets(subnetIDs)
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		return nil, err
	}

	var cloudSubnets []*types.CloudSubnet
	for _, tSubnet := range tSubnets {
		if vpcID != *tSubnet.VpcId || region != c.region {
			mf(metric.CloudOperationResultFailed)
			blog.Infof("vpcID or region of subnet does not match, expect (%s, %s), get (%s, %s)",
				vpcID, region, *tSubnet.VpcId, c.region)
			return nil, fmt.Errorf("vpcID or region of subnet does not match, expect (%s, %s), get (%s, %s)",
				vpcID, region, *tSubnet.VpcId, c.region)
		}
		cloudSubnets = append(cloudSubnets, &types.CloudSubnet{
			SubnetID:       *tSubnet.SubnetId,
			SubnetCidr:     *tSubnet.CidrBlock,
			VpcID:          *tSubnet.VpcId,
			Region:         region,
			Zone:           *tSubnet.Zone,
			AvailableIPNum: int64(*tSubnet.AvailableIpAddressCount),
		})
	}
	if len(cloudSubnets) == 0 {
		mf(metric.CloudOperationResultFailed)
		return nil, fmt.Errorf("subnets %+v not found", subnetIDs)
	}
	mf(metric.CloudOperationResultSuccess)
	return cloudSubnets, nil
}

// QueryEni query eni
func (c *Client) QueryEni(eniID string) (*types.EniObject, error) {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "QueryEni", result, startTime, time.Now())
	}

	req := vpc.NewDescribeNetworkInterfacesRequest()
	req.NetworkInterfaceIds = common.StringPtrs([]string{eniID})

	blog.V(3).Infof("DescribeNetworkInterfaces req: %s", req.ToJsonString())

	resp, err := c.vpcClient.DescribeNetworkInterfaces(req)
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		blog.Errorf("DescribeNetworkInterfaces failed, err %s", err.Error())
		return nil, fmt.Errorf("DescribeNetworkInterfaces failed, err %s", err.Error())
	}

	blog.V(3).Infof("DescribeNetworkInterfaces resp: %s", resp.ToJsonString())

	if len(resp.Response.NetworkInterfaceSet) != 1 {
		mf(metric.CloudOperationResultFailed)
		blog.Errorf(
			"DescribeNetworkInterfaces expect 1 eni, but get %d", len(resp.Response.NetworkInterfaceSet))
		return nil, fmt.Errorf(
			"DescribeNetworkInterfaces expect 1 eni, but get %d", len(resp.Response.NetworkInterfaceSet))
	}

	retEni := resp.Response.NetworkInterfaceSet[0]

	defer mf(metric.CloudOperationResultSuccess)
	return &types.EniObject{
		Region:   c.region,
		Zone:     *retEni.Zone,
		SubnetID: *retEni.SubnetId,
		VpcID:    *retEni.VpcId,
		EniID:    *retEni.NetworkInterfaceId,
		EniName:  *retEni.NetworkInterfaceName,
		MacAddr:  *retEni.MacAddress,
	}, nil
}

// QueryEniList query eni list
func (c *Client) QueryEniList(subnetID string) ([]*types.EniObject, error) {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "QueryEniList", result, startTime, time.Now())
	}
	var eniList []*types.EniObject
	limit := uint64(100)
	total, eniSet, err := c.queryENI(subnetID, "", "", "", 0, limit)
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		return nil, err
	}
	for _, retEni := range eniSet {
		newEniObject := &types.EniObject{
			Region:   c.region,
			Zone:     *retEni.Zone,
			SubnetID: *retEni.SubnetId,
			VpcID:    *retEni.VpcId,
			EniID:    *retEni.NetworkInterfaceId,
			EniName:  *retEni.NetworkInterfaceName,
			MacAddr:  *retEni.MacAddress,
		}
		for _, retIP := range retEni.PrivateIpAddressSet {
			newEniObject.IPs = append(newEniObject.IPs, &types.EniIPAddr{
				IP:        *retIP.PrivateIpAddress,
				IsPrimary: *retIP.Primary,
			})
		}
		eniList = append(eniList, newEniObject)
	}
	if total > 100 {
		for index := limit; index < total; index = index + 100 {
			_, eniSet, err := c.queryENI(subnetID, "", "", "", index, limit)
			if err != nil {
				mf(metric.CloudOperationResultFailed)
				return nil, err
			}
			for _, retEni := range eniSet {
				newEniObject := &types.EniObject{
					Region:   c.region,
					Zone:     *retEni.Zone,
					SubnetID: *retEni.SubnetId,
					VpcID:    *retEni.VpcId,
					EniID:    *retEni.NetworkInterfaceId,
					EniName:  *retEni.NetworkInterfaceName,
					MacAddr:  *retEni.MacAddress,
				}
				for _, retIP := range retEni.PrivateIpAddressSet {
					newEniObject.IPs = append(newEniObject.IPs, &types.EniIPAddr{
						IP:        *retIP.PrivateIpAddress,
						IsPrimary: *retIP.Primary,
					})
				}
				eniList = append(eniList, newEniObject)
			}
		}
	}
	mf(metric.CloudOperationResultSuccess)
	return eniList, nil
}

// AssignIPToEni assign ip to eni
func (c *Client) AssignIPToEni(ip, eniID string) (string, error) {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "AssignIPToEni", result, startTime, time.Now())
	}
	req := vpc.NewAssignPrivateIpAddressesRequest()
	req.NetworkInterfaceId = common.StringPtr(eniID)
	if len(ip) != 0 {
		req.PrivateIpAddresses = append(req.PrivateIpAddresses, &vpc.PrivateIpAddressSpecification{
			PrivateIpAddress: common.StringPtr(ip),
		})
	} else {
		req.SecondaryPrivateIpAddressCount = common.Uint64Ptr(1)
	}

	blog.V(3).Infof("AssignPrivateIpAddresses req: %s", req.ToJsonString())

	resp, err := c.vpcClient.AssignPrivateIpAddresses(req)
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		blog.Errorf("AssignPrivateIpAddresses failed, err %s", err.Error())
		return "", fmt.Errorf("AssignPrivateIpAddresses failed, err %s", err.Error())
	}

	blog.V(3).Infof("AssignPrivateIpAddresses resp: %s", resp.ToJsonString())

	if len(resp.Response.PrivateIpAddressSet) != 1 {
		mf(metric.CloudOperationResultFailed)
		blog.Errorf("AssignPrivateIpAddresses expect 1 ip, but get %d", len(resp.Response.PrivateIpAddressSet))
		return "", fmt.Errorf("AssignPrivateIpAddresses expect 1 ip, but get %d", len(resp.Response.PrivateIpAddressSet))
	}

	mf(metric.CloudOperationResultSuccess)
	return *resp.Response.PrivateIpAddressSet[0].PrivateIpAddress, nil

}

// UnassignIPFromEni unassign ip from eni
func (c *Client) UnassignIPFromEni(ips []string, eniID string) error {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "UnassignIPFromEni", result, startTime, time.Now())
	}
	req := vpc.NewUnassignPrivateIpAddressesRequest()
	req.NetworkInterfaceId = common.StringPtr(eniID)
	if len(ips) == 0 {
		mf(metric.CloudOperationResultFailed)
		return fmt.Errorf("ips to be unassigned cannot be empty")
	}
	for _, ip := range ips {
		req.PrivateIpAddresses = append(req.PrivateIpAddresses, &vpc.PrivateIpAddressSpecification{
			PrivateIpAddress: common.StringPtr(ip),
		})
	}

	blog.V(3).Infof("UnassignPrivateIpAddresses req: %s", req.ToJsonString())

	resp, err := c.vpcClient.UnassignPrivateIpAddresses(req)
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		blog.Errorf("UnassignPrivateIpAddresses failed, err %s", err.Error())
		return fmt.Errorf("UnassignPrivateIpAddresses failed, err %s", err.Error())
	}

	blog.V(3).Infof("UnassignPrivateIpAddresses resp: %s", resp.ToJsonString())
	mf(metric.CloudOperationResultSuccess)
	return nil
}

// MigrateIP migrate ip
func (c *Client) MigrateIP(ip, srcEniID, destEniID string) error {
	startTime := time.Now()
	mf := func(result string) {
		metric.DefaultCollector.StatCloudOperation(
			cloud.CloudProviderTencent, "MigrateIP", result, startTime, time.Now())
	}
	req := vpc.NewMigratePrivateIpAddressRequest()
	req.PrivateIpAddress = common.StringPtr(ip)
	req.SourceNetworkInterfaceId = common.StringPtr(srcEniID)
	req.DestinationNetworkInterfaceId = common.StringPtr(destEniID)

	blog.V(3).Infof("MigratePrivateIpAddress req: %s", req.ToJsonString())

	resp, err := c.vpcClient.MigratePrivateIpAddress(req)
	if err != nil {
		mf(metric.CloudOperationResultFailed)
		blog.Errorf("MigratePrivateIpAddress failed, err %s", err.Error())
		return fmt.Errorf("MigratePrivateIpAddress failed, err %s", err.Error())
	}

	blog.V(3).Infof("MigratePrivateIpAddress resp: %s", resp.ToJsonString())
	mf(metric.CloudOperationResultSuccess)
	return nil
}
