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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	UpdateDesiredSize(np string, desiredSize int) error
}

// NodePoolClient is client for nodegroup resource
type NodePoolClient struct {
	operator string
	url      string
	header   http.Header
}

// NewNodePoolClient init a new client
func NewNodePoolClient(operator, url, token string) (NodePoolClientInterface, error) {
	header := make(http.Header)
	header.Add("Accept", "application/json")
	header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	return &NodePoolClient{
		operator: operator,
		url:      url,
		header:   header,
	}, nil
}

// GetPool returns the nodegroup full config
func (npc *NodePoolClient) GetPool(np string) (*NodeGroup, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	contents, err := WithoutTLSClient(npc.header, npc.url).Get().WithContext(ctx).Resource("nodegroup").Name(np).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to finish http request: %v", err)
	}
	var pool NodeGroup
	res := GetNodeGroupResponse{Data: &pool}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("can not finish the request, err: %v, response message: %v", res.Message, string(contents))
	}

	return &pool, nil
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
	contents, err := WithoutTLSClient(npc.header, npc.url).Get().WithContext(ctx).Resource("nodegroup").Name(np).Resource("node").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to finish http request: %v", err)
	}
	res := ListNodesInGroupResponse{}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("can not finish the request, err: %v, response message: %v", res.Message, string(contents))
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
	contents, err := WithoutTLSClient(npc.header, npc.url).Get().WithContext(ctx).Resource("node").Name(ip).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to finish http request: %v", err)
	}
	res := GetNodeResponse{}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("can not finish the request, err: %v, response message: %v", res.Message, string(contents))
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
		Operator:    npc.operator,
	}
	byteReq, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	body := bytes.NewReader(byteReq)
	contents, err := WithoutTLSClient(npc.header, npc.url).POST().WithContext(ctx).
		Resource("nodegroup").Name(np).Resource("desirednode").Body(body).Do()
	if err != nil {
		return fmt.Errorf("failed to finish http request, err: %v, body: %v", err, string(contents))
	}
	res := UpdateGroupDesiredNodeResponse{}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return fmt.Errorf("can not finish the request UpdateDesiredNode, response: %v, err: %v",
			string(contents), res.Message)
	}
	if res.Code != 0 {
		return fmt.Errorf("can not finish the request, message: %v, response: %v", res.Message, string(contents))
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

	req := &CleanNodesInGroupRequest{
		ClusterID:   nodePool.ClusterID,
		Nodes:       ips,
		NodeGroupID: nodePool.NodeGroupID,
		Operator:    npc.operator,
	}
	byteReq, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	body := bytes.NewReader(byteReq)
	contents, err := WithoutTLSClient(npc.header, npc.url).DELETE().WithContext(ctx).
		Resource("nodegroup").Name(np).Body(body).Name("groupnode").Do()

	if err != nil {
		return fmt.Errorf("failed to finish http request, err: %v, body: %v", err, string(contents))
	}
	res := CleanNodesInGroupResponse{}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return fmt.Errorf("can not finish the request UpdateDesiredNode, response: %v, err: %v", string(contents), res.Message)
	}
	if res.Code != 0 {
		return fmt.Errorf("can not finish the request, message: %v, response: %v", res.Message, string(contents))
	}

	return nil
}

// UpdateDesiredSize sets the desiredSize of node group
func (npc *NodePoolClient) UpdateDesiredSize(np string, desiredSize int) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	req := &UpdateGroupDesiredSizeRequest{
		DesiredSize: uint32(desiredSize),
		Operator:    npc.operator,
	}
	byteReq, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	body := bytes.NewReader(byteReq)
	contents, err := WithoutTLSClient(npc.header, npc.url).POST().WithContext(ctx).
		Resource("nodegroup").Name(np).Resource("desiredsize").Body(body).Do()
	if err != nil {
		return fmt.Errorf("failed to finish http request, err: %v, body: %v", err, string(contents))
	}
	res := UpdateGroupDesiredSizeResponse{}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return fmt.Errorf("can not finish the request UpdateDesiredSize, response: %v, err: %v",
			string(contents), res.Message)
	}
	if res.Code != 0 {
		return fmt.Errorf("can not finish the request, message: %v, response: %v", res.Message, string(contents))
	}

	return nil
}
