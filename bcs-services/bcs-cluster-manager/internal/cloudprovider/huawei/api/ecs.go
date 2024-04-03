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
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// EcsClient ecs client
type EcsClient struct {
	*ecs.EcsClient
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

	auth, err := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).
		WithProjectId(projectID).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	hcClient, err := ecs.EcsClientBuilder().
		WithCredential(auth).
		WithRegion(rn). //指定region区域
		SafeBuild()
	if err != nil {
		return nil, err
	}

	return &EcsClient{&ecs.EcsClient{HcClient: hcClient}}, nil
}

// GetAllFlavors get all ecs flavors
func (e *EcsClient) GetAllFlavors(az string) (*[]model.Flavor, error) {
	rsp, err := e.ListFlavors(&model.ListFlavorsRequest{AvailabilityZone: &az})
	if err != nil {
		return nil, err
	}

	return rsp.Flavors, nil
}

// GetAvailabilityZones get all availability zones
func GetAvailabilityZones(opt *cloudprovider.CommonOption) ([]model.NovaAvailabilityZone, error) {
	client, err := NewEcsClient(opt)
	if err != nil {
		return nil, err
	}

	rsp, err := client.NovaListAvailabilityZones(&model.NovaListAvailabilityZonesRequest{})
	if err != nil {
		return nil, err
	}

	return *rsp.AvailabilityZoneInfo, nil
}

// ConvertPerformanceType convert performance type
func ConvertPerformanceType(source string) string {
	if v, ok := performanceTypeMap[source]; ok {
		return v
	}

	return ""
}
