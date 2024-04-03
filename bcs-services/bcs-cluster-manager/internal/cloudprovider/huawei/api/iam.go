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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// IamClient iam client
type IamClient struct {
	*iam.IamClient
}

// GetIamClient get iam client from common option
func GetIamClient(opt *cloudprovider.CommonOption) (*IamClient, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	auth, err := global.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region.SafeValueOf("cn-north-1")
	if err != nil {
		return nil, err
	}

	hcClient, err := iam.IamClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()

	// 创建IAM client
	return &IamClient{&iam.IamClient{HcClient: hcClient}}, nil
}

// GetCloudRegions get cloud all regions
func (cli *IamClient) GetCloudRegions() ([]model.Region, error) {
	rsp, err := cli.KeystoneListRegions(&model.KeystoneListRegionsRequest{})
	if err != nil {
		return nil, err
	}

	return *rsp.Regions, nil
}
