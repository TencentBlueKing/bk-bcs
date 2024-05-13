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

package eip

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/qcloud-eip/conf"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const (
	// ENIStatusPending eni pending status
	ENIStatusPending = "PENDING"
	// ENIStatusAvailable eni available status
	ENIStatusAvailable = "AVAILABLE"
	// ENIStatusAttaching eni attaching status
	ENIStatusAttaching = "ATTACHING"
	// ENIStatusDetaching eni detaching status
	ENIStatusDetaching = "DETACHING"
	// ENIStatusDeleting eni deleting status
	ENIStatusDeleting = "DELETING"

	// status determined by eni Attachment field

	// ENIStatusDetached eni detached status
	ENIStatusDetached = "DETACHED"
	// ENIStatusAttached eni attached status
	ENIStatusAttached = "ATTACHED"
)

type vpcClient struct {
	conf   *conf.NetConf
	vpcID  string
	client *vpc.Client
}

func newVPCClient(conf *conf.NetConf, vpcID string) *vpcClient {
	credential := common.NewCredential(
		conf.Secret,
		conf.UUID,
	)
	cpf := profile.NewClientProfile()

	// set tencentcloud domain
	if len(conf.TencentCloudVPCDomain) != 0 {
		cpf.HttpProfile.Endpoint = conf.TencentCloudVPCDomain
	}

	client, err := vpc.NewClient(credential, conf.Region, cpf)
	if err != nil {
		blog.Errorf("new vpc client failed, err %s", err.Error())
		return nil
	}
	return &vpcClient{
		conf:   conf,
		vpcID:  vpcID,
		client: client,
	}
}

// TakeOverENI take over network interface if the interface with name ifname is existed
// create a new network interface if the interface with name ifname is not existed
func (vc *vpcClient) TakeOverENI(instanceID string, privateIPNum uint64, ifname string) (*vpc.NetworkInterface, error) {
	enis, err := vc.queryENI("", "", ifname)
	if err != nil {
		blog.Warnf("query eni by interface-name %s failed, err %s", ifname, err.Error())
	}
	if len(enis) == 0 {
		// create, attach new eni
		// 1. create
		// 2. wait for available
		// 3. attach
		// 4. wait for attached
		blog.Infof("get no eni named %s", ifname)
		eniInterface, err := vc.createENI(instanceID, privateIPNum, ifname)
		if err != nil {
			blog.Errorf("create eni for ins %s with %d ips with name %s failed, err %s", instanceID, privateIPNum, ifname, err.Error())
			return nil, fmt.Errorf("create eni for ins %s with %d ips with name %s failed, err %s", instanceID, privateIPNum, ifname, err.Error())
		}
		err = vc.waitForAvailable(*eniInterface.NetworkInterfaceId, 5, 5)
		if err != nil {
			blog.Errorf("wait for created eni available failed, err %s", err.Error())
			return nil, fmt.Errorf("wait for created eni available failed, err %s", err.Error())
		}
		blog.Infof("attach eni %s to ins %s", *eniInterface.NetworkInterfaceId, instanceID)
		err = vc.attachENI(*eniInterface.NetworkInterfaceId, instanceID)
		if err != nil {
			blog.Errorf("attach eni %s to instance %s failed, err %s", *eniInterface.NetworkInterfaceId, instanceID, err.Error())
			return nil, fmt.Errorf("attach eni %s to instance %s failed, err %s", *eniInterface.NetworkInterfaceId, instanceID, err.Error())
		}
		err = vc.waitForAttached(*eniInterface.NetworkInterfaceId, 5, 5)
		if err != nil {
			blog.Errorf("wait for attached eni available failed, err %s", err.Error())
			return nil, fmt.Errorf("wait for attached eni available failed, err %s", err.Error())
		}
		blog.Infof("attach done")
		return eniInterface, nil
	}
	existedENI := enis[0]
	// existed but not attached
	if existedENI.Attachment == nil || existedENI.Attachment.InstanceId == nil {
		err := vc.attachENI(*existedENI.NetworkInterfaceId, instanceID)
		if err != nil {
			blog.Errorf("attach eni %s with ins %s failed, err %s", *existedENI.NetworkInterfaceId, instanceID, err.Error())
			return nil, fmt.Errorf("attach eni %s with ins %s failed, err %s", *existedENI.NetworkInterfaceId, instanceID, err.Error())
		}
		err = vc.waitForAvailable(*existedENI.NetworkInterfaceId, 5, 5)
		if err != nil {
			blog.Errorf("wait for eni %s available failed, err %s", *existedENI.NetworkInterfaceId, err.Error())
			return nil, fmt.Errorf("wait for eni %s available failed, err %s", *existedENI.NetworkInterfaceId, err.Error())
		}
	} else {
		// existed and binded
		if *existedENI.Attachment.InstanceId != instanceID {
			blog.Errorf("eni %s existed, already bind to ins %s, expected %s",
				*existedENI.NetworkInterfaceId,
				*existedENI.Attachment.InstanceId,
				instanceID)
			return nil, fmt.Errorf("eni %s existed, already bind to ins %s, expected %s",
				*existedENI.NetworkInterfaceId,
				*existedENI.Attachment.InstanceId,
				instanceID)
		}
		err = vc.waitForAvailable(*existedENI.NetworkInterfaceId, 5, 5)
		if err != nil {
			blog.Errorf("wait for eni %s available failed, err %s", *existedENI.NetworkInterfaceId, err.Error())
			return nil, fmt.Errorf("wait for eni %s available failed, err %s", *existedENI.NetworkInterfaceId, err.Error())
		}
	}
	// apply enough ips
	if len(existedENI.PrivateIpAddressSet) < int(privateIPNum+1) {
		_, err := vc.applyIPForENI(*existedENI.NetworkInterfaceId, int(privateIPNum+1)-len(existedENI.PrivateIpAddressSet))
		if err != nil {
			return nil, err
		}
		blog.Infof("apply ip for eni %s successfully", *existedENI.NetworkInterfaceId)
		enisAfterApplyIP, err := vc.queryENI("", "", ifname)
		if err != nil {
			blog.Errorf("query eni by interface-name %s failed, err %s", ifname, err.Error())
			return nil, fmt.Errorf("query eni by interface-name %s failed, err %s", ifname, err.Error())
		}
		existedENI = enisAfterApplyIP[0]
	}
	return existedENI, nil
}

// applyIPForENI apply extra private ip for eni
func (vc *vpcClient) applyIPForENI(eniID string, ipNum int) ([]*vpc.PrivateIpAddressSpecification, error) {
	request := vpc.NewAssignPrivateIpAddressesRequest()
	request.NetworkInterfaceId = common.StringPtr(eniID)
	request.SecondaryPrivateIpAddressCount = common.Uint64Ptr(uint64(ipNum))

	blog.V(3).Infof("apply ip for eni request:\n%s", request.ToJsonString())
	response, err := vc.client.AssignPrivateIpAddresses(request)
	if err != nil {
		blog.Errorf("apply ip for eni failed, err %s", err.Error())
		return nil, fmt.Errorf("apply ip for eni failed, err %s", err.Error())
	}
	blog.V(3).Infof("apply ip for eni response:\n%s", response.ToJsonString())
	return response.Response.PrivateIpAddressSet, nil
}

// createENI create eni with certain name
func (vc *vpcClient) createENI(instanceID string, privateIPNum uint64, ifname string) (*vpc.NetworkInterface, error) {
	request := vpc.NewCreateNetworkInterfaceRequest()
	request.VpcId = common.StringPtr(vc.vpcID)
	request.SubnetId = common.StringPtr(vc.conf.SubnetID)
	request.NetworkInterfaceName = common.StringPtr(ifname)
	request.SecondaryPrivateIpAddressCount = common.Uint64Ptr(privateIPNum)

	blog.V(3).Infof("create eni request:\n%s", request.ToJsonString())
	response, err := vc.client.CreateNetworkInterface(request)
	if err != nil {
		blog.Errorf("create network interface failed, err %s", err.Error())
		return nil, fmt.Errorf("create network interface failed, err %s", err.Error())
	}
	blog.V(3).Infof("create eni response:\n%s", response.ToJsonString())
	return response.Response.NetworkInterface, nil
}

// queryENI query eni, support query by eniID and eniName
// match eniID if eniID is not empty
// match eniName if eniName is not empty
func (vc *vpcClient) queryENI(eniID string, instanceID string, eniName string) ([]*vpc.NetworkInterface, error) {
	request := vpc.NewDescribeNetworkInterfacesRequest()
	request.Filters = make([]*vpc.Filter, 0)
	if len(eniID) != 0 {
		request.Filters = append(request.Filters, &vpc.Filter{
			Name: common.StringPtr("network-interface-id"),
			Values: []*string{
				common.StringPtr(eniID),
			},
		})
	}
	if len(instanceID) != 0 {
		request.Filters = append(request.Filters, &vpc.Filter{
			Name: common.StringPtr("attachment.instance-id"),
			Values: []*string{
				common.StringPtr(instanceID),
			},
		})
	}
	if len(eniName) != 0 {
		request.Filters = append(request.Filters, &vpc.Filter{
			Name: common.StringPtr("network-interface-name"),
			Values: []*string{
				common.StringPtr(eniName),
			},
		})
	}
	blog.V(3).Infof("describe enis request:\n%s", request.ToJsonString())
	response, err := vc.client.DescribeNetworkInterfaces(request)
	if err != nil {
		blog.Errorf("describe enis by id %s failed, err %s", eniID, err.Error())
		return nil, fmt.Errorf("describe enis by id %s failed, err %s", eniID, err.Error())
	}
	blog.V(3).Infof("describe enis response:\n%s", response.ToJsonString())

	if *(response.Response.TotalCount) == 0 {
		blog.Warnf("describe enis by id %s return zero result", eniID)
		return nil, nil
	}
	return response.Response.NetworkInterfaceSet, nil
}

// attachENI attach eni to cvm
func (vc *vpcClient) attachENI(eniID string, instanceID string) error {
	request := vpc.NewAttachNetworkInterfaceRequest()
	request.NetworkInterfaceId = common.StringPtr(eniID)
	request.InstanceId = common.StringPtr(instanceID)
	blog.V(3).Infof("attach request:\n%s", request.ToJsonString())
	response, err := vc.client.AttachNetworkInterface(request)
	if err != nil {
		blog.Errorf("attach eni %s to ins %s failed, err %s", eniID, instanceID, err.Error())
		return fmt.Errorf("attach eni %s to ins %s failed, err %s", eniID, instanceID, err.Error())
	}
	blog.V(3).Infof("attach response:\n%s", response.ToJsonString())
	return nil
}

// queryENIbyIP
func (vc *vpcClient) queryENIbyIP(eniIP string, instanceID string) (*vpc.NetworkInterface, error) {
	request := vpc.NewDescribeNetworkInterfacesRequest()
	request.Filters = make([]*vpc.Filter, 0)
	if len(instanceID) == 0 {
		blog.Errorf("query eni need instance id, but instance id is empty")
		return nil, fmt.Errorf("query eni need instance id, but instance id is empty")
	}
	request.Filters = append(request.Filters, &vpc.Filter{
		Name: common.StringPtr("attachment.instance-id"),
		Values: []*string{
			common.StringPtr(instanceID),
		},
	})
	blog.V(3).Infof("describe enis request:\n%s", request.ToJsonString())
	response, err := vc.client.DescribeNetworkInterfaces(request)
	if err != nil {
		blog.Errorf("describe enis by ip %s failed, err %s", eniIP, err.Error())
		return nil, fmt.Errorf("describe enis by ip %s failed, err %s", eniIP, err.Error())
	}
	blog.V(3).Infof("describe enis response:\n%s", response.ToJsonString())

	if *(response.Response.TotalCount) == 0 {
		blog.Warnf("describe enis by ip %s return zero result", eniIP)
		return nil, nil
	}
	for _, netif := range response.Response.NetworkInterfaceSet {
		for _, ip := range netif.PrivateIpAddressSet {
			if eniIP == *ip.PrivateIpAddress && *ip.Primary {
				return netif, nil
			}
		}
	}
	blog.Errorf("get no eni with ip %s and instanceid %s", eniIP, instanceID)
	return nil, fmt.Errorf("get no eni with ip %s and instanceid %s", eniIP, instanceID)
}

// detachENI delete eni by eniID
func (vc *vpcClient) detachENI(eniID string, instanceID string) error {
	request := vpc.NewDetachNetworkInterfaceRequest()
	request.NetworkInterfaceId = common.StringPtr(eniID)
	request.InstanceId = common.StringPtr(instanceID)
	blog.V(3).Infof("detach request:\n%s", request.ToJsonString())
	response, err := vc.client.DetachNetworkInterface(request)
	if err != nil {
		blog.Errorf("detach eni %s to ins %s failed, err %s", eniID, instanceID, err.Error())
		return fmt.Errorf("detach eni %s to ins %s failed, err %s", eniID, instanceID, err.Error())
	}
	blog.V(3).Infof("detach response:\n%s", response.ToJsonString())
	return nil
}

// deleteENI
func (vc *vpcClient) deleteENI(eniID string) error {
	request := vpc.NewDeleteNetworkInterfaceRequest()
	request.NetworkInterfaceId = common.StringPtr(eniID)
	blog.V(3).Infof("delete request:\n%s", request.ToJsonString())
	response, err := vc.client.DeleteNetworkInterface(request)
	if err != nil {
		blog.Errorf("delete eni %s failed, err %s", eniID, err.Error())
		return fmt.Errorf("delete eni %s failed, err %s", eniID, err.Error())
	}
	blog.V(3).Infof("delete response:\n%s", response.ToJsonString())
	return nil
}

func (vc *vpcClient) waitForAttached(eniID string, checkNum, checkInterval int) error {
	return vc.doWaitForStatus(eniID, checkNum, checkInterval, ENIStatusAttached)
}

func (vc *vpcClient) waitForDetached(eniID string, checkNum, checkInterval int) error {
	return vc.doWaitForStatus(eniID, checkNum, checkInterval, ENIStatusDetached)
}

// waitForAvailable
func (vc *vpcClient) waitForAvailable(eniID string, checkNum, checkInterval int) error {
	return vc.doWaitForStatus(eniID, checkNum, checkInterval, ENIStatusAvailable)
}

func (vc *vpcClient) doWaitForStatus(eniID string, checkNum, checkInterval int, finalStatus string) error {
	for i := 0; i < checkNum; i++ {
		time.Sleep(time.Second * time.Duration(checkInterval))
		enis, err := vc.queryENI(eniID, "", "")
		if err != nil {
			return err
		}
		for _, eni := range enis {
			if *eni.NetworkInterfaceId == eniID {
				switch *eni.State {
				case ENIStatusAvailable:
					switch finalStatus {
					case ENIStatusAttached:
						if eni.Attachment != nil && eni.Attachment.InstanceId != nil {
							blog.Infof("eni %s is attached", eniID)
							return nil
						}
						blog.Infof("eni %s is not attached", eniID)
					case ENIStatusDetached:
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
				case ENIStatusPending, ENIStatusAttaching, ENIStatusDetaching, ENIStatusDeleting:
					blog.Infof("eni %s is %s", eniID, *eni.State)
					break
				}
			}
		}
	}
	blog.Errorf("timeout when wait for eni %s", eniID)
	return fmt.Errorf("timeout when wait for eni %s", eniID)
}

// querySubnet
func (vc *vpcClient) querySubnet(subnetID string) (*vpc.Subnet, error) {
	request := vpc.NewDescribeSubnetsRequest()
	request.SubnetIds = []*string{
		common.StringPtr(subnetID),
	}
	blog.V(3).Infof("describe subnet request:\n%s", request.ToJsonString())
	response, err := vc.client.DescribeSubnets(request)
	if err != nil {
		blog.Errorf("describe subnet by id %s failed, err %s", subnetID, err.Error())
		return nil, fmt.Errorf("describe subnet by id %s failed, err %s", subnetID, err.Error())
	}
	blog.V(3).Infof("describe subnet response:\n%s", response.ToJsonString())
	if *response.Response.TotalCount == 0 {
		blog.Errorf("describe subnet by id %s return zero result", subnetID)
		return nil, fmt.Errorf("describe subnet by id %s return zero result", subnetID)
	}
	if len(response.Response.SubnetSet) != 1 {
		blog.Warnf("describe subnet by id %s return result more than 1", subnetID)

	}
	return response.Response.SubnetSet[0], nil
}
