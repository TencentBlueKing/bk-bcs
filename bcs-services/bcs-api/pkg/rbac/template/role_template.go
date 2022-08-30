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

package template

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	// ClusterRolePrefix xxx
	ClusterRolePrefix = "bke-"
)

// RoleTemplate xxx
type RoleTemplate struct {
	Name  string
	Rules []rbacv1.PolicyRule
}

// RoleTemplateStore xxx
var RoleTemplateStore map[string]*RoleTemplate

// InitRbacTemplates 初始化所有clusterrole角色，定义每个clusterrole的权限
func InitRbacTemplates() {
	RoleTemplateStore = make(map[string]*RoleTemplate)
	addRoleTemplate("cluster-manage", clusterManageRules)
	addRoleTemplate("cluster-readonly", clusterReadonlyRules)
	addRoleTemplate("services-view", servicesViewRules)
	addRoleTemplate("services-manage", servicesManageRules)
	addRoleTemplate("workloads-view", workloadsViewRules)
	addRoleTemplate("workloads-manage", workloadsManageRules)
}

func addRoleTemplate(roleName string, rules []rbacv1.PolicyRule) {
	roleTemplate := &RoleTemplate{
		Name:  ClusterRolePrefix + roleName,
		Rules: rules,
	}
	RoleTemplateStore[ClusterRolePrefix+roleName] = roleTemplate
}
