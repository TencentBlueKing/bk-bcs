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

// Package clusterManger cluster-service
package clusterManger

/*
	节点池
*/

const (
	// importClusterApi 集群创建(云凭证方式) post ()
	importClusterApi = "/clustermanager/v1/cluster/import"

	// createClusterApi  集群创建(直接创建) post ()
	createClusterApi = "/clustermanager/v1/cluster"

	// deleteClusterApi  删除集群 delete ( clusterID )
	deleteClusterApi = "/clustermanager/v1/cluster/%s"

	// updateClusterApi  更新集群 put ( clusterID )
	updateClusterApi = "/clustermanager/v1/cluster/%s"

	// getClusterApi  查询集群 get ( clusterID )
	getClusterApi = "/clustermanager/v1/cluster/%s"

	// listProjectClusterApi  查询某个项目下的Cluster列表 get ( projectID )
	listProjectClusterApi = "/clustermanager/v1/projects/%s/clusters"
)

// 集群管理类型
const (
	// IndependentCluster 独立集群
	IndependentCluster = "INDEPENDENT_CLUSTER"

	// ManagedCluster 云上托管集群
	ManagedCluster = "MANAGED_CLUSTER"
)
