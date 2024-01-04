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
	"sync"

	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/argoproj/argo-cd/v2/util/db"
	settings_util "github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Options for data storage
type Options struct {
	Service        string // storage host, split by comma
	User           string // storage user
	Pass           string // storage pass
	Cache          bool   // init cache for performance
	CacheHistory   bool
	AdminNamespace string

	// RepoServerUrl this parameter must be passed in only
	// when GetApplicationManifestsFromRepoServer needs to be called.
	RepoServerUrl string
}

// Store define data interface for argocd structure.
type Store interface {
	// store control interface
	Init() error
	InitArgoDB(ctx context.Context) error
	GetArgoDB() db.ArgoDB
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
	GetClusterFromDB(ctx context.Context, serverUrL string) (*v1alpha1.Cluster, error)
	ListCluster(ctx context.Context) (*v1alpha1.ClusterList, error)
	ListClustersByProject(ctx context.Context, project string) (*v1alpha1.ClusterList, error)
	UpdateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error
	DeleteCluster(ctx context.Context, name string) error

	// Repository interface
	GetRepository(ctx context.Context, repo string) (*v1alpha1.Repository, error)
	ListRepository(ctx context.Context, projNames []string) (*v1alpha1.RepositoryList, error)

	GetApplication(ctx context.Context, name string) (*v1alpha1.Application, error)
	GetApplicationRevisionsMetadata(ctx context.Context, repo, revision []string) ([]*v1alpha1.RevisionMetadata, error)
	GetApplicationResourceTree(ctx context.Context, name string) (*v1alpha1.ApplicationTree, error)
	ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (*v1alpha1.ApplicationList, error)
	DeleteApplicationResource(ctx context.Context, application *v1alpha1.Application,
		resources []*ApplicationResource) []ApplicationDeleteResourceResult
	GetApplicationManifests(ctx context.Context, name, revision string) (*apiclient.ManifestResponse, error)
	GetApplicationManifestsFromRepoServerWithMultiSources(ctx context.Context,
		application *v1alpha1.Application) ([]*apiclient.ManifestResponse, error)
	ApplicationNormalizeWhenDiff(app *v1alpha1.Application, target,
		live *unstructured.Unstructured, hideData bool) error

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

// NewArgoDB create the DB of argocd
func NewArgoDB(ctx context.Context, adminNamespace string) (db.ArgoDB, *settings_util.SettingsManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get in-cluster config failed")
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "create in-cluster client failed")
	}
	settingsMgr := settings_util.NewSettingsManager(ctx, kubeClient, adminNamespace)
	dbInstance := db.NewDB(adminNamespace, settingsMgr, kubeClient)
	return dbInstance, settingsMgr, nil
}
