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

package cluster

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// ResourceTypeIDMap xxx
var ResourceTypeIDMap = map[iam.TypeID]string{
	SysCluster: "集群",
}

const (
	// SysCluster resource cluster
	SysCluster iam.TypeID = "cluster"
)

// ClusterResourcePath build IAMPath for cluster resource
type ClusterResourcePath struct {
	ProjectID     string
	ClusterCreate bool
}

// BuildIAMPath build IAMPath, related resource project when clusterCreate
func (rp ClusterResourcePath) BuildIAMPath() string {
	if rp.ClusterCreate {
		return ""
	}
	return fmt.Sprintf("/project,%s/", rp.ProjectID)
}

// ClusterScopedResourcePath  build IAMPath for cluster scoped resource
type ClusterScopedResourcePath struct {
	ProjectID string
}

// BuildIAMPath build IAMPath
func (rp ClusterScopedResourcePath) BuildIAMPath() string {
	return fmt.Sprintf("/project,%s/", rp.ProjectID)
}

// ClusterResourceNode build cluster resourceNode
type ClusterResourceNode struct {
	IsCreateCluster bool

	SystemID  string
	ProjectID string
	ClusterID string
}

// BuildResourceNodes build cluster iam.ResourceNode
func (crn ClusterResourceNode) BuildResourceNodes() []iam.ResourceNode {
	if crn.IsCreateCluster {
		return []iam.ResourceNode{
			iam.ResourceNode{
				System:    crn.SystemID,
				RType:     string(project.SysProject),
				RInstance: crn.ProjectID,
				Rp: ClusterResourcePath{
					ClusterCreate: crn.IsCreateCluster,
				},
			},
		}
	}

	return []iam.ResourceNode{
		iam.ResourceNode{
			System:    crn.SystemID,
			RType:     string(SysCluster),
			RInstance: crn.ClusterID,
			Rp: ClusterResourcePath{
				ProjectID:     crn.ProjectID,
				ClusterCreate: false,
			},
		},
	}
}
