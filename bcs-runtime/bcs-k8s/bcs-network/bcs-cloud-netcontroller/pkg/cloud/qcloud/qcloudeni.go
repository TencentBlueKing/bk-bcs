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
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	ctrlcloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	cloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// Client qcloud client
type Client struct {
	// SecretID secret id
	SecretID string

	// SecretKey secret key
	SecretKey string

	// Region qcloud region
	Region string

	// VpcID qcloud vpc id
	VpcID string

	// VpcDomain domain for vpc api service
	VpcDomain string

	// CvmDomain domain for cvm api service
	CvmDomain string

	// SecurityGroups qcloud security groups
	SecurityGroups []string

	// SubnetIDs ids for subnet
	SubnetIDs []string

	// vpcClient client for tencent cloud vpc
	vpcClient *vpc.Client

	// cvmClient client for tencent cloud vpc
	cvmClient *cvm.Client
}

// New create client
func New() *Client {
	return &Client{}
}

func (c *Client) loadEnv() error {
	c.Region = os.Getenv(ENV_NAME_TENCENTCLOUD_REGION)
	c.VpcID = os.Getenv(ENV_NAME_TENCENTCLOUD_VPC)

	subnetsStr := os.Getenv(ENV_NAME_TENCENTCLOUD_SUBNETS)
	if len(subnetsStr) != 0 {
		strings.Replace(subnetsStr, ";", ",", -1)
		subnets := strings.Split(subnetsStr, ",")
		c.SubnetIDs = subnets
	}

	sGroupsStr := os.Getenv(ENV_NAME_TENCENTCLOUD_SECURITY_GROUPS)
	if len(sGroupsStr) != 0 {
		strings.Replace(sGroupsStr, ";", ",", -1)
		sGroups := strings.Split(sGroupsStr, ",")
		c.SecurityGroups = sGroups
	}

	c.SecretID = os.Getenv(ENV_NAME_TENCENTCLOUD_ACCESS_KEY_ID)
	secretKey := os.Getenv(ENV_NAME_TENCENTCLOUD_ACCESS_KEY)

	decryptSecretKey, err := encrypt.DesDecryptFromBase([]byte(secretKey))
	if err != nil {
		blog.Errorf("descrpt access secret key failed, err %s", err.Error())
		return fmt.Errorf("descrpt access secret key failed, err %s", err.Error())
	}
	c.SecretKey = string(decryptSecretKey)

	c.VpcDomain = os.Getenv(ENV_NAME_TENCENTCLOUD_VPC_DOMAIN)
	c.CvmDomain = os.Getenv(ENV_NAME_TENCENTCLOUD_CVM_DOMAIN)

	return nil
}

func (c *Client) validate() error {
	if len(c.Region) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_TENCENTCLOUD_REGION)
	}
	if len(c.VpcID) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_TENCENTCLOUD_VPC)
	}
	if len(c.SecretID) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_TENCENTCLOUD_ACCESS_KEY_ID)
	}
	if len(c.SecretKey) == 0 {
		return fmt.Errorf("%s cannot be empty", ENV_NAME_TENCENTCLOUD_ACCESS_KEY)
	}
	return nil
}

// GetENILimit get eni limit
func (c *Client) GetENILimit(instanceIP string) (int, int, error) {
	ins, err := c.queryInstance(instanceIP)
	if err != nil {
		return -1, -1, err
	}
	cores := *ins.CPU
	mem := *ins.Memory

	eniNum := getMaxENINumPerCVM(int(cores), int(mem))
	ipNum := getMaxPrivateIPNumPerENI(int(cores), int(mem))

	return eniNum, ipNum, nil
}

// Init client
func (c *Client) Init() error {

	if err := c.loadEnv(); err != nil {
		return err
	}

	if err := c.validate(); err != nil {
		return err
	}

	credential := common.NewCredential(
		c.SecretID,
		c.SecretKey,
	)
	cpfVpc := profile.NewClientProfile()
	if len(c.VpcDomain) != 0 {
		cpfVpc.HttpProfile.Endpoint = c.VpcDomain
	}
	vpcClient, err := vpc.NewClient(credential, c.Region, cpfVpc)
	if err != nil {
		blog.Errorf("new vpc client failed, err %s", err.Error())
		return fmt.Errorf("new vpc client failed, err %s", err.Error())
	}

	cpfCvm := profile.NewClientProfile()
	if len(c.CvmDomain) != 0 {
		cpfCvm.HttpProfile.Endpoint = c.CvmDomain
	}
	cvmClient, err := cvm.NewClient(credential, c.Region, cpfCvm)
	if err != nil {
		blog.Errorf("new cvm client failed, err %s", err.Error())
		return fmt.Errorf("new cvm client failed, err %s", err.Error())
	}

	c.vpcClient = vpcClient
	c.cvmClient = cvmClient

	return nil
}

// GetVMInfo get vm info
func (c *Client) GetVMInfo(instanceIP string) (*cloud.VMInfo, error) {
	ins, err := c.queryInstance(instanceIP)
	if err != nil {
		return nil, err
	}

	zone := *ins.Placement.Zone
	vpcID := *ins.VirtualPrivateCloud.VpcId
	subnetID := *ins.VirtualPrivateCloud.SubnetId
	instanceID := *ins.InstanceId
	return &cloud.VMInfo{
		NodeRegion:   c.Region,
		NodeZone:     zone,
		NodeVpcID:    vpcID,
		NodeSubnetID: subnetID,
		InstanceID:   instanceID,
		InstanceIP:   instanceIP,
	}, nil
}

// GetMaxENIIndex get max eni index, only for AWS attach eni, no need for tencent cloud
func (c *Client) GetMaxENIIndex(instanceIP string) (int, error) {
	// do nothing
	return 0, nil
}

// QueryENI query eni
func (c *Client) QueryENI(eniID string) (*cloud.ElasticNetworkInterface, error) {
	ifaceSet, err := c.queryENI(eniID, "", "")
	if err != nil {
		return nil, err
	}
	if len(ifaceSet) == 0 {
		return nil, ctrlcloud.ErrEniNotFound
	}
	if len(ifaceSet) != 1 {
		return nil, fmt.Errorf("found more than 1 eni with id %s", eniID)
	}
	iface := ifaceSet[0]
	netIf := &cloud.ElasticNetworkInterface{}
	netIf.EniName = *iface.NetworkInterfaceName
	netIf.EniSubnetID = *iface.SubnetId
	netIf.EniID = *iface.NetworkInterfaceId
	netIf.MacAddress = *iface.MacAddress

	if iface.Attachment != nil {
		netIf.Attachment = &cloud.NetworkInterfaceAttachment{
			Index:      int(*iface.Attachment.DeviceIndex),
			InstanceID: *iface.Attachment.InstanceId,
			EniID:      *iface.NetworkInterfaceId,
		}
	}

	// PrivateIpAddress in response contains both primary ip and secondary ips
	for _, ip := range iface.PrivateIpAddressSet {
		if *ip.Primary {
			netIf.Address = &cloud.IPAddress{
				IP:        *ip.PrivateIpAddress,
				IsPrimary: true,
			}
		} else {
			netIf.SecondaryAddresses = append(netIf.SecondaryAddresses, &cloud.IPAddress{
				IP:        *ip.PrivateIpAddress,
				IsPrimary: false,
			})
		}
	}
	return netIf, nil
}

// CreateENI create eni
func (c *Client) CreateENI(name, subnetID, addr string, ipNum int) (*cloud.ElasticNetworkInterface, error) {
	// query existed eni with certain name, if it is existed, reuse it
	ifaceSet, err := c.queryENI("", "", name)
	if err != nil {
		return nil, err
	}
	var iface *vpc.NetworkInterface
	if len(ifaceSet) == 0 {
		iface, err = c.createEni(name, subnetID, addr, ipNum)
		if err != nil {
			return nil, fmt.Errorf("createEni faile, err %s", err.Error())
		}

	} else if len(ifaceSet) != 1 {
		return nil, fmt.Errorf("found more than 1 eni named %s", name)

	} else {
		iface = ifaceSet[0]
		// if primary ip assigned, check it
		if len(addr) != 0 {
			for _, ip := range iface.PrivateIpAddressSet {
				if *ip.Primary && *ip.PrivateIpAddress != addr {
					return nil, fmt.Errorf("expect eni %s primary is %s but get %s", name, addr, *ip.PrivateIpAddress)
				}
			}
		}
		if len(iface.PrivateIpAddressSet)-1 < ipNum {
			err = c.assignIPsToEni(*iface.NetworkInterfaceId, ipNum-(len(iface.PrivateIpAddressSet)-1))
			if err != nil {
				return nil, fmt.Errorf("assign ip to %s failed, err %s", name, err.Error())
			}
		}

	}

	// wait for eni available
	err = c.waitForAvailable(*iface.NetworkInterfaceId, DEFAULT_CHECK_NUM, DEFAULT_CHECK_INTERVAL)
	if err != nil {
		return nil, fmt.Errorf("wait for available failed, err %s", err.Error())
	}

	subnet, err := c.querySubnet(*iface.SubnetId)
	if err != nil {
		return nil, fmt.Errorf("querySubnet failed, err %s", err.Error())
	}

	netIf := &cloud.ElasticNetworkInterface{}
	netIf.EniName = name
	netIf.EniSubnetID = *iface.SubnetId
	netIf.EniSubnetCidr = *subnet.CidrBlock
	netIf.IPNum = ipNum
	netIf.EniID = *iface.NetworkInterfaceId
	netIf.MacAddress = *iface.MacAddress

	if iface.Attachment != nil {
		netIf.Attachment = &cloud.NetworkInterfaceAttachment{
			Index:      int(*iface.Attachment.DeviceIndex),
			InstanceID: *iface.Attachment.InstanceId,
			EniID:      *iface.NetworkInterfaceId,
		}
	}

	// PrivateIpAddress in response contains both primary ip and secondary ips
	for _, ip := range iface.PrivateIpAddressSet {
		if *ip.Primary {
			netIf.Address = &cloud.IPAddress{
				IP:        *ip.PrivateIpAddress,
				IsPrimary: true,
			}
		} else {
			netIf.SecondaryAddresses = append(netIf.SecondaryAddresses, &cloud.IPAddress{
				IP:        *ip.PrivateIpAddress,
				IsPrimary: false,
			})
		}
	}

	return netIf, nil
}

// AttachENI attach eni
// [Attention] index no need for tencent cloud
func (c *Client) AttachENI(index int, eniID, instanceID, eniMac string) (*cloud.NetworkInterfaceAttachment, error) {
	err := c.attachENI(eniID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("attachENI failed, err %s", err.Error())
	}

	// wait for eni attached
	err = c.waitForAttached(eniID, DEFAULT_CHECK_NUM, DEFAULT_CHECK_INTERVAL)
	if err != nil {
		blog.Errorf("wait eni attached failed, err %s", err.Error())
		return nil, fmt.Errorf("wait for eni attached failed, err %s", err.Error())
	}

	return &cloud.NetworkInterfaceAttachment{
		EniID:      eniID,
		InstanceID: instanceID,
	}, nil
}

// DetachENI detach eni
func (c *Client) DetachENI(attachment *cloud.NetworkInterfaceAttachment) error {
	err := c.detachENI(attachment.EniID, attachment.InstanceID)
	if err != nil {
		return err
	}

	// wait for eni detached
	err = c.waitForDetached(attachment.EniID, DEFAULT_CHECK_NUM, DEFAULT_CHECK_INTERVAL)
	if err != nil {
		blog.Errorf("wait eni detached failed, err %s", err.Error())
		return fmt.Errorf("wait eni detached failed, err %s", err.Error())
	}

	return nil
}

// DeleteENI delete eni
func (c *Client) DeleteENI(eniID string) error {
	return c.deleteEni(eniID)
}
