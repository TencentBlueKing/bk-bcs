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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/utils"
)

// ProjectClusterData project&cluster
type ProjectClusterData struct {
	Project string
	Cluster string
}

// ClusterInstances build clusterInstances
// nolint
type ClusterInstances struct {
	IsCreateCluster bool
	Data            []ProjectClusterData
}

// BuildInstances for cluster resource
func (cls ClusterInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)
	if cls.IsCreateCluster && len(cls.Data) > 0 {
		for i := range cls.Data {
			iamInstances = append(iamInstances, []iam.Instance{
				{
					ResourceType: string(project.SysProject),
					ResourceID:   cls.Data[i].Project,
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
				ResourceType: string(SysCluster),
				ResourceID:   cls.Data[i].Cluster,
			},
		})
	}

	return iamInstances
}

// ClusterScopedInstances build clusterScopedInstances
// nolint
type ClusterScopedInstances struct {
	Data []ProjectClusterData
}

// BuildInstances for cluster scoped resource
func (cls ClusterScopedInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)

	for i := range cls.Data {
		iamInstances = append(iamInstances, []iam.Instance{
			{
				ResourceType: string(project.SysProject),
				ResourceID:   cls.Data[i].Project,
			},
			{
				ResourceType: string(SysCluster),
				ResourceID:   cls.Data[i].Cluster,
			},
		})
	}

	return iamInstances
}

// ClusterApplicationAction struct for clusterApplication
// nolint
type ClusterApplicationAction struct {
	IsCreateCluster bool
	ActionID        string
	Data            []ProjectClusterData
}

// BuildClusterApplicationInstance build cluster resource application
func BuildClusterApplicationInstance(clsAppAction ClusterApplicationAction) iam.ApplicationAction {
	clsApp := utils.ClusterApplication{ActionID: clsAppAction.ActionID}
	// cluster resource support one system, need to build multi instances if use extra system resource
	instances := ClusterInstances{
		IsCreateCluster: clsAppAction.IsCreateCluster,
		Data:            clsAppAction.Data,
	}.BuildInstances()

	resourceType := SysCluster
	if clsAppAction.IsCreateCluster {
		resourceType = project.SysProject
	}

	relatedResource := utils.BuildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)
	return utils.BuildIAMApplication(clsApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}

// BuildClusterSameInstanceApplication for same instanceSelection
func BuildClusterSameInstanceApplication(isCreate bool, actionIDs []string,
	data []ProjectClusterData) []iam.ApplicationAction {
	applications := make([]iam.ApplicationAction, 0)

	for i := range actionIDs {
		applications = append(applications, BuildClusterApplicationInstance(ClusterApplicationAction{
			IsCreateCluster: isCreate,
			ActionID:        actionIDs[i],
			Data:            data,
		}))
	}

	return applications
}

// ClusterScopedApplicationAction struct for clusterApplication
// nolint
type ClusterScopedApplicationAction struct {
	ActionID string
	Data     []ProjectClusterData
}

// BuildClusterScopedAppInstance build cluster scoped resource application
func BuildClusterScopedAppInstance(clsAppAction ClusterScopedApplicationAction) iam.ApplicationAction {
	nsApp := utils.ClusterApplication{ActionID: clsAppAction.ActionID}
	// cluster resource support one system, need to build multi instances if use extra system resource
	instances := ClusterScopedInstances{
		Data: clsAppAction.Data,
	}.BuildInstances()

	resourceType := SysCluster

	relatedResource := utils.BuildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)
	return utils.BuildIAMApplication(nsApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}
