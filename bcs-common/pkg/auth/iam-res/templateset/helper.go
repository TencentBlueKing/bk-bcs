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

package templateset

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
)

// ResourceTypeIDMap xxx
var ResourceTypeIDMap = map[iam.TypeID]string{
	SysTemplateSet: "模板集",
}

const (
	// SysTemplateSet resource templateSet
	SysTemplateSet iam.TypeID = "templateset"
)

// TemplateSetResourcePath build IAMPath for templateSet resource
// nolint
type TemplateSetResourcePath struct {
	ProjectID         string
	TemplateSetCreate bool
}

// BuildIAMPath build IAMPath, related resource project when clusterCreate
func (rp TemplateSetResourcePath) BuildIAMPath() string {
	if rp.TemplateSetCreate {
		return ""
	}
	return fmt.Sprintf("/project,%s/", rp.ProjectID)
}

// TemplateSetResourceNode build templateSet resourceNode
// nolint
type TemplateSetResourceNode struct {
	IsCreateTemplateSet bool

	SystemID  string
	ProjectID string
}

// BuildResourceNodes build templateSet iam.ResourceNode
func (trn TemplateSetResourceNode) BuildResourceNodes() []iam.ResourceNode {
	if trn.IsCreateTemplateSet {
		return []iam.ResourceNode{
			{
				System:    trn.SystemID,
				RType:     string(project.SysProject),
				RInstance: trn.ProjectID,
				Rp: TemplateSetResourcePath{
					TemplateSetCreate: trn.IsCreateTemplateSet,
				},
			},
		}
	}

	return []iam.ResourceNode{
		{
			System:    trn.SystemID,
			RType:     string(SysTemplateSet),
			RInstance: trn.ProjectID,
			Rp: TemplateSetResourcePath{
				ProjectID:         trn.ProjectID,
				TemplateSetCreate: false,
			},
		},
	}
}
