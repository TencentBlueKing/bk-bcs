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

	api "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"

	applicationpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	projectpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	repositorypkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/argoconn"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

type argo struct {
	option *Options
	token  string

	basicOpt      *api.ClientOptions
	conn          *grpc.ClientConn
	connCloser    io.Closer
	appClient     applicationpkg.ApplicationServiceClient
	appsetClient  appsetpkg.ApplicationSetServiceClient
	repoClient    repositorypkg.RepositoryServiceClient
	projectClient projectpkg.ProjectServiceClient
	clusterClient clusterpkg.ClusterServiceClient

	cacheSynced      atomic.Bool
	cacheApplication *sync.Map
}

// Init control interface
func (cd *argo) Init() error {
	initializer := []func() error{
		cd.initToken, cd.initBasicClient,
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

// Stop control interface
func (cd *argo) Stop() {
}

// GetOptions return the options of gitops
func (cd *argo) GetOptions() *Options {
	return cd.option
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

// CreateCluster interface
func (cd *argo) CreateCluster(ctx context.Context, cls *v1alpha1.Cluster) error {
	_, err := cd.clusterClient.Create(ctx, &cluster.ClusterCreateRequest{Cluster: cls})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("CreateCluster").Inc()
		}
		return errors.Wrapf(err, "argocd create cluster '%s' failed", cls.Name)
	}
	return nil
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

// ListRepository interface
func (cd *argo) ListRepository(ctx context.Context) (*v1alpha1.RepositoryList, error) {
	repos, err := cd.repoClient.List(ctx, &repository.RepoQuery{})
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerArgoOperateFailed.WithLabelValues("ListRepository").Inc()
		}
		return nil, errors.Wrapf(err, "argocd list repos failed")
	}
	return repos, nil
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
		apps := value.(map[string]*v1alpha1.Application)
		app, ok := apps[name]
		if ok {
			result = app
			return false
		}
		return true
	})
	if result == nil {
		blog.Warnf("argocd get application %s not found", name)
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
		projApps, ok := cd.cacheApplication.Load(projName)
		if !ok {
			continue
		}
		for _, v := range projApps.(map[string]*v1alpha1.Application) {
			if query.Name != nil && (*query.Name != "" && *query.Name != v.Name) {
				continue
			}
			if query.Repo != nil && (*query.Repo != "" && *query.Repo != v.Spec.Source.RepoURL) {
				continue
			}
			if query.AppNamespace != nil && (*query.AppNamespace != "" && *query.AppNamespace !=
				v.Spec.Destination.Namespace) {
				continue
			}
			if query.Selector != nil && (*query.Selector != "" && !selector.Matches(labels.Set(v.Labels))) {
				continue
			}
			result.Items = append(result.Items, *v)
		}
	}
	return result, nil
}

// GetToken authentication token
func (cd *argo) GetToken(ctx context.Context) string {
	return cd.token
}

// DeleteApplicationResource will delete all resources for application
func (cd *argo) DeleteApplicationResource(ctx context.Context, application *v1alpha1.Application) error {
	server := application.Spec.Destination.Server
	errs := make([]string, 0)
	for _, resource := range application.Status.Resources {
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
				blog.Warnf("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' "+
					"with status '%s', noneed care: %s",
					utils.RequestID(ctx), resource.Group, resource.Kind, resource.Name,
					server, application.Name, resource.Status, err.Error())
				continue
			}
			if utils.IsArgoResourceNotFound(err) {
				blog.Warnf("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' "+
					"got 'Not Found': %s",
					utils.RequestID(ctx), resource.Group, resource.Kind, resource.Name,
					server, application.Name, err.Error())
				continue
			}
			if !utils.IsContextCanceled(err) {
				metric.ManagerArgoOperateFailed.WithLabelValues("DeleteApplicationResource").Inc()
			}
			errs = append(errs, fmt.Sprintf("argocd delete resource '%s/%s/%s' failed for cluster '%s': %s",
				resource.Group, resource.Kind, resource.Name, server, err.Error()))
		} else {
			blog.Infof("RequestID[%s], delete resource '%s/%s/%s' for cluster '%s' with application '%s' success",
				utils.RequestID(ctx), resource.Group, resource.Kind, resource.Name, server, application.Name)
		}
	}
	if len(errs) != 0 {
		return errors.Errorf("delete application '%s' sub resource failed: %s",
			application.Name, strings.Join(errs, ","))
	}
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

func (cd *argo) initToken() error {
	// authorization doc: https://argo-cd.readthedocs.io/en/stable/developer-guide/api-docs/
	//$ curl $ARGOCD_SERVER/api/v1/session -d $'{"username":"admin","password":"password"}'
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
	result := make(map[string]string)
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return errors.Wrapf(err, "decode gitops session result fatal")
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
			projName := application.Spec.Project
			switch event.Type {
			case watch.Added, watch.Modified:
				projectApps, ok := cd.cacheApplication.Load(projName)
				if !ok {
					cd.cacheApplication.Store(projName, map[string]*v1alpha1.Application{
						application.Name: &application,
					})
				} else {
					projectApps.(map[string]*v1alpha1.Application)[application.Name] = &application
				}
			case watch.Deleted:
				projectApps, ok := cd.cacheApplication.Load(projName)
				if ok {
					delete(projectApps.(map[string]*v1alpha1.Application), application.Name)
				}
			}
		}
	}()
	return nil
}
