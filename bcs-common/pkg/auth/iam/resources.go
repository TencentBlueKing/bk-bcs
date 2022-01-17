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
 *
 */

package iam

var (
	clusterParent = Parent{
		SystemID:   SystemIDBKBCS,
		ResourceID: SysCluster,
	}
)

// ResourceTypeIDMap resource map name
var ResourceTypeIDMap = map[TypeID]string{
	SysCluster:   "集群",
	SysNamespace: "命名空间",
}

// GenerateResourceTypes generate resource types for register to iam
func GenerateResourceTypes() []ResourceType {
	resourceTypeList := make([]ResourceType, 0)

	// add register resource
	resourceTypeList = append(resourceTypeList, getPublicResource()...)

	return resourceTypeList
}

// GetResourceParentMap get resource parent for iam path
func GetResourceParentMap() map[TypeID][]TypeID {
	resourceParentMap := make(map[TypeID][]TypeID, 0)

	for _, resourceType := range GenerateResourceTypes() {
		for _, parent := range resourceType.Parents {
			resourceParentMap[resourceType.ID] = append(resourceParentMap[resourceType.ID], parent.ResourceID)
		}
	}

	return resourceParentMap
}

func getPublicResource() []ResourceType {
	return []ResourceType{
		{
			ID:            SysCluster,
			Name:          ResourceTypeIDMap[SysCluster],
			NameEn:        "bcs cluster",
			Description:   "集群",
			DescriptionEn: "kubernetes cluster",
			Parents:       nil,
			ProviderConfig: ResourceConfig{
				Path: "/usermanager/auth/v3/find/resource",
			},
			Version: 1,
		},
		{
			ID:            SysNamespace,
			Name:          ResourceTypeIDMap[SysNamespace],
			NameEn:        "bcs cluster namespace",
			Description:   "集群命名空间",
			DescriptionEn: "namespaces in bcs cluster",
			Parents:       []Parent{clusterParent},
			ProviderConfig: ResourceConfig{
				Path: "/usermanager/auth/v3/find/resource",
			},
			Version: 1,
		},
	}
}
