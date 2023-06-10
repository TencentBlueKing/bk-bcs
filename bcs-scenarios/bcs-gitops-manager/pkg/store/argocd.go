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
	"net/http"
	"net/url"
	"strings"

	api "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

type argo struct {
	option      *Options
	basicOpt    *api.ClientOptions
	basicClient api.Client
	token       string
}

// Init control interface
func (cd *argo) Init() error {
	initializer := []func() error{
		cd.initToken, cd.initBasicClient,
	}
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}

	return nil
}

// Stop control interface
func (cd *argo) Stop() {
}

// GetOptions ...
func (cd *argo) GetOptions() *Options {
	return cd.option
}

// CreateProject interface
func (cd *argo) CreateProject(ctx context.Context, pro *v1alpha1.AppProject) error {
	// NOTE: create new single connection per request
	// ! please make more attention to performance issue
	connection, client, err := cd.basicClient.NewProjectClient()
	if err != nil {
		return errors.Wrapf(err, "argocd init project client failed")
	}
	defer connection.Close() // nolint
	_, err = client.Create(ctx, &project.ProjectCreateRequest{Project: pro})
	if err != nil {
		return errors.Wrapf(err, "argocd create project '%s' failed", pro.GetName())
	}
	return nil
}

// UpdateProject interface
func (cd *argo) UpdateProject(ctx context.Context, pro *v1alpha1.AppProject) error {
	connection, client, err := cd.basicClient.NewProjectClient()
	if err != nil {
		return errors.Wrapf(err, "argocd init project client failed")
	}
	defer connection.Close() // nolint
	_, err = client.Update(ctx, &project.ProjectUpdateRequest{Project: pro})
	if err != nil {
		return errors.Wrapf(err, "argocd update project '%s' failed", pro.GetName())
	}
	return nil
}

// GetProject interface
func (cd *argo) GetProject(ctx context.Context, name string) (*v1alpha1.AppProject, error) {
	connection, client, err := cd.basicClient.NewProjectClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init project client failed")
	}
	defer connection.Close() // nolint
	pro, err := client.Get(ctx, &project.ProjectQuery{Name: name})
	if err != nil {
		// filter error that NotFound
		if strings.Contains(err.Error(), "code = NotFound") {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "argocd get project '%s' failed", name)
	}
	return pro, nil
}

// ListProjects interface
func (cd *argo) ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, error) {
	connection, client, err := cd.basicClient.NewProjectClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init project client failed")
	}
	defer connection.Close() // nolint
	pro, err := client.List(ctx, &project.ProjectQuery{})
	if err != nil {
		return nil, errors.Wrapf(err, "argocd list alll projects failed")
	}
	return pro, nil
}

// CreateCluster interface
func (cd *argo) CreateCluster(ctx context.Context, cls *v1alpha1.Cluster) error {
	// NOTE: create new single connection per request
	// ! please make more attention to performance issue
	connection, client, err := cd.basicClient.NewClusterClient()
	if err != nil {
		return errors.Wrapf(err, "argocd init cluster client failed")
	}
	defer connection.Close() // nolint
	_, err = client.Create(ctx, &cluster.ClusterCreateRequest{Cluster: cls})
	if err != nil {
		return errors.Wrapf(err, "argocd create cluster '%s' failed", cls.Name)
	}
	return nil
}

// DeleteCluster delete cluster by clusterID
func (cd *argo) DeleteCluster(ctx context.Context, name string) error {
	connection, client, err := cd.basicClient.NewClusterClient()
	if err != nil {
		return errors.Wrapf(err, "argocd init cluster client failed")
	}
	defer connection.Close() // nolint
	if _, err = client.Delete(ctx, &cluster.ClusterQuery{Name: name}); err != nil {
		// argocd return 403(PermissionDenied) when cluster do not exist
		// !make sure that gitops-manager has admin access
		if strings.Contains(err.Error(), "code = PermissionDenied") {
			blog.Warnf("argocd delete cluster %s warning: %s", name, err.Error())
			return nil
		}
		return errors.Wrapf(err, "delete cluster failed")
	}
	return nil
}

// GetCluster interface
func (cd *argo) GetCluster(ctx context.Context, name string) (*v1alpha1.Cluster, error) {
	// create new single connection per request
	// ! please make more attention to performance issue
	connection, client, err := cd.basicClient.NewClusterClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init cluster client failed")
	}
	defer connection.Close() // nolint
	cls, err := client.Get(ctx, &cluster.ClusterQuery{Name: name})
	if err != nil {
		// argocd return 403(PermissionDenied) when cluster do not exist
		// !make sure that gitops-manager has admin access
		if strings.Contains(err.Error(), "code = PermissionDenied") {
			blog.Warnf("argocd get cluster %s warning, No Cluster Found if admin access, %s", name, err.Error())
			return nil, nil
		}
		return nil, errors.Wrapf(err, "argocd get cluster '%s' failed", name)
	}
	return cls, nil
}

// UpdateCluster will update the annotation field
func (cd *argo) UpdateCluster(ctx context.Context, argoCluster *v1alpha1.Cluster) error {
	connection, client, err := cd.basicClient.NewClusterClient()
	if err != nil {
		return errors.Wrapf(err, "argocd init cluster client failed")
	}
	defer connection.Close() // nolint

	// UpdateFields: github.com/argoproj/argo-cd/server/cluster/cluster.go:#235
	if _, err := client.Update(ctx, &cluster.ClusterUpdateRequest{
		Cluster:       argoCluster,
		UpdatedFields: []string{"annotations"},
	}); err != nil {
		return errors.Wrapf(err, "update cluster '%s' failed", argoCluster.Name)
	}
	return nil
}

// ListCluster interface
func (cd *argo) ListCluster(ctx context.Context) (*v1alpha1.ClusterList, error) {
	connection, client, err := cd.basicClient.NewClusterClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init cluster client failed")
	}
	defer connection.Close() // nolint
	cls, err := client.List(ctx, &cluster.ClusterQuery{})
	if err != nil {
		return nil, errors.Wrapf(err, "argocd list all clusters failed")
	}
	return cls, nil
}

// ListClustersByProject will list clusters by project id
func (cd *argo) ListClustersByProject(ctx context.Context, project string) (*v1alpha1.ClusterList, error) {
	connection, client, err := cd.basicClient.NewClusterClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init cluster client failed")
	}
	defer connection.Close() // nolint
	cls, err := client.List(ctx, &cluster.ClusterQuery{})
	if err != nil {
		return nil, errors.Wrapf(err, "argocd list all clusters failed")
	}

	clusters := make([]v1alpha1.Cluster, 0, len(cls.Items))
	for _, item := range cls.Items {
		projectID, ok := item.Annotations[common.ProjectIDKey]
		if !ok || (projectID == "" && item.Name != common.InClusterName) {
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
	// create new single connection per request
	// ! please make more attention to performance issue
	connection, client, err := cd.basicClient.NewRepoClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init repo client failed")
	}
	defer connection.Close() // nolint
	repos, err := client.Get(ctx, &repository.RepoQuery{Repo: repo})
	if err != nil {
		if strings.Contains(err.Error(), "code = NotFound") {
			blog.Warnf("argocd get Repository %s warning, %s", repo, err.Error())
			return nil, nil
		}
		return nil, errors.Wrapf(err, "argocd get repo '%s' failed", repo)
	}
	return repos, nil
}

// ListRepository interface
func (cd *argo) ListRepository(ctx context.Context) (*v1alpha1.RepositoryList, error) {
	connection, client, err := cd.basicClient.NewRepoClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init repo client failed")
	}
	defer connection.Close() // nolint
	repos, err := client.List(ctx, &repository.RepoQuery{})
	if err != nil {
		return nil, errors.Wrapf(err, "argocd list repos failed")
	}
	return repos, nil
}

// GetApplication will return application by name
func (cd *argo) GetApplication(ctx context.Context, name string) (*v1alpha1.Application, error) {
	// create new single connection per request
	// ! please make more attention to performance issue
	connection, client, err := cd.basicClient.NewApplicationClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init application client failed")
	}
	defer connection.Close() // nolint
	app, err := client.Get(ctx, &appclient.ApplicationQuery{Name: &name})
	if err != nil {
		if strings.Contains(err.Error(), "code = NotFound") {
			blog.Warnf("argocd get application %s warning, %s", name, err.Error())
			return nil, nil
		}
		return nil, errors.Wrapf(err, "argocd get application '%s' failed", name)
	}
	return app, nil
}

// ListApplications interface
func (cd *argo) ListApplications(ctx context.Context, option *ListAppOptions) (*v1alpha1.ApplicationList, error) {
	connection, client, err := cd.basicClient.NewApplicationClient()
	if err != nil {
		return nil, errors.Wrapf(err, "argocd init application client failed")
	}
	defer connection.Close() // nolint
	apps, err := client.List(ctx, &appclient.ApplicationQuery{Projects: []string{option.Project}})
	if err != nil {
		return nil, errors.Wrapf(err, "argocd list application for project '%s' failed", option.Project)
	}
	return apps, nil
}

// GetToken authentication token
func (cd *argo) GetToken(ctx context.Context) string {
	return cd.token
}

// DeleteApplicationResource will delete all resources for application
func (cd *argo) DeleteApplicationResource(ctx context.Context, application *v1alpha1.Application) error {
	server := application.Spec.Destination.Server
	closer, appClient, err := cd.basicClient.NewApplicationClient()
	if err != nil {
		return errors.Wrapf(err, "argocd init application client failed")
	}
	defer closer.Close() // nolint
	errs := make([]string, 0)
	for _, resource := range application.Status.Resources {
		_, err = appClient.DeleteResource(ctx, &appclient.ApplicationResourceDeleteRequest{
			Name:         &application.Name,
			Kind:         &resource.Kind,
			Namespace:    &resource.Namespace,
			Group:        &resource.Group,
			Version:      &resource.Version,
			ResourceName: &resource.Name,
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("delete resource '%s/%s/%s' failed for cluster '%s': %s",
				resource.Group, resource.Kind, resource.Name, server, err.Error()))
		} else {
			blog.Infof("delete resource '%s/%s/%s' for cluster '%s' with application '%s' success",
				resource.Group, resource.Kind, resource.Name, server, application.Name)
		}
	}
	if len(errs) != 0 {
		return errors.Errorf("delete application '%s' sub resource failed: %s",
			application.Name, strings.Join(errs, ","))
	}
	return nil
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
		blog.Errorf("argocd proxy request session failure: %s", err.Error())
		return fmt.Errorf("argocd login session fatal")
	}
	defer response.Body.Close() // nolint
	result := make(map[string]string)
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		blog.Errorf("argocd store decode session response failure, %s", err.Error())
		return fmt.Errorf("decode gitops session result fatal")
	}
	t, ok := result["token"]
	if !ok {
		blog.Errorf("argocd store found no token in session response")
		return fmt.Errorf("found no login token in response")
	}
	blog.Infof("argocd store token session init OK, %s", t)
	cd.token = t
	return nil
}

func (cd *argo) initBasicClient() error {
	var err error
	// init basic client
	cd.basicOpt = &api.ClientOptions{
		ServerAddr: cd.option.Service,
		Insecure:   true,
		AuthToken:  cd.token,
	}
	cd.basicClient, err = api.NewClient(cd.basicOpt)
	if err != nil {
		blog.Errorf("argocd init client failure, %s", err.Error())
		return fmt.Errorf("argocd connection init failure")
	}
	return nil
}
