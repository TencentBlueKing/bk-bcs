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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetManagedClusterList 获取联邦集群子集群列表
func GetManagedClusterList(ctx context.Context, clusterID string) ([]string, error) {
	cacheKey := "bcs.GetManagedClusterList"
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.([]string), nil
	}
	client, err := GetClusterNetClientByClusterId(clusterID)
	if err != nil {
		return nil, err
	}

	clusterList, err := client.ClustersV1beta1().ManagedClusters("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, v := range clusterList.Items {
		clusters = append(clusters, strings.ToUpper(v.Name))
	}

	storage.LocalCache.Slot.Set(cacheKey, clusters, time.Minute*10)
	return clusters, nil
}
