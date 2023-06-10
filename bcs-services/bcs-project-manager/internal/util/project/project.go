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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// PatchBusinessName patch business name by business id for each project
func PatchBusinessName(projects []*proto.Project) error {
	bizIDs := []int{}
	for _, project := range projects {
		// 历史遗留原因，以前迁移的一部分项目未开启容器服务，但是却设置了业务ID为0
		if project.Kind == "k8s" && project.BusinessID != "" && project.BusinessID != "0" {
			bizID, err := strconv.Atoi(project.BusinessID)
			if err != nil {
				return err
			}
			bizIDs = append(bizIDs, bizID)
		}
	}
	details, err := cmdb.BatchSearchBusinessByBizIDs(bizIDs)
	if err != nil {
		return err
	}
	businessMap := make(map[string]string)
	for _, biz := range details.Info {
		businessMap[strconv.Itoa(int(biz.BKBizID))] = biz.BKBizName
	}
	for _, project := range projects {
		if _, ok := businessMap[project.BusinessID]; !ok {
			project.BusinessName = ""
		}
		project.BusinessName = businessMap[project.BusinessID]
	}
	return nil
}
