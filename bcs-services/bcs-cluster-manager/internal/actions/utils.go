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

// Package actions xxx
package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	corev1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// PermInfo for perm request
type PermInfo struct {
	ProjectID string
	UserID    string
	TenantID  string
}

// GetCloudAndCluster get relative cloud & cluster information
func GetCloudAndCluster(model store.ClusterManagerModel,
	cloudID string, clusterID string) (*proto.Cloud, *proto.Cluster, error) {
	// get relative Cluster for information injection
	cluster, err := model.GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("cluster %s err, %s", clusterID, err.Error())
	}

	if cloudID == "" {
		cloudID = cluster.Provider
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

// GetNodeTemplateByTemplateID get nodeTemplate info
func GetNodeTemplateByTemplateID(model store.ClusterManagerModel, templateID string) (*proto.NodeTemplate, error) {
	nodeTemplate, err := model.GetNodeTemplateByID(context.Background(), templateID)
	if err != nil {
		return nil, fmt.Errorf("nodeTemplate %s err, %s", templateID, err.Error())
	}

	return nodeTemplate, nil
}

// GetAsOptionByClusterID get asOption info
func GetAsOptionByClusterID(
	model store.ClusterManagerModel, clusterID string) (*proto.ClusterAutoScalingOption, error) {
	clsAsOption, err := model.GetAutoScalingOption(context.Background(), clusterID)
	if err != nil {
		return nil, fmt.Errorf("clusterAOption %s err, %s", clusterID, err.Error())
	}

	return clsAsOption, nil
}

// GetProjectClusters get project cluster list
func GetProjectClusters(ctx context.Context, model store.ClusterManagerModel, projectID string) (
	[]*proto.Cluster, error) {
	condCluster := operator.NewLeafCondition(operator.Eq, operator.M{
		"projectid": projectID,
	})
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)

	clusterList, err := model.ListCluster(ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("GetProjectClusters[%s] failed: %v", projectID, err)
		return nil, err
	}

	return clusterList, nil
}

// GetCloudClusters get project cluster list
func GetCloudClusters(ctx context.Context, model store.ClusterManagerModel, cloudID, accountID, vpcID string) (
	[]*proto.Cluster, error) {
	condCluster := operator.NewLeafCondition(operator.Eq, operator.M{
		"provider":       cloudID,
		"cloudaccountid": accountID,
		"vpcid":          vpcID,
	})
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)

	clusterList, err := model.ListCluster(ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("GetCloudClusters[%s] failed: %v", cloudID, err)
		return nil, err
	}

	return clusterList, nil
}

// GetClusterInfoByClusterID get cluster info
func GetClusterInfoByClusterID(model store.ClusterManagerModel, clusterID string) (*proto.Cluster, error) {
	cluster, err := model.GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster %s err, %s", clusterID, err.Error())
	}

	return cluster, nil
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

// UpdateNodeGroupDesiredSize update group desired size
func UpdateNodeGroupDesiredSize(model store.ClusterManagerModel, groupID string, nodeNum int, out bool) error {
	group, err := model.GetNodeGroup(context.Background(), groupID)
	if err != nil {
		blog.Errorf("updateNodeGroupDesiredSize failed when update group[%s] desiredSize: %v", groupID, err)
		return err
	}

	if out {
		if group.AutoScaling.DesiredSize >= uint32(nodeNum) {
			group.AutoScaling.DesiredSize -= uint32(nodeNum)
		} else {
			group.AutoScaling.DesiredSize = 0
			blog.Warnf("updateNodeGroupDesiredSize abnormal, desiredSize[%v] scaleNodesNum[%v]",
				group.AutoScaling.DesiredSize, nodeNum)
		}
	} else {
		group.AutoScaling.DesiredSize += uint32(nodeNum)
	}

	err = model.UpdateNodeGroup(context.Background(), group)
	if err != nil {
		blog.Errorf("updateNodeGroupDesiredSize failed when update group[%s] desiredSize: %v", err, groupID)
		return err
	}

	return nil
}

// TransNodeStatus 转换节点状态
func TransNodeStatus(cmNodeStatus string, k8sNode *corev1.Node) string {
	if cmNodeStatus == common.StatusInitialization || cmNodeStatus == common.StatusAddNodesFailed ||
		cmNodeStatus == common.StatusDeleting || cmNodeStatus == common.StatusRemoveNodesFailed ||
		cmNodeStatus == common.StatusRemoveCANodesFailed {
		return cmNodeStatus
	}
	for _, v := range k8sNode.Status.Conditions {
		if v.Type != corev1.NodeReady {
			continue
		}
		if v.Status == corev1.ConditionTrue {
			if k8sNode.Spec.Unschedulable {
				return common.StatusNodeRemovable
			}
			return common.StatusRunning
		}
		return common.StatusNodeNotReady
	}

	return common.StatusNodeUnknown
}
