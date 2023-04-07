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

// NodeManagement defines the set of node management services turned on for the node pool
type NodeManagement struct {
	// AutoRepair a flag that specifies whether the node auto-repair is enabled for the node pool
	AutoRepair bool `json:"autoRepair,omitempty"`
	// AutoUpgrade a flag that specifies whether node auto-upgrade is enabled for the node pool
	AutoUpgrade bool `json:"autoUpgrade,omitempty"`
}

// MaxPodsConstraint Constraints applied to pods
type MaxPodsConstraint struct {
	// MaxPodsPerNode constraint enforced on the max num of pods per node
	MaxPodsPerNode int64 `json:"maxPodsPerNode,omitempty,string"`
}

// UpgradeSettings these upgrade settings control the level of
// parallelism and the level of disruption caused by an upgrade
type UpgradeSettings struct {
	// MaxSurge the maximum number of nodes that can be created beyond the
	// current size of the node pool during the upgrade process
	MaxSurge int64 `json:"maxSurge,omitempty"`
	// MaxUnavailable the maximum number of nodes that can be
	// simultaneously unavailable during the upgrade process
	MaxUnavailable int64 `json:"maxUnavailable,omitempty"`
}

// NodePool contains the name and configuration for a cluster's node pool
type NodePool struct {
	// Autoscaling Autoscaler configuration for this NodePool
	Autoscaling *NodePoolAutoscaling `json:"autoscaling,omitempty"`
	// Config the node configuration of the pool
	Config *NodeConfig `json:"config,omitempty"`
	// InitialNodeCount the initial node count for the pool
	InitialNodeCount int64 `json:"initialNodeCount,omitempty"`
	// Management NodeManagement configuration for this NodePool
	Management *NodeManagement `json:"management,omitempty"`
	// MaxPodsConstraint the constraint on the maximum number of pods that
	// can be run simultaneously on a node in the node pool
	MaxPodsConstraint *MaxPodsConstraint `json:"maxPodsConstraint,omitempty"`
	// Name the name of the node pool.
	Name string `json:"name,omitempty"`
	// UpgradeSettings upgrade settings control disruption and speed of the upgrade
	UpgradeSettings *UpgradeSettings `json:"upgradeSettings,omitempty"`
	// Version the version of the Kubernetes of this node
	Version string `json:"version,omitempty"`
}

// Taint Kubernetes taint
type Taint struct {
	Effect string `json:"effect,omitempty"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
}

// NodeConfig describe the nodes in a cluster
type NodeConfig struct {
	// MachineType the name of a Google Compute Engine machine type
	MachineType string   `json:"machineType,omitempty"`
	OauthScopes []string `json:"oauthScopes,omitempty"`
	// DiskSizeGb size of the disk attached to each node, specified in GB
	DiskSizeGb int64 `json:"diskSizeGb,omitempty"`
	// DiskType type of the disk attached to each node
	DiskType string `json:"diskType,omitempty"`
	// ImageType the image type to use for this node
	ImageType string `json:"imageType,omitempty"`
	// Labels the map of Kubernetes labels to be applied to each node
	Labels map[string]string `json:"labels,omitempty"`
	Tags   []string          `json:"tags,omitempty"`
	Taints []*Taint          `json:"taints,omitempty"`
}

// NodePoolAutoscaling gke nodePool auto scaling
type NodePoolAutoscaling struct {
	Enabled      bool  `json:"enabled,omitempty"`
	MaxNodeCount int64 `json:"maxNodeCount,omitempty"`
	MinNodeCount int64 `json:"minNodeCount,omitempty"`
}

// CreateNodePoolRequest creates a node pool for a cluster
type CreateNodePoolRequest struct {
	// NodePool the node pool to create
	NodePool *NodePool `json:"nodePool,omitempty"`
	// Parent the parent where the node pool will be created
	// Specified in the format `projects/*/locations/*/clusters/*`
	Parent string `json:"parent,omitempty"`
}

// UpdateNodePoolRequest update a node pool's image and/or version
type UpdateNodePoolRequest struct {
	// ImageType the desired image type for the node pool
	ImageType string `json:"imageType,omitempty"`
	// Name is name of the node pool to update. Specified in the format
	// `projects/*/locations/*/clusters/*/nodePools/*`
	Name string `json:"name,omitempty"`
	// NodeVersion is the Kubernetes version to change the nodes to
	NodeVersion string `json:"nodeVersion,omitempty"`
}
