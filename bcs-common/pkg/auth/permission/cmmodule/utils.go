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

package cmmodule

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"
)

// ResourceNode interface for different resource IAMResourceNode
type ResourceNode interface {
	BuildResourceNodes() []iam.ResourceNode
}

// ProjectResourceNode build project resourceNode
type ProjectResourceNode struct {
	IsCreateProject bool

	SystemID  string
	ProjectID string
}

// BuildResourceNodes build project iam.ResourceNode
func (prn ProjectResourceNode) BuildResourceNodes() []iam.ResourceNode {
	if prn.IsCreateProject {
		return nil
	}

	return []iam.ResourceNode{
		iam.ResourceNode{
			System:    prn.SystemID,
			RType:     string(iam.SysProject),
			RInstance: prn.ProjectID,
			Rp:        iam.ProjectResourcePath{},
		},
	}
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
				RType:     string(iam.SysProject),
				RInstance: crn.ProjectID,
				Rp: iam.ClusterResourcePath{
					ClusterCreate: crn.IsCreateCluster,
				},
			},
		}
	}

	return []iam.ResourceNode{
		iam.ResourceNode{
			System:    crn.SystemID,
			RType:     string(iam.SysCluster),
			RInstance: crn.ClusterID,
			Rp: iam.ClusterResourcePath{
				ProjectID:     crn.ProjectID,
				ClusterCreate: false,
			},
		},
	}
}

// ResourceAction for multi action multi resources
type ResourceAction struct {
	Resource string
	Action   string
}

// CheckResourceRequest xxx
type CheckResourceRequest struct {
	Module    string
	Operation string
	User      string
}

// CheckResourcePerms check multi resources actions in perms
func CheckResourcePerms(req CheckResourceRequest, resources []ResourceAction,
	perms map[string]map[string]bool) (bool, error) {
	if len(perms) == 0 {
		return false, fmt.Errorf("checkResourcePerms get perm empty")
	}

	for _, r := range resources {
		perm, ok := perms[r.Resource]
		if !ok {
			blog.Errorf("%s %s user[%s] resource[%s] not exist in perms", req.Module,
				req.Operation, req.User)
			return false, nil
		}

		if !perm[r.Action] {
			blog.Infof("%s %s user[%s] resource[%v] action[%s] allow[%v]",
				req.Module, req.Operation, req.User, r.Resource, r.Action, perm[r.Action])
			return false, nil
		}
	}

	return true, nil
}

// ClusterApplication build iam.Application for ActionID
type ClusterApplication struct {
	ActionID string
}

// buildApplication only support same system same cluster
func buildClusterApplication(app ClusterApplication, resourceTypes []bkiam.ApplicationRelatedResourceType) iam.ApplicationAction {
	applicationAction := iam.ApplicationAction{
		ActionID:         app.ActionID,
		RelatedResources: make([]bkiam.ApplicationRelatedResourceType, 0),
	}
	if len(resourceTypes) > 0 {
		applicationAction.RelatedResources = append(applicationAction.RelatedResources, resourceTypes...)
	}

	return applicationAction
}

func buildRelatedSystemResource(systemID, resourceType string, instances [][]iam.Instance) bkiam.ApplicationRelatedResourceType {
	relatedResource := bkiam.ApplicationRelatedResourceType{
		SystemID:  systemID,
		Type:      resourceType,
		Instances: make([]bkiam.ApplicationResourceInstance, 0),
	}
	if len(instances) > 0 {
		for i := range instances {
			relatedResource.Instances = append(relatedResource.Instances, iam.BuildResourceInstance(instances[i]))
		}
	}

	return relatedResource
}

// ResourceInstance interface for build resource instance
type ResourceInstance interface {
	BuildInstances() [][]iam.Instance
}

// ProjectClusterData project&cluster
type ProjectClusterData struct {
	Project string
	Cluster string
}

// ClusterInstances build clusterInstances
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
				iam.Instance{
					ResourceType: string(iam.SysProject),
					ResourceID:   cls.Data[i].Project,
				},
			})
		}

		return iamInstances
	}

	for i := range cls.Data {
		iamInstances = append(iamInstances, []iam.Instance{
			iam.Instance{
				ResourceType: string(iam.SysProject),
				ResourceID:   cls.Data[i].Project,
			},
			iam.Instance{
				ResourceType: string(iam.SysCluster),
				ResourceID:   cls.Data[i].Cluster,
			},
		})
	}

	return iamInstances
}

// ProjectInstances build projectInstances
type ProjectInstances struct {
	IsCreateProject bool
	ProjectList     []string
}

// BuildInstances for project resource
func (cls ProjectInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)
	if cls.IsCreateProject && len(cls.ProjectList) > 0 {
		return iamInstances
	}

	for i := range cls.ProjectList {
		iamInstances = append(iamInstances, []iam.Instance{
			iam.Instance{
				ResourceType: string(iam.SysProject),
				ResourceID:   cls.ProjectList[i],
			},
		})
	}

	return iamInstances
}

// ClusterApplicationAction struct for clusterApplication
type ClusterApplicationAction struct {
	IsCreateCluster bool
	ActionID        string
	Data            []ProjectClusterData
}

// BuildClusterApplicationInstance build cluster resource application
func BuildClusterApplicationInstance(clsAppAction ClusterApplicationAction) iam.ApplicationAction {
	clsApp := ClusterApplication{ActionID: clsAppAction.ActionID}
	// cluster resource support one system, need to build multi instances if use extra system resource
	instances := ClusterInstances{
		IsCreateCluster: clsAppAction.IsCreateCluster,
		Data:            clsAppAction.Data,
	}.BuildInstances()

	resourceType := iam.SysCluster
	if clsAppAction.IsCreateCluster {
		resourceType = iam.SysProject
	}

	relatedResource := buildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)
	return buildClusterApplication(clsApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}

// ProjectApplicationAction for project application
type ProjectApplicationAction struct {
	IsCreateProject bool
	ActionID        string
	Data            []string
}

// BuildProjectApplicationInstance build project application
func BuildProjectApplicationInstance(proAppAction ProjectApplicationAction) iam.ApplicationAction {
	proApp := ClusterApplication{ActionID: proAppAction.ActionID}

	// project resource support one system, need to build multi instances if use extra system resource
	instances := ProjectInstances{
		IsCreateProject: proAppAction.IsCreateProject,
		ProjectList:     proAppAction.Data,
	}.BuildInstances()

	resourceType := iam.SysProject
	relatedResource := buildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)

	return buildClusterApplication(proApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}

// BuildClusterSameInstanceApplication for same instanceSelection
func BuildClusterSameInstanceApplication(isCreate bool, actionIDs []string, data []ProjectClusterData) []iam.ApplicationAction{
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
