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
	eip "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// EipClient eip client
type EipClient struct {
	eip *eip.EipClient
}

// NewEipClient new eip client
func NewEipClient(opt *cloudprovider.CommonOption) (*EipClient, error) {
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
	auth, err := getProjectAuth(opt.Account.SecretID, opt.Account.SecretKey, projectID)
	if err != nil {
		return nil, err
	}
	rn, err := region.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	hcClient, err := eip.EipClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &EipClient{&eip.EipClient{HcClient: hcClient}}, nil
}

// GetAllBandwidths get all bandwidths
func (e *EipClient) GetAllBandwidths() ([]model.BandwidthResponseBody, error) {
	rsp, err := e.eip.ListBandwidth(&model.ListBandwidthRequest{})
	if err != nil {
		return nil, err
	}

	return *rsp.EipBandwidths, nil
}
