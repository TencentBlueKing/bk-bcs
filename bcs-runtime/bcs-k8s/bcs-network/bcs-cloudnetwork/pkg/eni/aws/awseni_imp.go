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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// create eni
func (c *Client) createEni(name string, ipNum int) (*ec2.NetworkInterface, error) {
	// find available subnets, ipNum is secondary ip number
	subnet, err := c.getAvailableSubnet(ipNum)
	if err != nil {
		blog.Errorf("get available subnet when create eni")
		return nil, err
	}

	// use description for created eni name
	req := &ec2.CreateNetworkInterfaceInput{}
	req.SetDescription(name)
	req.SetSubnetId(aws.StringValue(subnet.SubnetId))
	req.SetSecondaryPrivateIpAddressCount(int64(ipNum))
	if len(c.SecurityGroups) != 0 {
		req.SetGroups(aws.StringSlice(c.SecurityGroups))
	}

	blog.V(2).Infof("aws CreateNetworkInterface request %+v", req)

	resp, err := c.ec2client.CreateNetworkInterface(req)
	if err != nil {
		blog.Errorf("aws CreateNetworkInterface failed, err %s", err.Error())
		return nil, err
	}

	blog.V(2).Infof("aws CreateNetworkInterface response %+v", resp)

	if resp.NetworkInterface == nil {
		blog.Errorf("aws CreateNetworkInterface failed, NetworkInterface in resp is empty")
		return nil, fmt.Errorf("aws CreateNetworkInterface failed, NetworkInterface in resp is empty")
	}

	return resp.NetworkInterface, nil
}

// modifyEniAttribute modify eni attribute
func (c *Client) modifyEniAttribute(eniID string, securityGroups []string, sourceDestCheckFlag bool) error {
	req := &ec2.ModifyNetworkInterfaceAttributeInput{}
	req.SetNetworkInterfaceId(eniID)
	if len(securityGroups) != 0 {
		req.SetGroups(aws.StringSlice(securityGroups))
	}
	req.SetSourceDestCheck(&ec2.AttributeBooleanValue{
		Value: aws.Bool(sourceDestCheckFlag),
	})

	blog.V(2).Infof("aws ModifyNetworkInterface Attribute request %+v", req)

	resp, err := c.ec2client.ModifyNetworkInterfaceAttribute(req)
	if err != nil {
		blog.Errorf("aws ModifyNetworkInterface failed, err %s", err.Error())
		return err
	}

	blog.V(2).Infof("aws ModifyNetworkInterface response %+v", resp)
	return nil
}

// query eni by eni description
func (c *Client) queryEni(eniName string) (*ec2.NetworkInterface, error) {
	req := &ec2.DescribeNetworkInterfacesInput{}
	req.SetFilters([]*ec2.Filter{
		{
			Name: aws.String("description"),
			Values: []*string{
				aws.String(eniName),
			},
		},
	})

	blog.V(2).Infof("aws DescribeNetworkInterfaces request %s", req.String())

	resp, err := c.ec2client.DescribeNetworkInterfaces(req)
	if err != nil {
		return nil, fmt.Errorf("describe network interface failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws DescribeNetworkInterfaces response %s", resp.String())

	if len(resp.NetworkInterfaces) == 0 {
		return nil, nil
	}
	if len(resp.NetworkInterfaces) > 1 {
		return nil, fmt.Errorf("eni with description %s more than 1", eniName)
	}

	return resp.NetworkInterfaces[0], nil
}

// assign private ip to network interface
func (c *Client) assignIPsToEni(eniID string, ipNum int) error {
	req := &ec2.AssignPrivateIpAddressesInput{}
	req.SetNetworkInterfaceId(eniID)
	req.SetSecondaryPrivateIpAddressCount(int64(ipNum))

	blog.V(2).Infof("aws AssignPrivateIpAddresses request %s", req.String())

	resp, err := c.ec2client.AssignPrivateIpAddresses(req)
	if err != nil {
		return fmt.Errorf("assign private ip addresses failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws AssignPrivateIpAddresses response %s", resp.String())
	return nil
}

// unassign private ip from network interface
func (c *Client) unassignIPsFromEni(eniID string, addrs []string) error {
	req := &ec2.UnassignPrivateIpAddressesInput{}
	req.SetNetworkInterfaceId(eniID)
	req.SetPrivateIpAddresses(aws.StringSlice(addrs))

	blog.V(2).Infof("aws UnassignPrivateIpAddresses request %s", req.String())

	resp, err := c.ec2client.UnassignPrivateIpAddresses(req)
	if err != nil {
		return fmt.Errorf("unassign private ip addresses failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws UnassignPrivateIpAddresses response %s", resp.String())
	return nil
}

func (c *Client) queryInstanceInfo() error {
	req := &ec2.DescribeInstancesInput{}
	req.SetFilters([]*ec2.Filter{
		{
			Name: aws.String("private-ip-address"),
			Values: []*string{
				aws.String(c.InstanceIP),
			},
		},
	})

	blog.V(2).Infof("aws DescribeInstances request %s", req.String())

	resp, err := c.ec2client.DescribeInstances(req)
	if err != nil {
		return fmt.Errorf("describe instance failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws DescribeInstances response %s", resp.String())

	if len(resp.Reservations) == 0 {
		return fmt.Errorf("no reservations info in DescribeInstances response")
	}

	reservation := resp.Reservations[0]
	if len(reservation.Instances) == 0 {
		return fmt.Errorf("no instance in reservation %s", aws.StringValue(reservation.ReservationId))
	}

	c.instance = reservation.Instances[0]
	return nil
}

func (c *Client) getAvailableSubnet(ipNum int) (*ec2.Subnet, error) {

	var subnetIDs []string
	if len(c.SubnetIDs) == 0 {
		subnetIDs = []string{aws.StringValue(c.instance.SubnetId)}
	} else {
		subnetIDs = c.SubnetIDs
	}

	req := &ec2.DescribeSubnetsInput{}
	req.SetSubnetIds(aws.StringSlice(subnetIDs))

	blog.V(2).Infof("aws DescribeSubnets request %s", req.String())

	resp, err := c.ec2client.DescribeSubnets(req)
	if err != nil {
		return nil, fmt.Errorf("describe subnets failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws DescribeSubnets response %s", resp.String())

	if len(resp.Subnets) == 0 {
		return nil, fmt.Errorf("no subnets in DescribeSubnets response")
	}
	for _, subnet := range resp.Subnets {
		if aws.Int64Value(subnet.AvailableIpAddressCount) > int64(ipNum+1) {
			return subnet, nil
		}
	}
	return nil, fmt.Errorf("no found available subnet")
}

func (c *Client) querySubent(subnetID string) (*ec2.Subnet, error) {
	req := &ec2.DescribeSubnetsInput{}
	req.SetSubnetIds(aws.StringSlice([]string{subnetID}))

	blog.V(2).Infof("aws DescribeSubnets request %s", req.String())

	resp, err := c.ec2client.DescribeSubnets(req)
	if err != nil {
		return nil, fmt.Errorf("describe subnets failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws DescribeSubnets response %s", resp.String())

	if len(resp.Subnets) == 0 {
		return nil, fmt.Errorf("no subnets in DescribeSubnets response")
	}

	return resp.Subnets[0], nil
}
