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

package project

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"
)

// ProjectInstances build projectInstances
type ProjectInstances struct {
	IsCreateProject bool
	ProjectList     []string
}

// BuildInstances for project resource
func (cls ProjectInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)
	if cls.IsCreateProject {
		return iamInstances
	}

	for i := range cls.ProjectList {
		iamInstances = append(iamInstances, []iam.Instance{
			iam.Instance{
				ResourceType: string(SysProject),
				ResourceID:   cls.ProjectList[i],
			},
		})
	}

	return iamInstances
}

// ProjectApplicationAction for project application
type ProjectApplicationAction struct {
	IsCreateProject bool
	ActionID        string
	Data            []string
}

// BuildProjectApplicationInstance build project application
func BuildProjectApplicationInstance(proAppAction ProjectApplicationAction) iam.ApplicationAction {
	proApp := utils.ClusterApplication{ActionID: proAppAction.ActionID}

	if proAppAction.IsCreateProject {
		return utils.BuildIAMApplication(proApp, []bkiam.ApplicationRelatedResourceType{})
	}

	// project resource support one system, need to build multi instances if use extra system resource
	instances := ProjectInstances{
		IsCreateProject: proAppAction.IsCreateProject,
		ProjectList:     proAppAction.Data,
	}.BuildInstances()

	resourceType := SysProject
	relatedResource := utils.BuildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)

	return utils.BuildIAMApplication(proApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}

// BuildProjectSameInstanceApplication for same instanceSelection
func BuildProjectSameInstanceApplication(isCreate bool, actionIDs []string, data []string) []iam.ApplicationAction {
	applications := make([]iam.ApplicationAction, 0)

	for i := range actionIDs {
		applications = append(applications, BuildProjectApplicationInstance(ProjectApplicationAction{
			IsCreateProject: isCreate,
			ActionID:        actionIDs[i],
			Data:            data,
		}))
	}

	return applications
}
