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

import "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

// ResourceTypeIDMap xxx
var ResourceTypeIDMap = map[iam.TypeID]string{
	SysProject: "项目",
}

const (
	// SysProject resource project
	SysProject iam.TypeID = "project"
)

// ProjectResourcePath build IAMPath for project resource
type ProjectResourcePath struct{}

// BuildIAMPath build IAMPath, related resource project
func (rp ProjectResourcePath) BuildIAMPath() string {
	return ""
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
			RType:     string(SysProject),
			RInstance: prn.ProjectID,
			Rp:        ProjectResourcePath{},
		},
	}
}
