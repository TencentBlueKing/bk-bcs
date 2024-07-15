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
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// EcsClient ecs client
type EcsClient struct {
	ecs *ecs.EcsClient
}

var (
	performanceTypeMap = map[string]string{
		"normal":            "通用型",
		"entry":             "通用入门型",
		"cpuv1":             "计算I型",
		"cpuv2":             "计算II型",
		"computingv3":       "通用计算增强型",
		"computingv3_c6ne":  "通用计算增强型",
		"kunpeng_computing": "鲲鹏通用计算增强型",
		"kunpeng_highmem":   "鲲鹏内存优化型",
		"kunpeng_highio":    "鲲鹏超高I/O型",
		"highmem":           "内存优化型",
		"saphana":           "大内存型",
		"diskintensive":     "磁盘增强型",
		"highio":            "超高I/O型",
		"ultracpu":          "超高性能计算型",
		"gpu":               "GPU加速型",
		"fpga":              "FPGA加速型",
		"ascend":            "AI加速型",
	}
)

// NewEcsClient new ecs client
func NewEcsClient(opt *cloudprovider.CommonOption) (*EcsClient, error) {
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

	hcClient, err := ecs.EcsClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &EcsClient{&ecs.EcsClient{HcClient: hcClient}}, nil
}

// GetAllFlavors get all ecs flavors
func (e *EcsClient) GetAllFlavors(az string) (*[]model.Flavor, error) {
	request := &model.ListFlavorsRequest{
		AvailabilityZone: func() *string {
			if az == "" {
				return nil
			}
			return &az
		}()}

	rsp, err := e.ecs.ListFlavors(request)
	if err != nil {
		return nil, err
	}

	return rsp.Flavors, nil
}

// ListAvailabilityZones get all availability zones
func (e *EcsClient) ListAvailabilityZones() ([]model.NovaAvailabilityZone, error) {
	rsp, err := e.ecs.NovaListAvailabilityZones(&model.NovaListAvailabilityZonesRequest{})
	if err != nil {
		return nil, err
	}

	return *rsp.AvailabilityZoneInfo, nil
}

// ShowServer server detail info
func (e *EcsClient) ShowServer(serverId string) (*model.ServerDetail, error) {
	request := &model.ShowServerRequest{ServerId: serverId}
	resp, err := e.ecs.ShowServer(request)
	if err != nil {
		return nil, err
	}

	return resp.Server, nil
}

// ListServerBlockDevices 查询弹性云服务器挂载磁盘列表详情信息
func (e *EcsClient) ListServerBlockDevices(serverId string) (*[]model.ServerBlockDevice, error) {
	request := &model.ListServerBlockDevicesRequest{ServerId: serverId}
	resp, err := e.ecs.ListServerBlockDevices(request)
	if err != nil {
		return nil, err
	}

	return resp.VolumeAttachments, nil
}

// ShowServerBlockDevice 查询弹性云服务器单个磁盘信息
func (e *EcsClient) ShowServerBlockDevice(serverId string, volumeId string) (*model.ServerBlockDevice, error) {
	request := &model.ShowServerBlockDeviceRequest{ServerId: serverId, VolumeId: volumeId}
	resp, err := e.ecs.ShowServerBlockDevice(request)
	if err != nil {
		return nil, err
	}

	return resp.VolumeAttachment, nil
}

// ConvertPerformanceType convert performance type
func ConvertPerformanceType(source string) string {
	if v, ok := performanceTypeMap[source]; ok {
		return v
	}

	return ""
}
