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

package k8sclient

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultTimeout = 20 * time.Second
)

// GetNodeList 获取集群节点列表
func GetNodeList(ctx context.Context, clusterID string, excludeMasterRole, filter bool) ([]string, []string, error) {
	client, err := GetK8SClientByClusterId(clusterID)
	if err != nil {
		return nil, nil, err
	}

	listOptions := metav1.ListOptions{}
	if excludeMasterRole {
		listOptions.LabelSelector = "node-role.kubernetes.io/master!=true"
	}

	nodeList, err := client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return nil, nil, err
	}

	nodeIPList := make([]string, 0)
	nodeNameList := make([]string, 0)
	for _, item := range nodeList.Items {
		// 过滤掉被标记的节点，该 annotation 表示该节点不参与资源调度
		if v, ok := item.Annotations["io.tencent.bcs.dev/filter-node-resource"]; ok && v == "true" && filter {
			continue
		}
		nodeNameList = append(nodeNameList, item.Name)
		for _, addr := range item.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				nodeIPList = append(nodeIPList, addr.Address)
			}
		}
	}
	return nodeIPList, nodeNameList, nil
}

// GetNodeByName 获取集群节点信息
func GetNodeByName(ctx context.Context, clusterId, name string) ([]string, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}
	node, err := client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	nodeIPList := make([]string, 0)

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeIPList = append(nodeIPList, addr.Address)
		}
	}
	return nodeIPList, nil
}

// GetNodeCRVersionByName 通过节点名称获取容器运行时版本
func GetNodeCRVersionByName(ctx context.Context, clusterId, name string) (string, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return "", err
	}
	node, err := client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return node.Status.NodeInfo.ContainerRuntimeVersion, nil
}

// GetNodeInfo 获取节点信息 返回相应的节点对象
func GetNodeInfo(ctx context.Context, clusterId, nodeName string) (*v1.Node, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return node, nil
}

// GetMasterNodeList 获取集群节点列表
func GetMasterNodeList(ctx context.Context, clusterID string) ([]string, []string, error) {
	client, err := GetK8SClientByClusterId(clusterID)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	if err != nil {
		return nil, nil, err
	}

	listOptions := metav1.ListOptions{}
	listOptions.LabelSelector = "node-role.kubernetes.io/master=true,io.tencent.bcs.dev/filter-node-resource!=true"

	nodeList, err := client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return nil, nil, err
	}

	nodeIPList := make([]string, 0)
	nodeNameList := make([]string, 0)
	for _, item := range nodeList.Items {
		nodeNameList = append(nodeNameList, item.Name)
		for _, addr := range item.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				nodeIPList = append(nodeIPList, addr.Address)
			}
		}
	}
	return nodeIPList, nodeNameList, nil
}
