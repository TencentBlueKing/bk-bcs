/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package project

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/projmgr"
)

// Project BCS 项目
type Project struct {
	ID    string
	Code  string
	BizID string
}

// GetProjectInfo ...
func GetProjectInfo(projectID string) (*Project, error) {
	projInfo, err := projmgr.FetchProjectInfo(projectID)
	if err != nil {
		return &Project{}, err
	}
	return &Project{
		ID:    projInfo["id"].(string),
		Code:  projInfo["code"].(string),
		BizID: projInfo["bizID"].(string),
	}, nil
}
