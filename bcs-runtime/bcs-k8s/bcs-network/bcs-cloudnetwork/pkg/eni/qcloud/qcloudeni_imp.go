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

package qcloud

import (
	"fmt"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// create eni
func (c *Client) createEni(name string, ipNum int) (*vpc.NetworkInterface, error) {
	// find available subnets, ipNum is secondary ip number
	subnet, err := c.getAvailableSubnet(ipNum)
	if err != nil {
		blog.Errorf("get available subnet failed when create eni, err %s", err.Error())
		return nil, fmt.Errorf("get available subnet failed when create eni, err %s", err.Error())
	}

	req := vpc.NewCreateNetworkInterfaceRequest()
	req.NetworkInterfaceName = common.StringPtr(name)
	req.SubnetId = subnet.SubnetId
	if len(c.SecurityGroups) != 0 {
		req.SecurityGroupIds = common.StringPtrs(c.SecurityGroups)
	}

	blog.V(2).Infof("tencentcloud CreateNetworkInterface request %s", req.ToJsonString())

	resp, err := c.vpcClient.CreateNetworkInterface(req)
	if err != nil {
		blog.Errorf("tencentcloud CreateNetworkInterface failed, err %s", err.Error())
	}

	blog.V(2).Infof("tencentcloud CreateNetworkInterface response %s", resp.ToJsonString())

	if resp.Response.NetworkInterface == nil {
		blog.Errorf("tencentcloud CreateNetworkInterface failed, NetworkInterface in resp is empty")
		return nil, fmt.Errorf("tencentcloud CreateNetworkInterface failed, NetworkInterface in resp is empty")
	}

	return resp.Response.NetworkInterface, nil
}

// queryENI query eni, support query by eniID and eniName
// match eniID if eniID is not empty
// match eniName if eniName is not empty
func (c *Client) queryENI(eniID string, instanceID string, eniName string) ([]*vpc.NetworkInterface, error) {
	req := vpc.NewDescribeNetworkInterfacesRequest()
	req.Filters = make([]*vpc.Filter, 0)
	if len(eniID) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name: common.StringPtr("network-interface-id"),
			Values: []*string{
				common.StringPtr(eniID),
			},
		})
	}
	if len(instanceID) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name: common.StringPtr("attachment.instance-id"),
			Values: []*string{
				common.StringPtr(instanceID),
			},
		})
	}
	if len(eniName) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name: common.StringPtr("network-interface-name"),
			Values: []*string{
				common.StringPtr(eniName),
			},
		})
	}

	blog.V(2).Infof("tencentcloud DescribeNetworkInterfaces request %s", req.ToJsonString())

	response, err := c.vpcClient.DescribeNetworkInterfaces(req)
	if err != nil {
		blog.Errorf("tencentcloud DescribeNetworkInterfaces failed, err %s", err.Error())
		return nil, fmt.Errorf("tencentcloud DescribeNetworkInterfaces failed, err %s", err.Error())
	}

	blog.V(2).Infof("tencentcloud DescribeNetworkInterfaces response: %s", response.ToJsonString())

	if *(response.Response.TotalCount) == 0 {
		blog.Warnf("tencentcloud DescribeNetworkInterfaces return zero result")
		return nil, nil
	}
	return response.Response.NetworkInterfaceSet, nil
}

// attachENI attach eni to cvm
func (c *Client) attachENI(eniID string, instanceID string) error {
	request := vpc.NewAttachNetworkInterfaceRequest()
	request.NetworkInterfaceId = common.StringPtr(eniID)
	request.InstanceId = common.StringPtr(instanceID)

	blog.V(2).Infof("tencentcloud AttachNetworkInterface request %s", request.ToJsonString())

	response, err := c.vpcClient.AttachNetworkInterface(request)
	if err != nil {
		blog.Errorf("tencentcloud AttachNetworkInterface %s to ins %s failed, err %s", eniID, instanceID, err.Error())
		return fmt.Errorf("tencentcloud AttachNetworkInterface %s to ins %s failed, err %s", eniID, instanceID, err.Error())
	}

	blog.V(3).Infof("tencentcloud AttachNetworkInterface response %s", response.ToJsonString())

	return nil
}

// detachENI delete eni by eniID
func (c *Client) detachENI(eniID string, instanceID string) error {
	request := vpc.NewDetachNetworkInterfaceRequest()
	request.NetworkInterfaceId = common.StringPtr(eniID)
	request.InstanceId = common.StringPtr(instanceID)

	blog.V(2).Infof("tencentcloud DetachNetworkInterface request %s", request.ToJsonString())

	response, err := c.vpcClient.DetachNetworkInterface(request)
	if err != nil {
		blog.Errorf("tencentcloud DetachNetworkInterface %s from ins %s failed, err %s", eniID, instanceID, err.Error())
		return fmt.Errorf("tencentcloud DetachNetworkInterface %s from ins %s failed, err %s", eniID, instanceID, err.Error())
	}
	blog.V(3).Infof("tencentcloud DetachNetworkInterface response %s", response.ToJsonString())
	return nil
}

// deleteEni delete eni
func (c *Client) deleteEni(eniID string) error {
	req := vpc.NewDeleteNetworkInterfaceRequest()
	req.NetworkInterfaceId = common.StringPtr(eniID)

	blog.V(2).Infof("tencentcloud DeleteNetworkInterface request %s", req.ToJsonString())

	resp, err := c.vpcClient.DeleteNetworkInterface(req)
	if err != nil {
		blog.Errorf("tencentcloud DeleteNetworkInterface failed %s", err.Error())
		return fmt.Errorf("tencentcloud DeleteNetworkInterface failed %s", err.Error())
	}

	blog.V(2).Infof("tencentcloud DeleteNetworkInterface response: %s", resp.ToJsonString())
	return nil
}

// assignIPsToEni assign private ip to network interface
func (c *Client) assignIPsToEni(eniID string, ipNum int) error {
	req := vpc.NewAssignPrivateIpAddressesRequest()
	req.NetworkInterfaceId = common.StringPtr(eniID)
	req.SecondaryPrivateIpAddressCount = common.Uint64Ptr(uint64(ipNum))

	blog.V(2).Infof("tencentcloud AssignPrivateIpAddresses request %s", req.ToJsonString())

	resp, err := c.vpcClient.AssignPrivateIpAddresses(req)
	if err != nil {
		blog.Errorf("tencentcloud AssignPrivateIpAddresses request %s", req.ToJsonString())
		return fmt.Errorf("tencentcloud AssignPrivateIpAddresses request %s", req.ToJsonString())
	}

	blog.V(2).Infof("tencentcloud AssignPrivateIpAddresses response %s", resp.ToJsonString())
	return nil
}

// unassignIPsFromEni unassign private ip from network interface
func (c *Client) unassignIPsFromEni(eniID string, addrs []string) error {
	req := vpc.NewUnassignPrivateIpAddressesRequest()
	req.NetworkInterfaceId = common.StringPtr(eniID)
	if len(addrs) == 0 {
		blog.Warnf("tencentcloud UnassignPrivateIpAddresses request with no addrs")
		return nil
	}
	for _, addr := range addrs {
		req.PrivateIpAddresses = append(req.PrivateIpAddresses, &vpc.PrivateIpAddressSpecification{
			PrivateIpAddress: common.StringPtr(addr),
		})
	}

	blog.V(2).Infof("tencentcloud UnassignPrivateIpAddresses request %s", req.ToJsonString())

	resp, err := c.vpcClient.UnassignPrivateIpAddresses(req)
	if err != nil {
		blog.Errorf("tencentcloud UnassignPrivateIpAddresses failed, err %s", err.Error())
		return fmt.Errorf("tencentcloud UnassignPrivateIpAddresses failed, err %s", err.Error())
	}

	blog.V(2).Infof("tencentcloud UnassignPrivateIpAddresses response %s", resp.ToJsonString())
	return nil
}

// queryInstanceInfo query instance info, and store it in Client object
func (c *Client) queryInstanceInfo() error {
	req := cvm.NewDescribeInstancesRequest()
	req.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("private-ip-address"),
			Values: common.StringPtrs([]string{c.InstanceIP}),
		},
	}

	blog.V(2).Infof("tencentcloud DescribeInstances request %s", req.ToJsonString())

	resp, err := c.cvmClient.DescribeInstances(req)
	if err != nil {
		return fmt.Errorf("tencentcloud DescribeInstances failed, err %s", err.Error())
	}

	blog.V(2).Infof("tencentcloud DescribeInstances response %s", resp.ToJsonString())

	if *resp.Response.TotalCount == 0 {
		return fmt.Errorf("no instances")
	}

	c.instance = resp.Response.InstanceSet[0]
	return nil
}

// getAvailableSubnet get available subnet id which meets request secondary ip
func (c *Client) getAvailableSubnet(ipNum int) (*vpc.Subnet, error) {
	var subnetIDs []*string
	if len(c.SubnetIDs) == 0 {
		subnetIDs = []*string{c.instance.VirtualPrivateCloud.SubnetId}
	} else {
		subnetIDs = common.StringPtrs(c.SubnetIDs)
	}

	req := vpc.NewDescribeSubnetsRequest()
	req.SubnetIds = subnetIDs

	blog.V(2).Infof("tencentcloud DescribeSubnets request %s", req.ToJsonString())

	resp, err := c.vpcClient.DescribeSubnets(req)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud DescribeSubnets failed, err %s", err.Error())
	}

	blog.V(2).Infof("tencentcloud DescribeSubnets response %s", resp.ToJsonString())

	if len(resp.Response.SubnetSet) == 0 {
		return nil, fmt.Errorf("no subnets in DescribeSubnets response")
	}
	for _, subnet := range resp.Response.SubnetSet {
		if int64(*subnet.AvailableIpAddressCount) > int64(ipNum+1) {
			return subnet, nil
		}
	}
	return nil, fmt.Errorf("no found available subnet")
}

func (c *Client) waitForAttached(eniID string, checkNum, checkInterval int) error {
	return c.doWaitForStatus(eniID, checkNum, checkInterval, ENI_STATUS_ATTACHED)
}

func (c *Client) waitForDetached(eniID string, checkNum, checkInterval int) error {
	return c.doWaitForStatus(eniID, checkNum, checkInterval, ENI_STATUS_DETACHED)
}

func (c *Client) waitForAvailable(eniID string, checkNum, checkInterval int) error {
	return c.doWaitForStatus(eniID, checkNum, checkInterval, ENI_STATUS_AVAILABLE)
}

// doWaitForStatus wait for the interface to reach a certain state
func (c *Client) doWaitForStatus(eniID string, checkNum, checkInterval int, finalStatus string) error {
	for i := 0; i < checkNum; i++ {
		time.Sleep(time.Second * time.Duration(checkInterval))
		enis, err := c.queryENI(eniID, "", "")
		if err != nil {
			return err
		}
		for _, eni := range enis {
			if *eni.NetworkInterfaceId == eniID {
				switch *eni.State {
				case ENI_STATUS_AVAILABLE:
					switch finalStatus {
					case ENI_STATUS_ATTACHED:
						if eni.Attachment != nil && eni.Attachment.InstanceId != nil {
							blog.Infof("eni %s is attached", eniID)
							return nil
						}
						blog.Infof("eni %s is not attached", eniID)
					case ENI_STATUS_DETACHED:
						if eni.Attachment == nil {
							blog.Infof("eni %s is detached", eniID)
							return nil
						}
						blog.Infof("eni %s is not detached", eniID)
					default:
						blog.Infof("eni %s is %s now", eniID, *eni.State)
						return nil
					}
					break
				case ENI_STATUS_PENDING, ENI_STATUS_ATTACHING, ENI_STATUS_DETACHING, ENI_STATUS_DELETING:
					blog.Infof("eni %s is %s", eniID, *eni.State)
					break
				}
			}
		}
	}
	blog.Errorf("timeout when wait for eni %s", eniID)
	return fmt.Errorf("timeout when wait for eni %s", eniID)
}

// querySubnet query subnet by subnetid
func (c *Client) querySubnet(subnetID string) (*vpc.Subnet, error) {
	request := vpc.NewDescribeSubnetsRequest()
	request.SubnetIds = []*string{
		common.StringPtr(subnetID),
	}

	blog.V(2).Infof("tencentcloud DescribeSubnets request %s", request.ToJsonString())

	response, err := c.vpcClient.DescribeSubnets(request)
	if err != nil {
		blog.Errorf("tencentcloud DescribeSubnets by id %s failed, err %s", subnetID, err.Error())
		return nil, fmt.Errorf("tencentcloud DescribeSubnets by id %s failed, err %s", subnetID, err.Error())
	}

	blog.V(2).Infof("tencentcloud DescribeSubnets response %s", response.ToJsonString())

	if *response.Response.TotalCount == 0 {
		blog.Errorf("tencentcloud DescribeSubnets by id %s return zero result", subnetID)
		return nil, fmt.Errorf("tencentcloud DescribeSubnets by id %s return zero result", subnetID)
	}
	if len(response.Response.SubnetSet) != 1 {
		blog.Errorf("tencentcloud DescribeSubnets by id %s return result more than 1", subnetID)
		return nil, fmt.Errorf("tencentcloud DescribeSubnets by id %s return result more than 1", subnetID)
	}
	return response.Response.SubnetSet[0], nil
}
