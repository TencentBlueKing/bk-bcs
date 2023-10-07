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

package manager

// GradeManagerInfo grade manager info
type GradeManagerInfo struct {
	// Name name
	Name string
	// Desc desc
	Desc string
	// Project project
	Project *Project
}

// UserGroupInfo userGroup basicInfo
type UserGroupInfo struct {
	// Name name
	Name string
	// Desc desc
	Desc string
	// Users members
	Users []string
	// Project project
	Project *Project
	// Cluster cluster
	Cluster *Cluster
	// Policy manager/viewer
	Policy PolicyType
}

// PolicyType xxx
type PolicyType string

// String toString
func (pt PolicyType) String() string {
	return string(pt)
}

var (
	// Manager xxx
	Manager PolicyType = "manager"
	// Viewer xxx
	Viewer PolicyType = "viewer"
)
