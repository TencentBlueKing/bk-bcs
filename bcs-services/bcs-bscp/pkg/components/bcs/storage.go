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

// Package bcs provides bcs api client.
package bcs

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
)

// QueryPodData 通过集群ID和 pod UID 查询 pod 信息返回
type QueryPodData struct {
	ClusterID    string      `json:"clusterID"`
	Namespace    string      `json:"namespace"`
	ResourceName string      `json:"resourceName"`
	ResourceType string      `json:"resourceType"`
	CreateTime   string      `json:"createTime"`
	UpdateTime   string      `json:"updateTime"`
	Data         *corev1.Pod `json:"data"`
}

// QueryPod 通过集群ID和 pod UID 查询 pod 信息
func QueryPod(ctx context.Context, clusterID string, uid string) (*corev1.Pod, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/storage/k8s/dynamic/all_resources/clusters/%s/Pod?data.metadata.uid=%s",
		cc.FeedServer().BCS.Host, clusterID, uid)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetAuthToken(cc.FeedServer().BCS.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	data := []*QueryPodData{}
	if err := components.UnmarshalBKResult(resp, &data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("pod %s not found in cluster %s", uid, clusterID)
	}

	return data[0].Data, nil
}

// QueryNodeData 通过集群ID和 Node name 查询 Node 信息
type QueryNodeData struct {
	ClusterID    string       `json:"clusterID"`
	Namespace    string       `json:"namespace"`
	ResourceName string       `json:"resourceName"`
	ResourceType string       `json:"resourceType"`
	CreateTime   string       `json:"createTime"`
	UpdateTime   string       `json:"updateTime"`
	Data         *corev1.Node `json:"data"`
}

// QueryNode 通过集群ID和 Node name 查询 Node 信息
func QueryNode(ctx context.Context, clusterID string, name string) (*corev1.Node, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/storage/k8s/dynamic/all_resources/clusters/%s/Node?data.metadata.name=%s",
		cc.FeedServer().BCS.Host, clusterID, name)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetAuthToken(cc.FeedServer().BCS.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	data := []*QueryNodeData{}
	if err := components.UnmarshalBKResult(resp, &data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("node %s not found in cluster %s", name, clusterID)
	}

	return data[0].Data, nil
}
