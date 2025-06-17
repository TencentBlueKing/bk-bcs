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

package istio

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
)

// GetIstioResources 获取集群中存在的 Istio 关联资源
// 返回存在的资源类型列表，格式为 "resource.group"
// 例如：["virtualservices.networking.istio.io", "gateways.networking.istio.io"]
func GetIstioResources(ctx context.Context, clusterID string) ([]string, error) {
	client, err := k8s.GetDynamicClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get dynamic client failed: %v", err)
	}

	discoveryClient, err := k8s.GetDiscoveryClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get discovery client failed: %v", err)
	}

	// 获取 Istio 关联资源
	resources, err := filterResources(discoveryClient)
	if err != nil {
		return nil, err
	}

	// 检查每个资源类型是否存在实例
	foundResources := checkInstances(ctx, client, resources)

	// 将结果转换为列表
	result := convertToResourceList(foundResources)

	return result, nil
}

// apiResource 表示一个 Istio API 资源
type apiResource struct {
	group    string
	version  string
	resource string
}

func filterResources(discoveryClient *discovery.DiscoveryClient) ([]apiResource, error) {
	// 获取所有 API 资源
	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("get server preferred resources failed: %v", err)
	}

	var resources []apiResource

	for _, apiResourceList := range apiResourceList {
		// 解析 group 和 version
		groupVersion := strings.Split(apiResourceList.GroupVersion, "/")
		if len(groupVersion) != 2 {
			continue
		}
		group := groupVersion[0]
		version := groupVersion[1]

		// 跳过非 Istio 组
		if !IsIstioGroup(group) {
			continue
		}

		// 遍历该组下的所有资源类型
		for _, resource := range apiResourceList.APIResources {

			// 跳过非 Istio 资源类型
			if !IsIstioKind(resource.Name) {
				continue
			}

			resources = append(resources, apiResource{
				group:    group,
				version:  version,
				resource: resource.Name,
			})
		}
	}

	return resources, nil
}

func checkInstances(ctx context.Context, client dynamic.Interface, resources []apiResource) map[string]bool {
	foundResources := make(map[string]bool)

	for _, resource := range resources {
		resourceKey := fmt.Sprintf("%s.%s", resource.resource, resource.group)

		// 如果已经找到该资源类型的实例，则跳过
		if foundResources[resourceKey] {
			continue
		}

		gvr := schema.GroupVersionResource{
			Group:    resource.group,
			Version:  resource.version,
			Resource: resource.resource,
		}

		list, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		if len(list.Items) > 0 {
			foundResources[resourceKey] = true
		}
	}

	return foundResources
}

func convertToResourceList(foundResources map[string]bool) []string {
	var result []string
	for resourceKey := range foundResources {
		result = append(result, resourceKey)
	}
	return result
}
