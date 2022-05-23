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

// GetCloudAndCluster get relative cloud & cluster information
func GetCloudAndCluster(model store.ClusterManagerModel,
	cloudID string, clusterID string) (*proto.Cloud, *proto.Cluster, error) {
	//get relative Cluster for information injection
	cluster, err := model.GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("cluster %s err, %s", clusterID, err.Error())
	}
	cloud, err := model.GetCloud(context.Background(), cloudID)
	if err != nil {
		return nil, nil, fmt.Errorf("cloud %s err, %s", cloudID, err.Error())
	}
	return cloud, cluster, nil
}

// GetCloudByCloudID get cloud info
func GetCloudByCloudID(model store.ClusterManagerModel, cloudID string) (*proto.Cloud, error) {
	cloud, err := model.GetCloud(context.Background(), cloudID)
	if err != nil {
		return nil, fmt.Errorf("cloud %s err, %s", cloudID, err.Error())
	}

	return cloud, nil
}

// GetProjectByProjectID get project info
func GetProjectByProjectID(model store.ClusterManagerModel, projectID string) (*proto.Project, error) {
	project, err := model.GetProject(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("project %s err, %s", projectID, err.Error())
	}

	return project, nil
}

// GetClusterInfoByClusterID get cluster info
func GetClusterInfoByClusterID(model store.ClusterManagerModel, clusterID string) (*proto.Cluster, error) {
	cluster, err := model.GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, fmt.Errorf("project %s err, %s", clusterID, err.Error())
	}

	return cluster, nil
}
