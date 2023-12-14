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

package sys

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"

// ResourceTypeIDMap resource type map.
var ResourceTypeIDMap = map[client.TypeID]string{
	Business:    "业务",
	Application: "服务",
}

// GenerateStaticResourceTypes generate all the resource types registered to IAM.
func GenerateStaticResourceTypes() []client.ResourceType {
	resourceTypeList := make([]client.ResourceType, 0)

	// add business resources
	resourceTypeList = append(resourceTypeList, genBusinessResources()...)
	return resourceTypeList
}

func genBusinessResources() []client.ResourceType {
	return []client.ResourceType{
		{
			ID:            Application,
			Name:          ResourceTypeIDMap[Application],
			NameEn:        "Application",
			Description:   "服务",
			DescriptionEn: "application under a business",
			Parents: []client.Parent{{
				SystemID:   SystemIDCMDB,
				ResourceID: Business,
			}},
			ProviderConfig: client.ResourceConfig{
				Path: "/api/v1/auth/iam/find/resource",
			},
			Version: 1,
		},
	}
}
