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
	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/manager"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
)

const (
	// ProjectCreate 创建项目
	ProjectCreate string = "project_create"
	// ProjectView 查看项目
	ProjectView string = "project_view"
	// ProjectEdit 编辑项目
	ProjectEdit string = "project_edit"
	// ProjectDelete 删除项目
	ProjectDelete string = "project_delete"

	// NamespaceCreate 创建命名空间
	NamespaceCreate string = "namespace_create"
	// NamespaceView 查看命名空间
	NamespaceView string = "namespace_view"
	// NamespaceUpdate 更新命名空间
	NamespaceUpdate string = "namespace_update"
	// NamespaceDelete 删除命名空间
	NamespaceDelete string = "namespace_delete"

	// NamespaceScopedCreate 资源创建(命名空间域)
	NamespaceScopedCreate string = "namespace_scoped_create"
	// NamespaceScopedView 资源查看(命名空间域)
	NamespaceScopedView string = "namespace_scoped_view"
	// NamespaceScopedUpdate 资源更新(命名空间域)
	NamespaceScopedUpdate string = "namespace_scoped_update"
	// NamespaceScopedDelete 资源删除(命名空间域)
	NamespaceScopedDelete string = "namespace_scoped_delete"
)

// GetProjectIamClient project iam client
func GetProjectIamClient(tenantId string) (*project.BCSProjectPerm, error) {
	iamClient, err := initPermClient(tenantId)
	if err != nil {
		return nil, err
	}

	return project.NewBCSProjectPermClient(iamClient), nil
}

// GetNamespaceIamClient namespace iam client
func GetNamespaceIamClient(tenantId string) (*namespace.BCSNamespacePerm, error) {
	iamClient, err := initPermClient(tenantId)
	if err != nil {
		return nil, err
	}

	return namespace.NewBCSNamespacePermClient(iamClient), nil
}

// GetPermManagerIamClient perm manager client
func GetPermManagerIamClient(tenantId string) (*manager.PermManager, error) {
	iamClient, err := initPermClient(tenantId)
	if err != nil {
		return nil, err
	}

	return manager.NewBCSPermManagerClient(iamClient), nil
}

func initPermClient(tenantId string) (bcsIAM.PermClient, error) {
	opts := &bcsIAM.Options{
		SystemID:    bcsIAM.SystemIDBKBCS,
		AppCode:     config.GlobalConf.App.Code,
		AppSecret:   config.GlobalConf.App.Secret,
		External:    !config.GlobalConf.IAM.UseGWHost,
		GateWayHost: config.GlobalConf.IAM.GatewayHost,
		IAMHost:     config.GlobalConf.IAM.IAMHost,
		BkiIAMHost:  config.GlobalConf.IAM.BKPaaSHost,
		Metric:      false,
		Debug:       config.GlobalConf.IAM.Debug,
		TenantId:    tenantId,
	}
	cli, err := bcsIAM.NewIamClient(opts)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

var (
	// ProjectIamClient iam client for project
	ProjectIamClient *project.BCSProjectPerm
	// NamespaceIamClient iam client for project
	NamespaceIamClient *namespace.BCSNamespacePerm
	// PermManagerClient iam client for manager
	PermManagerClient *manager.PermManager
)

// InitPermClient init perm client
func InitPermClient() error {
	opts := &bcsIAM.Options{
		SystemID:    bcsIAM.SystemIDBKBCS,
		AppCode:     config.GlobalConf.App.Code,
		AppSecret:   config.GlobalConf.App.Secret,
		External:    !config.GlobalConf.IAM.UseGWHost,
		GateWayHost: config.GlobalConf.IAM.GatewayHost,
		IAMHost:     config.GlobalConf.IAM.IAMHost,
		BkiIAMHost:  config.GlobalConf.IAM.BKPaaSHost,
		Metric:      false,
		Debug:       config.GlobalConf.IAM.Debug,
	}
	cli, err := bcsIAM.NewIamClient(opts)
	if err != nil {
		return err
	}
	ProjectIamClient = project.NewBCSProjectPermClient(cli)
	NamespaceIamClient = namespace.NewBCSNamespacePermClient(cli)
	PermManagerClient = manager.NewBCSPermManagerClient(cli)
	return nil
}
