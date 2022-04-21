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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/vishvananda/netlink"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/enilimit"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	cloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
)

const (
	// AWS_WAIT_ATTACHED_INTERVAL interval for waiting for aws eni attached
	AWS_WAIT_ATTACHED_INTERVAL = 5 * time.Second
	// AWS_WAIT_ATTACHED_MAX_RETRIES max retry times for waiting eni attached
	AWS_WAIT_ATTACHED_MAX_RETRIES = 10
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

func (c *Client) loadEnv() error {
	c.Region = os.Getenv(constant.ENV_NAME_AWS_REGION)
	c.VpcID = os.Getenv(constant.ENV_NAME_AWS_VPC)

	subnetsStr := os.Getenv(constant.ENV_NAME_AWS_SUBNETS)
	if len(subnetsStr) != 0 {
		strings.Replace(subnetsStr, ";", ",", -1)
		subnets := strings.Split(subnetsStr, ",")
		c.SubnetIDs = subnets
	}

	sGroupsStr := os.Getenv(constant.ENV_NAME_AWS_SECURITY_GROUPS)
	if len(sGroupsStr) != 0 {
		strings.Replace(sGroupsStr, ";", ",", -1)
		sGroups := strings.Split(sGroupsStr, ",")
		c.SecurityGroups = sGroups
	}

	c.AccessID = os.Getenv(constant.ENV_NAME_AWS_ACCESS_KEY_ID)
	accessSecret := os.Getenv(constant.ENV_NAME_AWS_SECRET_ACCESS_KEY)

	decryptSecret, err := encrypt.DesDecryptFromBase([]byte(accessSecret))
	if err != nil {
		blog.Errorf("decrypt access secret key failed, err %s", err.Error())
		return fmt.Errorf("decrypt access secret key failed, err %s", err.Error())
	}
	c.AccessSecret = string(decryptSecret)

	c.SessionToken = os.Getenv(constant.ENV_NAME_AWS_SESSION_TOKEN)
	return nil
}

func (c *Client) validate() error {
	if len(c.Region) == 0 {
		return fmt.Errorf("%s cannot be empty", constant.ENV_NAME_AWS_REGION)
	}
	if len(c.VpcID) == 0 {
		return fmt.Errorf("%s cannot be empty", constant.ENV_NAME_AWS_VPC)
	}
	if len(c.AccessID) == 0 {
		return fmt.Errorf("%s cannot be empty", constant.ENV_NAME_AWS_ACCESS_KEY_ID)
	}
	if len(c.AccessSecret) == 0 {
		return fmt.Errorf("%s cannot be empty", constant.ENV_NAME_AWS_SECRET_ACCESS_KEY)
	}
	return nil
}

// GetENILimit get eni limit
func (c *Client) GetENILimit() (int, int, error) {
	instanceType := aws.StringValue(c.instance.InstanceType)
	getter, err := enilimit.NewGetterFromEnv()
	if err != nil {
		blog.Warnf("get eni limit from env failed, err %s, lookup default map", err.Error())
	}
	if getter != nil {
		eniNum, ipNum, found := getter.GetLimit(instanceType)
		if found {
			return eniNum, ipNum, nil
		}
	}
	eniNum, ok := constant.AwsEniNumLimit[instanceType]
	if !ok {
		return -1, -1, fmt.Errorf("unknown instance type %s", instanceType)
	}
	ipNum, ok := constant.AwsIPNumLimit[instanceType]
	if !ok {
		return -1, -1, fmt.Errorf("unknown instance type %s", instanceType)
	}
	return eniNum, ipNum, nil
}

// Init implements eni interface
func (c *Client) Init() error {

	if err := c.loadEnv(); err != nil {
		return err
	}

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
		NodeRegion:   regionID,
		NodeVpcID:    vpcID,
		NodeSubnetID: subnetID,
		InstanceID:   instanceID,
		InstanceIP:   c.InstanceIP,
	}, nil
}

// GetMaxENIIndex get current eni max binding index
func (c *Client) GetMaxENIIndex() (int, error) {
	if c.instance == nil {
		return -1, fmt.Errorf("no vm info")
	}
	return len(c.instance.NetworkInterfaces) - 1, nil
}

// CreateENI create eni
func (c *Client) CreateENI(name string, ipNum int) (*cloud.ElasticNetworkInterface, error) {

	eni, err := c.queryEni(name)
	if err != nil {
		return nil, fmt.Errorf("queryEni failed, err %s", err.Error())
	}
	// take over existed eni
	if eni != nil {
		if len(eni.PrivateIpAddresses)-1 < ipNum {
			err = c.assignIPsToEni(aws.StringValue(eni.NetworkInterfaceId), ipNum-(len(eni.PrivateIpAddresses)-1))
			if err != nil {
				return nil, fmt.Errorf("assign ip to %s failed, err %s", name, err.Error())
			}
		} else if len(eni.PrivateIpAddresses)-1 > ipNum {
			var arrs []string
			for _, ipAddr := range eni.PrivateIpAddresses {
				if !aws.BoolValue(ipAddr.Primary) {
					arrs = append(arrs, aws.StringValue(ipAddr.PrivateIpAddress))
					if len(arrs) >= len(eni.PrivateIpAddresses)-1-ipNum {
						break
					}
				}
			}
			err := c.unassignIPsFromEni(aws.StringValue(eni.NetworkInterfaceId), arrs)
			if err != nil {
				return nil, fmt.Errorf("unassign ips %+v from %s failed, err %s", arrs, name, err.Error())
			}
		}
		if len(eni.PrivateIpAddresses)-1 != ipNum {
			eni, err = c.queryEni(name)
			if err != nil {
				return nil, fmt.Errorf("queryEni failed, err %s", err.Error())
			}
		}
		// create eni
	} else {
		eni, err = c.createEni(name, ipNum)
		if err != nil {
			return nil, fmt.Errorf("createEni failed, err %s", err.Error())
		}
	}

	// modify eni attribute for source dest check
	err = c.modifyEniAttribute(aws.StringValue(eni.NetworkInterfaceId), nil, false)
	if err != nil {
		return nil, fmt.Errorf("modifyEniAttribute failed, err %s", err.Error())
	}

	subnet, err := c.querySubent(aws.StringValue(eni.SubnetId))
	if err != nil {
		return nil, fmt.Errorf("querySubnet failed, err %s", err.Error())
	}

	netIf := &cloud.ElasticNetworkInterface{}
	netIf.EniName = name
	netIf.EniSubnetID = aws.StringValue(eni.SubnetId)
	netIf.EniSubnetCidr = aws.StringValue(subnet.CidrBlock)
	netIf.IPNum = ipNum
	netIf.EniID = aws.StringValue(eni.NetworkInterfaceId)
	netIf.MacAddress = aws.StringValue(eni.MacAddress)

	if eni.Attachment != nil {
		netIf.Attachment = &cloud.NetworkInterfaceAttachment{
			Index:        int(aws.Int64Value(eni.Attachment.DeviceIndex)),
			AttachmentID: aws.StringValue(eni.Attachment.AttachmentId),
			InstanceID:   aws.StringValue(eni.Attachment.InstanceId),
		}
	}

	// PrivateIpAddresses in response contains both primary ip and secondary ips
	for _, ip := range eni.PrivateIpAddresses {
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
func (c *Client) AttachENI(index int, eniID, instanceID, eniMac string) (*cloud.NetworkInterfaceAttachment, error) {
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

	// wait for real attachment
	err = c.waitForENIAttached(eniMac)
	if err != nil {
		blog.Errorf("wait for eni attached failed, err %s", err.Error())
		return nil, fmt.Errorf("wait for eni attached failed, err %s", err.Error())
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

	var err error
	var resp *ec2.DetachNetworkInterfaceOutput
	RetryWithBackoffTime(100, NewIncreseSeries(3*time.Second, 0.3, 0.3), func() bool {
		resp, err = c.ec2client.DetachNetworkInterface(req)
		if err != nil {
			blog.Errorf("aws DetachNetworkInterface failed, err %s", err.Error())
			return false
		}
		blog.V(2).Infof("aws DetachNetworkInterface response, %+v", resp)
		return true
	})
	if err != nil {
		return fmt.Errorf("aws DetachNetworkInterface failed, err %s", err.Error())
	}

	return nil
}

// DeleteENI delete eni from vm
func (c *Client) DeleteENI(eniID string) error {
	req := &ec2.DeleteNetworkInterfaceInput{}
	req.SetNetworkInterfaceId(eniID)

	blog.V(2).Infof("aws DeleteNetworkInterface request %+v", req)

	var err error
	var resp *ec2.DeleteNetworkInterfaceOutput
	RetryWithBackoffTime(100, NewIncreseSeries(3*time.Second, 0.3, 0.3), func() bool {
		resp, err = c.ec2client.DeleteNetworkInterface(req)
		if err != nil {
			blog.Errorf("aws DeleteNetworkInterface failed, err %s", err.Error())
			return false
		}

		blog.V(2).Infof("aws DeleteNetworkInterface response, %+v", resp)
		return true
	})
	if err != nil {
		return fmt.Errorf("aws DeleteNetworkInterface failed, err %s", err.Error())
	}

	return nil
}

// wait for eni attach
func (c *Client) waitForENIAttached(eniMac string) error {
	retries := 0
	for {
		linkList, err := netlink.LinkList()
		if err != nil {
			blog.Errorf("failed to list links, err %s", err.Error())
			return err
		}
		for _, link := range linkList {
			macFound := link.Attrs().HardwareAddr.String()
			linkName := link.Attrs().Name
			blog.V(3).Infof("link with mac: %s, name: %s", macFound, linkName)
			if strings.ToLower(macFound) == strings.ToLower(eniMac) {
				blog.V(3).Infof("found eni with mac %s", eniMac)
				return nil
			}
		}
		retries = retries + 1
		if retries > AWS_WAIT_ATTACHED_MAX_RETRIES {
			return fmt.Errorf("wait for eni attached failed, exceed max retries")
		}
		blog.V(3).Infof("%s not attached, retry (%d/%d)", eniMac, retries, AWS_WAIT_ATTACHED_MAX_RETRIES)
		time.Sleep(AWS_WAIT_ATTACHED_INTERVAL)
	}
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
