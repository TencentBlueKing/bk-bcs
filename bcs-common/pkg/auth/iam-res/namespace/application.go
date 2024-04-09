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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/cluster"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/utils"
)

// ProjectNamespaceData project&cluster&namespace
type ProjectNamespaceData struct {
	Project   string
	Cluster   string
	Namespace string
}

// NamespaceInstances build namespaceInstances
// nolint
type NamespaceInstances struct {
	IsClusterPerm bool
	Data          []ProjectNamespaceData
}

// BuildInstances for namespace resource
func (cls NamespaceInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)
	if cls.IsClusterPerm && len(cls.Data) > 0 {
		for i := range cls.Data {
			iamInstances = append(iamInstances, []iam.Instance{
				{
					ResourceType: string(project.SysProject),
					ResourceID:   cls.Data[i].Project,
				},
				{
					ResourceType: string(cluster.SysCluster),
					ResourceID:   cls.Data[i].Cluster,
				},
			})
		}

		return iamInstances
	}

	for i := range cls.Data {
		iamInstances = append(iamInstances, []iam.Instance{
			{
				ResourceType: string(project.SysProject),
				ResourceID:   cls.Data[i].Project,
			},
			{
				ResourceType: string(cluster.SysCluster),
				ResourceID:   cls.Data[i].Cluster,
			},
			{
				ResourceType: string(SysNamespace),
				ResourceID:   cls.Data[i].Namespace,
			},
		})
	}

	return iamInstances
}

// NamespaceScopedInstances build namespaceScopedInstances
// nolint
type NamespaceScopedInstances struct {
	Data []ProjectNamespaceData
}

// BuildInstances for namespace scoped resource
func (cls NamespaceScopedInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)

	for i := range cls.Data {
		iamInstances = append(iamInstances, []iam.Instance{
			{
				ResourceType: string(project.SysProject),
				ResourceID:   cls.Data[i].Project,
			},
			{
				ResourceType: string(cluster.SysCluster),
				ResourceID:   cls.Data[i].Cluster,
			},
			{
				ResourceType: string(SysNamespace),
				ResourceID:   cls.Data[i].Namespace,
			},
		})
	}

	return iamInstances
}

// NamespaceApplicationAction struct for namespaceApplication
// nolint
type NamespaceApplicationAction struct {
	IsClusterPerm bool
	ActionID      string
	Data          []ProjectNamespaceData
}

// BuildNamespaceApplicationInstance build namespace resource application
func BuildNamespaceApplicationInstance(nsAppAction NamespaceApplicationAction) iam.ApplicationAction {
	nsApp := utils.ClusterApplication{ActionID: nsAppAction.ActionID}
	// namespace resource support one system, need to build multi instances if use extra system resource
	instances := NamespaceInstances{
		IsClusterPerm: nsAppAction.IsClusterPerm,
		Data:          nsAppAction.Data,
	}.BuildInstances()

	resourceType := SysNamespace
	if nsAppAction.IsClusterPerm {
		resourceType = cluster.SysCluster
	}

	relatedResource := utils.BuildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)
	return utils.BuildIAMApplication(nsApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}

// NamespaceScopedApplicationAction struct for namespaceApplication
// nolint
type NamespaceScopedApplicationAction struct {
	ActionID string
	Data     []ProjectNamespaceData
}

// BuildNSScopedAppInstance build namespace scoped resource application
func BuildNSScopedAppInstance(nsAppAction NamespaceScopedApplicationAction) iam.ApplicationAction {
	nsApp := utils.ClusterApplication{ActionID: nsAppAction.ActionID}
	// namespace resource support one system, need to build multi instances if use extra system resource
	instances := NamespaceScopedInstances{
		Data: nsAppAction.Data,
	}.BuildInstances()

	resourceType := SysNamespace

	relatedResource := utils.BuildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)
	return utils.BuildIAMApplication(nsApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}
