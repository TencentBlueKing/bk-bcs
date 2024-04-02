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

// Package api xxx
package api

import (
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	cce "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// CceClient cce client
type CceClient struct {
	*cce.CceClient
}

// NewCceClient init cce client
func NewCceClient(opt *cloudprovider.CommonOption) (*CceClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	projectID, err := GetProjectIDByRegion(opt)
	if err != nil {
		return nil, err
	}

	auth, err := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).
		WithProjectId(projectID).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	// 创建CCE client
	hcClient, err := cce.CceClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &CceClient{cce.NewCceClient(hcClient)}, nil
}

// ListCceCluster get cce cluster list, region parameter init tke client
func (cli *CceClient) ListCceCluster() (*model.ListClustersResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ListClustersRequest{}
	rsp, err := cli.ListClusters(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// GetCceCluster get cce cluster
func (cli *CceClient) GetCceCluster(clusterID string) (*model.ShowClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ShowClusterRequest{
		ClusterId: clusterID,
	}
	rsp, err := cli.ShowCluster(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// ListClusterNodes get cluster all nodes
func (cli *CceClient) ListClusterNodes(clusterId string) ([]model.Node, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.ListNodes(&model.ListNodesRequest{
		ClusterId: clusterId,
	})
	if err != nil {
		return nil, err
	}

	return *rsp.Items, nil
}

// ListClusterNodePoolNodes get cluster node pool all nodes
func (cli *CceClient) ListClusterNodePoolNodes(clusterId, nodePoolId string) ([]model.Node, error) {
	nodes, err := cli.ListClusterNodes(clusterId)
	if err != nil {
		return nil, err
	}

	nodePoolNodes := make([]model.Node, 0)
	for _, v := range nodes {
		if id, ok := v.Metadata.Annotations["kubernetes.io/node-pool.id"]; ok {
			if id == nodePoolId {
				nodePoolNodes = append(nodePoolNodes, v)
			}
		}
	}

	return nodePoolNodes, nil
}

// CreateClusterNodePool create cluster node pool
func (cli *CceClient) CreateClusterNodePool(req *model.CreateNodePoolRequest) (*model.CreateNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.CreateNodePool(req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// GetClusterNodePool get cluster node pool
func (cli *CceClient) GetClusterNodePool(clusterId, nodePoolId string) (*model.ShowNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.ShowNodePool(&model.ShowNodePoolRequest{
		ClusterId:  clusterId,
		NodepoolId: nodePoolId,
	})
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// UpdateDesiredNodes update node pool InitialNodeCount
func (cli *CceClient) UpdateDesiredNodes(clusterId, nodePoolId string, nodeCount int32) (
	*model.UpdateNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	taints := make([]model.Taint, 0)
	k8sTags := map[string]string{}
	userTags := make([]model.UserTag, 0)

	nodePool, err := cli.GetClusterNodePool(clusterId, nodePoolId)
	if err != nil {
		return nil, fmt.Errorf("updateDesiredNodes get cluster nodePool err: %s", err)
	}

	if nodePool.Spec.NodeTemplate.Taints != nil {
		taints = *nodePool.Spec.NodeTemplate.Taints
	}

	if nodePool.Spec.NodeTemplate.K8sTags != nil {
		k8sTags = nodePool.Spec.NodeTemplate.K8sTags
	}

	if nodePool.Spec.NodeTemplate.UserTags != nil {
		userTags = *nodePool.Spec.NodeTemplate.UserTags
	}

	req := &model.UpdateNodePoolRequest{
		ClusterId:  clusterId,
		NodepoolId: nodePoolId,
		Body: &model.NodePoolUpdate{
			Metadata: &model.NodePoolMetadataUpdate{
				Name: nodePool.Metadata.Name,
			},
			Spec: &model.NodePoolSpecUpdate{
				NodeTemplate: &model.NodeSpecUpdate{
					Taints:   taints,
					K8sTags:  k8sTags,
					UserTags: userTags,
				},
				InitialNodeCount: nodeCount,
				Autoscaling:      &model.NodePoolNodeAutoscaling{},
			},
		},
	}

	rsp, err := cli.UpdateNodePool(req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// RemoveNodePoolNodes remove node pool nodes
func (cli *CceClient) RemoveNodePoolNodes(clusterId string, nodeIds []string, password string) error {
	if len(nodeIds) == 0 {
		return nil
	}

	pw, err := Crypt(password)
	if err != nil {
		return err
	}

	nodes := make([]model.NodeItem, 0)
	for _, v := range nodeIds {
		nodes = append(nodes, model.NodeItem{
			Uid: v,
		})
	}

	_, err = cli.RemoveNode(&model.RemoveNodeRequest{
		ClusterId: clusterId,
		Body: &model.RemoveNodesTask{
			Spec: &model.RemoveNodesSpec{
				Login: &model.Login{
					UserPassword: &model.UserPassword{
						Password: pw,
					},
				},
				Nodes: nodes,
			},
		},
	})

	return err
}

// CleanNodePoolNodes delete node pool nodes
func (cli *CceClient) CleanNodePoolNodes(clusterId string, nodeIds []string) error {
	if len(nodeIds) == 0 {
		return nil
	}

	for _, nodeId := range nodeIds {
		_, err := cli.DeleteNode(&model.DeleteNodeRequest{
			ClusterId: clusterId,
			NodeId:    nodeId,
		})
		if err != nil {
			return fmt.Errorf("删除节点[%s]失败, error: %s", nodeId, err)
		}
	}

	return nil
}
