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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
)

// ResourceTypeIDMap xxx
var ResourceTypeIDMap = map[iam.TypeID]string{
	SysCloudAccount: "云账号",
}

const (
	// SysCloudAccount resource account
	SysCloudAccount iam.TypeID = "cloud_account"
)

// AccountResourcePath build IAMPath for account resource
type AccountResourcePath struct {
	ProjectID string
	// 资源实例无关
	AccountCreate bool
}

// BuildIAMPath build IAMPath, related resource account when accountCreate
func (rp AccountResourcePath) BuildIAMPath() string {
	if rp.AccountCreate {
		return ""
	}
	return fmt.Sprintf("/project,%s/", rp.ProjectID)
}

// AccountResourceNode build account resourceNode
type AccountResourceNode struct {
	IsCreateAccount bool

	SystemID  string
	ProjectID string
	AccountID string
}

// BuildResourceNodes build account iam.ResourceNode
func (arn AccountResourceNode) BuildResourceNodes() []iam.ResourceNode {
	if arn.IsCreateAccount {
		return []iam.ResourceNode{
			{
				System:    arn.SystemID,
				RType:     string(project.SysProject),
				RInstance: arn.ProjectID,
				Rp: AccountResourcePath{
					AccountCreate: arn.IsCreateAccount,
				},
			},
		}
	}

	return []iam.ResourceNode{
		{
			System:    arn.SystemID,
			RType:     string(SysCloudAccount),
			RInstance: arn.AccountID,
			Rp: AccountResourcePath{
				ProjectID:     arn.ProjectID,
				AccountCreate: false,
			},
		},
	}
}
