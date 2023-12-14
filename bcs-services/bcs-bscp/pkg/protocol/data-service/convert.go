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

// Package pbds provides data service core protocol struct and convert functions.
package pbds

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"

// PbInstanceResources convert types InstanceResource to pb InstanceResource
func PbInstanceResources(instances []*types.InstanceResource) []*InstanceResource {
	resources := make([]*InstanceResource, 0)
	for _, instance := range instances {
		resources = append(resources, &InstanceResource{
			Id:   instance.ID,
			Name: instance.Name,
		})
	}

	return resources
}

// PbInstanceInfo convert types InstanceInfo to pb PbInstanceInfo
func PbInstanceInfo(instances []*types.InstanceInfo) []*InstanceInfo {
	infos := make([]*InstanceInfo, 0)
	for _, instance := range instances {
		infos = append(infos, &InstanceInfo{
			Id:          instance.ID,
			DisplayName: instance.DisplayName,
			Approver:    instance.Approver,
			Path:        instance.Path,
		})
	}
	return infos
}
