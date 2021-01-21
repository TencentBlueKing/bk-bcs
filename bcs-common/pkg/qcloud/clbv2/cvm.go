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

//CVMFilter filter for Describe cvm
//condition: https://cloud.tencent.com/document/api/213/9388
type CVMFilter struct {
	Name   string    `url:"Name"`
	Values GroupList `url:"Values"`
}

//CVMFilters define url interface for url encoding
type CVMFilters []CVMFilter

//EncodeValues interface for url encoding
func (fs CVMFilters) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range fs {
		subKey := fmt.Sprintf("%s.%d", key, i)
		name := fmt.Sprintf("%s.Name", subKey)
		urlv.Set(name, v.Name)
		value := fmt.Sprintf("%s.Values", subKey)
		v.Values.EncodeValues(value, urlv)
	}
	return nil
}

//DescribeCVM query qcloud vm info
type DescribeCVM struct {
	Action      string     `url:"Action"`
	Filters     CVMFilters `url:"Filters,omitempty"`
	InstanceIds GroupList  `url:"InstanceIds,omitempty"`
	Limit       int        `url:"Limit,omitempty"`
	Nonce       uint       `url:"Nonce"`
	Offset      int        `url:"Offset,omitempty"`
	Region      string     `url:"Region"`
	SecretID    string     `url:"SecretId"`
	Signature   string     `url:"Signature,omitempty"` //method is HmacSHA1
	Timestamp   uint       `url:"Timestamp"`
	Version     string     `url:"Version"` //version must be 2017-03-12
}

type DescribeResponse struct {
	Response CVMResponse `json:"Response"`
}

//CVMResponse content when DescribeCVM
type CVMResponse struct {
	TotalCnt    int         `json:"TotalCount"`
	InstanceSet []*Instance `json:"InstanceSet"`
	RequestID   string      `json:"RequestId"`
}

//PlaceMent section of host
type PlaceMent struct {
	Zone      string `json:"Zone"`
	HostID    string `json:"HostId,omitempty"`
	ProjectID int    `json:"ProjectId,omitempty"`
}

//VPCRef VirtualPrivateCloud referrence
type VPCRef struct {
	VpcID              string    `json:"VpcId"`
	SubnetID           string    `json:"SubnetId"`
	AsVpcGateway       bool      `json:"AsVpcGateway,omitempty"`
	PrivateIPAddresses GroupList `json:"PrivateIpAddresses,omitempty"`
}

//Instance for host
type Instance struct {
	Placement          PlaceMent `json:"Placement,omitempty"`
	InstanceID         string    `json:"InstanceId"`
	InstanceName       string    `json:"InstanceName"`
	InstanceType       string    `json:"InstanceType"`
	InstanceChargeType string    `json:"InstanceChargeType"`
	CPU                int       `json:"CPU"`
	Memory             int       `json:"Memory"`
	PrivateIPAddresses GroupList `json:"PrivateIpAddresses"`
	PublicIPAddresses  GroupList `json:"PublicIpAddresses,omitempty"`
	VPC                VPCRef    `json:"VirtualPrivateCloud"`
	CreatedTime        string    `json:"CreatedTime"`
	ExpiredTime        string    `json:"ExpiredTime"`
}

/////////////////////////moved from clb.go///////////////////////

//DescribeCVMInstanceInput describe cvm instance info
type DescribeCVMInstanceInput struct {
	APIMeta `url:",inline"`
	LanIP   string `url:"lanIps.0,omitempty"`
}

//DescribeCVMInstanceOutput describe cvm instance response
type DescribeCVMInstanceOutput struct {
	Response  `json:",inline"`
	Instances []CVMInstanceInfo `json:"instanceSet"`
}

//CVMInstanceInfo one cvm instance info
type CVMInstanceInfo struct {
	InstanceName string `json:"instanceName"`
	LanIP        string `json:"lanIp"`
	InstanceID   string `json:"unInstanceId"`
	SubnetID     string `json:"subnetId"`
	ProjectID    int    `json:"projectId"`
	Region       string `json:"Region"`
}

//DescribeCVMInstanceV3Input describe cvm instance info from v3 api
type DescribeCVMInstanceV3Input struct {
	APIMeta        `url:",inline"`
	FilterIPName   string `url:"Filters.0.Name,omitempty"`
	FilterIPValues FilterIPValueFieldList
	Version        string `url:"Version,omitempty"`
}

//FilterIPValueField filter ip field
type FilterIPValueField struct {
	IP string
}

//FilterIPValueFieldList filter ip field list
type FilterIPValueFieldList []FilterIPValueField

//EncodeValues encode filter ip field info into url format
func (ipList FilterIPValueFieldList) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range ipList {
		urlv.Set(fmt.Sprintf("Filters.0.Values.%d", i), fmt.Sprintf("%s", v.IP))
	}
	return nil
}

//DescribeCVMInstanceV3Output describe cvm instance result from v3 api
type DescribeCVMInstanceV3Output struct {
	CVMInfos CVMInstanceV3InfoList `json:"Response"`
}

//CVMInstanceV3InfoList cvm info list
type CVMInstanceV3InfoList struct {
	TotalCount int                 `json:"TotalCount"`
	CVMInfo    []CVMInstanceV3Info `json:"InstanceSet"`
}

//CVMInstanceV3Info cvm info
type CVMInstanceV3Info struct {
	InstanceName     string        `json:"InstanceName"`
	InstanceID       string        `json:"InstanceId"`
	PlaceInfo        PlacementInfo `json:"Placement"`
	SecurityGroupIds []string      `json:"SecurityGroupIds"`
}

//PlacementInfo cvm location info
type PlacementInfo struct {
	Zone      string `json:"Zone"`
	HostId    string `json:"HostId"`
	ProjectId int    `json:"ProjectId"`
}

//VirtualPrivateCloudInfo virtual network info about cvm instance
type VirtualPrivateCloudInfo struct {
	VpcId        string `json:"VpcId"`
	SubnetId     string `json:"SubnetId"`
	AsVpcGateway bool   `json:"AsVpcGateway"`
}
