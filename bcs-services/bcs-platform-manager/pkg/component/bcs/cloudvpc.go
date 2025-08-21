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

// Package bcs cloudvpc操作
package bcs

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// CidrState cidr state
type CidrState struct {
	Cidr  string `json:"cidr"`
	Block bool   `json:"block"`
}

// Cidr cidr
type Cidr struct {
	Cidrs         []*CidrState `json:"cidrs"`
	ReservedIPNum uint32       `json:"reservedIPNum"`
	ReservedCidrs []string     `json:"reservedCidrs"`
}

// CloudVPC for cloud vpc
type CloudVPC struct {
	CloudID       string `json:"cloudID"`
	Region        string `json:"region"`
	RegionName    string `json:"regionName"`
	NetworkType   string `json:"networkType"`
	VpcID         string `json:"vpcID"`
	VpcName       string `json:"vpcName"`
	Available     string `json:"available"`
	Extra         string `json:"extra"`
	Creator       string `json:"creator"`
	Updater       string `json:"updater"`
	CreatTime     string `json:"creatTime"`
	UpdateTime    string `json:"updateTime"`
	ReservedIPNum uint32 `json:"reservedIPNum"`
	BusinessID    string `json:"businessID"`
	Overlay       *Cidr  `json:"overlay"`
	Underlay      *Cidr  `json:"underlay"`
}

// CreateCloudVPCReq create cloud vpc request
type CreateCloudVPCReq struct {
	CloudID       string `json:"cloudID" validate:"required"`
	NetworkType   string `json:"networkType"`
	Region        string `json:"region"`
	RegionName    string `json:"regionName"`
	VpcName       string `json:"vpcName"`
	VpcID         string `json:"vpcID" validate:"required"`
	Available     string `json:"available"`
	Extra         string `json:"extra"`
	Creator       string `json:"creator"`
	ReservedIPNum uint32 `json:"reservedIPNum"`
	BusinessID    string `json:"businessID"`
	Overlay       *Cidr  `json:"overlay"`
	Underlay      *Cidr  `json:"underlay"`
}

// UpdateCloudVPCReq update cloud vpc request
type UpdateCloudVPCReq struct {
	CloudID       string `json:"cloudID" validate:"required"`
	NetworkType   string `json:"networkType"`
	Region        string `json:"region"`
	RegionName    string `json:"regionName"`
	VpcName       string `json:"vpcName"`
	VpcID         string `json:"vpcID" validate:"required"`
	Available     string `json:"available"`
	Updater       string `json:"updater"`
	ReservedIPNum uint32 `json:"reservedIPNum"`
	BusinessID    string `json:"businessID"`
	Overlay       *Cidr  `json:"overlay"`
	Underlay      *Cidr  `json:"underlay"`
}

// ListCloudVPC 获取cloud vpc列表
func ListCloudVPC(cloudID, region, vpcID, networkType, businessID string) ([]*CloudVPC, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cloudvpc", config.G.BCS.Host)

	queryParams := make(map[string]string)
	if cloudID != "" {
		queryParams["cloudID"] = cloudID
	}
	if region != "" {
		queryParams["region"] = region
	}
	if vpcID != "" {
		queryParams["vpcID"] = vpcID
	}
	if networkType != "" {
		queryParams["networkType"] = networkType
	}
	if businessID != "" {
		queryParams["businessID"] = businessID
	}

	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		SetQueryParams(queryParams).
		Get(url)

	if err != nil {
		blog.Errorf("list cloud vpc error, %s", err.Error())
		return nil, err
	}

	var result []*CloudVPC

	fmt.Printf("list cloud vpc response: %s", resp.String())
	if err = component.UnmarshalBKData(resp, &result); err != nil {
		blog.Errorf("unmarshal cloud vpc error, %s", err.Error())
		return nil, err
	}

	return result, nil
}

// CreateCloudVPC 创建cloud vpc
func CreateCloudVPC(vpc *CreateCloudVPCReq) (bool, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cloudvpc", config.G.BCS.Host)

	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		SetBody(vpc).
		Post(url)

	if err != nil {
		blog.Errorf("create cloud vpc error, %s", err.Error())
		return false, err
	}

	var result bool
	fmt.Printf("create cloud vpc response: %s", resp.String())
	if err = component.UnmarshalBKResult(resp, &result); err != nil {
		blog.Errorf("unmarshal cloud vpc error, %s", err.Error())
		return false, err
	}

	return result, nil
}

// UpdateCloudVPC 更新cloud vpc
func UpdateCloudVPC(req *UpdateCloudVPCReq) (*CloudVPC, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cloudvpc/%s/%s", config.G.BCS.Host, req.CloudID, req.VpcID)

	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		SetBody(req).
		Put(url)

	if err != nil {
		blog.Errorf("update cloud vpc error, %s", err.Error())
		return nil, err
	}

	var result *CloudVPC

	fmt.Printf("update cloud vpc response: %s", resp.String())
	if err = component.UnmarshalBKData(resp, result); err != nil {
		blog.Errorf("unmarshal cloud vpc error, %s", err.Error())
		return nil, err
	}

	return result, nil
}

// DeleteCloudVPC 删除cloud vpc
func DeleteCloudVPC(cloudID, vpcID string) (*CloudVPC, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cloudvpc/%s/%s", config.G.BCS.Host, cloudID, vpcID)

	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		Delete(url)

	if err != nil {
		blog.Errorf("delete cloud vpc error, %s", err.Error())
		return nil, err
	}

	var result *CloudVPC

	fmt.Printf("delete cloud vpc response: %s", resp.String())
	if err = component.UnmarshalBKData(resp, &result); err != nil {
		blog.Errorf("unmarshal cloud vpc error, %s", err.Error())
		return nil, err
	}

	return result, nil
}
