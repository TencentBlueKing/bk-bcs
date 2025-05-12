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

package manager

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/templateset"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// AuthorizationScopeInterface authorization scope interface
type AuthorizationScopeInterface interface {
	BuildScopePerm() []iam.AuthorizationScope
}

// AuthorizationScopeFunc authorization scope func
type AuthorizationScopeFunc func() iam.AuthorizationScope

// Project projectInfo
type Project struct {
	// ProjectID projectID
	ProjectID string
	// ProjectCode projectCode
	ProjectCode string
	// Name projectName
	Name string
	// TenantId tenantID
	TenantId string
}

// BuildScopePerm build project level perm scope
func (p *Project) BuildScopePerm() []iam.AuthorizationScope {
	scopeFuncs := make([]AuthorizationScopeFunc, 0)

	scopeFuncs = append(scopeFuncs, p.buildProjectCreateScopePerm, p.buildProjectOtherScope,
		p.buildClusterCreateScope, p.buildClusterOtherScope, p.buildClusterScopedScope,
		p.buildNamespaceCreateListScope, p.buildNamespaceOtherScope, p.buildNamespaceScopedScope,
		p.buildTemplateSetCreateScope, p.buildTemplateSetOtherScope,
		p.buildCloudAccountScope)

	authScopes := make([]iam.AuthorizationScope, 0)
	for i := range scopeFuncs {
		authScopes = append(authScopes, scopeFuncs[i]())
	}

	return authScopes
}

func (p *Project) validate() error {
	if p.ProjectID == "" || p.Name == "" || p.ProjectCode == "" {
		return fmt.Errorf("project object paras empty")
	}

	return nil
}

func (p *Project) getTenantId() string {
	if p == nil || p.TenantId == "" {
		return utils.DefaultTenantId
	}

	return p.TenantId
}

func (p *Project) buildProjectCreateScopePerm() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(project.SysProject, []iam.ActionID{
		project.ProjectCreate,
	}, nil)
}

func (p *Project) buildProjectOtherScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(project.SysProject, []iam.ActionID{
		project.ProjectView, project.ProjectEdit, project.ProjectDelete,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildProjectViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(project.SysProject, []iam.ActionID{
		project.ProjectView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildClusterCreateScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(project.SysProject, []iam.ActionID{
		cluster.ClusterCreate,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildClusterOtherScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterManage, cluster.ClusterDelete, cluster.ClusterView, cluster.ClusterUse,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildClusterViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildClusterScopedScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterScopedCreate, cluster.ClusterScopedDelete, cluster.ClusterScopedUpdate, cluster.ClusterScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildClusterScopedViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildNamespaceCreateListScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		namespace.NameSpaceCreate, namespace.NameSpaceList,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildNamespaceOtherScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceDelete, namespace.NameSpaceUpdate, namespace.NameSpaceView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildNamespaceListScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		namespace.NameSpaceList,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildNamespaceViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildNamespaceScopedScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceScopedCreate, namespace.NameSpaceScopedDelete, namespace.NameSpaceScopedUpdate,
		namespace.NameSpaceScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildNamespaceScopedViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildTemplateSetCreateScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(project.SysProject, []iam.ActionID{templateset.TemplateSetCreate},
		[]iam.LevelResource{
			{
				Type: string(project.SysProject),
				ID:   p.ProjectID,
				Name: p.Name,
			},
		})
}

func (p *Project) buildTemplateSetOtherScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(templateset.SysTemplateSet, []iam.ActionID{
		templateset.TemplateSetView, templateset.TemplateSetCopy, templateset.TemplateSetUpdate,
		templateset.TemplateSetDelete, templateset.TemplateSetInstantiate,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildTemplateSetViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(templateset.SysTemplateSet, []iam.ActionID{
		templateset.TemplateSetView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}

func (p *Project) buildCloudAccountScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cloudaccount.SysCloudAccount, []iam.ActionID{
		cloudaccount.AccountUse, cloudaccount.AccountManage,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   p.ProjectID,
			Name: p.Name,
		},
	})
}
