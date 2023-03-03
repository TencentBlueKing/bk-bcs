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
	"encoding/json"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var tracer = otel.Tracer("k8s_client")

// GetNodeList 获取集群节点列表
func GetNodeList(ctx context.Context, clusterId string, excludeMasterRole bool) ([]string, []string, error) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("clusterId", clusterId),
		attribute.Bool("excludeMasterRole", excludeMasterRole),
	}
	ctx, span := tracer.Start(ctx, "GetNodeList", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	listOptions := metav1.ListOptions{}
	if excludeMasterRole {
		listOptions.LabelSelector = "node-role.kubernetes.io/master!=true"
	}

	nodeList, err := client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	nodeIPListStr, _ := json.Marshal(nodeIPList)
	nodeNameListStr, _ := json.Marshal(nodeNameList)
	// 设置额外标签
	span.SetAttributes(attribute.String("nodeIPList", string(nodeIPListStr)))
	span.SetAttributes(attribute.String("nodeNameList", string(nodeNameListStr)))
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
