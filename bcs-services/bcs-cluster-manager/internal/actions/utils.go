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

	corev1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// PermInfo for perm request
type PermInfo struct {
	ProjectID string
	UserID    string
}

// GetCloudAndCluster get relative cloud & cluster information
func GetCloudAndCluster(model store.ClusterManagerModel,
	cloudID string, clusterID string) (*proto.Cloud, *proto.Cluster, error) {
	// get relative Cluster for information injection
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

// GetNodeGroupByGroupID get nodeGroup info
func GetNodeGroupByGroupID(model store.ClusterManagerModel, groupID string) (*proto.NodeGroup, error) {
	nodeGroup, err := model.GetNodeGroup(context.Background(), groupID)
	if err != nil {
		return nil, fmt.Errorf("nodeGroup %s err, %s", groupID, err.Error())
	}

	return nodeGroup, nil
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

// GetDefaultClusterAutoScalingOption get default cluster auto scaling option
func GetDefaultClusterAutoScalingOption(clusterID string) *proto.ClusterAutoScalingOption {
	return &proto.ClusterAutoScalingOption{
		Expander:            "random",
		BufferResourceRatio: 100,
		ClusterID:           clusterID,
	}
}

// TaintToK8sTaint convert taint to k8s taint
func TaintToK8sTaint(taint []*proto.Taint) []corev1.Taint {
	taints := make([]corev1.Taint, 0)
	for _, v := range taint {
		taints = append(taints, corev1.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: corev1.TaintEffect(v.Effect),
		})
	}
	return taints
}

// K8sTaintToTaint convert k8s taint to taint
func K8sTaintToTaint(taint []corev1.Taint) []*proto.Taint {
	taints := make([]*proto.Taint, 0)
	for _, v := range taint {
		taints = append(taints, &proto.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: string(v.Effect),
		})
	}
	return taints
}
