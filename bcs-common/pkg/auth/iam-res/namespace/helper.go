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

package namespace

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/cluster"
)

// ResourceTypeIDMap xxx
var ResourceTypeIDMap = map[iam.TypeID]string{
	SysNamespace:       "命名空间",
	SysNamespaceScoped: "命名空间域资源",
}

const (
	// SysNamespace resource namespace
	SysNamespace iam.TypeID = "namespace"
	// SysNamespaceScoped resource namespace
	SysNamespaceScoped iam.TypeID = "namespace_scoped"
)

// NamespaceResourcePath  build IAMPath for namespace resource
// nolint
type NamespaceResourcePath struct {
	ProjectID     string
	ClusterID     string
	IsClusterPerm bool
}

// BuildIAMPath build IAMPath
func (rp NamespaceResourcePath) BuildIAMPath() string {
	// special case to handle create namespace resource
	if rp.IsClusterPerm {
		return fmt.Sprintf("/project,%s/", rp.ProjectID)
	}
	return fmt.Sprintf("/project,%s/cluster,%s/", rp.ProjectID, rp.ClusterID)
}

// NamespaceScopedResourcePath  build IAMPath for namespace scoped resource
// nolint
type NamespaceScopedResourcePath struct {
	ProjectID string
	ClusterID string
}

// BuildIAMPath build IAMPath
func (rp NamespaceScopedResourcePath) BuildIAMPath() string {
	return fmt.Sprintf("/project,%s/cluster,%s/", rp.ProjectID, rp.ClusterID)
}

// NamespaceResourceNode build namespace resourceNode
// nolint
type NamespaceResourceNode struct {
	IsClusterPerm bool

	SystemID  string
	ProjectID string
	ClusterID string
	Namespace string
}

// BuildResourceNodes build namespace iam.ResourceNode
func (nrn NamespaceResourceNode) BuildResourceNodes() []iam.ResourceNode {
	if nrn.IsClusterPerm {
		return []iam.ResourceNode{
			{
				System:    nrn.SystemID,
				RType:     string(cluster.SysCluster),
				RInstance: nrn.ClusterID,
				Rp: NamespaceResourcePath{
					ProjectID:     nrn.ProjectID,
					ClusterID:     nrn.ClusterID,
					IsClusterPerm: nrn.IsClusterPerm,
				},
			},
		}
	}

	return []iam.ResourceNode{
		{
			System:    nrn.SystemID,
			RType:     string(SysNamespace),
			RInstance: nrn.Namespace,
			Rp: NamespaceResourcePath{
				ProjectID: nrn.ProjectID,
				ClusterID: nrn.ClusterID,
			},
		},
	}
}

// NamespaceScopedResourceNode build namespace scoped resourceNode
// nolint
type NamespaceScopedResourceNode struct {
	SystemID  string
	ProjectID string
	ClusterID string
	Namespace string
}

// BuildResourceNodes build namespace scoped iam.ResourceNode
func (nrn NamespaceScopedResourceNode) BuildResourceNodes() []iam.ResourceNode {
	return []iam.ResourceNode{
		{
			System:    nrn.SystemID,
			RType:     string(SysNamespace),
			RInstance: nrn.Namespace,
			Rp: NamespaceScopedResourcePath{
				ProjectID: nrn.ProjectID,
				ClusterID: nrn.ClusterID,
			},
		},
	}
}
