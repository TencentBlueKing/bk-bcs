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

	evs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/region"
)

// NewEvsClient init evs client
func NewEvsClient(opt *cloudprovider.CommonOption) (*EvsClient, error) {
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

	// 创建hc client
	hcClient, err := evs.EvsClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &EvsClient{evs.NewEvsClient(hcClient)}, nil
}

// EvsClient evs client
type EvsClient struct {
	evs *evs.EvsClient
}

// CinderListVolumeTypes get evs volume types
func (cli *EvsClient) CinderListVolumeTypes() (*[]model.VolumeType, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	request := &model.CinderListVolumeTypesRequest{}
	rsp, err := cli.evs.CinderListVolumeTypes(request)
	if err != nil {
		return nil, err
	}

	return rsp.VolumeTypes, nil
}

// ShowVolume 查询单个云硬盘详情
func (cli *EvsClient) ShowVolume(volumeId string) (*model.VolumeDetail, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	request := &model.ShowVolumeRequest{VolumeId: volumeId}
	rsp, err := cli.evs.ShowVolume(request)
	if err != nil {
		return nil, err
	}

	return rsp.Volume, nil
}
