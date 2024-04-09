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

// Package manager xxx
package manager

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/cluster"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/namespace"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
)

// Cluster clusterInfo
type Cluster struct {
	// ProjectID id
	ProjectID string
	// ProjectName name
	ProjectName string
	// ClusterID id
	ClusterID string
	// ClusterName name
	ClusterName string
}

// BuildScopePerm build cluster level perm scope
func (c *Cluster) BuildScopePerm() []iam.AuthorizationScope {
	scopeFuncs := make([]AuthorizationScopeFunc, 0)

	scopeFuncs = append(scopeFuncs, c.buildClusterCreateScope, c.buildClusterOtherScope, c.buildClusterScopedScope,
		c.buildNamespaceCreateListScope, c.buildNamespaceOtherScope, c.buildNamespaceScopedScope)

	authScopes := make([]iam.AuthorizationScope, 0)
	for i := range scopeFuncs {
		authScopes = append(authScopes, scopeFuncs[i]())
	}

	return authScopes
}

func (c *Cluster) validate() error {
	if c.ProjectID == "" || c.ProjectName == "" || c.ClusterID == "" || c.ClusterName == "" {
		return fmt.Errorf("cluster object paras empty")
	}

	return nil
}

func (c *Cluster) buildClusterCreateScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(project.SysProject, []iam.ActionID{
		cluster.ClusterCreate,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
	})
}

func (c *Cluster) buildClusterOtherScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterManage, cluster.ClusterDelete, cluster.ClusterView, cluster.ClusterUse,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

// NOCC:golint/unused(误报)
// nolint
func (c *Cluster) buildClusterViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterView},
		[]iam.LevelResource{
			{
				Type: string(project.SysProject),
				ID:   c.ProjectID,
				Name: c.ProjectName,
			},
			{
				Type: string(cluster.SysCluster),
				ID:   c.ClusterID,
				Name: c.ClusterName,
			},
		})
}

func (c *Cluster) buildClusterScopedScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterScopedCreate, cluster.ClusterScopedDelete, cluster.ClusterScopedUpdate, cluster.ClusterScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

// NOCC:golint/unused(误报)
// nolint
func (c *Cluster) buildClusterScopedViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		cluster.ClusterScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

func (c *Cluster) buildNamespaceCreateListScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		namespace.NameSpaceCreate, namespace.NameSpaceList,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

// NOCC:golint/unused(误报)
// nolint
func (c *Cluster) buildNamespaceListScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(cluster.SysCluster, []iam.ActionID{
		namespace.NameSpaceList,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

func (c *Cluster) buildNamespaceOtherScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceDelete, namespace.NameSpaceUpdate, namespace.NameSpaceView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

// NOCC:golint/unused(误报)
// nolint
func (c *Cluster) buildNamespaceViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

func (c *Cluster) buildNamespaceScopedScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceScopedCreate, namespace.NameSpaceScopedDelete, namespace.NameSpaceScopedUpdate,
		namespace.NameSpaceScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}

// NOCC:golint/unused(误报)
// nolint
func (c *Cluster) buildNamespaceScopedViewScope() iam.AuthorizationScope {
	return iam.BuildAuthorizationScope(namespace.SysNamespace, []iam.ActionID{
		namespace.NameSpaceScopedView,
	}, []iam.LevelResource{
		{
			Type: string(project.SysProject),
			ID:   c.ProjectID,
			Name: c.ProjectName,
		},
		{
			Type: string(cluster.SysCluster),
			ID:   c.ClusterID,
			Name: c.ClusterName,
		},
	})
}
