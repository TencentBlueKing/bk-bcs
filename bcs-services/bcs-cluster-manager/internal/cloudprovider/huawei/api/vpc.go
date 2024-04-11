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
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	vpc2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	model2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	region2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/region"
	vpc3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3"
	region3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// Vpc2Client vpc v2 client
type Vpc2Client struct {
	*vpc2.VpcClient
}

// Vpc2Client vpc v2 client
type Vpc3Client struct {
	*vpc3.VpcClient
}

// GetVpc2Client get vpc client from common option
func GetVpc2Client(opt *cloudprovider.CommonOption) (*Vpc2Client, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	auth, err := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region2.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	hcClient, err := vpc2.VpcClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &Vpc2Client{vpc2.NewVpcClient(hcClient)}, nil
}

// GetVpc3Client get vpc client from common option
func GetVpc3Client(opt *cloudprovider.CommonOption) (*Vpc3Client, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	auth, err := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region3.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	hcClient, err := vpc3.VpcClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &Vpc3Client{vpc3.NewVpcClient(hcClient)}, nil
}

// ListVpcsByID 获取vpc
func (v *Vpc2Client) ListVpcsByID(vpcID string) ([]model2.Vpc, error) {
	rsp, err := v.ListVpcs(&model2.ListVpcsRequest{Id: &vpcID})
	if err != nil {
		return nil, err
	}

	return *rsp.Vpcs, nil
}

// CalculateAvailableIp 计算子网可用ip数
func (v *Vpc2Client) CalculateAvailableIp(networkId string) (int32, error) {
	rsp, err := v.ShowNetworkIpAvailabilities(&model2.ShowNetworkIpAvailabilitiesRequest{NetworkId: networkId})
	if err != nil {
		return 0, err
	}

	if rsp.NetworkIpAvailability == nil {
		return 0, fmt.Errorf("network ip availability is nil")
	}

	return rsp.NetworkIpAvailability.TotalIps - rsp.NetworkIpAvailability.UsedIps, nil
}
