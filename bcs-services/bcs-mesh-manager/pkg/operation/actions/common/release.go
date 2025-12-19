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

// Package common 异步操作通用函数
package common

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// GetReleaseName 获取指定集群和组件的release名称
func GetReleaseName(releaseNames map[string]map[string]string, cluster, component, meshID string) (string, error) {
	if releaseNames[cluster] == nil {
		blog.Errorf("[%s]release names not found for cluster: %s", meshID, cluster)
		return "", fmt.Errorf("release names not found for cluster: %s", cluster)
	}
	releaseName, ok := releaseNames[cluster][component]
	if !ok {
		// 网关组件的releaseName获取失败则表示网关未安装
		if component == common.ComponentIstioGateway {
			return "", nil
		}
		blog.Errorf("[%s]get %s release name failed, clusterID: %s", meshID, component, cluster)
		return "", fmt.Errorf("get %s release name failed, clusterID: %s", component, cluster)
	}
	return releaseName, nil
}
