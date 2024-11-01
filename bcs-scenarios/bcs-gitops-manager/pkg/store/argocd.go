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
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/argoproj/argo-cd/v2/applicationset/generators"
	"github.com/argoproj/argo-cd/v2/applicationset/services"
	appsetutils "github.com/argoproj/argo-cd/v2/applicationset/utils"
	argocommon "github.com/argoproj/argo-cd/v2/common"
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
	argopkg "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	argoutil "github.com/argoproj/argo-cd/v2/util/argo"
	utilargo "github.com/argoproj/argo-cd/v2/util/argo"
	"github.com/argoproj/argo-cd/v2/util/argo/normalizers"
	"github.com/argoproj/argo-cd/v2/util/db"
	settings_util "github.com/argoproj/argo-cd/v2/util/settings"
	gitopsdiff "github.com/argoproj/gitops-engine/pkg/diff"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
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
	argoK8SClient *argopkg.ArgoprojV1alpha1Client

	cacheSynced      atomic.Bool
	cacheApplication *sync.Map
	cacheAppSet      *sync.Map
	cacheAppProject  *sync.Map
}

// Init control interface
func (cd *argo) Init() error {
	initializer := []func() error{
		cd.initToken, cd.initBasicClient, cd.initArgoK8SClient, cd.initAppHistoryStore,
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
		if err := cd.handleAppSetWatch(); err != nil {
			return errors.Wrapf(err, "handle applicationset watch failed")
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

// GetArgoDB return argodb object
func (cd *argo) GetArgoDB() db.ArgoDB {
	return cd.argoDB
}

// GetAppClient return argo app client
func (cd *argo) GetAppClient() applicationpkg.ApplicationServiceClient {
	return cd.appClient
}

// Stop control interface
func (cd *argo) Stop() {
}

// GetOptions return the options of gitops
func (cd *argo) GetOptions() *Options {
	return cd.option
}

// ReturnArgoK8SClient return the argo k8s client
func (cd *argo) ReturnArgoK8SClient() *argopkg.ArgoprojV1alpha1Client {
	return cd.argoK8SClient
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
	proj, err := cd.projectClient.Create(ctx, &project.ProjectCreateRequest{Project: pro})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("CreateProject").Inc()
		}
		return errors.Wrapf(err, "argocd create project '%s' failed", pro.GetName())
	}
	if cd.option.Cache {
		cd.cacheAppProject.Store(proj.Name, proj)
	}
	return nil
}

// UpdateProject interface
func (cd *argo) UpdateProject(ctx context.Context, pro *v1alpha1.AppProject) error {
	newProj, err := cd.projectClient.Update(ctx, &project.ProjectUpdateRequest{Project: pro})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("UpdateProject").Inc()
		}
		return errors.Wrapf(err, "argocd update project '%s' failed", pro.GetName())
	}
	if cd.option.Cache {
		cd.cacheAppProject.Store(newProj.Name, newProj)
	}
	return nil
}

// GetProject interface
func (cd *argo) GetProject(ctx context.Context, name string) (*v1alpha1.AppProject, error) {
	if cd.cacheSynced.Load() {
		obj, ok := cd.cacheAppProject.Load(name)
		if ok {
			proj := obj.(*v1alpha1.AppProject)
			return proj.DeepCopy(), nil
		}
	}

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
	cd.cacheAppProject.Store(name, pro)
	return pro, nil
}

// ListProjects interface
func (cd *argo) ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, error) {
	if cd.cacheSynced.Load() {
		items := make([]v1alpha1.AppProject, 0)
		cd.cacheAppProject.Range(func(k, v any) bool {
			proj := v.(*v1alpha1.AppProject)
			items = append(items, *proj.DeepCopy())
			return true
		})
		return &v1alpha1.AppProjectList{Items: items}, nil
	}

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
			// ignore this metric
			// metric.ManagerArgoOperateFailed.WithLabelValues("CreateCluster").Inc()
			return errors.Wrapf(err, "cluster not normal")
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
func (cd *argo) ListClustersByProject(ctx context.Context, projID string) (*v1alpha1.ClusterList, error) {
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
		if projectID == projID {
			clusters = append(clusters, item)
		}
	}
	cls.Items = clusters
	return cls, nil
}

// ListClustersByProjectName will list clusters by project name
func (cd *argo) ListClustersByProjectName(ctx context.Context, projectName string) (*v1alpha1.ClusterList, error) {
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
		if item.Project == projectName {
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
	allApps := cd.getAllApplications()
	for _, v := range allApps {
		for _, app := range v {
			result = append(result, app)
		}
	}
	return result
}

// TerminateAppOperation terminate application operation
func (cd *argo) TerminateAppOperation(ctx context.Context, req *applicationpkg.OperationTerminateRequest) error {
	_, err := cd.appClient.TerminateOperation(ctx, req)
	return err
}

// GetApplication will return application by name
func (cd *argo) GetApplication(ctx context.Context, name string) (*v1alpha1.Application, error) {
	if !cd.cacheSynced.Load() {
		app, err := cd.appClient.Get(ctx, &appclient.ApplicationQuery{Name: &name})
		if err != nil {
			if utils.IsArgoResourceNotFound(err) {
				return nil, nil
			}
			if !utils.IsContextCanceled(err) {
				metric.ManagerArgoOperateFailed.WithLabelValues("GetApplication").Inc()
			}
			return nil, errors.Wrapf(err, "argocd get application '%s' failed", name)
		}
		return app, nil
	}
	argoAppList, err := cd.ListProjects(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "argocd list projects failed")
	}
	var result *v1alpha1.Application
	for i := range argoAppList.Items {
		argoProj := &argoAppList.Items[i]
		if !strings.HasPrefix(name, argoProj.Name) {
			continue
		}
		apps := cd.getProjectApplications(argoProj.Name)
		app, ok := apps[name]
		if ok {
			result = app.DeepCopy()
			break
		}
	}
	return result, nil
}

// GetApplicationRevisionsMetadata get revisions metadata for repos, adapt to multiple sources
func (cd *argo) GetApplicationRevisionsMetadata(ctx context.Context, repos,
	revisions []string) ([]*v1alpha1.RevisionMetadata, error) {
	repoClientSet := apiclient.NewRepoServerClientset(cd.option.RepoServerUrl, 300,
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

// PatchApplicationResource patch application resources
func (cd *argo) PatchApplicationResource(ctx context.Context, appName string, resource *v1alpha1.ResourceStatus,
	patch, patchType string) error {
	_, err := cd.appClient.PatchResource(ctx, &appclient.ApplicationResourcePatchRequest{
		Name:         &appName,
		Namespace:    &resource.Namespace,
		ResourceName: &resource.Name,
		Version:      &resource.Version,
		Group:        &resource.Group,
		Kind:         &resource.Kind,
		PatchType:    &patchType,
		Patch:        &patch,
	})
	if err != nil {
		return errors.Wrapf(err, "patch '%s/%s/%s-%s' failed", resource.Group, resource.Version,
			resource.Kind, resource.Name)
	}
	return nil
}

// PatchApplicationAnnotation will update application annotations
func (cd *argo) PatchApplicationAnnotation(ctx context.Context, appName, namespace string,
	annotations map[string]interface{}) error {
	patch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})
	if err != nil {
		return errors.Wrap(err, "marshal annotation error")
	}
	if _, err := cd.argoK8SClient.Applications(namespace).
		Patch(ctx, appName, types.MergePatchType, patch,
			metav1.PatchOptions{}); err != nil {
		return errors.Wrapf(err, "patch application '%s' annotations '%v' error", appName, annotations)
	}
	return nil
}

// GetApplicationManifestsFromRepoServerWithMultiSources returns the manifests result of application which not
// created. This function will direct call reposerver of argocd
// nolint
func (cd *argo) GetApplicationManifestsFromRepoServerWithMultiSources(ctx context.Context,
	application *v1alpha1.Application) ([]*apiclient.ManifestResponse, error) {
	repoUrl := application.Spec.GetSource().RepoURL
	repo, err := cd.argoDB.GetRepository(ctx, repoUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "get repo '%s' failed", repoUrl)
	}
	if repo == nil {
		return nil, errors.Wrapf(err, "get repo '%s' not found", repoUrl)
	}
	repoClientSet := apiclient.NewRepoServerClientset(cd.option.RepoServerUrl, 300,
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
	sources = append(sources, application.Spec.GetSource())

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
			KustomizeOptions: &v1alpha1.KustomizeOptions{
				BuildOptions: "--enable-alpha-plugins",
			},
			ApplicationSource: &sources[i],
		})
		if err != nil {
			return nil, errors.Wrapf(err, "generate manifests failed")
		}
		result = append(result, resp)
	}
	return result, nil
}

func (cd *argo) checkAppMatchQuery(app *v1alpha1.Application, query *appclient.ApplicationQuery,
	selector labels.Selector) bool {
	if query.Name != nil && (*query.Name != "" && *query.Name != app.Name) {
		return false
	}
	if query.Repo != nil {
		queryRepo := *query.Repo
		if app.Spec.HasMultipleSources() {
			consistent := false
			for _, source := range app.Spec.Sources {
				if queryRepo == source.RepoURL {
					consistent = true
					break
				}
			}
			if !consistent {
				return false
			}
		} else if queryRepo != app.Spec.Source.RepoURL {
			return false
		}
	}
	if query.AppNamespace != nil && (*query.AppNamespace != "" && *query.AppNamespace !=
		app.Spec.Destination.Namespace) {
		return false
	}
	if query.Selector != nil && (*query.Selector != "" && !selector.Matches(labels.Set(app.Labels))) {
		return false
	}
	return true
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

	blog.Infof("RequestID[%s] query applications with params: %s", ctxutils.RequestID(ctx),
		utils.MarshalObject(query))
	result := &v1alpha1.ApplicationList{
		Items: make([]v1alpha1.Application, 0),
	}
	if len(query.Projects) == 0 {
		allApps := cd.getAllApplications()
		for _, apps := range allApps {
			for _, app := range apps {
				if cd.checkAppMatchQuery(app, query, selector) {
					result.Items = append(result.Items, *app)
				}
			}
		}
		return result, nil
	}
	for i := range query.Projects {
		projName := query.Projects[i]
		projApps := cd.getProjectApplications(projName)
		for _, app := range projApps {
			if cd.checkAppMatchQuery(app, query, selector) {
				result.Items = append(result.Items, *app)
			}
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
	SyncWave     int64  `json:"syncWave,omitempty"`
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
		resources = make([]*ApplicationResource, 0)
	}
	for _, resource := range application.Status.Resources {
		resources = append(resources, &ApplicationResource{
			ResourceName: resource.Name,
			Kind:         resource.Kind,
			Namespace:    resource.Namespace,
			Group:        resource.Group,
			Version:      resource.Version,
			SyncWave:     resource.SyncWave,
		})
	}

	// 按照 syncwave 值进行排序
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].SyncWave > resources[j].SyncWave
	})

	// 按syncwave从大到小的顺序删除资源
	for _, rw := range resources {
		key := buildResourceKeyWithCustomResource(rw)
		if err := cd.deleteApplicationResource(ctx, application, &v1alpha1.ResourceStatus{
			Name:      rw.ResourceName,
			Kind:      rw.Kind,
			Namespace: rw.Namespace,
			Group:     rw.Group,
			Version:   rw.Version,
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
	retryNum := 5
	requestID := ctx.Value(traceconst.RequestIDHeaderKey).(string)
	var err error
	for i := 0; i < retryNum; i++ {
		if err = cd.handleDeleteAppResource(ctx, application, resource); err == nil {
			return nil
		}
		blog.Errorf("RequestID[%s] %s (retry: %d)", requestID, err.Error(), i)
	}
	return err
}

func (cd *argo) handleDeleteAppResource(ctx context.Context, application *v1alpha1.Application,
	resource *v1alpha1.ResourceStatus) error {
	server := application.Spec.Destination.Server
	requestID := ctx.Value(traceconst.RequestIDHeaderKey).(string)
	newCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	needForce := true
	_, err := cd.appClient.DeleteResource(newCtx, &appclient.ApplicationResourceDeleteRequest{
		Name:         &application.Name,
		Kind:         &resource.Kind,
		Namespace:    &resource.Namespace,
		Group:        &resource.Group,
		Version:      &resource.Version,
		ResourceName: &resource.Name,
		Force:        &needForce,
	})
	if err != nil {
		// 存在一些没有Health字段的资源,所以需要利用短路求值提前对health进行非空判断,防止出现panic
		if resource.Health != nil && resource.Health.Status == health.HealthStatusMissing {
			// nolint goconst
			blog.Warnf("RequestID[%s], resource '%s/%s/%s' for cluster '%s' with application '%s' is missing, "+
				"noneed care: %s", requestID, resource.Group, resource.Kind, resource.Name,
				server, application.Name, err.Error())
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
	if cd.cacheSynced.Load() {
		v, ok := cd.cacheAppSet.Load(name)
		if !ok {
			return nil, nil
		}
		return v.(*v1alpha1.ApplicationSet), nil
	}

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
	if cd.cacheSynced.Load() {
		appSets := cd.AllApplicationSets()
		result := make([]v1alpha1.ApplicationSet, 0)
		for _, appset := range appSets {
			if len(query.Projects) != 0 {
				if !slices.Contains(query.Projects, appset.Spec.Template.Spec.Project) {
					continue
				}
			}
			if query.Selector != "" {
				selector, err := labels.Parse(query.Selector)
				if err != nil {
					return nil, errors.Wrapf(err, "parse selector '%s' failed", query.Selector)
				}
				if !selector.Matches(labels.Set(appset.Labels)) {
					continue
				}
			}
			result = append(result, *appset)
		}
		return &v1alpha1.ApplicationSetList{Items: result}, nil
	}

	appsets, err := cd.appsetClient.List(ctx, query)
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListApplicationSets").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list applicationsets by project '%v' failed", *query)
	}
	return appsets, nil
}

// AllApplicationSets return app the applicationsets
func (cd *argo) AllApplicationSets() []*v1alpha1.ApplicationSet {
	result := make([]*v1alpha1.ApplicationSet, 0)
	cd.cacheAppSet.Range(func(key, value any) bool {
		appSet := value.(*v1alpha1.ApplicationSet)
		result = append(result, appSet.DeepCopy())
		return true
	})
	return result
}

var (
	render = appsetutils.Render{}
)

func (cd *argo) ApplicationSetDryRun(appSet *v1alpha1.ApplicationSet) ([]*v1alpha1.Application, error) {
	repoClientSet := apiclient.NewRepoServerClientset(cd.option.RepoServerUrl, 300,
		apiclient.TLSConfiguration{
			DisableTLS:       false,
			StrictValidation: false,
		})
	argoCDService, _ := services.NewArgoCDService(cd.argoDB, true, repoClientSet, false)
	// this will render the Applications by ApplicationSet's generators
	// refer to:
	// https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L499
	results := make([]*v1alpha1.Application, 0)
	for i := range appSet.Spec.Generators {
		generator := appSet.Spec.Generators[i]
		if generator.List == nil && generator.Git == nil && generator.Matrix == nil && generator.Merge == nil {
			continue
		}
		listGenerator := generators.NewListGenerator()
		gitGenerator := generators.NewGitGenerator(argoCDService)
		terminalGenerators := map[string]generators.Generator{
			"List": listGenerator,
			"Git":  gitGenerator,
		}
		tsResult, err := generators.Transform(generator, map[string]generators.Generator{
			"List":   listGenerator,
			"Git":    gitGenerator,
			"Matrix": generators.NewMatrixGenerator(terminalGenerators),
			"Merge":  generators.NewMergeGenerator(terminalGenerators),
		}, appSet.Spec.Template, appSet, map[string]interface{}{})
		if err != nil {
			return nil, errors.Wrapf(err, "transform generator[%d] failed", i)
		}
		for j := range tsResult {
			ts := tsResult[j]
			tmplApplication := getTempApplication(ts.Template)
			if tmplApplication.Labels == nil {
				tmplApplication.Labels = make(map[string]string)
			}
			for _, p := range ts.Params {
				var app *v1alpha1.Application
				app, err = render.RenderTemplateParams(tmplApplication, appSet.Spec.SyncPolicy,
					p, appSet.Spec.GoTemplate, nil)
				if err != nil {
					return nil, errors.Wrap(err, "error generating application from params")
				}
				results = append(results, app)
			}
		}
	}
	return results, nil
}

// refer to:
// https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L487
func getTempApplication(applicationSetTemplate v1alpha1.ApplicationSetTemplate) *v1alpha1.Application {
	var tmplApplication v1alpha1.Application
	tmplApplication.Annotations = applicationSetTemplate.Annotations
	tmplApplication.Labels = applicationSetTemplate.Labels
	tmplApplication.Namespace = applicationSetTemplate.Namespace
	tmplApplication.Name = applicationSetTemplate.Name
	tmplApplication.Spec = applicationSetTemplate.Spec
	tmplApplication.Finalizers = applicationSetTemplate.Finalizers

	return &tmplApplication
}

// RefreshApplicationSet refresh appset trigger it to generate applications
func (cd *argo) RefreshApplicationSet(namespace, name string) error {
	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				argocommon.AnnotationApplicationSetRefresh: "true",
			},
		},
	}
	patch, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrapf(err, "error marshaling metadata")
	}
	for attempt := 0; attempt < 5; attempt++ {
		_, err = cd.argoK8SClient.ApplicationSets(namespace).Patch(context.Background(), name,
			types.MergePatchType, patch, metav1.PatchOptions{})
		if err == nil {
			return nil
		}
		if !apierr.IsConflict(err) {
			return errors.Wrapf(err, "error patching annotations for appset %s/%s", namespace, name)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.Wrapf(err, "error paching annotation for appset %s/%s with conflict", namespace, name)
}

// DeleteApplicationSetOrphan delete application-set with orphan
func (cd *argo) DeleteApplicationSetOrphan(ctx context.Context, name string) error {
	deleteOption := metav1.DeletePropagationOrphan
	if err := cd.argoK8SClient.ApplicationSets(cd.option.AdminNamespace).
		Delete(ctx, name, metav1.DeleteOptions{
			PropagationPolicy: &deleteOption,
		}); err != nil {
		return errors.Wrapf(err, "delete application-set failed")
	}
	return nil
}

var (
	commitSHARegex          = regexp.MustCompile("^[0-9A-Fa-f]{40}$")
	truncatedCommitSHARegex = regexp.MustCompile("^[0-9A-Fa-f]{7,}$")
)

// isTruncatedCommitSHA returns whether or not a string is a truncated  SHA-1
func isTruncatedCommitSHA(sha string) bool {
	return truncatedCommitSHARegex.MatchString(sha)
}

// isCommitSHA returns whether or not a string is a 40 character SHA-1
func isCommitSHA(sha string) bool {
	return commitSHARegex.MatchString(sha)
}

// GetRepoLastCommitID get the last commit-id
func (cd *argo) GetRepoLastCommitID(ctx context.Context, repoUrl, revision string) (string, error) {
	if isCommitSHA(revision) {
		return revision, nil
	}
	repoAuth, err := cd.buildRepoAuth(ctx, repoUrl)
	if err != nil {
		return "", err
	}
	memStore := memory.NewStorage()
	remoteRepo := git.NewRemote(memStore, &config.RemoteConfig{
		Name: repoUrl,
		URLs: []string{repoUrl},
	})
	refs, err := remoteRepo.List(&git.ListOptions{
		Auth: repoAuth,
	})
	if err != nil {
		return "", errors.Wrapf(err, "git repo '%s' fetch refs failed", repoUrl)
	}
	// refToHash keeps a maps of remote refs to their hash
	// (e.g. refs/heads/master -> a67038ae2e9cb9b9b16423702f98b41e36601001)
	refToHash := make(map[string]string)
	// refToResolve remembers ref name of the supplied revision if we determine the revision is a
	// symbolic reference (like HEAD), in which case we will resolve it from the refToHash map
	refToResolve := ""
	for _, ref := range refs {
		refName := ref.Name().String()
		hash := ref.Hash().String()
		if ref.Type() == plumbing.HashReference {
			refToHash[refName] = hash
		}
		if ref.Name().Short() == revision || refName == revision {
			if ref.Type() == plumbing.HashReference {
				return hash, nil
			}
			if ref.Type() == plumbing.SymbolicReference {
				refToResolve = ref.Target().String()
			}
		}
	}
	if refToResolve != "" {
		// If refToResolve is non-empty, we are resolving symbolic reference (e.g. HEAD).
		// It should exist in our refToHash map
		if hash, ok := refToHash[refToResolve]; ok {
			return hash, nil
		}
	}
	if isTruncatedCommitSHA(revision) {
		return revision, nil
	}

	return "", errors.Errorf("unable to resolve '%s' to a commit SHA", revision)
}

func (cd *argo) buildRepoAuth(ctx context.Context, repoUrl string) (transport.AuthMethod, error) {
	argoRepo, err := cd.argoDB.GetRepository(ctx, repoUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "get repository '%s' from argo db failed", repoUrl)
	}
	if argoRepo == nil {
		return nil, errors.Errorf("repository '%s' not found", repoUrl)
	}
	if argoRepo.Username != "" && argoRepo.Password != "" {
		return &githttp.BasicAuth{
			Username: argoRepo.Username,
			Password: argoRepo.Password,
		}, nil
	}
	if argoRepo.SSHPrivateKey != "" {
		var publicKeys *ssh.PublicKeys
		publicKeys, err = ssh.NewPublicKeys("git", []byte(argoRepo.SSHPrivateKey), "")
		if err != nil {
			return nil, errors.Wrapf(err, "create public keys failed")
		}
		return publicKeys, nil
	}
	return nil, errors.Errorf("not https/ssh authentication")
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

func (cd *argo) initArgoK8SClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get k8s incluster config failed")
	}
	cd.argoK8SClient, err = argopkg.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create argo k8s client failed")
	}
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
	blog.Infof("[store] init cache application success.")

	appsetList, err := cd.appsetClient.List(context.Background(), &appsetpkg.ApplicationSetListQuery{})
	if err != nil {
		return errors.Wrapf(err, "list applicationsets failed when init watch")
	}
	for i := range appsetList.Items {
		appSet := appsetList.Items[i]
		cd.cacheAppSet.Store(appSet.Name, &appSet)
	}
	blog.Infof("[store] init cache appset success")

	projList, err := cd.projectClient.List(context.Background(), &projectpkg.ProjectQuery{})
	if err != nil {
		return errors.Wrapf(err, "list projec failed when init watch")
	}
	for i := range projList.Items {
		proj := projList.Items[i]
		cd.cacheAppProject.Store(proj.Name, &proj)
	}
	blog.Infof("[store] init cache project success.")
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
				if cd.historyStore != nil {
					cd.historyStore.enqueue(application)
				}
			case watch.Deleted:
				cd.deleteApplication(&application)
			}
		}
	}()
	return nil
}

// handleAppSetWatch create the appset watch client to watch appset changes
// and store the changed appset into cache
func (cd *argo) handleAppSetWatch() error {
	watchClient, err := cd.argoK8SClient.ApplicationSets(cd.option.AdminNamespace).
		Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "create applicationset watch client failed")
	}
	go func() {
		defer watchClient.Stop()

		for {
			for e := range watchClient.ResultChan() {
				eventAppSet, ok := e.Object.(*v1alpha1.ApplicationSet)
				if !ok {
					blog.Errorf("[store] appset event object convert type failed")
					continue
				}
				if e.Type != watch.Modified {
					blog.Infof("[store] appset watch received: %s/%s", string(e.Type), eventAppSet.Name)
				}
				switch e.Type {
				case watch.Added, watch.Modified:
					cd.cacheAppSet.Store(eventAppSet.Name, eventAppSet)
				case watch.Deleted:
					cd.cacheAppSet.Delete(eventAppSet.Name)
				}
			}

			// reconnect appset watch if channel is closed
			blog.Errorf("[store] appset watch chan is closed")
			for {
				if watchClient != nil {
					watchClient.Stop()
				}
				time.Sleep(5 * time.Second)
				watchClient, err = cd.argoK8SClient.ApplicationSets(cd.option.AdminNamespace).
					Watch(context.Background(), metav1.ListOptions{})
				if err != nil {
					blog.Error("[store] appset watch re-connect failed: %s", err.Error())
					continue
				}
				blog.Infof("[store] appset watch re-connect success")
				break
			}
		}
	}()
	return nil
}

func (cd *argo) getAllApplications() map[string]map[string]*v1alpha1.Application {
	cd.RLock()
	defer cd.RUnlock()
	result := make(map[string]map[string]*v1alpha1.Application)
	cd.cacheApplication.Range(func(proj, projApps interface{}) bool {
		p := proj.(string)
		result[p] = make(map[string]*v1alpha1.Application)
		apps := projApps.(map[string]*v1alpha1.Application)
		for appName, app := range apps {
			result[p][appName] = app.DeepCopy()
		}
		return true
	})
	return result
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
