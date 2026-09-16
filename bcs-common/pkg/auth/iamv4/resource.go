/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package iamv4

import "fmt"

// TypeID 资源类型 ID
type TypeID string

// String 返回资源类型字符串
func (t TypeID) String() string { return string(t) }

// ActionID 操作 ID
type ActionID string

// String 返回操作字符串
func (a ActionID) String() string { return string(a) }

// RoleID 角色 ID
type RoleID string

// String 返回角色字符串
func (r RoleID) String() string { return string(r) }

const (
	// SysProject 项目
	SysProject TypeID = "project"
	// SysCluster 集群
	SysCluster TypeID = "cluster"
	// SysNamespace 命名空间
	SysNamespace TypeID = "namespace"
	// SysTemplateSet 模板集
	SysTemplateSet TypeID = "templateset"
	// SysCloudAccount 云账号
	SysCloudAccount TypeID = "cloud_account"
)

const (
	// ProjectCreate 项目创建（无关联资源）
	ProjectCreate ActionID = "project_create"
	// ProjectView 项目查看
	ProjectView ActionID = "project_view"
	// ProjectEdit 项目编辑
	ProjectEdit ActionID = "project_edit"
	// ProjectDelete 项目删除
	ProjectDelete ActionID = "project_delete"

	// ClusterCreate 集群创建（关联项目）
	ClusterCreate ActionID = "cluster_create"
	// ClusterView 集群查看
	ClusterView ActionID = "cluster_view"
	// ClusterManage 集群管理
	ClusterManage ActionID = "cluster_manage"
	// ClusterDelete 集群删除
	ClusterDelete ActionID = "cluster_delete"
	// ClusterUse 集群使用
	ClusterUse ActionID = "cluster_use"

	// ClusterScopedCreate 集群域资源创建
	ClusterScopedCreate ActionID = "cluster_scoped_create"
	// ClusterScopedView 集群域资源查看
	ClusterScopedView ActionID = "cluster_scoped_view"
	// ClusterScopedUpdate 集群域资源更新
	ClusterScopedUpdate ActionID = "cluster_scoped_update"
	// ClusterScopedDelete 集群域资源删除
	ClusterScopedDelete ActionID = "cluster_scoped_delete"

	// NamespaceCreate 命名空间创建（关联集群）
	NamespaceCreate ActionID = "namespace_create"
	// NamespaceList 命名空间列举（关联集群）
	NamespaceList ActionID = "namespace_list"
	// NamespaceView 命名空间查看
	NamespaceView ActionID = "namespace_view"
	// NamespaceUpdate 命名空间更新
	NamespaceUpdate ActionID = "namespace_update"
	// NamespaceDelete 命名空间删除
	NamespaceDelete ActionID = "namespace_delete"
	// NamespaceUse 命名空间使用
	NamespaceUse ActionID = "namespace_use"

	// NamespaceScopedCreate 命名空间域资源创建
	NamespaceScopedCreate ActionID = "namespace_scoped_create"
	// NamespaceScopedView 命名空间域资源查看
	NamespaceScopedView ActionID = "namespace_scoped_view"
	// NamespaceScopedUpdate 命名空间域资源更新
	NamespaceScopedUpdate ActionID = "namespace_scoped_update"
	// NamespaceScopedDelete 命名空间域资源删除
	NamespaceScopedDelete ActionID = "namespace_scoped_delete"

	// TemplateSetCreate 模板集创建（关联项目）
	TemplateSetCreate ActionID = "templateset_create"
	// TemplateSetView 模板集查看
	TemplateSetView ActionID = "templateset_view"
	// TemplateSetCopy 模板集复制
	TemplateSetCopy ActionID = "templateset_copy"
	// TemplateSetUpdate 模板集更新
	TemplateSetUpdate ActionID = "templateset_update"
	// TemplateSetDelete 模板集删除
	TemplateSetDelete ActionID = "templateset_delete"
	// TemplateSetInstantiate 模板集实例化
	TemplateSetInstantiate ActionID = "templateset_instantiate"

	// CloudAccountCreate 云账号创建（关联项目）
	CloudAccountCreate ActionID = "cloud_account_create"
	// CloudAccountManage 云账号管理
	CloudAccountManage ActionID = "cloud_account_manage"
	// CloudAccountUse 云账号使用
	CloudAccountUse ActionID = "cloud_account_use"
)

const (
	// RoleReadOnly 业务只读
	RoleReadOnly RoleID = "read_only"
	// RoleDeveloper 业务开发
	RoleDeveloper RoleID = "developer"
	// RoleOperator 业务运维
	RoleOperator RoleID = "operator"
	// RoleProjectViewer 项目查看者
	RoleProjectViewer RoleID = "project_viewer"
	// RoleProjectManager 项目管理员
	RoleProjectManager RoleID = "project_manager"
	// RoleClusterViewer 集群资源查看
	RoleClusterViewer RoleID = "cluster_viewer"
	// RoleClusterManager 集群资源管理
	RoleClusterManager RoleID = "cluster_manager"
	// RoleNamespaceViewer 命名空间资源查看
	RoleNamespaceViewer RoleID = "namespace_viewer"
	// RoleNamespaceManager 命名空间资源管理
	RoleNamespaceManager RoleID = "namespace_manager"
	// RoleTemplateSetViewer 模板集查看
	RoleTemplateSetViewer RoleID = "templateset_viewer"
	// RoleTemplateSetManager 模板集管理
	RoleTemplateSetManager RoleID = "templateset_manager"
	// RoleCloudAccountViewer 云账号使用
	RoleCloudAccountViewer RoleID = "cloud_account_viewer"
	// RoleCloudAccountManager 云账号管理
	RoleCloudAccountManager RoleID = "cloud_account_manager"
)

// ClusterIAMPath 集群实例的祖先路径（项目）
func ClusterIAMPath(projectID string) string {
	return fmt.Sprintf("/project,%s/", projectID)
}

// NamespaceIAMPath 命名空间实例的祖先路径（项目 / 集群）
func NamespaceIAMPath(projectID, clusterID string) string {
	return fmt.Sprintf("/project,%s/cluster,%s/", projectID, clusterID)
}

// TemplateSetIAMPath 模板集实例的祖先路径（项目）
func TemplateSetIAMPath(projectID string) string {
	return ClusterIAMPath(projectID)
}

// CloudAccountIAMPath 云账号实例的祖先路径（项目）
func CloudAccountIAMPath(projectID string) string {
	return ClusterIAMPath(projectID)
}

// AuthResourceWithPath 构造带拓扑路径的鉴权资源
func AuthResourceWithPath(id, iamPath string) AuthResource {
	res := AuthResource{ID: id}
	if iamPath != "" {
		res.Attributes = map[string]interface{}{IAMPathAttr: iamPath}
	}
	return res
}

// actionResourceTypes 对齐 migrations-v4/0002_actions.up.json。
// 值为空字符串表示操作不关联资源；不在表中则为未知操作。
var actionResourceTypes = map[string]string{
	string(ProjectCreate): "",
	string(ProjectView):   string(SysProject),
	string(ProjectEdit):   string(SysProject),
	string(ProjectDelete): string(SysProject),

	string(ClusterCreate):       string(SysProject),
	string(ClusterView):         string(SysCluster),
	string(ClusterManage):       string(SysCluster),
	string(ClusterDelete):       string(SysCluster),
	string(ClusterUse):          string(SysCluster),
	string(ClusterScopedCreate): string(SysCluster),
	string(ClusterScopedView):   string(SysCluster),
	string(ClusterScopedUpdate): string(SysCluster),
	string(ClusterScopedDelete): string(SysCluster),

	string(NamespaceCreate):       string(SysCluster),
	string(NamespaceList):         string(SysCluster),
	string(NamespaceView):         string(SysNamespace),
	string(NamespaceUpdate):       string(SysNamespace),
	string(NamespaceDelete):       string(SysNamespace),
	string(NamespaceUse):          string(SysNamespace),
	string(NamespaceScopedCreate): string(SysNamespace),
	string(NamespaceScopedView):   string(SysNamespace),
	string(NamespaceScopedUpdate): string(SysNamespace),
	string(NamespaceScopedDelete): string(SysNamespace),

	string(TemplateSetCreate):      string(SysProject),
	string(TemplateSetView):        string(SysTemplateSet),
	string(TemplateSetCopy):        string(SysTemplateSet),
	string(TemplateSetUpdate):      string(SysTemplateSet),
	string(TemplateSetDelete):      string(SysTemplateSet),
	string(TemplateSetInstantiate): string(SysTemplateSet),

	string(CloudAccountCreate): string(SysProject),
	string(CloudAccountManage): string(SysCloudAccount),
	string(CloudAccountUse):    string(SysCloudAccount),
}

// LookupActionResourceType 返回操作关联的资源类型。
// known=false 表示未注册；known=true 且 resourceType="" 表示无关联资源。
func LookupActionResourceType(actionID string) (resourceType string, known bool) {
	resourceType, known = actionResourceTypes[actionID]
	return resourceType, known
}
