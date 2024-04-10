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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	api "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	applicationpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	projectpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	repositorypkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	argoutil "github.com/argoproj/argo-cd/v2/util/argo"
	utilargo "github.com/argoproj/argo-cd/v2/util/argo"
	"github.com/argoproj/argo-cd/v2/util/argo/normalizers"
	"github.com/argoproj/argo-cd/v2/util/db"
	settings_util "github.com/argoproj/argo-cd/v2/util/settings"
	gitopsdiff "github.com/argoproj/gitops-engine/pkg/diff"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/argoconn"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

type argo struct {
	sync.RWMutex

	option *Options
	token  string

	basicOpt      *api.ClientOptions
	conn          *grpc.ClientConn
	connCloser    io.Closer
	argoDB        db.ArgoDB
	settingMgr    *settings_util.SettingsManager
	appClient     applicationpkg.ApplicationServiceClient
	appsetClient  appsetpkg.ApplicationSetServiceClient
	repoClient    repositorypkg.RepositoryServiceClient
	projectClient projectpkg.ProjectServiceClient
	clusterClient clusterpkg.ClusterServiceClient
	historyStore  *appHistoryStore

	cacheSynced      atomic.Bool
	cacheApplication *sync.Map
}

// Init control interface
func (cd *argo) Init() error {
	initializer := []func() error{
		cd.initToken, cd.initBasicClient, cd.initAppHistoryStore,
	}
	if cd.option.Cache {
		initializer = append(initializer, cd.initCache)
	}
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}
	if cd.option.Cache {
		if err := cd.handleApplicationWatch(); err != nil {
			return errors.Wrapf(err, "handle application watch failed")
		}
	}
	return nil
}

// InitArgoDB used to init the DB of argocd
func (cd *argo) InitArgoDB(ctx context.Context) error {
	argoDB, settingMgr, err := NewArgoDB(ctx, cd.option.AdminNamespace)
	if err != nil {
		return errors.Wrapf(err, "create argo db failed")
	}
	cd.argoDB = argoDB
	cd.settingMgr = settingMgr
	return nil
}

func (cd *argo) GetArgoDB() db.ArgoDB {
	return cd.argoDB
}

// Stop control interface
func (cd *argo) Stop() {
}

// GetOptions return the options of gitops
func (cd *argo) GetOptions() *Options {
	return cd.option
}

func (cd *argo) ApplicationNormalizeWhenDiff(app *v1alpha1.Application, target,
	live *unstructured.Unstructured, hideData bool) error {
	var err error
	if hideData {
		target, live, err = gitopsdiff.HideSecretData(target, live)
		if err != nil {
			return fmt.Errorf("error hiding secret data: %s", err)
		}
	}

	resourceOverrides, err := cd.settingMgr.GetResourceOverrides()
	if err != nil {
		return fmt.Errorf("error getting resource overrides: %s", err)
	}
	ignoreNormalizer, err := normalizers.NewIgnoreNormalizer(app.Spec.IgnoreDifferences, resourceOverrides)
	if err != nil {
		return errors.Wrapf(err, "create ignore normalizer failed")
	}
	knownTypeNorm, err := normalizers.NewKnownTypesNormalizer(resourceOverrides)
	if err != nil {
		return errors.Wrapf(err, "create known type normalizer failed")
	}
	if err = ignoreNormalizer.Normalize(target); err != nil {
		return errors.Wrapf(err, "ignore normalizer target failed")
	}
	if err = ignoreNormalizer.Normalize(live); err != nil {
		return errors.Wrapf(err, "ignore normalizer live failed")
	}
	if err = knownTypeNorm.Normalize(target); err != nil {
		return errors.Wrapf(err, "known normalizer target failed")
	}
	if err = knownTypeNorm.Normalize(live); err != nil {
		return errors.Wrapf(err, "known normalizer target failed")
	}

	appLabelKey, err := cd.settingMgr.GetAppInstanceLabelKey()
	if err != nil {
		return fmt.Errorf("error getting app instance label key: %s", err)
	}
	trackingMethod, err := cd.settingMgr.GetTrackingMethod()
	if err != nil {
		return fmt.Errorf("error getting tracking method: %s", err)
	}
	resourceTracking := argoutil.NewResourceTracking()
	if err = resourceTracking.Normalize(target, live, appLabelKey, trackingMethod); err != nil {
		blog.Warnf("resource tracking normalize failed: %s", err.Error())
	}
	return nil
}

// CreateProject interface
func (cd *argo) CreateProject(ctx context.Context, pro *v1alpha1.AppProject) error {
	_, err := cd.projectClient.Create(ctx, &project.ProjectCreateRequest{Project: pro})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("CreateProject").Inc()
		}
		return errors.Wrapf(err, "argocd create project '%s' failed", pro.GetName())
	}
	return nil
}

// UpdateProject interface
func (cd *argo) UpdateProject(ctx context.Context, pro *v1alpha1.AppProject) error {
	_, err := cd.projectClient.Update(ctx, &project.ProjectUpdateRequest{Project: pro})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("UpdateProject").Inc()
		}
		return errors.Wrapf(err, "argocd update project '%s' failed", pro.GetName())
	}
	return nil
}

// GetProject interface
func (cd *argo) GetProject(ctx context.Context, name string) (*v1alpha1.AppProject, error) {
	pro, err := cd.projectClient.Get(ctx, &project.ProjectQuery{Name: name})
	if err != nil {
		if utils.IsArgoResourceNotFound(err) {
			return nil, nil
		}
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("GetProject").Inc()
		}
		return nil, errors.Wrapf(err, "argocd get project '%s' failed", name)
	}
	return pro, nil
}

// ListProjects interface
func (cd *argo) ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, error) {
	pro, err := cd.projectClient.List(ctx, &project.ProjectQuery{})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListProjects").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list alll projects failed")
	}
	return pro, nil
}

// ListProjectsWithoutAuth return projects all
func (cd *argo) ListProjectsWithoutAuth(ctx context.Context) (*v1alpha1.AppProjectList, error) {
	projectList, err := cd.ListProjects(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "list projects failed")
	}
	result := make([]v1alpha1.AppProject, 0, len(projectList.Items))
	for i := range projectList.Items {
		appProj := projectList.Items[i]
		projectID := common.GetBCSProjectID(appProj.Annotations)
		if projectID == "" {
			continue
		}
		result = append(result, appProj)
	}
	projectList.Items = result
	return projectList, nil
}

// CreateCluster interface
func (cd *argo) CreateCluster(ctx context.Context, cls *v1alpha1.Cluster) error {
	_, err := cd.clusterClient.Create(ctx, &cluster.ClusterCreateRequest{Cluster: cls})
	if err != nil {
		if !utils.IsContextCanceled(err) && !utils.IsPermissionDenied(err) && !utils.IsUnauthenticated(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("CreateCluster").Inc()
		}
		return errors.Wrapf(err, "argocd create cluster '%s' failed", cls.Name)
	}
	return nil
}

// GetClusterFromDB get cluster info from ArgoDB which will return the token
func (cd *argo) GetClusterFromDB(ctx context.Context, serverUrl string) (*v1alpha1.Cluster, error) {
	argoCluster, err := cd.argoDB.GetCluster(ctx, serverUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster '%s' failed", serverUrl)
	}
	return argoCluster, nil
}

// DeleteCluster delete cluster by clusterID
func (cd *argo) DeleteCluster(ctx context.Context, name string) error {
	if _, err := cd.clusterClient.Delete(ctx, &cluster.ClusterQuery{Name: name}); err != nil {
		// argocd return 403(PermissionDenied) when cluster do not exist
		// !make sure that gitops-manager has admin access
		if strings.Contains(err.Error(), "code = PermissionDenied") {
			blog.Warnf("argocd delete cluster %s warning: %s", name, err.Error())
			return nil
		}
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("DeleteCluster").Inc()
		}
		return errors.Wrapf(err, "argocd delete cluster failed")
	}
	return nil
}

// GetCluster interface
func (cd *argo) GetCluster(ctx context.Context, query *cluster.ClusterQuery) (*v1alpha1.Cluster, error) {
	cls, err := cd.clusterClient.Get(ctx, query)
	if err != nil {
		// argocd return 403(PermissionDenied) when cluster do not exist
		// !make sure that gitops-manager has admin access
		if strings.Contains(err.Error(), "code = PermissionDenied") {
			blog.Warnf("argocd get cluster '%v' warning, No Cluster Found if admin access, %s",
				*query, err.Error())
			return nil, nil
		}
		if utils.IsArgoResourceNotFound(err) {
			return nil, nil
		}
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("GetCluster").Inc()
		}
		return nil, errors.Wrapf(err, "argocd get cluster '%v' failed", *query)
	}
	return cls, nil
}

// UpdateCluster will update the annotation field
func (cd *argo) UpdateCluster(ctx context.Context, argoCluster *v1alpha1.Cluster) error {
	if _, err := cd.clusterClient.Update(ctx, &cluster.ClusterUpdateRequest{
		Cluster:       argoCluster,
		UpdatedFields: []string{"annotations"},
	}); err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("UpdateCluster").Inc()
		}
		return errors.Wrapf(err, "argocd update cluster '%s' failed", argoCluster.Name)
	}
	return nil
}

// ListCluster interface
func (cd *argo) ListCluster(ctx context.Context) (*v1alpha1.ClusterList, error) {
	cls, err := cd.clusterClient.List(ctx, &cluster.ClusterQuery{})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListCluster").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list all clusters failed")
	}
	return cls, nil
}

// ListClustersByProject will list clusters by project id
func (cd *argo) ListClustersByProject(ctx context.Context, project string) (*v1alpha1.ClusterList, error) {
	cls, err := cd.clusterClient.List(ctx, &cluster.ClusterQuery{})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListClustersByProject").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list all clusters failed")
	}

	clusters := make([]v1alpha1.Cluster, 0, len(cls.Items))
	for _, item := range cls.Items {
		projectID, ok := item.Annotations[common.ProjectIDKey]
		if item.Name != common.InClusterName && (!ok || projectID == "") {
			blog.Errorf("cluster '%s' not have project id annotation", item.Name)
			continue
		}
		if projectID == project {
			clusters = append(clusters, item)
		}
	}
	cls.Items = clusters
	return cls, nil
}

// repo name perhaps encoded, such as: https%253A%252F%252Fgit.fake.com%252Ftest%252Fhelloworld.git.
// So we should urldecode the repo name twice. It also works fine when repo is normal.
func (cd *argo) decodeRepoUrl(repo string) (string, error) {
	t, err := url.PathUnescape(repo)
	if err != nil {
		return "", errors.Wrapf(err, "decode failed")
	}
	result, err := url.PathUnescape(t)
	if err != nil {
		return "", errors.Wrapf(err, "decode second failed")
	}
	return result, nil
}

// GetRepository interface
func (cd *argo) GetRepository(ctx context.Context, repo string) (*v1alpha1.Repository, error) {
	var err error
	repo, err = cd.decodeRepoUrl(repo)
	if err != nil {
		return nil, errors.Wrapf(err, "get repository failed with decode repo '%s'", repo)
	}
	repos, err := cd.repoClient.Get(ctx, &repository.RepoQuery{Repo: repo})
	if err != nil {
		if utils.IsArgoResourceNotFound(err) {
			blog.Warnf("argocd get Repository %s warning, %s", repo, err.Error())
			return nil, nil
		}
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("GetRepository").Inc()
		}
		return nil, errors.Wrapf(err, "argocd get repo '%s' failed", repo)
	}
	return repos, nil
}

// ListRepository list repository with project names
func (cd *argo) ListRepository(ctx context.Context, projNames []string) (*v1alpha1.RepositoryList, error) {
	repos, err := cd.repoClient.List(ctx, &repository.RepoQuery{})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListRepository").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list repos failed")
	}
	if len(projNames) == 0 {
		return repos, nil
	}

	// filter specified project
	items := v1alpha1.Repositories{}
	for _, repo := range repos.Items {
		if sliceutils.StringInSlice(repo.Project, projNames) {
			items = append(items, repo)
		}
	}
	if len(items) == 0 {
		return &v1alpha1.RepositoryList{}, nil
	}
	repos.Items = items
	return repos, nil
}

// AllApplications return all applications cached
func (cd *argo) AllApplications() []*v1alpha1.Application {
	result := make([]*v1alpha1.Application, 0)
	cd.cacheApplication.Range(func(key, value any) bool {
		projectApps := value.(map[string]*v1alpha1.Application)
		for _, app := range projectApps {
			result = append(result, app.DeepCopy())
		}
		return true
	})
	return result
}

// GetApplication will return application by name
func (cd *argo) GetApplication(ctx context.Context, name string) (*v1alpha1.Application, error) {
	if !cd.cacheSynced.Load() {
		app, err := cd.appClient.Get(ctx, &appclient.ApplicationQuery{Name: &name})
		if err != nil {
			if utils.IsArgoResourceNotFound(err) {
				blog.Warnf("argocd get application %s not found: %s", name, err.Error())
				return nil, nil
			}
			if !utils.IsContextCanceled(err) {
				metric.ManagerArgoOperateFailed.WithLabelValues("GetApplication").Inc()
			}
			return nil, errors.Wrapf(err, "argocd get application '%s' failed", name)
		}
		return app, nil
	}
	var result *v1alpha1.Application
	cd.cacheApplication.Range(func(key, value any) bool {
		apps := cd.getProjectApplications(key.(string))
		app, ok := apps[name]
		if ok {
			result = app.DeepCopy()
			return false
		}
		return true
	})
	if result == nil {
		blog.Warnf("argocd get application %s not found", name)
	}
	return result, nil
}

// GetApplicationRevisionsMetadata get revisions metadata for repos, adapt to multiple sources
func (cd *argo) GetApplicationRevisionsMetadata(ctx context.Context, repos,
	revisions []string) ([]*v1alpha1.RevisionMetadata, error) {
	repoClientSet := apiclient.NewRepoServerClientset(cd.option.RepoServerUrl, 60,
		apiclient.TLSConfiguration{
			DisableTLS:       false,
			StrictValidation: false,
		})
	repoCloser, repoClient, err := repoClientSet.NewRepoServerClient()
	if err != nil {
		return nil, errors.Wrapf(err, "create reposerver client failed")
	}
	defer repoCloser.Close() // nolint

	result := make([]*v1alpha1.RevisionMetadata, 0, len(repos))
	for i, repo := range repos {
		argoRepo, err := cd.argoDB.GetRepository(ctx, repo)
		if err != nil {
			return nil, errors.Wrapf(err, "get repo '%s' from db failed", repo)
		}
		revisionMetadata, err := repoClient.GetRevisionMetadata(ctx, &apiclient.RepoServerRevisionMetadataRequest{
			Repo:           argoRepo,
			Revision:       revisions[i],
			CheckSignature: false,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "get revision metadata '%s/%s' failed", repo, revisions[i])
		}
		result = append(result, revisionMetadata)
	}
	return result, nil
}

// UpdateApplicationSpec will return application by name
func (cd *argo) UpdateApplicationSpec(
	ctx context.Context, spec *applicationpkg.ApplicationUpdateSpecRequest) (*v1alpha1.ApplicationSpec, error) {
	resp, err := cd.appClient.UpdateSpec(ctx, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "update application spec '%s' failed", *spec.Name)
	}
	return resp, nil
}

// GetApplicationResourceTree returns the resource tree of application
func (cd *argo) GetApplicationResourceTree(ctx context.Context, name string) (*v1alpha1.ApplicationTree, error) {
	resp, err := cd.appClient.ResourceTree(ctx, &applicationpkg.ResourcesQuery{
		ApplicationName: &name,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get application '%s' resource tree failed", name)
	}
	return resp, nil
}

// GetApplicationManifests returns the manifests result of application
func (cd *argo) GetApplicationManifests(
	ctx context.Context, name, revision string) (*apiclient.ManifestResponse, error) {
	resp, err := cd.appClient.GetManifests(ctx, &appclient.ApplicationManifestQuery{
		Name:     &name,
		Revision: &revision,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get manifests for app '%s/%s' failed", name, revision)
	}
	return resp, nil
}

// GetApplicationManifestsFromRepoServerWithMultiSources returns the manifests result of application which not
// created. This function will direct call reposerver of argocd
// nolint
func (cd *argo) GetApplicationManifestsFromRepoServerWithMultiSources(ctx context.Context,
	application *v1alpha1.Application) ([]*apiclient.ManifestResponse, error) {
	repoUrl := application.Spec.Source.RepoURL
	repo, err := cd.argoDB.GetRepository(ctx, repoUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "get repo '%s' failed", repoUrl)
	}
	if repo == nil {
		return nil, errors.Wrapf(err, "get repo '%s' not found", repoUrl)
	}
	repoClientSet := apiclient.NewRepoServerClientset(cd.option.RepoServerUrl, 60,
		apiclient.TLSConfiguration{
			DisableTLS:       false,
			StrictValidation: false,
		})
	repoCloser, repoClient, err := repoClientSet.NewRepoServerClient()
	if err != nil {
		return nil, errors.Wrapf(err, "create reposerver client failed")
	}
	defer repoCloser.Close()

	// Store the map of all sources having ref field into a map for applications with sources field
	refSources, err := utilargo.GetRefSources(context.Background(), application.Spec, cd.argoDB)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ref sources")
	}
	sources := make([]v1alpha1.ApplicationSource, 0)
	if application.Spec.HasMultipleSources() {
		sources = append(sources, application.Spec.Sources...)
	} else {
		sources = append(sources, *application.Spec.Source)
	}

	result := make([]*apiclient.ManifestResponse, 0)
	for i := range sources {
		var resp *apiclient.ManifestResponse
		resp, err = repoClient.GenerateManifest(ctx, &apiclient.ManifestRequest{
			Repo:               repo,
			Revision:           sources[i].TargetRevision,
			Namespace:          application.Spec.Destination.Namespace,
			AppName:            application.Name,
			RefSources:         refSources,
			HasMultipleSources: application.Spec.HasMultipleSources(),
			ApplicationSource:  &sources[i],
		})
		if err != nil {
			return nil, errors.Wrapf(err, "generate manifests failed")
		}
		result = append(result, resp)
	}
	return result, nil
}

// ListApplications interface
func (cd *argo) ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (
	*v1alpha1.ApplicationList, error) {
	if !cd.cacheSynced.Load() {
		apps, err := cd.appClient.List(ctx, query)
		if err != nil {
			if !utils.IsContextCanceled(err) {
				metric.ManagerArgoOperateFailed.WithLabelValues("ListApplications").Inc()
			}
			return nil, errors.Wrapf(err, "argocd list application for project '%v' failed", query.Projects)
		}
		return apps, nil
	}
	var (
		selector labels.Selector
		err      error
	)
	if query.Selector != nil && *query.Selector != "" {
		selector, err = labels.Parse(*query.Selector)
		if err != nil {
			return nil, errors.Wrapf(err, "argocd list application for project '%v' failed", query.Projects)
		}
	}
	result := &v1alpha1.ApplicationList{
		Items: make([]v1alpha1.Application, 0),
	}
	for i := range query.Projects {
		projName := query.Projects[i]
		projApps := cd.getProjectApplications(projName)
		for _, v := range projApps {
			if query.Name != nil && (*query.Name != "" && *query.Name != v.Name) {
				continue
			}
			if query.Repo != nil {
				queryRepo := *query.Repo
				if v.Spec.HasMultipleSources() {
					consistent := false
					for _, source := range v.Spec.Sources {
						if queryRepo == source.RepoURL {
							consistent = true
							break
						}
					}
					if !consistent {
						continue
					}
				} else if queryRepo != v.Spec.Source.RepoURL {
					continue
				}
			}
			if query.AppNamespace != nil && (*query.AppNamespace != "" && *query.AppNamespace !=
				v.Spec.Destination.Namespace) {
				continue
			}
			if query.Selector != nil && (*query.Selector != "" && !selector.Matches(labels.Set(v.Labels))) {
				continue
			}
			result.Items = append(result.Items, *v.DeepCopy())
		}
	}
	return result, nil
}

// GetToken authentication token
func (cd *argo) GetToken(ctx context.Context) string {
	return cd.token
}

// ApplicationResource defines the source of application, it same to ResourceStatus
type ApplicationResource struct {
	ResourceName string `json:"resourceName"`
	Kind         string `json:"kind"`
	Namespace    string `json:"namespace"`
	Group        string `json:"group"`
	Version      string `json:"version"`
}

// ApplicationDeleteResourceResult defines the resource deletion result
type ApplicationDeleteResourceResult struct {
	Succeeded  bool   `json:"succeeded"`
	ErrMessage string `json:"errMessage"`
}

// nolint unused
func buildResourceKeyWithResourceStatus(resource *v1alpha1.ResourceStatus) string {
	return strings.Join([]string{resource.Kind, resource.Namespace, resource.Group,
		resource.Version, resource.Name}, "/")
}

func buildResourceKeyWithCustomResource(resource *ApplicationResource) string {
	return strings.Join([]string{resource.Kind, resource.Namespace, resource.Group,
		resource.Version, resource.ResourceName}, "/")
}

// DeleteApplicationResource will delete all resources for application
func (cd *argo) DeleteApplicationResource(ctx context.Context, application *v1alpha1.Application,
	resources []*ApplicationResource) []ApplicationDeleteResourceResult {
	var result []ApplicationDeleteResourceResult
	if len(resources) == 0 {
		result = make([]ApplicationDeleteResourceResult, 0, len(application.Status.Resources))
		for _, resource := range application.Status.Resources {
			if err := cd.deleteApplicationResource(ctx, application, &resource); err != nil {
				result = append(result, ApplicationDeleteResourceResult{
					Succeeded:  false,
					ErrMessage: err.Error(),
				})
			} else {
				result = append(result, ApplicationDeleteResourceResult{
					Succeeded: true,
				})
			}
		}
		return result
	}
	result = make([]ApplicationDeleteResourceResult, 0, len(resources))
	for _, appResource := range resources {
		key := buildResourceKeyWithCustomResource(appResource)
		if err := cd.deleteApplicationResource(ctx, application, &v1alpha1.ResourceStatus{
			Name:      appResource.ResourceName,
			Kind:      appResource.Kind,
			Namespace: appResource.Namespace,
			Group:     appResource.Group,
			Version:   appResource.Version,
		}); err != nil {
			result = append(result, ApplicationDeleteResourceResult{
				Succeeded:  false,
				ErrMessage: fmt.Sprintf("resource '%s' delete failed: %s", key, err.Error()),
			})
		} else {
			result = append(result, ApplicationDeleteResourceResult{
				Succeeded: true,
			})
		}
	}
	return result
}

func (cd *argo) deleteApplicationResource(ctx context.Context, application *v1alpha1.Application,
	resource *v1alpha1.ResourceStatus) error {
	server := application.Spec.Destination.Server
	requestID := ctx.Value(traceconst.RequestIDHeaderKey).(string)
	_, err := cd.appClient.DeleteResource(ctx, &appclient.ApplicationResourceDeleteRequest{
		Name:         &application.Name,
		Kind:         &resource.Kind,
		Namespace:    &resource.Namespace,
		Group:        &resource.Group,
		Version:      &resource.Version,
		ResourceName: &resource.Name,
	})
	if err != nil {
		if resource.Status != v1alpha1.SyncStatusCodeSynced {
			// nolint goconst
			blog.Warnf("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' "+
				"with status '%s', noneed care: %s",
				requestID, resource.Group, resource.Kind, resource.Name,
				server, application.Name, resource.Status, err.Error())
			return nil
		}
		if utils.IsArgoResourceNotFound(err) {
			blog.Warnf("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' "+
				"got 'Not Found': %s",
				requestID, resource.Group, resource.Kind, resource.Name,
				server, application.Name, err.Error())
			return nil
		}
		if utils.IsArgoNotFoundAsPartOf(err) {
			blog.Warnf("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' "+
				"got 'not found as part of': %s",
				requestID, resource.Group, resource.Kind, resource.Name,
				server, application.Name, err.Error())
			return nil
		}
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("DeleteApplicationResource").Inc()
		}
		return errors.Wrapf(err, "argocd delete resource '%s/%s/%s' failed for cluster '%s'",
			resource.Group, resource.Kind, resource.Name, server)
	}
	blog.Infof("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' success",
		requestID, resource.Group, resource.Kind, resource.Name, server, application.Name)
	return nil
}

// GetApplicationSet query the ApplicationSet by name
func (cd *argo) GetApplicationSet(ctx context.Context, name string) (*v1alpha1.ApplicationSet, error) {
	appset, err := cd.appsetClient.Get(ctx, &appsetpkg.ApplicationSetGetQuery{
		Name: name,
	})
	if err != nil {
		if utils.IsArgoResourceNotFound(err) {
			return nil, nil
		}
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("GetApplicationSet").Inc()
		}
		return nil, errors.Wrapf(err, "argocd get applicationset '%s' failed", name)
	}
	return appset, nil
}

// ListApplicationSets list applicationsets by projects
func (cd *argo) ListApplicationSets(ctx context.Context, query *appsetpkg.ApplicationSetListQuery) (
	*v1alpha1.ApplicationSetList, error) {
	appsets, err := cd.appsetClient.List(ctx, query)
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListApplicationSets").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list applicationsets by project '%v' failed", *query)
	}
	return appsets, nil
}

func (cd *argo) initAppHistoryStore() error {
	if !cd.option.CacheHistory {
		return nil
	}
	cd.historyStore = &appHistoryStore{
		num:       10,
		appClient: cd.appClient,
		db:        dao.GlobalDB(),
	}
	cd.historyStore.init()
	return nil
}

func (cd *argo) initToken() error {
	// authorization doc: https://argo-cd.readthedocs.io/en/stable/developer-guide/api-docs/
	// $ curl $ARGOCD_SERVER/api/v1/session -d $'{"username":"admin","password":"password"}'
	// {"token":"...jwttoken info..."}
	// set token to http request header
	req := map[string]string{
		"username": cd.option.User,
		"password": cd.option.Pass,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("auth session request format error")
	}
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
		},
	}
	argoUrl := fmt.Sprintf("https://%s/api/v1/session", cd.option.Service)
	response, err := client.Post(
		argoUrl,
		"application/json",
		bytes.NewBuffer(reqBytes),
	)
	if err != nil {
		return errors.Wrapf(err, "argocd login session fatal")
	}
	defer response.Body.Close() // nolint
	bs, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Wrapf(err, "init token response body read failed")
	}
	result := make(map[string]string)
	if err = json.Unmarshal(bs, &result); err != nil {
		return errors.Wrapf(err, "decode gitops session result '%s' fatal", string(bs))
	}
	t, ok := result["token"]
	if !ok {
		return fmt.Errorf("found no login token in response")
	}
	blog.Infof("[store] argocd token session init OK, %s", t)
	cd.token = t
	return nil
}

// initBasicClient 会创建与 argocd server 的链接，并维护链接的状态. 当检测到状态不正常时，将链接
// 重置并进行重连
func (cd *argo) initBasicClient() error {
	var err error
	// init basic client
	cd.basicOpt = &api.ClientOptions{
		ServerAddr: cd.option.Service,
		AuthToken:  cd.token,
	}
	if err = cd.connect(); err != nil {
		return errors.Wrapf(err, "argocd connect grpc failed")
	}
	go func() {
		for {
			// 当链接发生故障，则进行重连; 重连成功，则持续检查状态; 重连失败则间隔重连
			if cd.conn == nil {
				metric.ManagerArgoConnectionNum.WithLabelValues().Inc()
				if err = cd.connect(); err != nil {
					blog.Errorf("[store] argocd grpc connection connect failed: %s", err.Error())
					time.Sleep(5 * time.Second)
					continue
				}
				blog.Infof("[store] argocd grpc connection re-connect success")
			}
			checkTicker := time.NewTicker(3 * time.Second)
			for range checkTicker.C {
				// 检测 GRPC 连接状态是否断开
				state := cd.conn.GetState()
				if state != connectivity.TransientFailure && state != connectivity.Shutdown {
					metric.ManagerArgoConnectionStatus.WithLabelValues().Set(0)
					continue
				}
				metric.ManagerArgoConnectionStatus.WithLabelValues().Set(1)
				blog.Errorf("[store] argocd grpc connection disconnect: %s", state.String())
				// 发生链接断开问题，则重置链接
				if cd.connCloser != nil {
					cd.connCloser.Close() // nolint
				}
				cd.conn = nil
				checkTicker.Stop()
				break
			}
		}
	}()
	return nil
}

// connect 创建到 argocd server 的 gRPC 链接，并根据链接创建对应的操作 Client
func (cd *argo) connect() error {
	var err error
	cd.conn, cd.connCloser, err = argoconn.NewConn(cd.basicOpt)
	if err != nil {
		return errors.Wrapf(err, "create connection to argocd failed")
	}
	blog.Infof("[store] create connection success")
	cd.appClient = applicationpkg.NewApplicationServiceClient(cd.conn)
	cd.repoClient = repositorypkg.NewRepositoryServiceClient(cd.conn)
	cd.projectClient = projectpkg.NewProjectServiceClient(cd.conn)
	cd.clusterClient = clusterpkg.NewClusterServiceClient(cd.conn)
	cd.appsetClient = appsetpkg.NewApplicationSetServiceClient(cd.conn)
	return nil
}

// initCache 初始化缓存, 当前缓存为: ApplicationCache
func (cd *argo) initCache() error {
	list, err := cd.appClient.List(context.Background(), &applicationpkg.ApplicationQuery{})
	if err != nil {
		return errors.Wrapf(err, "list applications failed when init watch")
	}
	for i := range list.Items {
		app := list.Items[i]
		projectApps, ok := cd.cacheApplication.Load(app.Spec.Project)
		if !ok {
			cd.cacheApplication.Store(app.Spec.Project, map[string]*v1alpha1.Application{
				app.Name: &app,
			})
		} else {
			projectApps.(map[string]*v1alpha1.Application)[app.Name] = &app
		}
	}
	blog.Infof("[store] init cache success.")
	cd.cacheSynced.Store(true)
	return nil
}

// handleApplicationWatch 监听 Application 的事件，并刷新缓存中的数据
func (cd *argo) handleApplicationWatch() error {
	watchClient, err := cd.appClient.Watch(context.Background(), &applicationpkg.ApplicationQuery{})
	if err != nil {
		return errors.Wrapf(err, "init application watch failed")
	}
	go func() {
		blog.Infof("[store] application watch started")
		for {
			// 如果 watchClient 是空的，我们需要去重连
			if watchClient == nil {
				cd.cacheSynced.Store(false)
				watchClient, err = cd.appClient.Watch(context.Background(), &applicationpkg.ApplicationQuery{})
				if err != nil {
					blog.Error("[store] application watch client recreated failed")
					// 重连如果失败，在 10s 后继续重连
					time.Sleep(10 * time.Second)
					continue
				}
				blog.Infof("[store] application watch client re-created")
				// 当重连成功之后，刷新所有缓存，并重新开始监听事件
				if err = cd.initCache(); err != nil {
					blog.Errorf("[store] cache synced failed: %s", err.Error())
				} else {
					blog.Infof("[store] cache synced success")
				}
			}
			var event *v1alpha1.ApplicationWatchEvent
			if event, err = watchClient.Recv(); err != nil {
				blog.Errorf("[store] application watch received error: %s", err.Error())
				// 如果 watchClient 接收到错误，直接设置为空进行重连逻辑
				watchClient = nil
				continue
			}

			if event.Type != watch.Modified {
				blog.Infof("[store] application watch received: %s/%s", string(event.Type), event.Application.Name)
			}
			application := event.Application
			switch event.Type {
			case watch.Added, watch.Modified:
				cd.storeApplication(&application)
				cd.historyStore.enqueue(application)
			case watch.Deleted:
				cd.deleteApplication(&application)
			}
		}
	}()
	return nil
}

func (cd *argo) getProjectApplications(projName string) map[string]*v1alpha1.Application {
	cd.RLock()
	defer cd.RUnlock()
	v, ok := cd.cacheApplication.Load(projName)
	if !ok {
		return map[string]*v1alpha1.Application{}
	}
	result := v.(map[string]*v1alpha1.Application)
	newResult := make(map[string]*v1alpha1.Application)
	for appName, app := range result {
		newResult[appName] = app.DeepCopy()
	}
	return newResult
}

func (cd *argo) storeApplication(application *v1alpha1.Application) {
	cd.Lock()
	defer cd.Unlock()
	projName := application.Spec.Project
	projectApps, ok := cd.cacheApplication.Load(projName)
	if !ok {
		cd.cacheApplication.Store(projName, map[string]*v1alpha1.Application{
			application.Name: application.DeepCopy(),
		})
	} else {
		projectApps.(map[string]*v1alpha1.Application)[application.Name] = application.DeepCopy()
	}
}

func (cd *argo) deleteApplication(application *v1alpha1.Application) {
	cd.Lock()
	defer cd.Unlock()
	projName := application.Spec.Project
	projectApps, ok := cd.cacheApplication.Load(projName)
	if ok {
		delete(projectApps.(map[string]*v1alpha1.Application), application.Name)
	}
}
