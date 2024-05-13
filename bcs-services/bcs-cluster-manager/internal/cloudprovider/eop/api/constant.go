/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

const (
	// ClusterStatusRunning running
	ClusterStatusRunning = "K8S_CLUSTER_STATE_RUNNING"
	// ClusterStatusInitial initial
	ClusterStatusInitial = "K8S_CLUSTER_STATE_INITIAL"
	// ClusterStatusCreateFailed create failed
	ClusterStatusCreateFailed = "K8S_CLUSTER_STATE_CREATE_FAILED"

	// MasterNodePoolName master node pool name
	MasterNodePoolName = "system-nodepool"
	// NodePoolStatusInitial nodepool initial state
	NodePoolStatusInitial = "NODE_POOL_STATE_INITIAL"
	// NodePoolStatusCreateFailed nodepool create failed state
	NodePoolStatusCreateFailed = "NODE_POOL_STATE_CREATE_FAILED"
	// NodePoolStatusActive nodepool active state
	NodePoolStatusActive = "NODE_POOL_STATE_ACTIVATED"

	// NodeStatusRunning node running state
	NodeStatusRunning = "NODE_STATE_RUNNING"
	// NodeStatusUnknown node unknown state
	NodeStatusUnknown = "NODE_STATE_UNKNOWN"

	// NodeRoleMaster master node
	NodeRoleMaster = "NODE_ROLE_MASTER"
	// NodeRoleWorker worker node
	NodeRoleWorker = "NODE_ROLE_WORKER"

	// DisKIOTypeNormal normal IO disk
	DisKIOTypeNormal = "DISK_IO_TYPE_NORMAL"
	// DisKIOTypeHigh high IO disk
	DisKIOTypeHigh = "DISK_IO_TYPE_HIGH"
	// DisKIOTypeUltra ultra IO disk
	DisKIOTypeUltra = "DISK_IO_TYPE_ULTRA"
)
