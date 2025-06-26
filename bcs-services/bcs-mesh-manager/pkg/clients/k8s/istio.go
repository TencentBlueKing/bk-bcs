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

package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CheckIstioResourceExists 检查集群中是否存在任意 Istio 关联资源
// 存在则返回true，否则返回false
func CheckIstioResourceExists(ctx context.Context, clusterID string) (bool, error) {
	client, err := GetDynamicClient(clusterID)
	if err != nil {
		return false, fmt.Errorf("get dynamic client failed: %v", err)
	}
	discoveryClient, err := GetDiscoveryClient(clusterID)
	if err != nil {
		return false, fmt.Errorf("get discovery client failed: %v", err)
	}

	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return false, fmt.Errorf("get server preferred resources failed: %v", err)
	}
	// 编译 apiResourceList,如果是Istio资源，则查询资源是否存在
	for _, apiList := range apiResourceList {
		groupVersion := strings.Split(apiList.GroupVersion, "/")
		if len(groupVersion) != 2 {
			continue
		}
		group, version := groupVersion[0], groupVersion[1]
		if !IsIstioGroup(group) {
			continue
		}
		// 遍历apiList.APIResources，查询资源是否存在
		for _, res := range apiList.APIResources {
			gvr := schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: res.Name,
			}
			list, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
			blog.Warnf("check istio resource exists, clusterID: %s, group: %s, version: %s, resource: %s, err: %v",
				clusterID, group, version, res.Name, err)
			if err != nil {
				blog.Errorf("check istio resource exists, clusterID: %s, group: %s, version: %s, resource: %s, err: %v",
					clusterID, group, version, res.Name, err)
				return false, fmt.Errorf("list istio resource failed: %v", err)
			}
			if len(list.Items) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}
