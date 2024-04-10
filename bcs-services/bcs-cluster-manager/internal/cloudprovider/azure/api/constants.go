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
	// ManagedClusterProvisioningStateCreating cluster provisioning state creating
	ManagedClusterProvisioningStateCreating = "Creating"
	// ManagedClusterPodIdentityProvisioningStateSucceeded cluster provisioning state succeeded
	ManagedClusterPodIdentityProvisioningStateSucceeded = "Succeeded"
	// ManagedClusterPodIdentityProvisioningStateDeleting cluster provisioning state deleting
	ManagedClusterPodIdentityProvisioningStateDeleting = "Deleting"
	// ManagedClusterPodIdentityProvisioningStateUpdating cluster provisioning state updating
	ManagedClusterPodIdentityProvisioningStateUpdating = "Updating"
	// ManagedClusterPodIdentityProvisioningStateFailed cluster provisioning state failed
	ManagedClusterPodIdentityProvisioningStateFailed = "Failed"

	// AgentPoolProvisioningStateCreating cluster provisioning state creating
	AgentPoolProvisioningStateCreating = "Creating"
	// AgentPoolPodIdentityProvisioningStateSucceeded cluster provisioning state succeeded
	AgentPoolPodIdentityProvisioningStateSucceeded = "Succeeded"
	// AgentPoolPodIdentityProvisioningStateDeleting cluster provisioning state deleting
	AgentPoolPodIdentityProvisioningStateDeleting = "Deleting"
	// AgentPoolPodIdentityProvisioningStateUpdating cluster provisioning state updating
	AgentPoolPodIdentityProvisioningStateUpdating = "Updating"
	// AgentPoolPodIdentityProvisioningStateFailed cluster provisioning state failed
	AgentPoolPodIdentityProvisioningStateFailed = "Failed"

	// VMProvisioningStateSucceeded vm provisioning state succeeded
	VMProvisioningStateSucceeded = "Succeeded"
	// VMProvisioningStateFailed vm provisioning state failed
	VMProvisioningStateFailed = "Failed"

	// NodeGroupLifeStateCreating node group life state creating
	NodeGroupLifeStateCreating = "creating"
	// NodeGroupLifeStateNormal node group life state normal
	NodeGroupLifeStateNormal = "normal"
	// NodeGroupLifeStateUpdating node group life state updating
	NodeGroupLifeStateUpdating = "updating"
	// NodeGroupLifeStateDeleting node group life state deleting
	NodeGroupLifeStateDeleting = "deleting"
	// NodeGroupLifeStateDeleted node group life state deleted
	NodeGroupLifeStateDeleted = "deleted"
)
