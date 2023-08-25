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
	// NodeGroupStatusCreating node group status creating
	NodeGroupStatusCreating = "CREATING"
	// NodeGroupStatusActive node group status active
	NodeGroupStatusActive = "ACTIVE"
	// NodeGroupStatusUpdating node group status updating
	NodeGroupStatusUpdating = "UPDATING"
	// NodeGroupStatusDeleting node group status deleting
	NodeGroupStatusDeleting = "DELETING"
	// NodeGroupStatusCreateFailed node group status create failed
	NodeGroupStatusCreateFailed = "CREATE_FAILED"
	// NodeGroupStatusDeleteFailed node group status delete failed
	NodeGroupStatusDeleteFailed = "DELETE_FAILED"
	// NodeGroupStatusDegraded node group status degraded
	NodeGroupStatusDegraded = "DEGRADED"
)

const (
	// InstanceLifecycleStateInService instance life cycle state InService
	InstanceLifecycleStateInService = "InService"
	// InstanceLifecycleStatePending instance life cycle state Pending
	InstanceLifecycleStatePending = "Pending"
	// InstanceLifecycleStateTerminating instance life cycle state Terminating
	InstanceLifecycleStateTerminating = "Terminating"
	// InstanceLifecycleStateTerminated instance life cycle state Terminated
	InstanceLifecycleStateTerminated = "Terminated"
	// InstanceLifecycleStateEnteringStandby instance life cycle state EnteringStandby
	InstanceLifecycleStateEnteringStandby = "EnteringStandby"
	// InstanceLifecycleStateStandby instance life cycle state service
	InstanceLifecycleStateStandby = "Standby"
	// InstanceLifecycleStateQuarantined instance life cycle state Quarantined
	InstanceLifecycleStateQuarantined = "Quarantined"
	// InstanceLifecycleStateDetaching instance life cycle state Detaching
	InstanceLifecycleStateDetaching = "Detaching"
	// InstanceLifecycleStateDetached instance life cycle state Detached
	InstanceLifecycleStateDetached = "Detached"
)

const (
	// InstanceStateRunning instance state running
	InstanceStateRunning = "running"
	// InstanceStatePending instance state pending
	InstanceStatePending = "pending"
	// InstanceStateShuttingDown instance state shutting down
	InstanceStateShuttingDown = "shutting-down"
	// InstanceStateTerminated instance state terminated
	InstanceStateTerminated = "terminated"
	// InstanceStateStopping instance state stopping
	InstanceStateStopping = "stopping"
	// InstanceStateStopped instance state stopped
	InstanceStateStopped = "stopped"
)
