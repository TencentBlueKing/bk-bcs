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

package auth

import (
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// ActionPermissions action 对应权限中心的权限
var ActionPermissions = map[string]string{
	// project
	"BCSProject.GetProject":             project.CanViewProjectOperation,
	"BCSProject.ListAuthorizedProjects": project.CanViewProjectOperation,
	"BCSProject.ListProjects":           project.CanViewProjectOperation,
	"BCSProject.CreateProject":          project.CanCreateProjectOperation,
	"BCSProject.UpdateProject":          project.CanEditProjectOperation,
	"BCSProject.DeleteProject":          project.CanDeleteProjectOperation,
	// business
	"Business.GetBusiness":         project.CanViewProjectOperation,
	"Business.ListBusiness":        project.CanViewProjectOperation,
	"Business.GetBusinessTopology": project.CanViewProjectOperation,
	// variable
	"Variable.CreateVariable":            project.CanViewProjectOperation,
	"Variable.UpdateVariable":            project.CanViewProjectOperation,
	"Variable.ListVariableDefinitions":   project.CanViewProjectOperation,
	"Variable.DeleteVariableDefinitions": project.CanViewProjectOperation,
	"Variable.ListClustersVariables":     project.CanViewProjectOperation,
	"Variable.ListNamespacesVariables":   project.CanViewProjectOperation,
	"Variable.UpdateClustersVariables":   project.CanViewProjectOperation,
	"Variable.UpdateNamespacesVariables": project.CanViewProjectOperation,
	"Variable.ListClusterVariables":      project.CanViewProjectOperation,
	"Variable.ListNamespaceVariables":    project.CanViewProjectOperation,
	"Variable.UpdateClusterVariables":    project.CanViewProjectOperation,
	"Variable.UpdateNamespaceVariables":  project.CanViewProjectOperation,
	"Variable.ImportVariables":           project.CanViewProjectOperation,
	"Variable.RenderVariables":           project.CanViewProjectOperation,
	// Namespace
	"Namespace.SyncNamespace":        namespace.CanCreateNamespaceOperation,
	"Namespace.CreateNamespace":      namespace.CanCreateNamespaceOperation,
	"Namespace.UpdateNamespace":      namespace.CanUpdateNamespaceOperation,
	"Namespace.GetNamespace":         namespace.CanViewNamespaceOperation,
	"Namespace.ListNamespaces":       namespace.CanListNamespaceOperation,
	"Namespace.ListNativeNamespaces": namespace.CanListNamespaceOperation,
	"Namespace.DeleteNamespace":      namespace.CanDeleteNamespaceOperation,
}
