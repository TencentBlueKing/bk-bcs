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

package store

import (
	"context"

	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"

	"sync"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

// Options for data storage
type Options struct {
	Service string // storage host, split by comma
	User    string // storage user
	Pass    string // storage pass
	Cache   bool   // init cache for performance
}

// Store define data interface for argocd structure.
type Store interface {
	// store control interface
	Init() error
	Stop()
	GetOptions() *Options

	// Project interface
	CreateProject(ctx context.Context, pro *v1alpha1.AppProject) error
	UpdateProject(ctx context.Context, pro *v1alpha1.AppProject) error
	GetProject(ctx context.Context, name string) (*v1alpha1.AppProject, error)
	ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, error)

	// Cluster interface
	CreateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error
	GetCluster(ctx context.Context, query *cluster.ClusterQuery) (*v1alpha1.Cluster, error)
	ListCluster(ctx context.Context) (*v1alpha1.ClusterList, error)
	ListClustersByProject(ctx context.Context, project string) (*v1alpha1.ClusterList, error)
	UpdateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error
	DeleteCluster(ctx context.Context, name string) error

	// Repository interface
	GetRepository(ctx context.Context, repo string) (*v1alpha1.Repository, error)
	ListRepository(ctx context.Context) (*v1alpha1.RepositoryList, error)

	GetApplication(ctx context.Context, name string) (*v1alpha1.Application, error)
	ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (*v1alpha1.ApplicationList, error)
	DeleteApplicationResource(ctx context.Context, application *v1alpha1.Application) error

	GetApplicationSet(ctx context.Context, name string) (*v1alpha1.ApplicationSet, error)
	ListApplicationSets(ctx context.Context, query *appsetpkg.ApplicationSetListQuery) (
		*v1alpha1.ApplicationSetList, error)

	// authentication token
	GetToken(ctx context.Context) string
}

// NewStore create storage client
func NewStore(opt *Options) Store {
	return &argo{
		option:           opt,
		cacheApplication: &sync.Map{},
	}
}
