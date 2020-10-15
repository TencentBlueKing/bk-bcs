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

package tencentcloud

import (
	"fmt"
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
)

// Client client for tencentcloud
type Client struct {
	// tencentcloud vpc domain
	vpcDomain string

	// tencentcloud region
	region string

	// secret id for tencent cloud
	secretID string

	// secret key for tencent cloud
	secretKey string

	// vpcClient client for tencent cloud vpc
	vpcClient *vpc.Client
}

// NewClient create new client
func NewClient() (*Client, error) {
	c := &Client{}
	err := c.loadEnv()
	if err != nil {
		return nil, err
	}

	credential := common.NewCredential(
		c.secretID,
		c.secretKey,
	)
	cpf := profile.NewClientProfile()
	if len(c.vpcDomain) != 0 {
		cpf.HttpProfile.Endpoint = c.vpcDomain
	}
	client, err := vpc.NewClient(credential, c.region, cpf)
	if err != nil {
		blog.Errorf("new vpc client failed, err %s", err.Error())
		return nil, fmt.Errorf("new vpc client failed, err %s", err.Error())
	}
	c.vpcClient = client
	return c, nil
}

func (c *Client) loadEnv() error {
	c.vpcDomain = os.Getenv(ENV_NAME_TENCENTCLOUD_VPC_DOMAIN)
	c.region = os.Getenv(ENV_NAME_TENCENTCLOUD_REGION)
	c.secretID = os.Getenv(ENV_NAME_TENCENTCLOUD_ACCESS_KEY_ID)
	secretKey := os.Getenv(ENV_NAME_TENCENTCLOUD_ACCESS_KEY)

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
	tSubnets, err := c.describeSubnets([]string{subnetID})
	if err != nil {
		return nil, err
	}
	// expect just one subnet
	if len(tSubnets) != 1 {
		return nil, fmt.Errorf("describeSubnet expect 1 subnet")
	}
	tSubnet := tSubnets[0]

	if vpcID != *tSubnet.VpcId || region != c.region {
		blog.Infof("vpcID or region of subnet does not match, expect (%s, %s), get (%s, %s)",
			vpcID, region, *tSubnet.VpcId, c.region)
		return nil, fmt.Errorf("vpcID or region of subnet does not match, expect (%s, %s), get (%s, %s)",
			vpcID, region, *tSubnet.VpcId, c.region)
	}
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
	tSubnets, err := c.describeSubnets(subnetIDs)
	if err != nil {
		return nil, err
	}

	var cloudSubnets []*types.CloudSubnet
	for _, tSubnet := range tSubnets {
		if vpcID != *tSubnet.VpcId || region != c.region {
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
		return nil, fmt.Errorf("subnets %+v not found", subnetIDs)
	}
	return cloudSubnets, nil
}

// QueryEni query eni
func (c *Client) QueryEni(eniID string) (*types.EniObject, error) {
	req := vpc.NewDescribeNetworkInterfacesRequest()
	req.NetworkInterfaceIds = common.StringPtrs([]string{eniID})

	blog.V(3).Infof("DescribeNetworkInterfaces req: %s", req.ToJsonString())

	resp, err := c.vpcClient.DescribeNetworkInterfaces(req)
	if err != nil {
		blog.Errorf("DescribeNetworkInterfaces failed, err %s", err.Error())
		return nil, fmt.Errorf("DescribeNetworkInterfaces failed, err %s", err.Error())
	}

	blog.V(3).Infof("DescribeNetworkInterfaces resp: %s", resp.ToJsonString())

	if len(resp.Response.NetworkInterfaceSet) != 1 {
		blog.Errorf("DescribeNetworkInterfaces expect 1 eni, but get %d", len(resp.Response.NetworkInterfaceSet))
		return nil, fmt.Errorf("DescribeNetworkInterfaces expect 1 eni, but get %d", len(resp.Response.NetworkInterfaceSet))
	}

	retEni := resp.Response.NetworkInterfaceSet[0]

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

// AssignIPToEni assign ip to eni
func (c *Client) AssignIPToEni(ip, eniID string) (string, error) {
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
		blog.Errorf("AssignPrivateIpAddresses failed, err %s", err.Error())
		return "", fmt.Errorf("AssignPrivateIpAddresses failed, err %s", err.Error())
	}

	blog.V(3).Infof("AssignPrivateIpAddresses resp: %s", resp.ToJsonString())

	if len(resp.Response.PrivateIpAddressSet) != 1 {
		blog.Errorf("AssignPrivateIpAddresses expect 1 ip, but get %d", len(resp.Response.PrivateIpAddressSet))
		return "", fmt.Errorf("AssignPrivateIpAddresses expect 1 ip, but get %d", len(resp.Response.PrivateIpAddressSet))
	}

	return *resp.Response.PrivateIpAddressSet[0].PrivateIpAddress, nil

}

// UnassignIPFromEni unassign ip from eni
func (c *Client) UnassignIPFromEni(ip, eniID string) error {
	req := vpc.NewUnassignPrivateIpAddressesRequest()
	req.NetworkInterfaceId = common.StringPtr(eniID)
	req.PrivateIpAddresses = append(req.PrivateIpAddresses, &vpc.PrivateIpAddressSpecification{
		PrivateIpAddress: common.StringPtr(ip),
	})

	blog.V(3).Infof("UnassignPrivateIpAddresses req: %s", req.ToJsonString())

	resp, err := c.vpcClient.UnassignPrivateIpAddresses(req)
	if err != nil {
		blog.Errorf("UnassignPrivateIpAddresses failed, err %s", err.Error())
		return fmt.Errorf("UnassignPrivateIpAddresses failed, err %s", err.Error())
	}

	blog.V(3).Infof("UnassignPrivateIpAddresses resp: %s", resp.ToJsonString())
	return nil
}

// MigrateIP migrate ip
func (c *Client) MigrateIP(ip, srcEniID, destEniID string) error {
	req := vpc.NewMigratePrivateIpAddressRequest()
	req.PrivateIpAddress = common.StringPtr(ip)
	req.SourceNetworkInterfaceId = common.StringPtr(srcEniID)
	req.DestinationNetworkInterfaceId = common.StringPtr(destEniID)

	blog.V(3).Infof("MigratePrivateIpAddress req: %s", req.ToJsonString())

	resp, err := c.vpcClient.MigratePrivateIpAddress(req)
	if err != nil {
		blog.Errorf("MigratePrivateIpAddress failed, err %s", err.Error())
		return fmt.Errorf("MigratePrivateIpAddress failed, err %s", err.Error())
	}

	blog.V(3).Infof("MigratePrivateIpAddress resp: %s", resp.ToJsonString())
	return nil
}
