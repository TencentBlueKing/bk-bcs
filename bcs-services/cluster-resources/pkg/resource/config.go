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

package resource

import (
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
)

// ClusterConf 集群配置信息
type ClusterConf struct {
	Rest      *rest.Config
	ClusterID string
}

// NewClusterConfig 生成 ClusterConf 对象
func NewClusterConfig(clusterID string) *ClusterConf {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return NewMockClusterConfig(clusterID)
	}
	return &ClusterConf{
		Rest: &rest.Config{
			Host:            envs.BCSApiGWHost + "/clusters/" + clusterID,
			BearerToken:     envs.BCSApiGWAuthToken,
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		},
		ClusterID: clusterID,
	}
}
