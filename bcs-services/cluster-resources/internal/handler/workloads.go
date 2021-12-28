/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * workloads.go 工作负载类接口
 */

package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"

	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListWorkloadDeploy 获取工作负载 Deployment 列表
func (cr *ClusterResources) ListWorkloadDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) error {
	// 替换成实际集群查询结果
	m, _ := structpb.NewValue(map[string]interface{}{
		"firstName": "John",
		"lastName":  "Smith",
		"isAlive":   true,
		"age":       27,
		"address": map[string]interface{}{
			"streetAddress": "21 2nd Street",
			"city":          "New York",
			"state":         "NY",
			"postalCode":    "10021-xxx",
		},
		"phoneNumbers": []interface{}{
			map[string]interface{}{
				"type":   "home",
				"number": "212 xxx-1234",
			},
			map[string]interface{}{
				"type":   "office",
				"number": "646 xxx-4567",
			},
		},
		"children": []interface{}{},
		"spouse":   nil,
	})
	resp.Data = &structpb.Struct{Fields: map[string]*structpb.Value{"manifest": m}}
	return nil
}
