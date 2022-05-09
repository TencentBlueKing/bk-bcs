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
 *
 */

package config

// ClusterKind 集群类型
type ClusterKind string

var (
	IsolatedCLuster ClusterKind = "isolated"
	SharedCluster   ClusterKind = "shared"
	FederatedCluter ClusterKind = "federated"
)

// ClusterResource 集群配置
type ClusterResource struct {
	Kind      string   `yaml:"kind"`
	Member    string   `yaml:"member"`
	Members   []string `yaml:"members"`
	Master    string   `yaml:"master"`
	ClusterId string   `yaml:"cluster_id"`
}
