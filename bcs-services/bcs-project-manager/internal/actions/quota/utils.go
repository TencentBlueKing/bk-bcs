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

package quota

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
)

func checkProjectValidate(model store.ProjectModel, projectId, projectCode, name string) (*pm.Project, error) {
	if len(projectId) == 0 && len(projectCode) == 0 && len(strings.TrimSpace(name)) == 0 {
		return nil, fmt.Errorf("project id/code/name field all empty")
	}

	p, err := model.GetProjectByField(context.Background(), &pm.ProjectField{ProjectID: projectId,
		ProjectCode: projectCode, Name: name})
	if err != nil {
		return nil, fmt.Errorf("projectId(%s) projectCode(%s) projectName(%s) is invalid",
			projectId, projectCode, name)
	}

	return p, nil
}

func checkClusterValidate(clusterId string) (*clustermanager.Cluster, error) {
	if len(strings.TrimSpace(clusterId)) == 0 {
		return nil, nil
	}

	cls, err := clustermanager.GetCluster(clusterId)
	if err != nil {
		return nil, err
	}

	return cls, nil
}
