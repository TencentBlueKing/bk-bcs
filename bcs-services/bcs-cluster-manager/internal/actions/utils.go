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

package actions

import (
	"context"
	"fmt"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetProjectAndCloud get relative cloud & project information
func GetProjectAndCloud(model store.ClusterManagerModel,
	projectID, cloudID string) (*proto.Cloud, *proto.Project, error) {
	//get relative Project for information injection
	project, err := model.GetProject(context.Background(), projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("project %s err, %s", projectID, err.Error())
	}
	cloud, err := model.GetCloud(context.Background(), cloudID)
	if err != nil {
		return nil, nil, fmt.Errorf("cloud %s err, %s", cloudID, err.Error())
	}
	return cloud, project, nil
}
