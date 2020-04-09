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

package aws

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"bk-bcs/bcs-common/common/blog"
	cloud "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis/cloud/v1"
)

// Client client for aws eni
type Client struct {
	// AccessID access id
	AccessID string
	// AccessSecret access secret
	AccessSecret string
	// SessionToken session token
	SessionToken string
	// Region aws region
	Region string
	// VpcID aws vpc id
	VpcID string

	// Instance IP
	InstanceIP string
	// SecurityGroups aws security groups
	SecurityGroups []string
	// SubnetIDs ids for subnet
	SubnetIDs []string

	instance *ec2.Instance

	ec2client *ec2.EC2
}

// New create aws client
func New(instanceIP string) *Client {
	return &Client{
		InstanceIP: instanceIP,
	}
}

func (c *Client) loadEnv() {
	c.Region = os.Getenv(ENV_NAME_AWS_REGION)
	c.VpcID = os.Getenv(ENV_NAME_AWS_VPC)

	subnetsStr := os.Getenv(ENV_NAME_AWS_SUBNETS)
	if len(subnetsStr) != 0 {
		strings.Replace(subnetsStr, ";", ",", -1)
		subnets := strings.Split(subnetsStr, ",")
		c.SubnetIDs = subnets
	}

	c.AccessID = os.Getenv(ENV_NAME_AWS_ACCESS_KEY_ID)
	c.AccessSecret = os.Getenv(ENV_NAME_AWS_SECRET_ACCESS_KEY)
	c.SessionToken = os.Getenv(ENV_NAME_AWS_SESSION_TOKEN)
}

func (c *Client) validate() error {
	if len(c.Region) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_AWS_REGION)
	}
	if len(c.VpcID) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_AWS_VPC)
	}
	if len(c.AccessID) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_AWS_ACCESS_KEY_ID)
	}
	if len(c.AccessSecret) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_AWS_SECRET_ACCESS_KEY)
	}
	return nil
}

// GetENILimit get eni limit
func (c *Client) GetENILimit() (int, int, error) {
	instanceType := aws.StringValue(c.instance.InstanceType)
	eniNum, ok := EniNumLimit[instanceType]
	if !ok {
		return -1, -1, fmt.Errorf("unknown instance type %s", instanceType)
	}
	ipNum, ok := IPNumLimit[instanceType]
	if !ok {
		return -1, -1, fmt.Errorf("unknown instance type %s", instanceType)
	}
	return eniNum, ipNum, nil
}

// Init implements eni interface
func (c *Client) Init() error {

	c.loadEnv()

	if err := c.validate(); err != nil {
		return err
	}

	session, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(c.Region),
			Credentials: credentials.NewStaticCredentials(c.AccessID, c.AccessSecret, c.SessionToken),
		},
	)
	if err != nil {
		return fmt.Errorf("create aws session failed, err %s", err.Error())
	}
	c.ec2client = ec2.New(session)

	err = c.queryInstanceInfo()
	if err != nil {
		return err
	}

	return nil
}

// GetVMInfo get vm info
func (c *Client) GetVMInfo() (*cloud.VMInfo, error) {
	if c.instance == nil {
		return nil, fmt.Errorf("no vm info")
	}
	regionID := aws.StringValue(c.instance.Placement.AvailabilityZone)
	vpcID := aws.StringValue(c.instance.VpcId)
	subnetID := aws.StringValue(c.instance.SubnetId)
	instanceID := aws.StringValue(c.instance.InstanceId)
	return &cloud.VMInfo{
		NodeRegionID: regionID,
		NodeVpcID:    vpcID,
		NodeSubnetID: subnetID,
		InstanceID:   instanceID,
	}, nil
}

// CreateENI create eni
func (c *Client) CreateENI(name string, ipNum int) (*cloud.ElasticNetworkInterface, error) {
	// find available subnets
	subnetID, err := c.getAvailableSubnet(ipNum)
	if err != nil {
		blog.Errorf("get available subnet when create eni")
		return nil, err
	}

	// use description for created eni name
	req := &ec2.CreateNetworkInterfaceInput{}
	req.SetDescription(name)
	req.SetSubnetId(subnetID)
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

	netIf := &cloud.ElasticNetworkInterface{}
	netIf.EniName = name
	netIf.EniSubnetID = subnetID
	netIf.IPNum = ipNum
	netIf.EniID = aws.StringValue(resp.NetworkInterface.NetworkInterfaceId)
	netIf.MacAddress = aws.StringValue(resp.NetworkInterface.MacAddress)

	// PrivateIpAddresses in response contains both primary ip and secondary ips
	for _, ip := range resp.NetworkInterface.PrivateIpAddresses {
		if aws.BoolValue(ip.Primary) {
			netIf.Address = &cloud.IPAddress{
				IP:        aws.StringValue(ip.PrivateIpAddress),
				DNSName:   aws.StringValue(ip.PrivateDnsName),
				IsPrimary: aws.BoolValue(ip.Primary),
			}
		} else {
			netIf.SecondaryAddresses = append(netIf.SecondaryAddresses, &cloud.IPAddress{
				IP:        aws.StringValue(ip.PrivateIpAddress),
				DNSName:   aws.StringValue(ip.PrivateDnsName),
				IsPrimary: aws.BoolValue(ip.Primary),
			})
		}

	}

	return netIf, nil
}

// AttachENI attach eni to vm
func (c *Client) AttachENI(index int, eniID, instanceID string) (*cloud.NetworkInterfaceAttachment, error) {
	req := &ec2.AttachNetworkInterfaceInput{}
	req.SetNetworkInterfaceId(eniID)
	req.SetInstanceId(instanceID)
	req.SetDeviceIndex(int64(index))

	blog.V(2).Infof("aws AttachNetworkInterface request %+v", req)

	resp, err := c.ec2client.AttachNetworkInterface(req)
	if err != nil {
		blog.Errorf("aws AttachNetworkInterface failed, err %s", err.Error())
		return nil, err
	}

	blog.V(2).Infof("aws AttachNetworkInterface response %+v", resp)

	if resp.AttachmentId == nil {
		blog.Errorf("aws AttachNetworkInterface, AttachmentId in resp is empty")
		return nil, fmt.Errorf("aws AttachNetworkInterface, AttachmentId in resp is empty")
	}
	return &cloud.NetworkInterfaceAttachment{
		Index:        index,
		AttachmentID: aws.StringValue(resp.AttachmentId),
		InstanceID:   instanceID,
	}, nil
}

// DetachENI detach eni from vm
func (c *Client) DetachENI(attachment *cloud.NetworkInterfaceAttachment) error {
	req := &ec2.DetachNetworkInterfaceInput{}
	req.SetAttachmentId(attachment.AttachmentID)

	blog.V(2).Infof("aws DetachNetworkInterface request %+v", req)

	resp, err := c.ec2client.DetachNetworkInterface(req)
	if err != nil {
		blog.Errorf("aws DetachNetworkInterface failed, err %s", err.Error())
		return err
	}

	blog.V(2).Infof("aws DetachNetworkInterface response, %+v", resp)
	return nil
}

// DeleteENI delete eni from vm
func (c *Client) DeleteENI(eniID string) error {
	req := &ec2.DeleteNetworkInterfaceInput{}
	req.SetNetworkInterfaceId(eniID)

	blog.V(2).Infof("aws DeleteNetworkInterface request %+v", req)

	resp, err := c.ec2client.DeleteNetworkInterface(req)
	if err != nil {
		blog.Errorf("aws DeleteNetworkInterface failed, err %s", err.Error())
		return err
	}

	blog.V(2).Infof("aws DeleteNetworkInterface response, %+v", resp)
	return nil
}

// ListENIs list enis of a vm
func (c *Client) ListENIs(eniIDs []string) ([]*cloud.ElasticNetworkInterface, error) {
	req := &ec2.DescribeNetworkInterfacesInput{}
	req.SetNetworkInterfaceIds(aws.StringSlice(eniIDs))

	blog.V(2).Infof("aws DescribeNetworkInterfaces request %+v", req)

	resp, err := c.ec2client.DescribeNetworkInterfaces(req)
	if err != nil {
		blog.Errorf("aws DescribeNetworkInterfaces failed, err %s", err.Error())
		return nil, err
	}

	blog.V(2).Infof("aws DescribeNetworkInterfaces response, %+v", resp)

	var ifs []*cloud.ElasticNetworkInterface
	for _, netif := range resp.NetworkInterfaces {
		tmpIf := &cloud.ElasticNetworkInterface{
			EniID:       aws.StringValue(netif.NetworkInterfaceId),
			EniName:     aws.StringValue(netif.Description),
			EniSubnetID: aws.StringValue(netif.SubnetId),
			MacAddress:  aws.StringValue(netif.MacAddress),
			IPNum:       len(netif.PrivateIpAddresses),
		}
		if netif.Attachment != nil {
			tmpIf.Attachment = &cloud.NetworkInterfaceAttachment{
				Index:        int(aws.Int64Value(netif.Attachment.DeviceIndex)),
				AttachmentID: aws.StringValue(netif.Attachment.AttachmentId),
				InstanceID:   aws.StringValue(netif.Attachment.InstanceId),
			}
		}
		for _, ip := range netif.PrivateIpAddresses {
			if aws.BoolValue(ip.Primary) {
				tmpIf.Address = &cloud.IPAddress{
					IP:        aws.StringValue(ip.PrivateIpAddress),
					DNSName:   aws.StringValue(ip.PrivateDnsName),
					IsPrimary: aws.BoolValue(ip.Primary),
				}
			} else {
				tmpIf.SecondaryAddresses = append(tmpIf.SecondaryAddresses, &cloud.IPAddress{
					IP:        aws.StringValue(ip.PrivateIpAddress),
					DNSName:   aws.StringValue(ip.PrivateDnsName),
					IsPrimary: aws.BoolValue(ip.Primary),
				})
			}
		}
		ifs = append(ifs, tmpIf)
	}
	return ifs, nil
}
