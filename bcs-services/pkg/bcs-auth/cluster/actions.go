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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

// cluster resource actions
const (
	// ClusterCreate xxx
	ClusterCreate iam.ActionID = "cluster_create"
	// ClusterView xxx
	ClusterView iam.ActionID = "cluster_view"
	// ClusterManage xxx
	ClusterManage iam.ActionID = "cluster_manage"
	// ClusterDelete xxx
	ClusterDelete iam.ActionID = "cluster_delete"
	// ClusterUse xxx
	ClusterUse iam.ActionID = "cluster_use"
)

// cluster scoped resource actions
const (
	// ClusterScopedCreate xxx
	ClusterScopedCreate iam.ActionID = "cluster_scoped_create"
	// ClusterScopedView xxx
	ClusterScopedView iam.ActionID = "cluster_scoped_view"
	// ClusterScopedUpdate xxx
	ClusterScopedUpdate iam.ActionID = "cluster_scoped_update"
	// ClusterScopedDelete xxx
	ClusterScopedDelete iam.ActionID = "cluster_scoped_delete"
)

// ActionIDNameMap map ActionID to name
var ActionIDNameMap = map[iam.ActionID]string{
	ClusterCreate: "集群创建",
	ClusterView:   "集群查看",
	ClusterManage: "集群管理",
	ClusterUse:    "集群使用",
	ClusterDelete: "集群删除",

	ClusterScopedCreate: "资源创建",
	ClusterScopedUpdate: "资源更新",
	ClusterScopedDelete: "资源删除",
	ClusterScopedView:   "资源查看",
}
