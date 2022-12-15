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

package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// RSClient ReplicaSet Client
type RSClient struct {
	ResClient
}

// NewRSClient ...
func NewRSClient(ctx context.Context, conf *res.ClusterConf) *RSClient {
	rsRes, _ := res.GetGroupVersionResource(ctx, conf, resCsts.RS, "")
	return &RSClient{ResClient{NewDynamicClient(conf), conf, rsRes}}
}

// NewRSCliByClusterID ...
func NewRSCliByClusterID(ctx context.Context, clusterID string) *RSClient {
	return NewRSClient(ctx, res.NewClusterConf(clusterID))
}

// List 获取 ReplicaSet 列表
func (c *RSClient) List(
	ctx context.Context, namespace, ownerName string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	ret, err := c.ResClient.List(ctx, namespace, opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 只有指定 OwnerReferences 信息才会再过滤
	if ownerName == "" {
		return manifest, nil
	}

	ownerRefs := []map[string]string{{"kind": resCsts.Deploy, "name": ownerName}}
	manifest["items"] = filterByOwnerRefs(mapx.GetList(manifest, "items"), ownerRefs)
	return manifest, nil
}
