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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	kps "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kps/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kps/v3/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kps/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// KpsClient dew client
type KpsClient struct {
	*kps.KpsClient
}

// NewKpsClient new kps client
func NewKpsClient(opt *cloudprovider.CommonOption) (*KpsClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	projectID, err := GetProjectIDByRegion(opt)
	if err != nil {
		return nil, err
	}

	auth, err := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).
		WithProjectId(projectID).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	hcClient, err := kps.KpsClientBuilder().
		WithCredential(auth).
		WithRegion(rn). //指定region区域
		SafeBuild()
	if err != nil {
		return nil, err
	}

	return &KpsClient{&kps.KpsClient{HcClient: hcClient}}, nil
}

// GetAllUsableKeypairs get all usable keypairs
func (k *KpsClient) GetAllUsableKeypairs() ([]model.Keypairs, error) {
	rsp, err := k.ListKeypairs(&model.ListKeypairsRequest{})
	if err != nil {
		return nil, err
	}

	kps := make([]model.Keypairs, 0)
	for _, kp := range *rsp.Keypairs {
		if *kp.Keypair.FrozenState == "0" {
			kps = append(kps, kp)
		}
	}

	return *rsp.Keypairs, nil
}
