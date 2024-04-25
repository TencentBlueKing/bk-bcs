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

// Package api xxx
package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	vpc2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	modelv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	vpc3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3"
	modelv3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
	regionv3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// NewVpcClient get vpc client from common option
func NewVpcClient(opt *cloudprovider.CommonOption) (*VpcClient, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	projectID, err := GetProjectIDByRegion(opt)
	if err != nil {
		return nil, err
	}

	auth, err := getProjectAuth(opt.Account.SecretID, opt.Account.SecretKey, projectID)
	if err != nil {
		return nil, err
	}
	rn, err := regionv3.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	vpc2HcClient, err := vpc2.VpcClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	vpc3HcClient, err := vpc3.VpcClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &VpcClient{vpc2: vpc2.NewVpcClient(vpc2HcClient), vpc3: vpc3.NewVpcClient(vpc3HcClient)}, nil
}

// VpcClient client
type VpcClient struct {
	vpc2 *vpc2.VpcClient
	vpc3 *vpc3.VpcClient
}

// ListVpcs 获取账号下所有的vpc
func (v *VpcClient) ListVpcs(vpcIds []string) ([]modelv3.Vpc, error) {

	var (
		vpcs = make([]modelv3.Vpc, 0)
		// 分页参数
		limit  = int32(200) // 每页显示的记录数
		marker *string      // 分页标记，即下一页的起始位置
	)

	for {
		request := &modelv3.ListVpcsRequest{
			Limit:  &limit,
			Marker: marker,
			Id: func() *[]string {
				if len(vpcIds) == 0 {
					return nil
				}
				return &vpcIds
			}(),
		}
		response, errLocal := v.vpc3.ListVpcs(request)
		if errLocal != nil {
			blog.Infof("ListVpcs failed: %v", errLocal)
			return nil, errLocal
		}

		// 处理当前页的VPC列表
		for _, vpc := range *response.Vpcs {
			vpcs = append(vpcs, vpc)
		}

		// 如果没有更多的页，则退出循环
		if response.PageInfo.NextMarker == nil || *response.PageInfo.NextMarker == "" {
			break
		}
		marker = response.PageInfo.NextMarker // 准备获取下一页
	}

	return vpcs, nil
}

// ShowVpc 查询vpc详情
func (v *VpcClient) ShowVpc(vpcId string) (*modelv3.Vpc, error) {
	request := &modelv3.ShowVpcRequest{VpcId: vpcId}

	rsp, err := v.vpc3.ShowVpc(request)
	if err != nil {
		return nil, err
	}

	return rsp.Vpc, nil
}

// ListSubnets 获取账号下所有的subnets
func (v *VpcClient) ListSubnets(vpcId string) ([]modelv2.Subnet, error) {
	var (
		subnets = make([]modelv2.Subnet, 0)
		// 分页参数
		limit  = int32(2000) // 每页显示的记录数
		marker *string       // 分页标记，即下一页的起始位置
	)

	// 目前华为云v2接口不支持分页返回，只能查询最大limit的子网个数
	request := &modelv2.ListSubnetsRequest{
		Limit:  &limit,
		Marker: marker,
		VpcId: func() *string {
			if len(vpcId) > 0 {
				return &vpcId
			}
			return nil
		}(),
	}
	response, errLocal := v.vpc2.ListSubnets(request)
	if errLocal != nil {
		blog.Infof("ListSubnets failed: %v", errLocal)
		return nil, errLocal
	}

	// 处理当前页的VPC列表
	for _, subnet := range *response.Subnets {
		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

// ShowSubnet 查询subnet详情
func (v *VpcClient) ShowSubnet(subnetId string) (*modelv2.Subnet, error) {
	request := &modelv2.ShowSubnetRequest{SubnetId: subnetId}

	rsp, err := v.vpc2.ShowSubnet(request)
	if err != nil {
		return nil, err
	}

	return rsp.Subnet, nil
}

// ShowNetworkIpAvailabilities 查询网络IP使用情况
func (v *VpcClient) ShowNetworkIpAvailabilities(subnetId string) (*modelv2.NetworkIpAvailability, error) {
	request := &modelv2.ShowNetworkIpAvailabilitiesRequest{
		NetworkId: subnetId,
	}
	response, err := v.vpc2.ShowNetworkIpAvailabilities(request)
	if err != nil {
		return nil, err
	}

	return response.NetworkIpAvailability, nil
}

// ListSecurityGroups 查询安全组列表
func (v *VpcClient) ListSecurityGroups(secIds []string) ([]modelv3.SecurityGroup, error) {

	var (
		secs = make([]modelv3.SecurityGroup, 0)
		// 分页参数
		limit  = int32(200) // 每页显示的记录数
		marker *string      // 分页标记，即下一页的起始位置
	)

	for {
		request := &modelv3.ListSecurityGroupsRequest{
			Limit:  &limit,
			Marker: marker,
			Id: func() *[]string {
				if len(secIds) == 0 {
					return nil
				}
				return &secIds
			}(),
		}
		response, errLocal := v.vpc3.ListSecurityGroups(request)
		if errLocal != nil {
			blog.Infof("ListSecurityGroups failed: %v", errLocal)
			return nil, errLocal
		}

		// 处理当前页的安全组列表
		for _, sec := range *response.SecurityGroups {
			secs = append(secs, sec)
		}

		// 如果没有更多的页，则退出循环
		if response.PageInfo.NextMarker == nil || *response.PageInfo.NextMarker == "" {
			break
		}
		marker = response.PageInfo.NextMarker // 准备获取下一页
	}

	return secs, nil
}
