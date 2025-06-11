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
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const cacheExpireDuration = time.Hour * 24 // 缓存过期时间, 现在的场景主要是获取不可变的 lowerPodID

// Workload 简化版 worload
type Workload struct {
	Kind       string    `json:"kind"`
	ApiVersion string    `json:"apiVersion"`
	Metadata   *Metadata `json:"metadata"`
	Items      []*Item   `json:"items,omitempty"` // 类型是 NamespaceList 等有值
}

// Metadata :
type Metadata struct {
	Name        string `json:"name"`
	Labels      Set    `json:"labels"`
	Annotations Set    `json:"annotations"`
}

// Item :
type Item struct {
	Metadata *Metadata `json:"metadata"`
}

// Set is a map of label:value. It implements Labels.
// https://github.com/kubernetes/apimachinery/blob/master/pkg/labels/labels.go
type Set map[string]string

// GetPodEntryValue :
func GetPodEntryValue(ctx context.Context, clusterID, namespace, podname, key string) (string, error) {
	pod, err := GetPod(ctx, clusterID, namespace, podname)
	if err != nil {
		return "", err
	}

	value, ok := pod.Metadata.Annotations[key]
	if ok {
		return value, nil
	}

	value, ok = pod.Metadata.Labels[key]
	if ok {
		return value, nil
	}

	return "", errors.Errorf("key %s not in annotations or labels", key)
}

// GetPod 单个Pod, 查询缓存
func GetPod(ctx context.Context, clusterID, namespace, podname string) (*Workload, error) {
	cacheKey := fmt.Sprintf("components.k8sclient.GetPodLabel:%s.%s.%s", clusterID, namespace, podname)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.(*Workload), nil
	}

	url := fmt.Sprintf("%s/clusters/%s/api/v1/namespaces/%s/pods/%s", config.G.BCS.Host, clusterID, namespace, podname)

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)). // 泳道特性
		SetAuthToken(config.G.BCS.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	workload := &Workload{}
	err = json.Unmarshal(resp.Body(), workload)
	if err != nil {
		return nil, err
	}

	// 保存缓存
	storage.LocalCache.Slot.Set(cacheKey, workload, cacheExpireDuration)
	return workload, nil
}

// GetNamespaces 获取集群的namespace列表
func GetNamespaces(ctx context.Context, clusterID string) ([]string, error) {
	cacheKey := fmt.Sprintf("components.k8sclient.GetNamespaces:%s", clusterID)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.([]string), nil
	}

	url := fmt.Sprintf("%s/clusters/%s/api/v1/namespaces", config.G.BCS.Host, clusterID)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)). // 泳道特性
		SetAuthToken(config.G.BCS.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	workload := &Workload{}
	err = json.Unmarshal(resp.Body(), workload)
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(workload.Items))
	for _, item := range workload.Items {
		if item.Metadata.Name == "" {
			continue
		}
		namespaces = append(namespaces, item.Metadata.Name)
	}

	// 保存缓存
	storage.LocalCache.Slot.Set(cacheKey, namespaces, time.Minute*5)
	return namespaces, nil
}
