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

package clustermanager

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// NodePoolClientInterface defines the interface of node pool client
type NodePoolClientInterface interface {
	GetPool(np string) (*NodeGroup, error)
	GetPoolConfig(np string) (*AutoScalingGroup, error)
	GetPoolNodeTemplate(np string) (*LaunchConfiguration, error)
	GetNodes(np string) ([]*Node, error)
	GetAutoScalingNodes(np string) ([]*Node, error)
	GetNode(ip string) (*Node, error)
	UpdateDesiredNode(np string, desiredNode int) error
	RemoveNodes(np string, ips []string) error
}

// NodePoolClient is client for nodegroup resource
type NodePoolClient struct {
	client ClusterManagerClient
}

// NewNodePoolClient init a new client
func NewNodePoolClient(endpoint string, opts []grpc.DialOption) (NodePoolClientInterface, error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
	}()

	client := NewClusterManagerClient(conn)
	return &NodePoolClient{
		client: client,
	}, nil
}

// GetPool returns the nodegroup full config
func (npc *NodePoolClient) GetPool(np string) (*NodeGroup, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	req := &GetNodeGroupRequest{
		NodeGroupID: np,
	}
	res, err := npc.client.GetNodeGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to finish grpc request: %v", err)
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("can not finish the request, err: %v, response message: %+v", res.Message, res)
	}
	return res.Data, nil
}

// GetPoolConfig returns the nodegroup scaling config
func (npc *NodePoolClient) GetPoolConfig(np string) (*AutoScalingGroup, error) {
	pool, err := npc.GetPool(np)
	if err != nil {
		return nil, err
	}
	if pool.AutoScaling == nil {
		return nil, fmt.Errorf("node group is not scalable")
	}
	return pool.AutoScaling, nil
}

// GetPoolNodeTemplate returns the node template config
func (npc *NodePoolClient) GetPoolNodeTemplate(np string) (*LaunchConfiguration, error) {
	pool, err := npc.GetPool(np)
	if err != nil {
		return nil, err
	}
	if pool.LaunchTemplate == nil {
		return nil, fmt.Errorf("launch template is not set")
	}
	return pool.LaunchTemplate, nil
}

// GetNodes returns the nodes of a specified node group
func (npc *NodePoolClient) GetNodes(np string) ([]*Node, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	req := &GetNodeGroupRequest{
		NodeGroupID: np,
	}
	res, err := npc.client.ListNodesInGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to finish grpc request: %v", err)
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("can not finish the request, err: %v, response message: %+v", res.Message, res)
	}
	return res.Data, nil
}

// GetAutoScalingNodes returns the nodes of a specified node group
func (npc *NodePoolClient) GetAutoScalingNodes(np string) ([]*Node, error) {
	_, err := npc.GetPoolConfig(np)
	if err != nil {
		return nil, err
	}
	nodes, err := npc.GetNodes(np)
	if err != nil {
		return nil, err
	}
	retNodes := make([]*Node, 0)
	for _, node := range nodes {
		if node.NodeGroupID == "" {
			continue
		}
		retNodes = append(retNodes, node)
	}
	return retNodes, nil
}

// GetNode returns the node of the given ip
func (npc *NodePoolClient) GetNode(ip string) (*Node, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	req := &GetNodeRequest{
		InnerIP: ip,
	}
	res, err := npc.client.GetNode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to finish grpc request: %v", err)
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("can not finish the request, err: %v, response message: %+v", res.Message, res)
	}
	if res.Data == nil {
		return nil, nil
	}
	return res.Data[0], nil
}

// UpdateDesiredNode sets the desiredNode number of node group
func (npc *NodePoolClient) UpdateDesiredNode(np string, desiredNode int) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	req := &UpdateGroupDesiredNodeRequest{
		NodeGroupID: np,
		DesiredNode: uint32(desiredNode),
	}
	res, err := npc.client.UpdateGroupDesiredNode(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to finish grpc request: %v", err)
	}
	if res.Code != 0 {
		return fmt.Errorf("can not finish the request, err: %v, response message: %+v", res.Message, res)
	}
	if !res.Result {
		return fmt.Errorf("update node group desired node failed, err: %v, response message: %+v", res.Message, res)
	}
	return nil
}

// RemoveNodes removes the ips from the node group
func (npc *NodePoolClient) RemoveNodes(np string, ips []string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	nodePool, err := npc.GetPool(np)
	if err != nil {
		return err
	}

	req := &RemoveNodesFromGroupRequest{
		ClusterID:   nodePool.ClusterID,
		Nodes:       ips,
		NodeGroupID: nodePool.NodeGroupID,
	}
	res, err := npc.client.RemoveNodesFromGroup(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to finish grpc request: %v", err)
	}
	if res.Code != 0 {
		return fmt.Errorf("can not finish the request, err: %v, response message: %+v", res.Message, res)
	}
	if !res.Result {
		return fmt.Errorf("remove nodes from node group failed, err: %v, response message: %+v", res.Message, res)
	}
	return nil
}
