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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodeList 获取集群节点列表
func GetNodeList(ctx context.Context, clusterId string, excludeMasterRole bool) ([]string, []string, error) {
	client, err := GetK8SClientByClusterId(clusterId)
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

	nodeIPList := make([]string, 0, len(nodeList.Items))
	nodeNameList := make([]string, 0, len(nodeList.Items))
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
