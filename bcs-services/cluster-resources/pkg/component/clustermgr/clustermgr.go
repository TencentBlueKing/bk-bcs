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

package clustermgr

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
)

// FetchClusterInfo 获取集群信息
func FetchClusterInfo(clusterID string) (map[string]interface{}, error) {
	if runtime.RunMode == runmode.UnitTest {
		return fetchMockClusterInfo(clusterID)
	}
	return fetchClusterInfo(clusterID)
}

func fetchClusterInfo(clusterID string) (map[string]interface{}, error) {
	// TODO 这里判断集群类型，是根据配置来的，后续切换成实际获取集群信息逻辑（调用 clustermgr api ?）
	return fetchMockClusterInfo(clusterID)
}
