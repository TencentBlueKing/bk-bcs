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

package api

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// NewCBSClient get cvm client from common option
func NewCBSClient(opt *cloudprovider.CommonOption) (*CBSClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)

	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.MachineDomain
	}

	cli, err := cbs.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &CBSClient{cbs: cli}, nil
}

// CBSClient is the client for as
type CBSClient struct {
	cbs *cbs.Client
}

// GetDiskTypes get available disk types for instance types and zones
func (c *CBSClient) GetDiskTypes(instanceTypes []string, zones []string, diskChargeType string, cpu, memory uint64) (
	[]*cbs.DiskConfig, error) {
	request := cbs.NewDescribeDiskConfigQuotaRequest()
	request.InquiryType = common.StringPtr("INQUIRY_CVM_CONFIG")
	if len(instanceTypes) > 0 {
		request.InstanceFamilies = common.StringPtrs(instanceTypes)
	}

	if len(zones) > 0 {
		request.Zones = common.StringPtrs(zones)
	}

	request.DiskChargeType = common.StringPtr(diskChargeType)
	request.CPU = common.Uint64Ptr(cpu)
	request.Memory = common.Uint64Ptr(memory)

	rsp, err := c.cbs.DescribeDiskConfigQuota(request)
	if err != nil {
		return nil, err
	}

	return rsp.Response.DiskConfigSet, nil
}
