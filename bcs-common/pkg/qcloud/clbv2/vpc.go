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
 *
 */

package qcloud

import (
	"fmt"
	"net/url"
)

/*
Network Interface object
url: https://vpc.api.qcloud.com/v2/index.php
*/

//PrivateIPAddressSet address for New NIC application
type PrivateIPAddressSet struct {
	Primary          bool   `url:"primary" json:"primary"`
	PrivateIPAddress string `url:"privateIpAddress" json:"privateIpAddress"`
	Description      string `url:"description,omitempty" json:"description,omitempty"`
	WanIp            string `url:"wanIp,omitempty" json:"wanIp,omitempty"`
	IsWanIpBlocked   bool   `url:"isWanIpBlocked,omitempty" json:"isWanIpBlocked,omitempty"`
	EipId            string `url:"eipId,omitempty" json:"eipId,omitempty"`
}

//IPSet for ip address list url encoding
type IPSet []PrivateIPAddressSet

//EncodeValues interface for url encoding
func (ips IPSet) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range ips {
		primary := fmt.Sprintf("%s.%d.primary", key, i)
		urlv.Set(primary, fmt.Sprintf("%v", v.Primary))
		ipAddr := fmt.Sprintf("%s.%d.privateIpAddress", key, i)
		urlv.Set(ipAddr, v.PrivateIPAddress)
	}
	return nil
}

//CreateNIC object for qcloud api to create new network interface
type CreateNIC struct {
	APIMeta                        `url:",inline"`
	EniDescription                 string    `url:"eniDescription,omitempty"`
	EniName                        string    `url:"eniName"`
	IPAddressSet                   IPSet     `url:"privateIpAddressSet,omitempty"`
	SecondaryPrivateIPAddressCount int       `url:"secondaryPrivateIpAddressCount,omitempty"`
	SgIds                          GroupList `url:"sgIds,omitempty"`
	SubnetID                       string    `url:"subnetId"`
	VpcID                          string    `url:"vpcId"`
}

//AssignPrivateIPAddr assign more ip address to network interface
type AssignPrivateIPAddr struct {
	APIMeta                        `url:",inline"`
	VpcID                          string `url:"vpcId"`
	NetworkInterfaceID             string `url:"networkInterfaceId"`
	SecondaryPrivateIPAddressCount int    `url:"secondaryPrivateIpAddressCount"`
}

//ModifyNIC object for modifing via api
type ModifyNIC struct {
	EniName            string `url:"eniName,omitempty"`
	EniDescription     string `url:"eniDescription,omitempty"`
	NetworkInterfaceID string `url:"networkInterfaceId"`
	VpcID              string `url:"vpcId"`
}

//DescribeNIC query NIC info in cloud host
type DescribeNIC struct {
	APIMeta            `url:",inline"`
	EniDescription     string `url:"eniDescription,omitempty"`
	EniName            string `url:"eniName,omitempty"`
	InstanceID         string `url:"instanceId,omitempty"`
	Limit              int    `url:"limit,omitempty"`
	NetworkInterfaceID string `url:"networkInterfaceId,omitempty"`
	Offset             int    `url:"offset,omitempty"`
	OrderDirection     string `url:"orderDirection,omitempty"`
	OrderField         string `url:"orderField,omitempty"`
	VpcID              string `url:"vpcId,omitempty"`
}

//NICAttachHost attach one nic to host
type NICAttachHost struct {
	APIMeta            `url:",inline"`
	InstanceID         string `url:"instanceId"`
	NetworkInterfaceID string `url:"networkInterfaceId"`
	VpcID              string `url:"vpcId"`
}

//InstanceRef info for host
type InstanceRef struct {
	InstanceID string `json:"instanceId"`
	AttachTime string `json:"attachTime"`
}

//SecureGroup group
type SecureGroup struct {
	SgID      string `json:"sgId"`
	SgName    string `json:"sgName"`
	ProjectID int    `json:"projectId"`
}

//NIC network interface card
type NIC struct {
	VpcID              string        `json:"vpcId"`
	VpcName            string        `json:"vpcName,omitempty"`
	SubnetID           string        `json:"subnetId"`
	ZoneID             int           `json:"zoneId,omitempty"`
	EniName            string        `json:"eniName"`
	EniDescription     string        `json:"eniDescription,omitempty"`
	NetworkInterfaceID string        `json:"networkInterfaceId"`
	Primary            bool          `json:"primary,omitempty"`
	MacAddress         string        `json:"macAddress"`
	IPAddressSet       IPSet         `json:"privateIpAddressesSet,omitempty"`
	InstanceSet        InstanceRef   `json:"instanceSet,omitempty"`
	GroupSet           []SecureGroup `json:"groupSet,omitempty"`
}

//NICResponse response for DescribeNetworkInterface
type NICResponse struct {
	Response `json:",inline"`
	Data     NICList `json:"data"`
}

//NICList list of nic
type NICList struct {
	TotalCnt int    `json:"totalNum"`
	Data     []*NIC `json:"data"`
}
