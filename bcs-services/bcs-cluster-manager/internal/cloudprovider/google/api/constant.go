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

package api

const (
	// NodeGroupStatusProvisioning indicates the node pool is being created
	NodeGroupStatusProvisioning = "PROVISIONING"
	// NodeGroupStatusRunning indicates the node pool has been created and is fully usable
	NodeGroupStatusRunning = "RUNNING"
	// NodeGroupStatusStopping indicates the node pool is being deleted
	NodeGroupStatusStopping = "STOPPING"
	// NodeGroupStatusError indicates the node pool may be unusable
	NodeGroupStatusError = "ERROR"
	// NodeGroupStatusReconciling indicates that some work is actively
	// being done on the node pool, such as upgrading node software
	NodeGroupStatusReconciling = "RECONCILING"
	// NodeGroupStatusUnspecified node group state not set.
	NodeGroupStatusUnspecified = "STATUS_UNSPECIFIED"
	// NodeGroupStatusRunningWithError indicates the node pool has been created and is partially usable
	NodeGroupStatusRunningWithError = "RUNNING_WITH_ERROR"
)

const (
	// InstanceStatusProvisioning indicates the instance is being created
	InstanceStatusProvisioning = "PROVISIONING"
	// InstanceStatusStaging indicates the instance is being staged
	InstanceStatusStaging = "STAGING"
	// InstanceStatusRunning indicates the instance is running
	InstanceStatusRunning = "RUNNING"
	// InstanceStatusStopping indicates the instance is stopping
	InstanceStatusStopping = "STOPPING"
	// InstanceStatusSuspending indicates the instance is being suspended
	InstanceStatusSuspending = "SUSPENDING"
	// InstanceStatusSuspended indicates the instance is suspended
	InstanceStatusSuspended = "SUSPENDED"
	// InstanceStatusRepairing indicates the instance is being repaired
	InstanceStatusRepairing = "REPAIRING"
	// InstanceStatusTerminated indicates the instance is terminated
	InstanceStatusTerminated = "TERMINATED"
)
