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

package cloudprovider

import (
	"errors"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

var (
	//ErrCloudCredentialLost credential lost in option
	ErrCloudCredentialLost = errors.New("credential info lost")
	//ErrCloudRegionLost region information lost in option
	ErrCloudRegionLost = errors.New("region info lost")
	//ErrCloudLostResponse lost response information in cloud response
	ErrCloudLostResponse = errors.New("lost response information")
	//ErrCloudNoHost no specified instance
	ErrCloudNoHost = errors.New("no such host in region")
	//ErrCloudNoProvider no specified cloud provider
	ErrCloudNoProvider = errors.New("no such cloudprovider")
	//ErrCloudNotImplemented no implementation
	ErrCloudNotImplemented = errors.New("not implemented")
	// ErrCloudInitFailed init failed
	ErrCloudInitFailed = errors.New("failed to init cloud client")
	// ErrServerIsNil server nil
	ErrServerIsNil = errors.New("server is nil")
	//ErrCloudNodeVPCDiffWithClusterResponse for node VPC different cluster VPC
	ErrCloudNodeVPCDiffWithClusterResponse = "node[%s] VPC is different from cluster VPC"
)

const (
	//StatusInitialization node initialization
	StatusInitialization = "INITIALIZATION"
	//StatusFailed status failed
	StatusFailed = "FAILURE"
	//StatusRunning status running
	StatusRunning = "RUNNING"
	//StatusDeleting status deleting for scaling down
	StatusDeleting = "DELETING"
	//StatusCreating node status creating for scaling up
	StatusCreating = "CREATING"
	//StatusDeleted status deleted
	StatusDeleted = "DELETED"

	// TaskStatusInit INIT task status
	TaskStatusInit = "INITIALIZING"
	// TaskStatusRunning running task status
	TaskStatusRunning = "RUNNING"
	// TaskStatusSuccess task success
	TaskStatusSuccess = "SUCCESS"
	// TaskStatusFailure task failed
	TaskStatusFailure = "FAILURE"
	// TaskStatusTimeout task run timeout
	TaskStatusTimeout = "TIMEOUT"
	// TaskStatusForceTerminate force task terminate
	TaskStatusForceTerminate = "FORCETERMINATE"
	// TaskStatusNotStarted force task terminate
	TaskStatusNotStarted = "NOTSTARTED"
)

// CommonOption for all option
type CommonOption struct {
	//request ID for tracing
	RequestID string
	//key and secret for api authentication
	Key    string // tencentcloudSecretID
	Secret string // tencentcloudSecretKey
	// region information for cloudprovider
	// region is unnecessary in some api
	Region string
}

// InitClusterConfigOption init cluster default cloud config
type InitClusterConfigOption struct {
	// CommonOption for common options
	Common *CommonOption
	// Cloud for cluster
	Cloud *proto.Cloud
	// ClusterVersion for cluster version
	ClusterVersion string
}

// SyncClusterCloudInfoOption sync cluster cloud info
type SyncClusterCloudInfoOption struct {
	// CommonOption for common options
	Common *CommonOption
	// Cloud for cluster
	Cloud *proto.Cloud
	// ImportMode import mode
	ImportMode *proto.ImportCloudMode
	// ClusterVersion for cluster version
	ClusterVersion string
}

// GetNodeOption for GetNodeByIP options
type GetNodeOption struct {
	// CommonOption for common options
	Common *CommonOption
	// ClusterVPCID for cluster vpc
	ClusterVPCID string
}

// ListNodesOption for ListNodesByIP options
type ListNodesOption struct {
	// CommonOption for common options
	Common *CommonOption
	// ClusterVPCID for cluster vpc
	ClusterVPCID string
	// ClusterID for cluster id
	ClusterID string
}

// TaskOptions option for create specified task
type TaskOptions struct {
	Cloud    *proto.Cloud
	Project  *proto.Project
	Cluster  *proto.Cluster
	Operator string
}

// CreateClusterOption create cluster option
type CreateClusterOption struct {
	CommonOption
	Reinstall    bool
	InitPassword string
	Operator     string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
}

// ImportClusterOption import cluster option
type ImportClusterOption struct {
	CommonOption
	// cloud is used for cloudprovider template
	Cloud     *proto.Cloud
	CloudMode *proto.ImportCloudMode
	Operator  string
}

// DeleteMode xxx
type DeleteMode string

// String to string
func (dm DeleteMode) String() string {
	return string(dm)
}

const (
	// Terminate terminate mode
	Terminate DeleteMode = "terminate"
	// Retain retain mode
	Retain DeleteMode = "retain"
)

// DeleteClusterOption delete cluster option
type DeleteClusterOption struct {
	CommonOption
	// force delete cluster
	IsForce bool
	// DeleteMode instance deleteMode(terminate/retain)
	DeleteMode DeleteMode
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// Cluster used for cloudprovider
	Cluster *proto.Cluster
}

// GetNodesOption create cluster option
type GetNodesOption struct {
	CommonOption
}

// GetClusterOption create cluster option
type GetClusterOption struct {
	CommonOption
}

// CheckClusterCIDROption check cluster CIDR
type CheckClusterCIDROption struct {
	CommonOption
	CurrentNodeCnt  uint64
	IncomingNodeCnt uint64
}

// AddNodesOption add nodes to cluster option
type AddNodesOption struct {
	CommonOption
	Reinstall    bool
	InitPassword string
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// setting NodeGroupID means add to specified NodeGroup
	NodeGroupID string
}

// DeleteNodesOption create cluster option
type DeleteNodesOption struct {
	CommonOption
	// Operator user
	Operator   string
	IsForce    bool
	DeleteMode string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
}

// CreateNodeGroupOption create nodegroup option
type CreateNodeGroupOption struct {
	CommonOption
}

// DeleteNodeGroupOption create nodegroup option
type DeleteNodeGroupOption struct {
	CommonOption
	IsForce bool
	// reserve all nodes in cluster
	ReserveNodesInCluster bool
	// move all nodes out of cluster, clean all containers
	// but keep nodes running
	ReservedNodeInstance bool
	// move all node out of cluster and return all nodes
	CleanInstanceInCluster bool
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// Operator
	Operator string
}

// MoveNodesOption move nodes to NodeGroup management
type MoveNodesOption struct {
	CommonOption
	Cluster *proto.Cluster
}

// RemoveNodesOption remove nodes from NodeGroup,
// nodes are still in Cluster
type RemoveNodesOption struct {
	CommonOption
	Cloud   *proto.Cloud
	Cluster *proto.Cluster
}

// CleanNodesOption clean nodes in NodeGroup option
type CleanNodesOption struct {
	CommonOption
	Cluster  *proto.Cluster
	Cloud    *proto.Cloud
	Operator string
}

//CleanNodesResponse response for clean nodes in NodeGroup
type CleanNodesResponse struct {
	ClusterID       string
	ResponseOrderID []string
	SuccNodes       []string
	FailedNodes     []string
}

// UpdateDesiredNodeOption update desired node
type UpdateDesiredNodeOption struct {
	CommonOption
	Cluster *proto.Cluster
	Cloud   *proto.Cloud
}

// ScalingResponse response for UpdateDesired nodes
type ScalingResponse struct {
	ResponseID   string
	ScalingUp    uint32
	CapableNodes []string
	Data         string
}

// CreateScalingOption create NodeGroup option
type CreateScalingOption struct {
	CommonOption
}

// DeleteScalingOption create NodeGroup option
type DeleteScalingOption struct {
	CommonOption
}

// UpdateScalingOption create NodeGroup option
type UpdateScalingOption struct {
	CommonOption
}
