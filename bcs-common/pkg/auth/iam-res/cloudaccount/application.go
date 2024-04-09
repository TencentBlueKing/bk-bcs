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

package cloudaccount

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/utils"
)

// ProjectAccountData project&account
type ProjectAccountData struct {
	Project string
	Account string
}

// AccountInstances build accountInstances
type AccountInstances struct {
	IsCreateAccount bool
	Data            []ProjectAccountData
}

// BuildInstances for account resource
func (cls AccountInstances) BuildInstances() [][]iam.Instance {
	iamInstances := make([][]iam.Instance, 0)
	if cls.IsCreateAccount && len(cls.Data) > 0 {
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
				ResourceType: string(SysCloudAccount),
				ResourceID:   cls.Data[i].Account,
			},
		})
	}

	return iamInstances
}

// AccountApplicationAction struct for accountApplication
type AccountApplicationAction struct {
	IsCreateAccount bool
	ActionID        string
	Data            []ProjectAccountData
}

// BuildAccountApplicationInstance build account resource application
func BuildAccountApplicationInstance(accountAppAction AccountApplicationAction) iam.ApplicationAction {
	accountApp := utils.ClusterApplication{ActionID: accountAppAction.ActionID}
	// account resource support one system, need to build multi instances if use extra system resource
	instances := AccountInstances{
		IsCreateAccount: accountAppAction.IsCreateAccount,
		Data:            accountAppAction.Data,
	}.BuildInstances()

	resourceType := SysCloudAccount
	if accountAppAction.IsCreateAccount {
		resourceType = project.SysProject
	}

	relatedResource := utils.BuildRelatedSystemResource(iam.SystemIDBKBCS, string(resourceType), instances)
	return utils.BuildIAMApplication(accountApp, []bkiam.ApplicationRelatedResourceType{relatedResource})
}

// BuildAccountSameInstanceApplication for same instanceSelection
func BuildAccountSameInstanceApplication(isCreate bool, actionIDs []string,
	data []ProjectAccountData) []iam.ApplicationAction {
	applications := make([]iam.ApplicationAction, 0)

	for i := range actionIDs {
		applications = append(applications, BuildAccountApplicationInstance(AccountApplicationAction{
			IsCreateAccount: isCreate,
			ActionID:        actionIDs[i],
			Data:            data,
		}))
	}

	return applications
}
