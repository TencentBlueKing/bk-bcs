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

// Package storage xxx
package storage

// ListOptions for list operation
type ListOptions struct {
	Limit                  int
	Page                   int
	ReturnSoftDeletedItems bool
	DoPagination           bool
}

// CreateOptions for create strategy
type CreateOptions struct {
	OverWriteIfExist bool
}

// UpdateOptions for update strategy
type UpdateOptions struct {
	CreateIfNotExist        bool
	OverwriteZeroOrEmptyStr bool
}

// DeleteOptions for delete strategy
type DeleteOptions struct {
	ErrIfNotExist bool
}

// GetOptions for get single data
type GetOptions struct {
	ErrIfNotExist  bool
	GetSoftDeleted bool
}

// Storage interface define data object store behavior
// that is independent of any kind of implementation,
// such as MySQL, MongoDB
type Storage interface {
	// ListNodeGroupStrategies operations
	ListNodeGroupStrategies(opt *ListOptions) ([]*NodeGroupMgrStrategy, error)
	ListNodeGroupStrategiesByType(strategyType string, opt *ListOptions) ([]*NodeGroupMgrStrategy, error)
	GetNodeGroupStrategy(name string, opt *GetOptions) (*NodeGroupMgrStrategy, error)
	CreateNodeGroupStrategy(strategy *NodeGroupMgrStrategy, opt *CreateOptions) error
	UpdateNodeGroupStrategy(strategy *NodeGroupMgrStrategy, opt *UpdateOptions) (*NodeGroupMgrStrategy, error)
	DeleteNodeGroupStrategy(name string, opt *DeleteOptions) (*NodeGroupMgrStrategy, error)

	// ListNodeGroups information operations
	ListNodeGroups(opt *ListOptions) ([]*NodeGroup, error)
	GetNodeGroup(nodegroupID string, opt *GetOptions) (*NodeGroup, error)
	CreateNodeGroup(nodegroup *NodeGroup, opt *CreateOptions) error
	UpdateNodeGroup(nodegroup *NodeGroup, opt *UpdateOptions) (*NodeGroup, error)
	DeleteNodeGroup(nodegroupID string, opt *DeleteOptions) (*NodeGroup, error)

	// ListNodeGroupAction list action
	// NodeGroup scaleUp or scaleDown action operations
	// ScaleUp and ScaleDown will happened at the same time sometimes.
	ListNodeGroupAction(nodeGroupID string, opt *ListOptions) ([]*NodeGroupAction, error)
	ListNodeGroupActionByEvent(event string, opt *ListOptions) ([]*NodeGroupAction, error)
	ListNodeGroupActionByTaskID(taskID string, opt *ListOptions) ([]*NodeGroupAction, error)
	GetNodeGroupAction(nodeGroupID, event string, opt *GetOptions) (*NodeGroupAction, error)
	CreateNodeGroupAction(action *NodeGroupAction, opt *CreateOptions) error
	UpdateNodeGroupAction(action *NodeGroupAction, opt *UpdateOptions) (*NodeGroupAction, error)
	DeleteNodeGroupAction(action *NodeGroupAction, opt *DeleteOptions) (*NodeGroupAction, error)

	// ListNodeGroupEvent list event
	// tracing Event for nodegroup
	ListNodeGroupEvent(nodeGroupID string, opt *ListOptions) ([]*NodeGroupEvent, error)
	CreateNodeGroupEvent(event *NodeGroupEvent, opt *CreateOptions) error

	// CreateTask task operation
	CreateTask(task *ScaleDownTask, opt *CreateOptions) error
	UpdateTask(task *ScaleDownTask, opt *UpdateOptions) (*ScaleDownTask, error)
	GetTask(taskID string, opt *GetOptions) (*ScaleDownTask, error)
	ListTasks(opt *ListOptions) ([]*ScaleDownTask, error)
	ListTasksByStrategy(strategyName string, opt *ListOptions) ([]*ScaleDownTask, error)
	DeleteTask(taskID string, opt *DeleteOptions) (*ScaleDownTask, error)
}
